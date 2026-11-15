// Checks that tie the help documents to the code, run at build time: a section must describe a
// field the form has, and a dialog must name a document that exists.
import { HelpDocError } from './parse.mjs';

const DOC_KEY = /^[a-z0-9]+(-[a-z0-9]+)*$/;

/**
 * The `data-testid` values written literally in the template of a form. Bound ones
 * (`:data-testid="…"`) are computed at run time and can not anchor a section.
 *
 * @param {string} formSource text of the single-file component
 * @returns {Set<string>}
 */
export function formFields(formSource) {
  return new Set([...formSource.matchAll(/(?<![:@\w-])data-testid="([^"]+)"/g)].map((m) => m[1]));
}

/**
 * Fails when a section of the document is anchored to a field the form does not have.
 *
 * @param {{ sections: { field: string, title: string, line: number }[], form: string }} doc
 * @param {string} formSource text of the form the document names
 * @param {string} file name of the document, for the error message
 * @throws {HelpDocError}
 */
export function checkFields(doc, formSource, file) {
  const fields = formFields(formSource);
  const unknown = doc.sections.find((section) => !fields.has(section.field));
  if (unknown) {
    throw new HelpDocError(
      file,
      unknown.line,
      `the section "${unknown.title}" is anchored to "${unknown.field}", `
        + `but ${doc.form} has no data-testid="${unknown.field}"`,
    );
  }
}

/**
 * The keys of help documents a single-file component passes to a dialog: the value of
 * `help-doc="key"` and every string literal inside `:help-doc="expression"`.
 *
 * @param {string} vueSource text of the single-file component
 * @returns {{ key: string, line: number }[]}
 */
export function findDocRefs(vueSource) {
  const refs = [];
  [...vueSource.matchAll(/(?<![\w-])(:|v-bind:)?help-doc="([^"]*)"/g)].forEach((m) => {
    const line = vueSource.slice(0, m.index).split('\n').length;
    const keys = m[1]
      ? [...m[2].matchAll(/'([^']*)'|`([^`$]*)`/g)].map((literal) => literal[1] ?? literal[2])
      : [m[2]];
    keys.filter((key) => key !== '').forEach((key) => refs.push({ key, line }));
  });
  return refs;
}

/**
 * Whether a file name (without the extension) can be the key of a document.
 *
 * @param {string} key
 * @returns {boolean}
 */
export function isDocKey(key) {
  return DOC_KEY.test(key);
}

/**
 * What of a parsed document goes to the browser: the tree, without the build-time fields.
 *
 * @param {object} doc result of parseHelpDoc
 * @param {string} key name of the document
 * @returns {{ key: string, title: string, intro: object[],
 *   sections: { field: string, title: string, blocks: object[] }[] }}
 */
export function toModule(doc, key) {
  return {
    key,
    title: doc.title,
    intro: doc.intro,
    sections: doc.sections.map(({ field, title, blocks }) => ({ field, title, blocks })),
  };
}
