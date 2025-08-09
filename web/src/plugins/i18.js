import { createI18n } from 'vue-i18n';
import { messages } from '../lang';

let locale = localStorage.getItem('lang');

if (!locale) {
  locale = navigator.language.replace('-', '_').toLocaleLowerCase();
}

export default createI18n({
  legacy: false,
  fallbackLocale: 'en',
  locale,
  messages,
});
