import { createI18n } from 'vue-i18n';
import { messages } from '../lang';

let locale = localStorage.getItem('lang');

if (!locale) {
  locale = navigator.language.replace('-', '_').toLocaleLowerCase();
}

const i18n = createI18n({
  legacy: true,
  fallbackLocale: 'en',
  locale,
  messages,
  silentFallbackWarn: true,
});

export default i18n;
