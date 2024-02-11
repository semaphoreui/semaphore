<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <IntegrationExtractorCrumb/>
    <IntegrationExtractValue/>
    <IntegrationMatcher/>
  </div>
</template>
<script>
import IntegrationExtractorsBase from '@/components/IntegrationExtractorsBase';
import IntegrationExtractorBase from '@/components/IntegrationExtractorBase';

import IntegrationExtractValue from './IntegrationExtractValue.vue';
import IntegrationMatcher from './IntegrationMatcher.vue';
import IntegrationExtractorCrumb from './IntegrationExtractorCrumb.vue';

export default {
  mixins: [IntegrationExtractorsBase, IntegrationExtractorBase],
  components: { IntegrationMatcher, IntegrationExtractValue, IntegrationExtractorCrumb },
  computed: {
    integrationId() {
      if (/^-?\d+$/.test(this.$route.params.integrationId)) {
        return parseInt(this.$route.params.integrationId, 10);
      }
      return this.$route.params.integrationId;
    },
    extractorId() {
      if (/^-?\d+$/.test(this.$route.params.extractorId)) {
        return parseInt(this.$route.params.extractorId, 10);
      }
      return this.$route.params.extractorId;
    },
  },
  methods: {
    allowActions() {
      return true;
    },
  },
};
</script>
