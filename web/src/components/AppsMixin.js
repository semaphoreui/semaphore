import axios from 'axios';
import { APP_ICONS, APP_TITLE } from '../lib/constants';

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
    async loadAppsDataFromBackend() {
      return (await axios({
        method: 'get',
        url: '/api/apps',
        responseType: 'json',
      })).data;
    },

    getAppColor(id) {
      if (APP_ICONS[id]) {
        return this.$vuetify.theme.dark ? APP_ICONS[id].darkColor : APP_ICONS[id].color;
      }

      if (this.appsMixin.apps[id]) {
        return this.appsMixin.apps[id].color || 'gray';
      }

      return 'gray';
    },

    getAppTitle(id) {
      if (APP_TITLE[id]) {
        return APP_TITLE[id];
      }

      if (this.appsMixin.apps[id]) {
        return this.appsMixin.apps[id].title;
      }

      return '';
    },

    getAppIcon(id) {
      if (APP_ICONS[id]) {
        return APP_ICONS[id].icon;
      }

      if (this.appsMixin.apps[id]) {
        return `mdi-${this.appsMixin.apps[id].icon}`;
      }

      return 'mdi-help';
    },

  },
};
