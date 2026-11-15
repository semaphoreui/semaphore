// The help documents are translated into every language of the interface, and a section names
// its field or button as the interface does in that language.
//
// A translator checks one language with
//   npx vitest run tests/unit/help-translations.spec.js -t "locale=de "
import { expect } from 'chai';
import { languages, messages } from '@/lang';
import { helpDocLocales, loadHelpDoc } from '@/help';
import {
  bareLabel, labelOf, readSource, sourceDocs,
} from './helpers/helpSources';

// The label of a section is read from the form (`:label="$t('key')"`, a literal label, the text
// of a button). Where the form computes it, or the element has none, a label is given here.
const LABELS = {
  'inventory-ansible': {
    'inventoryForm-inventoryCode': { key: 'inventory' },
  },
  'project-settings': {
    'newProject-name': { key: 'projectName' },
  },
};

// The title of a document is the title of its dialog in that language, where the dialog builds
// it from a fixed pattern.
const TITLES = {
  'inventory-ansible': (t) => `${t('edit')} Ansible Inventory`,
};

// vue-i18n falls back to English for a key a language does not have, and to the key itself for
// a key English does not have either.
function translator(locale) {
  return (key) => (messages[locale] && messages[locale][key]) || messages.en[key] || key;
}

const DOCS = sourceDocs().map(({ key, doc }) => {
  const sources = doc.forms.map(readSource);
  const labels = Object.fromEntries(doc.sections.map((s) => [
    s.field,
    (LABELS[key] && LABELS[key][s.field]) || labelOf(sources, s.field),
  ]));
  return { key, labels };
});

describe('help translations', () => {
  it('know the labels they override', () => {
    Object.entries(LABELS).forEach(([key, labels]) => {
      const doc = DOCS.find((d) => d.key === key);
      expect(doc, key).to.not.equal(undefined);
      Object.keys(labels).forEach((field) => expect(doc.labels, key).to.have.property(field));
    });
  });

  languages.forEach((locale) => {
    describe(`locale=${locale} `, () => {
      const t = translator(locale);

      DOCS.forEach(({ key, labels }) => {
        it(`${key} is written in this language`, () => {
          expect(helpDocLocales(key), `src/help/*/${key}.${locale}.md`).to.include(locale);
        });

        it(`${key} names the fields and buttons as the interface does`, async () => {
          const doc = await loadHelpDoc(key, locale);
          expect(doc.locale, 'the translation is missing').to.equal(locale);

          const titles = doc.sections.map((s) => {
            const label = labels[s.field];
            if (!label) return s.title;
            return bareLabel(label.text || t(label.key));
          });
          expect(doc.sections.map((s) => s.title)).to.deep.equal(titles);

          if (TITLES[key]) {
            expect(doc.title).to.equal(TITLES[key](t));
          }
        });
      });
    });
  });
});
