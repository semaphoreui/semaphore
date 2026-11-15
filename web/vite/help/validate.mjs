// Checks that tie the help documents to the code, run at build time: a section must describe an
// element its form has (or one of its forms: a page and the form it shows), and a dialog or a
// page must name a document that exists.
import { HelpDocError } from './parse.mjs';

const DOC_KEY = /^[a-z0-9]+(-[a-z0-9]+)*$/;

// The name of a document file without `.md`: `<key>` for the English source, `<key>.<locale>` for
// a translation, the locale being the name of a file of src/lang (`de`, `pt_br`).
const DOC_NAME = /^([a-z0-9]+(?:-[a-z0-9]+)*)(?:\.([a-z]{2}(?:_[a-z]{2})?))?$/;

/**
 * Splits the name of a document file (without `.md`) into its key and its locale.
 *
 * @param {string} name `inventory-ansible` or `inventory-ansible.pt_br`
 * @returns {{ key: string, locale: string|null }|null} null for a name that is neither; the
 *   locale is null for the English source
 */
export function splitDocName(name) {
  const m = DOC_NAME.exec(name);
  return m ? { key: m[1], locale: m[2] || null } : null;
}

// Every fenced code block of a parsed document, in order.
function codeBlocks(doc) {
  return [doc.intro, ...doc.sections.map((section) => section.blocks)]
    .flat()
    .filter((block) => block.type === 'code')
    .map((block) => block.text);
}

// Every link of a parsed document, as a sorted list of distinct addresses.
function links(doc) {
  const inlines = [doc.intro, ...doc.sections.map((section) => section.blocks)]
    .flat()
    .flatMap((block) => {
      if (block.type === 'p') return block.inline;
      if (block.type === 'ul') return block.items.flat();
      return [];
    });
  return [...new Set(inlines.filter((node) => node.type === 'link').map((node) => node.href))].sort();
}

/**
 * Fails when a translation does not match its English source: it must describe the same
 * elements in the same order, name the same forms, and keep the code and the links untouched.
 *
 * @param {object} source the parsed English document
 * @param {object} translation the parsed translation
 * @param {string} file name of the translation, for the error message
 * @param {string} sourceName name of the source, for the error message
 * @throws {HelpDocError}
 */
export function checkTranslation(source, translation, file, sourceName) {
  if (translation.forms.join() !== source.forms.join()) {
    throw new HelpDocError(file, 1, `"form" must be the same as in ${sourceName}: ${source.forms.join(', ')}`);
  }
  if (translation.skip.join() !== source.skip.join()) {
    throw new HelpDocError(file, 1, `"skip" must be the same as in ${sourceName}: ${source.skip.join(', ') || 'none'}`);
  }

  const count = Math.max(source.sections.length, translation.sections.length);
  for (let i = 0; i < count; i += 1) {
    const expected = source.sections[i];
    const actual = translation.sections[i];
    if (!actual) {
      throw new HelpDocError(file, 1, `the section {#${expected.field}} of ${sourceName} is missing`);
    }
    if (!expected) {
      throw new HelpDocError(file, actual.line, `the section {#${actual.field}} is not in ${sourceName}`);
    }
    if (actual.field !== expected.field) {
      throw new HelpDocError(
        file,
        actual.line,
        `the section {#${actual.field}} stands where ${sourceName} has {#${expected.field}}: `
          + 'a translation keeps the sections of its source in their order',
      );
    }
  }

  if (codeBlocks(translation).join('\u0000') !== codeBlocks(source).join('\u0000')) {
    throw new HelpDocError(file, 1, `the code blocks must be the same as in ${sourceName}: code is not translated`);
  }

  const expectedLinks = links(source);
  const actualLinks = links(translation);
  if (actualLinks.join() !== expectedLinks.join()) {
    throw new HelpDocError(
      file,
      1,
      `the links must be the same as in ${sourceName}: ${expectedLinks.join(', ') || 'none'}; `
        + `found ${actualLinks.join(', ') || 'none'}`,
    );
  }
}

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
 * Fails when a section of the document is anchored to an element none of its forms has.
 *
 * @param {{ sections: { field: string, title: string, line: number }[], forms: string[] }} doc
 * @param {string[]} formSources texts of the files the document names, in the order of `forms`
 * @param {string} file name of the document, for the error message
 * @throws {HelpDocError}
 */
export function checkFields(doc, formSources, file) {
  const fields = new Set(formSources.flatMap((source) => [...formFields(source)]));
  const where = doc.forms.length === 1
    ? `${doc.forms[0]} has no`
    : `none of ${doc.forms.join(', ')} has`;
  const skipped = (doc.skip || []).find((field) => !fields.has(field));
  if (skipped) {
    throw new HelpDocError(file, 1, `"skip" names "${skipped}", but ${where} data-testid="${skipped}"`);
  }
  const unknown = doc.sections.find((section) => !fields.has(section.field));
  if (unknown) {
    throw new HelpDocError(
      file,
      unknown.line,
      `the section "${unknown.title}" is anchored to "${unknown.field}", `
        + `but ${where} data-testid="${unknown.field}"`,
    );
  }
}

/**
 * The keys of help documents a single-file component passes to a dialog or a page: the value of
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
 * @param {string} [locale] the language of the document
 * @returns {{ key: string, locale: string, title: string, intro: object[],
 *   sections: { field: string, title: string, blocks: object[] }[] }}
 */
export function toModule(doc, key, locale = 'en') {
  return {
    key,
    locale,
    title: doc.title,
    intro: doc.intro,
    sections: doc.sections.map(({ field, title, blocks }) => ({ field, title, blocks })),
  };
}
