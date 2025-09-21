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
    try {
      const apps = await this.loadAppsDataFromBackend();

      this.appsMixin.activeAppIds = apps.filter((app) => app.active).map((app) => app.id);

      this.appsMixin.apps = apps.reduce((prev, app) => ({
        ...prev,
        [app.id]: app,
      }), {});
    } catch (err) {
      console.error('Failed to load apps data:', err);
      // Set defaults to ensure the page can still render
      this.appsMixin.activeAppIds = [];
      this.appsMixin.apps = {};
    }
  },

  computed: {
    isAppsLoaded() {
      return this.appsMixin.apps != null;
    },
  },

  methods: {
    async loadAppsDataFromBackend() {
      try {
        return (await axios({
          method: 'get',
          url: '/api/apps',
          responseType: 'json',
        })).data;
      } catch (err) {
        console.error('Failed to load apps from backend:', err);
        return [];
      }
    },

    getAppColor(id) {
      if (this.appsMixin.apps[id]?.color) {
        return this.appsMixin.apps[id].color || 'gray';
      }

      if (APP_ICONS[id]) {
        return this.$vuetify.theme.dark ? APP_ICONS[id].darkColor : APP_ICONS[id].color;
      }

      return 'gray';
    },

    getAppTitle(id, short = false) {
      if (this.appsMixin.apps[id]?.title) {
        return this.appsMixin.apps[id].title;
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
      if (this.appsMixin.apps[id]?.icon) {
        return `mdi-${this.appsMixin.apps[id].icon}`;
      }

      if (APP_ICONS[id]) {
        return APP_ICONS[id].icon;
      }

      return 'mdi-help';
    },

  },
};
