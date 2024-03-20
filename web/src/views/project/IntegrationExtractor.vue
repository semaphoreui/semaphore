<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <IntegrationExtractorCrumb/>

    <div class="px-4 py-3">
      <div v-for="alias of (aliases || [])" :key="alias.id">
        <code class="mr-2">{{ alias.url }}</code>
        <v-btn icon>
          <v-icon>mdi-content-copy</v-icon>
        </v-btn>
        <v-btn icon @click="deleteAlias(alias.id)">
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </div>

      <v-btn color="primary" @click="addAlias()" :disabled="aliases == null">
        {{ aliases == null ? 'Loading aliases...' : 'Add Alias' }}
      </v-btn>
    </div>

    <IntegrationExtractValue/>
    <IntegrationMatcher/>
  </div>
</template>
<script>
import IntegrationExtractorsBase from '@/components/IntegrationExtractorsBase';
import IntegrationsBase from '@/views/project/IntegrationsBase';
import IntegrationExtractValue from './IntegrationExtractValue.vue';
import IntegrationMatcher from './IntegrationMatcher.vue';
import IntegrationExtractorCrumb from './IntegrationExtractorCrumb.vue';

export default {
  mixins: [IntegrationExtractorsBase, IntegrationsBase],
  components: { IntegrationMatcher, IntegrationExtractValue, IntegrationExtractorCrumb },

  methods: {
    allowActions() {
      return true;
    },
  },
};
</script>
