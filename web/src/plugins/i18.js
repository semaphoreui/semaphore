import { createI18n } from 'vue-i18n';
import { messages } from '../lang';

let locale = localStorage.getItem('lang');

if (!locale) {
  locale = navigator.language.replace('-', '_').toLocaleLowerCase();
}

// Composition mode: `$t`, `$te` and `$i18n.locale` are injected into every component.
export default createI18n({
  legacy: false,
  globalInjection: true,
  fallbackLocale: 'en',
  locale,
  messages,
  fallbackWarn: false,
  warnHtmlMessage: false,
});
