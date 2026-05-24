export default {
  props: {
    systemInfo: Object,
  },

  computed: {

    features() {
      return this.systemInfo?.features || {};
    },

  },
};
