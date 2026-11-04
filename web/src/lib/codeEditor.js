// CodeMirror 6 setup of the code editors (vue-codemirror6 `<CodeMirror>` with `minimal`):
// line numbers, a language and the Tab key as in the former CodeMirror 5 options.
// The values are marked raw: CodeMirror compares extensions by identity, a Vue proxy breaks it.

import { markRaw } from 'vue';
import { lineNumbers, keymap } from '@codemirror/view';
import { indentMore, insertTab } from '@codemirror/commands';
import { json } from '@codemirror/lang-json';
import { yaml } from '@codemirror/lang-yaml';

// Tab of CodeMirror 5 by default: indent the selected lines, or insert a tab character.
const defaultTab = { key: 'Tab', run: insertTab };

// Tab of the inventory editor: indent the selected lines, or insert two spaces.
const twoSpacesTab = {
  key: 'Tab',
  run(view) {
    if (!view.state.selection.main.empty) {
      return indentMore(view);
    }
    view.dispatch(view.state.replaceSelection('  '));
    return true;
  },
};

const LANGUAGES = {
  json,
  yaml,
};

// language: 'json', 'yaml' or none (plain text); tab: 'default' or 'spaces'.
export default function codeEditorExtensions({ language, tab = 'default' } = {}) {
  const extensions = [
    lineNumbers(),
    keymap.of([tab === 'spaces' ? twoSpacesTab : defaultTab]),
  ];
  if (LANGUAGES[language]) {
    extensions.push(LANGUAGES[language]());
  }
  return markRaw(extensions);
}
