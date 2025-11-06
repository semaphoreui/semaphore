import Vue from 'vue';
import axios from 'axios';
import { AnsiUp } from 'ansi_up';
import { Line, Bar } from 'vue-chartjs/legacy';

import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import localizedFormat from 'dayjs/plugin/localizedFormat';
import durationPlugin from 'dayjs/plugin/duration';

import App from './App.vue';
import router from './router';
import vuetify from './plugins/vuetify';
import './assets/scss/main.scss';
import i18n from './plugins/i18';

const convert = new AnsiUp();
convert.ansi_colors = [
  [
    { rgb: [85, 85, 85], class_name: 'ansi-black' },
    { rgb: [170, 0, 0], class_name: 'ansi-red' },
    { rgb: [0, 170, 0], class_name: 'ansi-green' },
    { rgb: [255, 204, 102], class_name: 'ansi-yellow' },
    { rgb: [33, 150, 243], class_name: 'ansi-blue' },
    { rgb: [170, 0, 170], class_name: 'ansi-magenta' },
    { rgb: [0, 170, 170], class_name: 'ansi-cyan' },
    { rgb: [170, 170, 170], class_name: 'ansi-white' },
  ],
  [
    { rgb: [85, 85, 85], class_name: 'ansi-bright-black' },
    { rgb: [255, 85, 85], class_name: 'ansi-bright-red' },
    { rgb: [85, 255, 85], class_name: 'ansi-bright-green' },
    { rgb: [255, 255, 85], class_name: 'ansi-bright-yellow' },
    { rgb: [85, 85, 255], class_name: 'ansi-bright-blue' },
    { rgb: [255, 85, 255], class_name: 'ansi-bright-magenta' },
    { rgb: [85, 255, 255], class_name: 'ansi-bright-cyan' },
    { rgb: [255, 255, 255], class_name: 'ansi-bright-white' },
  ],
];

axios.defaults.baseURL = document.baseURI;
Vue.config.productionTip = false;

//
// Dates
//

// install needed plugins:
// npm install dayjs dayjs-plugin-relativeTime dayjs-plugin-localizedFormat dayjs-plugin-duration

// extend Day.js
dayjs.extend(relativeTime);
dayjs.extend(localizedFormat);
dayjs.extend(durationPlugin);

Vue.filter('formatDate2', (value) => (value
  ? dayjs(String(value)).format('LL')
  : '—'));

// formatDate: “from now” if today, else localized date+time
Vue.filter('formatDate', (value) => {
  if (!value) return '—';
  const date = dayjs(value);
  const now = dayjs();

  if (now.isSame(date, 'day')) {
    return `${date.fromNow()} (${date.format('HH:mm')})`;
  }
  return date.format('L HH:mm');
});

// formatTime: localized time with seconds
Vue.filter('formatTime', (value) => (value ? dayjs(String(value)).format('LTS') : '—'));

// formatLog: unchanged (ANSI → HTML)
Vue.filter('formatLog', (value) => (value ? convert.ansi_to_html(String(value)) : value));

// formatMilliseconds: humanize a duration or a start/end pair
Vue.filter('formatMilliseconds', (value) => {
  if (value == null || value === '') return '—';

  let ms;

  if (typeof value === 'string') {
    ms = parseInt(value, 10);
  } else if (typeof value === 'number') {
    ms = value;
  } else if (Array.isArray(value)) {
    if (value.length !== 2) {
      throw new Error('formatMilliseconds: invalid value format');
    }
    const [startRaw, endRaw] = value;
    if (startRaw == null || startRaw === '') return '—';
    const start = dayjs(startRaw);
    const end = endRaw == null || endRaw === '' ? dayjs() : dayjs(endRaw);
    ms = end.valueOf() - start.valueOf();
  } else {
    throw new Error('formatMilliseconds: unsupported value type');
  }

  return dayjs.duration(ms).humanize();
});

//
// -------------
//

Vue.component('LineChartGenerator', Line);
Vue.component('BarChartGenerator', Bar);

new Vue({
  router,
  vuetify,
  i18n,
  render: (h) => h(App),
}).$mount('#app');
