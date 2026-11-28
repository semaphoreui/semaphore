<template>
  <div v-if="integration != null" data-testid="integration-header">
    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="integration-drawerToggle"
      ></v-app-bar-nav-icon>

      <v-toolbar-title class="breadcrumbs flex-0-1" data-testid="integration-breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/integrations/`"
          data-testid="integration-backToList"
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
        <span
          class="breadcrumbs__item"
          data-testid="integration-title"
        >{{ integration.name }}</span>
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
