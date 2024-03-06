<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <IntegrationExtractorCrumb/>
    <IntegrationExtractValue/>
    <IntegrationMatcher/>
  </div>
</template>
<script>
import IntegrationExtractorsBase from '@/components/IntegrationExtractorsBase';
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
  methods: {
    allowActions() {
      return true;
    },
  },
};
</script>
