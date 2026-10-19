<template>
  <div>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('noProjectsTitle') }}</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <v-divider/>

    <div class="text-center" style="margin: auto; max-width: 600px; padding: 64px 16px;">
      <v-icon size="64" color="grey">mdi-folder-outline</v-icon>
      <div class="text-h6 mt-4">{{ $t('noProjectsTitle') }}</div>
      <div class="mt-2">{{ $t('noProjectsText') }}</div>
      <i18n
        v-if="user"
        :path="user.email ? 'signedInAs' : 'signedInAsNoEmail'"
        tag="div"
        class="mt-6 grey--text"
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
