const prefersDarkMode = window.matchMedia('(prefers-color-scheme: dark)');

export default {
  data() {
    return {
      darkModeListener: null,
    };
  },

  watch: {
    darkMode(val) {
      this.$vuetify.theme.dark = val;
      if (val && !prefersDarkMode.matches) {
        localStorage.setItem('darkMode', '1');
      } else if (!val && prefersDarkMode.matches) {
        localStorage.setItem('darkMode', '0');
      } else {
        localStorage.removeItem('darkMode');
      }
    },
  },

  methods: {
    initDarkMode() {
      const isDarkMode = localStorage.getItem('darkMode');
      if (isDarkMode !== null) {
        this.darkMode = isDarkMode === '1';
      } else {
        this.darkModeListener = (e) => {
          this.darkMode = e.matches;
        };
        prefersDarkMode.addEventListener('change', this.darkModeListener);

        if (prefersDarkMode.matches) {
          this.darkMode = true;
        }
      }
    },
  },

  beforeDestroy() {
    if (this.darkModeListener) {
      prefersDarkMode.removeEventListener('change', this.darkModeListener);
    }
  },
};
