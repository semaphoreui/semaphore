<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <IntegrationExtractorCrumb v-if="integration != null" :integration="integration"/>

    <div class="px-4 py-3">
      <div v-for="alias of (aliases || [])" :key="alias.id">
        <code class="mr-2">{{ alias.url }}</code>
        <v-btn icon
               @click="copyToClipboard(
                 alias.url, 'The alias URL  has been copied to the clipboard.')">
          <v-icon>mdi-content-copy</v-icon>
        </v-btn>
        <v-btn icon @click="deleteAlias(alias.id)">
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </div>

      <v-btn color="primary" @click="addAlias()" :disabled="aliases == null">
        {{ aliases == null ? 'Loading aliases...' : 'Add Alias' }}
      </v-btn>

      <v-checkbox
        v-model="integration.searchable"
        label="Available by project and global alias"
        @change="updateIntegration()"
      />
    </div>

    <IntegrationExtractValue/>
    <IntegrationMatcher/>
  </div>
</template>
<script>
import IntegrationExtractorsBase from '@/components/IntegrationExtractorsBase';
import IntegrationsBase from '@/views/project/IntegrationsBase';
import copyToClipboard from '@/lib/copyToClipboard';
import axios from 'axios';
import IntegrationExtractValue from './IntegrationExtractValue.vue';
import IntegrationMatcher from './IntegrationMatcher.vue';
import IntegrationExtractorCrumb from './IntegrationExtractorCrumb.vue';

export default {
  mixins: [IntegrationExtractorsBase, IntegrationsBase],
  components: { IntegrationMatcher, IntegrationExtractValue, IntegrationExtractorCrumb },

  data() {
    return {
      integration: null,
    };
  },

  async created() {
    this.integration = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/integrations/${this.integrationId}`,
      responseType: 'json',
    })).data;
  },

  methods: {
    copyToClipboard,
    allowActions() {
      return true;
    },
    async updateIntegration() {
      await axios({
        method: 'put',
        url: `/api/project/${this.projectId}/integrations/${this.integrationId}`,
        responseType: 'json',
        data: this.integration,
      });
    },
  },
};
</script>
