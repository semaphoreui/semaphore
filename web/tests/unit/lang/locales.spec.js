import { expect } from 'chai';
import { messages, languages } from '@/lang';

describe('lang/', () => {
  const en = messages.en;
  const enKeys = new Set(Object.keys(en));

  it('exposes English as a language', () => {
    expect(languages).to.include('en');
    expect(enKeys.size).to.be.greaterThan(0);
  });

  Object.keys(messages).forEach((lang) => {
    describe(lang, () => {
      it('has no keys missing from en (dead translations)', () => {
        const extra = Object.keys(messages[lang]).filter((k) => !enKeys.has(k));
        expect(extra, `keys not present in en.js: ${extra.join(', ')}`).to.deep.equal([]);
      });

      it('contains only non-empty string values', () => {
        Object.entries(messages[lang]).forEach(([key, value]) => {
          expect(value, `${lang}.${key}`).to.be.a('string');
          expect(value.trim(), `${lang}.${key} is empty`).to.not.equal('');
        });
      });

      it('keeps the same interpolation placeholders as en', () => {
        const placeholders = (s) => (s.match(/\{[a-zA-Z0-9_]+\}/g) || []).sort();
        Object.entries(messages[lang]).forEach(([key, value]) => {
          if (typeof en[key] !== 'string') {
            return;
          }
          expect(placeholders(value), `${lang}.${key}`).to.deep.equal(placeholders(en[key]));
        });
      });
    });
  });
});
