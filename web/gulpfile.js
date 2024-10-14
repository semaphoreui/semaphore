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
  zh_cn: 'Simplified Chinese',
  zh_tw: 'Traditional Chinese',
  nl: 'Dutch (Netherlands)',
  pl: 'Polish',
  pt_br: 'Brazilian Portuguese',
};

function tr() {
  return Object.keys(LANG_NAMES).filter((lang) => lang !== 'en').map((lang) => src('src/lang/en.js')
    .pipe(gptTranslate({
      apiKey: process.env.OPENAI_API_KEY,
      targetLanguage: LANG_NAMES[lang],
      messages: [
        'Translate values of the JS object fields.',
        'Preserve file format. Do not wrap result to markdown tag. Result must be valid js file.',
      ],
    }))
    .pipe(rename({ basename: lang }))
    .pipe(dest('src/lang')));
}

module.exports = {
  tr,
};
