// Built-in help: the documents of dialogs and pages, by key and language. A document is a file in
// ./dialogs or ./pages, compiled at build time into a tree (web/vite/help), and is loaded on first
// use: every document is a chunk of its own, served by the Semaphore server like the rest of the
// application.
//
// `<key>.md` is the English source; `<key>.<locale>.md` next to it is its translation into a
// language of the interface (src/lang).
//
// To give a dialog its help: add ./dialogs/<key>.md and pass `help-doc="<key>"` to EditDialog.
// To give a page its help: add ./pages/<key>.md and wrap the page in `<HelpPage help-doc="<key>">`.
const FILE = /^\.\/(?:dialogs|pages)\/([^.]+)(?:\.([^.]+))?\.md$/;

const loaders = (() => {
  const byKey = {};
  Object.entries(import.meta.glob(['./dialogs/*.md', './pages/*.md'], { import: 'default' }))
    .forEach(([file, load]) => {
      const [, key, locale = 'en'] = FILE.exec(file);
      byKey[key] = { ...byKey[key], [locale]: load };
    });
  Object.values(byKey).forEach(Object.freeze);
  return Object.freeze(byKey);
})();

const has = (object, key) => Object.prototype.hasOwnProperty.call(object, key);

/**
 * @param {string|null|undefined} key
 * @returns {boolean} whether there is a help document of that key
 */
export function hasHelpDoc(key) {
  return typeof key === 'string' && has(loaders, key) && has(loaders[key], 'en');
}

/**
 * The language a document is shown in for a language of the interface: the same one, else the
 * language without its region (`de` for `de_de`), else English.
 *
 * @param {string} key
 * @param {string|null|undefined} locale the language of the interface (`$i18n.locale`)
 * @returns {string|null} null when there is no document of that key
 */
export function helpDocLocale(key, locale) {
  if (!hasHelpDoc(key)) {
    return null;
  }
  const wanted = String(locale || '').replace('-', '_').toLowerCase();
  return [wanted, wanted.split('_')[0]].find((l) => l !== '' && has(loaders[key], l)) || 'en';
}

/**
 * @param {string} key
 * @param {string} [locale] the language of the interface
 * @returns {Promise<{ key: string, locale: string, title: string, intro: object[],
 *   sections: { field: string, title: string, blocks: object[] }[] }>}
 */
export function loadHelpDoc(key, locale = 'en') {
  if (!hasHelpDoc(key)) {
    return Promise.reject(new Error(`No help document "${key}"`));
  }
  return loaders[key][helpDocLocale(key, locale)]();
}

/**
 * @returns {string[]} keys of all help documents
 */
export function helpDocKeys() {
  return Object.keys(loaders).filter(hasHelpDoc);
}

/**
 * @param {string} key
 * @returns {string[]} the languages a document is written in, English included
 */
export function helpDocLocales(key) {
  return hasHelpDoc(key) ? Object.keys(loaders[key]).sort() : [];
}
