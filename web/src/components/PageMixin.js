export default {
  props: {
    systemInfo: Object,
  },

  computed: {

    premiumFeatures() {
      return this.systemInfo?.features || {};
    },

  },
};
