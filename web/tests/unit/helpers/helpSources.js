// Reading the sources of the built-in help for the specs: the documents with their front
// matter, and the fields and labels of the forms they describe.
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { parseHelpDoc } from '../../../vite/help/parse.mjs';

export const WEB = path.join(path.dirname(fileURLToPath(import.meta.url)), '../../..');
export const SRC = path.join(WEB, 'src');

const DOC_DIRS = ['src/help/dialogs', 'src/help/pages'];

export function readSource(file) {
  return fs.readFileSync(path.join(WEB, file), 'utf8');
}

/**
 * The English source documents, parsed, with their front matter.
 *
 * @returns {{ key: string, file: string, doc: object }[]}
 */
export function sourceDocs() {
  return DOC_DIRS.flatMap((dir) => fs.readdirSync(path.join(WEB, dir))
    .filter((name) => /^[a-z0-9-]+\.md$/.test(name))
    .map((name) => {
      const file = `${dir}/${name}`;
      return { key: name.replace(/\.md$/, ''), file, doc: parseHelpDoc(readSource(file), file) };
    }));
}

// The template of a single-file component (the text before its first <script> or <style>).
function templateOf(source) {
  const ends = ['<script', '<style'].map((tag) => source.indexOf(tag)).filter((i) => i >= 0);
  return ends.length ? source.slice(0, Math.min(...ends)) : source;
}

// The opening tag that starts at `start`: up to the first `>` outside a quoted attribute value
// (a rule such as `v => !!v` has one inside).
function openingTag(text, start) {
  let quote = null;
  let end = start;
  while (end < text.length && (quote || text[end] !== '>')) {
    const ch = text[end];
    if (quote) {
      quote = ch === quote ? null : quote;
    } else if (ch === '"' || ch === "'") {
      quote = ch;
    }
    end += 1;
  }
  return text.slice(start, end + 1);
}

// The components a user types into or picks with.
const FIELD_TAG = /<(v-text-field|v-textarea|v-select|v-autocomplete|v-combobox|v-checkbox|v-switch|v-radio-group|v-file-input|v-slider|v-number-input|CodeMirror|codemirror)\b/g;

/**
 * The fields of a form with a literal `data-testid`, in the order of the template.
 *
 * @param {string} source text of the single-file component
 * @returns {string[]}
 */
export function formFieldIds(source) {
  const template = templateOf(source);
  return [...template.matchAll(FIELD_TAG)]
    .map((tag) => /\sdata-testid="([^"]+)"/.exec(openingTag(template, tag.index)))
    .filter(Boolean)
    .map((m) => m[1]);
}

/**
 * The label the interface shows for the element of a `data-testid`, when it is fixed: the key of
 * `$t('key')` in its `:label` (or in the text of a button), or a literal label. Null when the
 * label is computed or there is none.
 *
 * @param {string[]} sources texts of the files the element may be in
 * @param {string} testId
 * @returns {{ key: string }|{ text: string }|null}
 */
export function labelOf(sources, testId) {
  const found = sources
    .map((source) => {
      const template = templateOf(source);
      return { template, at: template.indexOf(`data-testid="${testId}"`) };
    })
    .find(({ at }) => at >= 0);
  if (!found) {
    return null;
  }

  const { template, at } = found;
  const start = template.lastIndexOf('<', at);
  const tag = openingTag(template, start);
  const bound = /\s:label="\$t\('([^']+)'\)(?:\s*\+\s*'[^']*')?"/.exec(tag);
  if (bound) {
    return { key: bound[1] };
  }
  const literal = /\slabel="([^"]+)"/.exec(tag);
  if (literal) {
    return { text: literal[1] };
  }
  if (tag.startsWith('<v-btn')) {
    const close = template.indexOf('</v-btn>', start + tag.length);
    const text = template.slice(start + tag.length, close).trim();
    const t = /^\{\{\s*\$t\('([^']+)'\)\s*\}\}$/.exec(text);
    if (t) {
      return { key: t[1] };
    }
    if (text && !/[<{]/.test(text)) {
      return { text: text.replace(/\s+/g, ' ') };
    }
  }
  return null;
}

/**
 * A label without what the interface adds to it: a trailing "(Optional)" in any language and
 * bracket, a trailing "*" of a required field.
 *
 * @param {string} label
 * @returns {string}
 */
export function bareLabel(label) {
  return label
    .replace(/\s*\*\s*$/, '')
    .replace(/\s*[(（][^()（）]*[)）]\s*$/, '')
    .replace(/\s*\*\s*$/, '')
    .trim();
}
