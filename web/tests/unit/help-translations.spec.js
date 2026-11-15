// The help documents are translated into every language of the interface, and a translation
// names the fields and the buttons as the interface does in that language.
//
// A translator checks one language with
//   npx vitest run tests/unit/help-translations.spec.js -t "locale=de "
import { expect } from 'chai';
import { languages, messages } from '@/lang';
import { helpDocKeys, helpDocLocales, loadHelpDoc } from '@/help';

// The interface labels the sections of a document stand for, by anchor: a key of src/lang, or a
// string the interface shows untranslated. Every section of every document is listed: a new
// section without an entry fails the spec below.
const LABELS = {
  'inventory-ansible': {
    'inventoryForm-name': { key: 'name' },
    'inventoryForm-runnerTag': { key: 'runner_tag' },
    'inventoryForm-sshKeyId': { key: 'userCredentials' },
    'inventoryForm-becomeKeyId': { key: 'sudoCredentialsOptional' },
    'inventoryForm-type': { key: 'type' },
    'inventoryForm-inventoryPath': { key: 'pathToInventoryFile' },
    'inventoryForm-repositoryId': { key: 'repository' },
    'inventoryForm-inventoryCode': { key: 'inventory' },
  },
  'project-settings': {
    'newProject-name': { key: 'projectName' },
    'newProject-maxParallelTasks': { key: 'maxNumberOfParallelTasksOptional' },
    'newProject-tg': { key: 'telegramChatIdOptional' },
    'newProject-alert': { key: 'allowAlertsForThisProject' },
    // Settings.vue writes the label of this button in English.
    'settings-testAlerts': { text: 'Test Alerts' },
    'settings-exportProject': { key: 'backup' },
    'settings-clearCache': { key: 'clear_cache' },
    'settings-deleteProject': { key: 'deleteProject2' },
  },
};

// The title of the document of a dialog is the title of the dialog in that language.
const TITLES = {
  'inventory-ansible': (t) => `${t('edit')} Ansible Inventory`,
};

// vue-i18n falls back to English for a key a language does not have.
function translator(locale) {
  return (key) => (messages[locale] && messages[locale][key]) || messages.en[key];
}

// A label without the "(Optional)" the interface adds to it, in any language and bracket.
function bare(label) {
  return label.replace(/\s*[(（][^()（）]*[)）]\s*$/, '').trim();
}

describe('help translations', () => {
  it('cover every document of the registry and every section of it', async () => {
    expect(helpDocKeys().sort()).to.deep.equal(Object.keys(LABELS).sort());
    const docs = await Promise.all(Object.keys(LABELS).map((key) => loadHelpDoc(key)));
    docs.forEach((doc) => {
      expect(doc.sections.map((s) => s.field), doc.key).to.deep.equal(Object.keys(LABELS[doc.key]));
    });
  });

  languages.forEach((locale) => {
    describe(`locale=${locale} `, () => {
      const t = translator(locale);

      Object.keys(LABELS).forEach((key) => {
        it(`${key} is written in this language`, () => {
          expect(helpDocLocales(key), `src/help/*/${key}.${locale}.md`).to.include(locale);
        });

        it(`${key} names the fields and buttons as the interface does`, async () => {
          const doc = await loadHelpDoc(key, locale);
          expect(doc.locale, 'the translation is missing').to.equal(locale);

          const expected = doc.sections.map((s) => {
            const label = LABELS[key][s.field];
            return label.text || bare(t(label.key));
          });
          expect(doc.sections.map((s) => s.title)).to.deep.equal(expected);

          if (TITLES[key]) {
            expect(doc.title).to.equal(TITLES[key](t));
          }
          doc.sections.forEach((s) => {
            expect(s.blocks, `${key}.${locale} ${s.field}`).to.not.have.length(0);
          });
        });
      });
    });
  });
});
