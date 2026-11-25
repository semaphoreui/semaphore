// CodeMirror 6 setup of the code editors (vue-codemirror6 `<CodeMirror>` with `minimal`):
// line numbers, a language and the Tab key as in the former CodeMirror 5 options.
// The values are marked raw: CodeMirror compares extensions by identity, a Vue proxy breaks it.

import { markRaw } from 'vue';
import { lineNumbers, keymap, EditorView } from '@codemirror/view';
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

// The editor as CodeMirror 5 drew it in Vuetify 2 (`codemirror.css`): black text on white in the
// line height of the card text (22 px, not 1.4), lines padded 4 px, a 30 px gutter of #f7f7f7 with
// numbers in #999. A theme, not a style block: CodeMirror 6 mounts its own styles outside the CSS
// layers, where they win over any layered rule.
export const CODEMIRROR5_LOOK = EditorView.theme({
  '&': {
    color: '#000',
    backgroundColor: '#fff',
  },
  '.cm-scroller': {
    lineHeight: 'inherit',
  },
  '.cm-line': {
    padding: '0 4px',
  },
  '.cm-gutters': {
    backgroundColor: '#f7f7f7',
    color: '#999',
  },
  '.cm-lineNumbers .cm-gutterElement': {
    minWidth: '29px',
  },
});
