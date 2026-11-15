// Built-in help: the documents of dialogs, by key. A document is a file in ./dialogs, compiled at
// build time into a tree (web/vite/help), and is loaded on first use: the documents are in
// chunks of their own, served by the Semaphore server like the rest of the application.
//
// To give a dialog its help: add ./dialogs/<key>.md and pass `help-doc="<key>"` to EditDialog.
const loaders = Object.freeze(Object.fromEntries(
  Object.entries(import.meta.glob('./dialogs/*.md', { import: 'default' }))
    .map(([file, load]) => [file.replace(/^\.\/dialogs\/|\.md$/g, ''), load]),
));

/**
 * @param {string|null|undefined} key
 * @returns {boolean} whether there is a help document of that key
 */
export function hasHelpDoc(key) {
  return typeof key === 'string' && Object.prototype.hasOwnProperty.call(loaders, key);
}

/**
 * @param {string} key
 * @returns {Promise<{ key: string, title: string, intro: object[],
 *   sections: { field: string, title: string, blocks: object[] }[] }>}
 */
export function loadHelpDoc(key) {
  if (!hasHelpDoc(key)) {
    return Promise.reject(new Error(`No help document "${key}"`));
  }
  return loaders[key]();
}

/**
 * @returns {string[]} keys of all help documents
 */
export function helpDocKeys() {
  return Object.keys(loaders);
}
