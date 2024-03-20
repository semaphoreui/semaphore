<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <IntegrationExtractorCrumb/>
    <div class="px-4 py-3">
      <div v-if="aliasURL == null">Loading...</div>
      <div v-else-if="aliasURL === ''">
        <span class="mr-2">No public URL</span>
        <v-btn color="primary" @click="generateAlias()">Generate</v-btn>
      </div>
      <div v-else>
        <span class="mr-2">Public URL:</span>
        <code class="mr-2">{{ aliasURL }}</code>
        <v-btn icon>
          <v-icon>mdi-content-copy</v-icon>
        </v-btn>
        <v-btn icon>
          <v-icon>mdi-refresh</v-icon>
        </v-btn>
        <v-btn icon>
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </div>
    </div>
    <IntegrationExtractValue/>
    <IntegrationMatcher/>
  </div>
</template>
<script>
import IntegrationExtractorsBase from '@/components/IntegrationExtractorsBase';
import axios from 'axios';
import IntegrationExtractValue from './IntegrationExtractValue.vue';
import IntegrationMatcher from './IntegrationMatcher.vue';
import IntegrationExtractorCrumb from './IntegrationExtractorCrumb.vue';

export default {
  mixins: [IntegrationExtractorsBase],
  components: { IntegrationMatcher, IntegrationExtractValue, IntegrationExtractorCrumb },
  computed: {
    integrationId() {
      if (/^-?\d+$/.test(this.$route.params.integrationId)) {
        return parseInt(this.$route.params.integrationId, 10);
      }
      return this.$route.params.integrationId;
    },
  },

  data() {
    return {
      aliasURL: null,
    };
  },

  async created() {
    await this.loadData();
  },

  methods: {
    allowActions() {
      return true;
    },

    async generateAlias() {
      this.aliasURL = (await axios({
        method: 'post',
        url: `/api/project/${this.projectId}/integrations/${this.integrationId}/alias`,
        responseType: 'json',
      })).data.url;
    },

    async loadData() {
      this.aliasURL = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/integrations/${this.integrationId}/alias`,
        responseType: 'json',
      })).data.url;
    },
  },
};
</script>
