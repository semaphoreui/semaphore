<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="integration != null">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>

      <v-toolbar-title class="breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/integrations/`"
        >
          {{ $t('integrations') }}
        </router-link>
        <template v-if="group != null">
          <v-icon>mdi-chevron-right</v-icon>
          <router-link
            class="breadcrumbs__item breadcrumbs__item--link"
            :to="`/project/${projectId}/integrations/`"
          >
            {{ group.name }}
          </router-link>
        </template>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ integration.name }}</span>
      </v-toolbar-title>

      <v-spacer></v-spacer>
    </v-toolbar>
  </div>
</template>
<script>
import EventBus from '@/event-bus';

export default {
  props: {
    integration: Object,
    /** The integration's group; null while it loads. */
    group: Object,
    projectId: Number,
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },
  },
};
</script>
