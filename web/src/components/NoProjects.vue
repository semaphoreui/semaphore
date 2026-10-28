<template>
  <div data-testid="noProjects-page">
    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="noProjects-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title data-testid="noProjects-title">{{ $t('noProjectsTitle') }}</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <v-divider/>

    <div
      class="text-center"
      style="margin: auto; max-width: 600px; padding: 64px 16px;"
      data-testid="noProjects-empty"
    >
      <v-icon size="64" color="grey" data-testid="noProjects-icon">mdi-folder-outline</v-icon>
      <div class="text-h6 mt-4" data-testid="noProjects-heading">{{ $t('noProjectsTitle') }}</div>
      <div class="mt-2" data-testid="noProjects-text">{{ $t('noProjectsText') }}</div>
      <i18n
        v-if="user"
        data-testid="noProjects-signedInAs"
        :path="user.email ? 'signedInAs' : 'signedInAsNoEmail'"
        tag="div"
        class="mt-6 text-grey"
      >
        <template #username><b>{{ user.username }}</b></template>
        <template #email>{{ user.email }}</template>
      </i18n>
    </div>
  </div>
</template>
<script>
import EventBus from '@/event-bus';

export default {
  props: {
    user: Object,
  },
  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },
  },
};
</script>
