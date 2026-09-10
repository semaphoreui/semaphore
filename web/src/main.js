import Vue from 'vue';
import axios from 'axios';
import { Line, Bar } from 'vue-chartjs/legacy';

import App from './App.vue';
import router from './router';
import vuetify from './plugins/vuetify';
import './assets/scss/main.scss';
import i18n from './plugins/i18';
import filtersPlugin from './plugins/filters';

axios.defaults.baseURL = document.baseURI;
Vue.config.productionTip = false;

Vue.use(filtersPlugin);

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
