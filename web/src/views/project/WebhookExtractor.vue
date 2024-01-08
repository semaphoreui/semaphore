<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <WebhookExtractorCrumb/>
    <WebhookExtractValue/>
    <WebhookMatcher/>
  </div>
</template>
<script>
import WebhookExtractorsBase from '@/components/WebhookExtractorsBase';
import WebhookExtractorBase from '@/components/WebhookExtractorBase';

import WebhookExtractValue from './WebhookExtractValue.vue';
import WebhookMatcher from './WebhookMatcher.vue';
import WebhookExtractorCrumb from './WebhookExtractorCrumb.vue';

export default {
  mixins: [WebhookExtractorsBase, WebhookExtractorBase],
  components: { WebhookMatcher, WebhookExtractValue, WebhookExtractorCrumb },
  computed: {
    webhookId() {
      if (/^-?\d+$/.test(this.$route.params.webhookId)) {
        return parseInt(this.$route.params.webhookId, 10);
      }
      return this.$route.params.webhookId;
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
