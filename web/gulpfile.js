const { src, dest } = require('gulp');
const rename = require('gulp-rename');
require('dotenv').config();
const gptTranslate = require('./gulp-gpt-translate');

const LANG_NAMES = {
  en: 'English',
  ru: 'Russian',
  es: 'Spanish',
  fr: 'French',
  de: 'German',
  it: 'Italian',
  ja: 'Japanese',
  ko: 'Korean',
  pt: 'Portuguese',
  zh: 'Chinese',
};

function tr() {
  return Object.keys(LANG_NAMES).filter((lang) => lang !== 'en').map((lang) => src('src/lang/en.js')
    .pipe(gptTranslate({
      apiKey: process.env.OPENAI_API_KEY,
      targetLanguage: LANG_NAMES[lang],
      messages: [
        'This content contains Ansible Semaphore strings.',
        'Do not wrap result to any formatting tags. Result must be valid json file.',
      ],
    }))
    .pipe(rename({ basename: lang }))
    .pipe(dest('src/lang')));
}

module.exports = {
  tr,
};
