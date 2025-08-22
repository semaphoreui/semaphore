export default {
  props: {
    systemInfo: Object,
  },

  computed: {

    premiumFeatures() {
      return this.systemInfo?.premium_features || {};
    },

  },
};
