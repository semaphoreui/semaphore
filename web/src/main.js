import Vue from 'vue';
import moment from 'moment';
import axios from 'axios';
import Convert from 'ansi-to-html';
import App from './App.vue';
import router from './router';
import vuetify from './plugins/vuetify';
import './assets/scss/main.scss';
import i18n from './plugins/i18';

const convert = new Convert();

axios.defaults.baseURL = document.baseURI;
Vue.config.productionTip = false;

Vue.filter('formatDate', (value) => {
  if (!value) {
    return '—';
  }
  const date = moment(value);
  const now = moment();

  if (now.isSame(date, 'day')) {
    return `${date.fromNow()} (${date.format('LT')})`; // Display only time if today
  }
  return date.format('L LT'); // Display only date otherwise
});
Vue.filter('formatTime', (value) => (value ? moment(String(value)).format('LTS') : '—'));
Vue.filter('formatLog', (value) => (value ? convert.toHtml(String(value)) : value));

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
