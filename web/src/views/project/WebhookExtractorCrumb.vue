<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="extractor != null">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title class="breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/webhook/${this.webhookId}`"
          >
          {{ webhook.name }}
        </router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ extractor.name }}</span>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">Extractor Configuration</span>
      </v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>
  </div>
</template>
<script>
import axios from 'axios';

import ItemListPageBase from '@/components/ItemListPageBase';

import WebhookExtractorsBase from '@/components/WebhookExtractorsBase';
import WebhookExtractorBase from '@/components/WebhookExtractorBase';

export default {
  mixins: [ItemListPageBase, WebhookExtractorsBase, WebhookExtractorBase],
  components: { },
  data() {
    return {
      webhook: null,
      extractor: null,
    };
  },
  async created() {
    this.webhook = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/webhook/${this.webhookId}`,
      responseType: 'json',
    })).data;

    this.extractor = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/webhook/${this.webhookId}/extractor/${this.extractorId}`,
      responseType: 'json',
    })).data;
  },

  computed: {
    projectId() {
      if (/^-?\d+$/.test(this.$route.params.projectId)) {
        return parseInt(this.$route.params.projectId, 10);
      }
      return this.$route.params.projectId;
    },
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
    getHeaders() {
      return [];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/webhook/${this.webhookId}/extractors`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/webhook/${this.webhookId}/extractor/${this.extractorId}`;
    },
    getEventName() {
      return 'w-webhook-matcher';
    },
  },
};
</script>
