import axios from 'axios';

export default {

  data() {
    return {
      aliases: null,
    };
  },

  async created() {
    await this.loadAliases();
  },

  computed: {
    integrationId() {
      if (/^-?\d+$/.test(this.$route.params.integrationId)) {
        return parseInt(this.$route.params.integrationId, 10);
      }
      return this.$route.params.integrationId;
    },

    aliasPath() {
      if (!this.integrationId) {
        return `/api/project/${this.projectId}/integrations/aliases`;
      }
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/aliases`;
    },
  },

  methods: {

    async loadAliases() {
      this.aliases = (await axios({
        method: 'get',
        url: this.aliasPath,
        responseType: 'json',
      })).data;
    },

    async deleteAlias(id) {
      await axios({
        method: 'delete',
        url: `${this.aliasPath}/${id}`,
      });
      await this.loadAliases();
    },

    async addAlias() {
      await axios({
        method: 'post',
        url: this.aliasPath,
        responseType: 'json',
      });
      await this.loadAliases();
    },
  },
};
