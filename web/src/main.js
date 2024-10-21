import Vue from 'vue';
import moment from 'moment';
import axios from 'axios';
import { AnsiUp } from 'ansi_up';
import App from './App.vue';
import router from './router';
import vuetify from './plugins/vuetify';
import './assets/scss/main.scss';
import i18n from './plugins/i18';

const convert = new AnsiUp();
convert.ansi_colors = [
  [
    { rgb: [0, 0, 0], class_name: 'ansi-black' },
    { rgb: [170, 0, 0], class_name: 'ansi-red' },
    { rgb: [0, 170, 0], class_name: 'ansi-green' },
    { rgb: [170, 85, 0], class_name: 'ansi-yellow' },
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

Vue.filter('formatDate', (value) => {
  if (!value) {
    return '—';
  }
  const date = moment(value);
  const now = moment();

  if (now.isSame(date, 'day')) {
    return `${date.fromNow()} (${date.format('HH:mm')})`; // Display only time if today
  }
  return date.format('L HH:mm'); // Display only date otherwise
});
Vue.filter('formatTime', (value) => (value ? moment(String(value)).format('LTS') : '—'));
Vue.filter('formatLog', (value) => (value ? convert.ansi_to_html(String(value)) : value));

Vue.filter('formatMilliseconds', (value) => {
  if (value == null || value === '') {
    return '—';
  }

  let duration;
  if (typeof value === 'string') {
    duration = parseInt(value, 10);
  } else if (typeof value === 'number') {
    duration = value;
  } else if (Array.isArray(value)) {
    if (value.length !== 2) {
      throw new Error('formatMilliseconds: invalid value format');
    }

    if (value[0] == null || value[0] === '') {
      return '—';
    }
    const start = typeof value[0] === 'string' ? new Date(value[0]) : value[0];
    let end;

    if (value[1] == null || value[1] === '') {
      end = Date.now();
    } else {
      end = typeof value[1] === 'string' ? new Date(value[1]) : value[1];
    }

    duration = end - start;
  }
  return moment.duration(duration, 'milliseconds').humanize();
});

new Vue({
  router,
  vuetify,
  i18n,
  render: (h) => h(App),
}).$mount('#app');
