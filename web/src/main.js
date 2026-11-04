// Vuetify base styles first: component and application styles come after them.
// eslint-disable-next-line import/extensions
import 'vuetify/styles';
import { createApp } from 'vue';
import axios from 'axios';
import { Line, Bar } from 'vue-chartjs';

import App from './App.vue';
import router from './router';
import vuetify from './plugins/vuetify';
import '@mdi/font/css/materialdesignicons.css';
import './assets/scss/main.scss';
import i18n from './plugins/i18';
import { filters } from './plugins/filters';

axios.defaults.baseURL = document.baseURI;

const app = createApp(App);

app.use(router);
app.use(vuetify);
app.use(i18n);

app.config.globalProperties.$filters = filters;

//
// -------------
//

app.component('LineChartGenerator', Line);
app.component('BarChartGenerator', Bar);

// Mount after the initial navigation: App.vue reads the real `$route` in data() and created().
router.isReady().then(() => app.mount('#app'));
