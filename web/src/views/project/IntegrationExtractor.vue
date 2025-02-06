<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="integration != null">
    <IntegrationExtractorCrumb :integration="integration"/>

    <div class="px-4 pt-3 pb-2">
      <v-switch
         class="mt-0"
        v-model="integration.searchable"
        :label="$t('globalAlias')"
        @change="updateIntegration()"
      />
    </div>

    <div v-if="integration.searchable" class="px-4">
      <v-alert type="info" text class="d-inline-block">
        Matchers allow the integration to be found by a project alias.
      </v-alert>
    </div>

    <div v-else class="px-4 pb-6">
      <div class="mb-3 pl-1" v-if="(aliases || []).length === 0">There is no aliases.</div>

      <div v-else v-for="alias of (aliases || [])" :key="alias.id">
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

      <v-btn
        color="primary"
        @click="addAlias()"
        :disabled="aliases == null"
      >
        {{ aliases == null ? $t('LoadAlias') : $t('AddAlias') }}
      </v-btn>
    </div>

    <v-divider />

    <IntegrationMatcher class="mb-6" v-if="integration.searchable" />

    <IntegrationExtractValue/>

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
