import axios from 'axios';
import EventBus from '@/event-bus';

export default {
  props: {
    projectId: Number,
  },

  methods: {
    async loadEndpoint(endpoint) {
      return (
        await axios({
          method: 'get',
          url: endpoint,
          responseType: 'json',
        })
      ).data;
    },

    upgradeToPro(feature) {
      EventBus.$emit('i-subscription', { feature });
    },

    async loadProjectEndpoint(endpoint) {
      return this.loadEndpoint(`/api/project/${this.projectId}${endpoint}`);
    },

    async loadProjectResources(name) {
      return this.loadProjectEndpoint(`/${name}`);
    },

    async loadProjectResource(name, id) {
      return this.loadProjectEndpoint(`/${name}/${id}`);
    },
  },
};
