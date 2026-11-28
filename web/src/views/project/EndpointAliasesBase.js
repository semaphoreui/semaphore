import axios from 'axios';

/**
 * Public webhook URLs (aliases) of endpoints, keyed by endpoint id.
 * The host component must provide `projectId`.
 */
export default {

  data() {
    return {
      /** endpoint id → array of `{id, url}`; absent key = not loaded yet. */
      aliases: {},
    };
  },

  methods: {

    aliasPath(endpointId) {
      return `/api/project/${this.projectId}/endpoints/${endpointId}/aliases`;
    },

    endpointAliases(endpointId) {
      return this.aliases[endpointId] || null;
    },

    async loadAliases(endpointId) {
      const data = (await axios({
        method: 'get',
        url: this.aliasPath(endpointId),
        responseType: 'json',
      })).data;
      this.aliases[endpointId] = data || [];
    },

    async deleteAlias(endpointId, aliasId) {
      await axios({
        method: 'delete',
        url: `${this.aliasPath(endpointId)}/${aliasId}`,
      });
      await this.loadAliases(endpointId);
    },

    async addAlias(endpointId) {
      await axios({
        method: 'post',
        url: this.aliasPath(endpointId),
        responseType: 'json',
      });
      await this.loadAliases(endpointId);
    },
  },
};
