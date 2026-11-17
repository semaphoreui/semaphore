// Every dialog a user types into has built-in help, and every field of it has a section.
import fs from 'node:fs';
import path from 'node:path';
import { expect } from 'chai';
import {
  SRC, formFieldIds, readSource, sourceDocs,
} from './helpers/helpSources';

// Dialogs on EditDialog that need no help: nothing in them to type or to pick. By file and
// `test-id`.
const EDIT_DIALOGS_WITHOUT_HELP = [
  'components/TaskLogDialog.vue#taskLogDialog', // the log of a task
  'views/project/Keys.vue#keystorePublicKeyDialog', // shows a generated public key
  'views/Runners.vue#runnerTokenDialog', // shows the token of a runner
];

// Files with a plain `v-dialog` (not EditDialog, not HelpDialog), and why it needs no help.
const PLAIN_DIALOGS_WITHOUT_HELP = {
  'components/EditDialog.vue': 'the dialog every form dialog is built on',
  'components/HelpDialog.vue': 'the plain dialog with help',
  'components/YesNoDialog.vue': 'a confirmation',
  'components/ObjectRefsDialog.vue': 'lists what uses an object',
  'components/SystemInfoDialog.vue': 'shows information about the server',
  'components/SystemSettingsDialog.vue': 'shows the settings of the server',
  'components/RichEditor.vue': 'a code editor opened over a field that has help of its own',
  'components/ArgsPicker.vue': 'edits one argument of a list that has help of its own',
  'views/Auth.vue': 'help on how to fix a sign-in, itself',
  'components/TemplateForm.vue': 'its three help dialogs are out of scope (TASK@A3B2A4AC7F)',
};

function vueFiles(dir) {
  return fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) return vueFiles(full);
    return entry.name.endsWith('.vue') ? [full] : [];
  });
}

const FILES = vueFiles(SRC).map((full) => ({
  name: path.relative(SRC, full).split(path.sep).join('/'),
  source: fs.readFileSync(full, 'utf8'),
}));

describe('built-in help coverage', () => {
  describe('every EditDialog a user types into names its help document', () => {
    FILES.forEach(({ name, source }) => {
      [...source.matchAll(/<EditDialog\b[\s\S]*?>/g)].forEach((m, i) => {
        const testId = /\stest-id="([^"]+)"/.exec(m[0]);
        const id = `${name}#${testId ? testId[1] : i}`;
        if (EDIT_DIALOGS_WITHOUT_HELP.includes(id)) {
          return;
        }
        it(id, () => {
          expect(m[0], `${id} has no help-doc`).to.match(/\s:?help-doc="/);
        });
      });
    });
  });

  it('every plain v-dialog is a HelpDialog, or is listed with the reason it needs no help', () => {
    const plain = FILES.filter(({ source }) => /<v-dialog\b/.test(source)).map(({ name }) => name);
    expect(plain.sort()).to.deep.equal(Object.keys(PLAIN_DIALOGS_WITHOUT_HELP).sort());
  });

  it('every HelpDialog names its help document', () => {
    FILES.forEach(({ name, source }) => {
      [...source.matchAll(/<HelpDialog\b[\s\S]*?>/g)].forEach((m) => {
        expect(m[0], `a HelpDialog of ${name} has no help-doc`).to.match(/\s:?help-doc="/);
      });
    });
  });

  describe('every field of a form has a section, or is skipped on purpose', () => {
    sourceDocs().forEach(({ key, doc }) => {
      it(key, () => {
        const described = new Set([...doc.sections.map((s) => s.field), ...doc.skip]);
        const fields = doc.forms.flatMap((form) => formFieldIds(readSource(form)));
        const missing = [...new Set(fields)].filter((field) => !described.has(field));
        expect(missing, `${key}: fields without a section (add one, or list them in "skip")`)
          .to.deep.equal([]);
      });
    });
  });
});
