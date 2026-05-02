<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="integration != null">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>

      <v-toolbar-title class="breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/integrations/`"
        >
          Integrations
        </router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ integration.name }}</span>
      </v-toolbar-title>

      <v-spacer></v-spacer>
    </v-toolbar>
  </div>
</template>
<script>

import ItemListPageBase from '@/components/ItemListPageBase';

export default {
  mixins: [ItemListPageBase],
  components: { },
  props: {
    integration: Object,
    integrationId: Number,
  },
  data() {
    return {
      extractor: null,
    };
  },

  methods: {
    allowActions() {
      return true;
    },
    getHeaders() {
      return [];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/integrations`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.integrationId}`;
    },
    getEventName() {
      return 'w-integration-matcher';
    },
  },
};
</script>
