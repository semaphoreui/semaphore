import EventBus from '@/event-bus';

export default {
  props: {
    systemInfo: Object,
  },

  computed: {
    features() {
      return this.systemInfo?.features || {};
    },
  },

  methods: {
    upgradeToPro(feature) {
      EventBus.$emit('i-subscription', { feature });
    },
  },
};
