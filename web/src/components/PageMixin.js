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
    upgradeToPro(source) {
      EventBus.$emit('i-subscription', { source });
    },
  },
};
