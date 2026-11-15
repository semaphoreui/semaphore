// Vite plugin of the built-in help. A document (src/help/dialogs/<key>.md) is compiled into a
// module that exports its tree, so the help is part of the frontend build and of the binary the
// build is embedded in. The build fails when a document leaves the supported subset, describes a
// field its form does not have, or a dialog names a document that does not exist.
import fs from 'node:fs';
import path from 'node:path';
import { HelpDocError, parseHelpDoc } from './parse.mjs';
import {
  checkFields, findDocRefs, isDocKey, toModule,
} from './validate.mjs';

export const HELP_DOCS_DIR = 'src/help/dialogs';

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
 * @returns {{ formFile: string, doc: object }} the form the document depends on and its tree
 * @throws {HelpDocError}
 */
export function compileHelpDoc(root, file, source) {
  const name = slashes(path.relative(root, file));
  const key = path.basename(file, '.md');
  if (!isDocKey(key)) {
    throw new HelpDocError(name, 1, `the name of a document is lower-case words joined by "-", got "${key}"`);
  }

  const doc = parseHelpDoc(source, name);
  const formFile = path.join(root, doc.form);
  if (!fs.existsSync(formFile)) {
    throw new HelpDocError(name, 1, `the form "${doc.form}" does not exist`);
  }
  checkFields(doc, fs.readFileSync(formFile, 'utf8'), name);

  return { formFile, doc: toModule(doc, key) };
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
  const missing = findDocRefs(source)
    .find((ref) => !isDocKey(ref.key) || !fs.existsSync(path.join(root, HELP_DOCS_DIR, `${ref.key}.md`)));
  if (missing) {
    throw new HelpDocError(
      slashes(path.relative(root, file)),
      missing.line,
      `help document "${missing.key}" does not exist (${HELP_DOCS_DIR}/${missing.key}.md)`,
    );
  }
}

/**
 * @param {string} root directory of the frontend (web/)
 * @returns {import('vite').Plugin}
 */
export default function helpDocs(root) {
  const docsDir = `${slashes(path.join(root, HELP_DOCS_DIR))}/`;

  return {
    name: 'semaphore:help-docs',
    // Ahead of the Vue plugin: a component is checked as it is written.
    enforce: 'pre',

    transform(code, id) {
      const [file, query] = id.split('?');

      if (file.startsWith(docsDir) && file.endsWith('.md')) {
        const { formFile, doc } = compileHelpDoc(root, file, code);
        // A field renamed in the form must fail the document in the dev server as well.
        this.addWatchFile(formFile);
        return { code: `export default ${JSON.stringify(doc)};\n`, map: null };
      }

      if (!query && file.endsWith('.vue') && !file.includes('/node_modules/')) {
        checkDocRefs(root, file, code);
      }
      return null;
    },
  };
}
