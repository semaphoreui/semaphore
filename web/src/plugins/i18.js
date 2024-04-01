import Vue from 'vue';
import VueI18n from 'vue-i18n';
import { messages } from '../lang';

Vue.use(VueI18n);

let locale = localStorage.getItem('lang');

if (!locale) {
  locale = navigator.language.split('-')[0];
}

export default new VueI18n({
  fallbackLocale: 'en',
  locale,
  messages,
  silentFallbackWarn: true,
});
