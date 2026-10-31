import axios from 'axios';
import { APP_ICONS, APP_SHORT_TITLE, APP_TITLE } from '../lib/constants';

export default {
  data() {
    return {
      appsMixin: {
        activeAppIds: [],
        apps: null,
      },
    };
  },

  async created() {
    const apps = await this.loadAppsDataFromBackend();

    this.appsMixin.activeAppIds = apps.filter((app) => app.active).map((app) => app.id);

    this.appsMixin.apps = apps.reduce((prev, app) => ({
      ...prev,
      [app.id]: app,
    }), {});
  },

  computed: {
    isAppsLoaded() {
      return this.appsMixin.apps != null;
    },
  },

  methods: {
    // The app of the backend (GET /api/apps), or undefined until it is in: a page may draw its
    // rows before the request returns, and falls back to the built-in icon, colour and title.
    getBackendApp(id) {
      return this.appsMixin.apps?.[id];
    },

    async loadAppsDataFromBackend() {
      return (await axios({
        method: 'get',
        url: '/api/apps',
        responseType: 'json',
      })).data;
    },

    getAppColor(id) {
      if (this.getBackendApp(id)?.color) {
        return this.getBackendApp(id).color || 'gray';
      }

      if (APP_ICONS[id]) {
        return this.$vuetify.theme.current.dark ? APP_ICONS[id].darkColor : APP_ICONS[id].color;
      }

      return 'gray';
    },

    getAppTitle(id, short = false) {
      if (this.getBackendApp(id)?.title) {
        return this.getBackendApp(id).title;
      }

      if (short) {
        if (APP_SHORT_TITLE[id]) {
          return APP_SHORT_TITLE[id];
        }
      }

      if (APP_TITLE[id]) {
        return APP_TITLE[id];
      }

      return '';
    },

    getAppIcon(id) {
      if (this.getBackendApp(id)?.icon) {
        return `mdi-${this.getBackendApp(id).icon}`;
      }

      if (APP_ICONS[id]) {
        return APP_ICONS[id].icon;
      }

      return 'mdi-help';
    },

  },
};
