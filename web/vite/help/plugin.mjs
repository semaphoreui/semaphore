// Vite plugin of the built-in help. A document (src/help/dialogs/<key>.md for a dialog,
// src/help/pages/<key>.md for a page) is compiled into a module that exports its tree, so the
// help is part of the frontend build and of the binary the build is embedded in. The build fails
// when a document leaves the supported subset, describes an element its forms do not have, or a
// dialog or a page names a document that does not exist.
import fs from 'node:fs';
import path from 'node:path';
import { HelpDocError, parseHelpDoc } from './parse.mjs';
import {
  checkFields, findDocRefs, isDocKey, toModule,
} from './validate.mjs';

// A key names one document, whichever folder it is in.
export const HELP_DOCS_DIRS = ['src/help/dialogs', 'src/help/pages'];

// Module ids of Vite use forward slashes on every platform.
function slashes(file) {
  return file.split(path.sep).join('/');
}

/**
 * Compiles one document.
 *
 * @param {string} root directory of the frontend (web/)
 * @param {string} file absolute path of the document
 * @param {string} source text of the document
 * @returns {{ formFiles: string[], doc: object }} the files the document depends on, and its tree
 * @throws {HelpDocError}
 */
export function compileHelpDoc(root, file, source) {
  const name = slashes(path.relative(root, file));
  const key = path.basename(file, '.md');
  if (!isDocKey(key)) {
    throw new HelpDocError(name, 1, `the name of a document is lower-case words joined by "-", got "${key}"`);
  }

  const twin = HELP_DOCS_DIRS
    .map((dir) => path.join(root, dir, `${key}.md`))
    .find((other) => other !== file && fs.existsSync(other));
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

  return { formFiles, doc: toModule(doc, key) };
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
        const { formFiles, doc } = compileHelpDoc(root, file, code);
        // A field renamed in a form must fail the document in the dev server as well.
        formFiles.forEach((formFile) => this.addWatchFile(formFile));
        return { code: `export default ${JSON.stringify(doc)};\n`, map: null };
      }

      if (!query && file.endsWith('.vue') && !file.includes('/node_modules/')) {
        checkDocRefs(root, file, code);
      }
      return null;
    },
  };
}
