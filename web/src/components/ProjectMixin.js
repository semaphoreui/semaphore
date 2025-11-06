import axios from 'axios';

export default {
  props: {
    projectId: Number,
  },

  methods: {
    async loadEndpoint(endpoint) {
      return (await axios({
        method: 'get',
        url: endpoint,
        responseType: 'json',
      })).data;
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
