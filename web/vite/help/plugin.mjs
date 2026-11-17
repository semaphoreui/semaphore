// Vite plugin of the built-in help. A document (src/help/dialogs/<key>.md for a dialog,
// src/help/pages/<key>.md for a page, in English) and its translations (<key>.<locale>.md next to
// it) are compiled into modules that export their trees, so the help is part of the frontend build
// and of the binary the build is embedded in. The build fails when a document leaves the supported
// subset, describes an element its forms do not have, a translation does not match its source, or
// a dialog or a page names a document that does not exist.
import fs from 'node:fs';
import path from 'node:path';
import { HelpDocError, parseHelpDoc } from './parse.mjs';
import {
  checkFields, checkTranslation, findDocRefs, isDocKey, splitDocName, toModule,
} from './validate.mjs';

// A key names one document, whichever folder it is in.
export const HELP_DOCS_DIRS = ['src/help/dialogs', 'src/help/pages'];

// The languages of the interface: a translation is made for one of them.
export const LOCALES_DIR = 'src/lang';

/**
 * The locales of the interface: the names of the files of src/lang.
 *
 * @param {string} root directory of the frontend (web/)
 * @returns {string[]}
 */
export function interfaceLocales(root) {
  return fs.readdirSync(path.join(root, LOCALES_DIR))
    .filter((file) => file.endsWith('.js') && file !== 'index.js')
    .map((file) => file.replace(/\.js$/, ''))
    .sort();
}

// Module ids of Vite use forward slashes on every platform.
function slashes(file) {
  return file.split(path.sep).join('/');
}

/**
 * Compiles one document: an English source, or a translation checked against its source.
 *
 * @param {string} root directory of the frontend (web/)
 * @param {string} file absolute path of the document
 * @param {string} source text of the document
 * @returns {{ formFiles: string[], dependencies: string[], doc: object }} the forms of the
 *   document, every file it depends on (the forms, and the source of a translation), its tree
 * @throws {HelpDocError}
 */
export function compileHelpDoc(root, file, source) {
  const name = slashes(path.relative(root, file));
  const parts = splitDocName(path.basename(file, '.md'));
  if (!parts) {
    throw new HelpDocError(
      name,
      1,
      'the name of a document is lower-case words joined by "-", '
        + `with ".<locale>" for a translation, got "${path.basename(file)}"`,
    );
  }
  const { key, locale } = parts;

  if (locale != null) {
    // The two call each other: a translation compiles its source first.
    // eslint-disable-next-line no-use-before-define
    return compileTranslation(root, file, source, key, locale);
  }

  // Not `!==`: Vite names the file with forward slashes on every platform, `path.join` with the
  // separator of the platform. `path.relative` normalises both (on Windows the letter case too)
  // and is empty for the file itself.
  const twin = HELP_DOCS_DIRS
    .map((dir) => path.join(root, dir, `${key}.md`))
    .find((other) => path.relative(other, file) !== '' && fs.existsSync(other));
  if (twin) {
    throw new HelpDocError(name, 1, `another document has the same name: ${slashes(path.relative(root, twin))}`);
  }

  const doc = parseHelpDoc(source, name);
  const formFiles = doc.forms.map((form) => path.join(root, form));
  doc.forms.forEach((form, i) => {
    if (!fs.existsSync(formFiles[i])) {
      throw new HelpDocError(name, 1, `the form "${form}" does not exist`);
    }
  });
  checkFields(doc, formFiles.map((formFile) => fs.readFileSync(formFile, 'utf8')), name);

  return { formFiles, dependencies: formFiles, doc: toModule(doc, key) };
}

// A translation: of an existing source, in a language of the interface, matching the source.
function compileTranslation(root, file, text, key, locale) {
  const name = slashes(path.relative(root, file));
  const locales = interfaceLocales(root);
  if (locale === 'en') {
    throw new HelpDocError(name, 1, `the English document is ${key}.md, without a locale`);
  }
  if (!locales.includes(locale)) {
    throw new HelpDocError(
      name,
      1,
      `"${locale}" is not a language of the interface (${LOCALES_DIR}): ${locales.join(', ')}`,
    );
  }

  const sourceFile = path.join(path.dirname(file), `${key}.md`);
  const sourceName = slashes(path.relative(root, sourceFile));
  if (!fs.existsSync(sourceFile)) {
    throw new HelpDocError(name, 1, `a translation of a document that does not exist: ${sourceName}`);
  }

  const source = compileHelpDoc(root, sourceFile, fs.readFileSync(sourceFile, 'utf8'));
  const sourceDoc = parseHelpDoc(fs.readFileSync(sourceFile, 'utf8'), sourceName);
  const doc = parseHelpDoc(text, name);
  checkTranslation(sourceDoc, doc, name, sourceName);

  return {
    formFiles: source.formFiles,
    // A change of the source must check the translation again in the dev server.
    dependencies: [...source.formFiles, sourceFile],
    doc: toModule(doc, key, locale),
  };
}

/**
 * Fails when a single-file component names a help document that does not exist.
 *
 * @param {string} root directory of the frontend (web/)
 * @param {string} file absolute path of the component
 * @param {string} source text of the component
 * @throws {HelpDocError}
 */
export function checkDocRefs(root, file, source) {
  const exists = (key) => HELP_DOCS_DIRS
    .some((dir) => fs.existsSync(path.join(root, dir, `${key}.md`)));
  const missing = findDocRefs(source).find((ref) => !isDocKey(ref.key) || !exists(ref.key));
  if (missing) {
    const expected = HELP_DOCS_DIRS.map((dir) => `${dir}/${missing.key}.md`).join(' or ');
    throw new HelpDocError(
      slashes(path.relative(root, file)),
      missing.line,
      `help document "${missing.key}" does not exist (${expected})`,
    );
  }
}

/**
 * @param {string} root directory of the frontend (web/)
 * @returns {import('vite').Plugin}
 */
export default function helpDocs(root) {
  const docsDirs = HELP_DOCS_DIRS.map((dir) => `${slashes(path.join(root, dir))}/`);

  return {
    name: 'semaphore:help-docs',
    // Ahead of the Vue plugin: a component is checked as it is written.
    enforce: 'pre',

    transform(code, id) {
      const [file, query] = id.split('?');

      if (file.endsWith('.md') && docsDirs.some((dir) => file.startsWith(dir))) {
        const { dependencies, doc } = compileHelpDoc(root, file, code);
        // A field renamed in a form, or a section moved in the source of a translation, must fail
        // the document in the dev server as well.
        dependencies.forEach((dependency) => this.addWatchFile(dependency));
        return { code: `export default ${JSON.stringify(doc)};\n`, map: null };
      }

      if (!query && file.endsWith('.vue') && !file.includes('/node_modules/')) {
        checkDocRefs(root, file, code);
      }
      return null;
    },
  };
}
