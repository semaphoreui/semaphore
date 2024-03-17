<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <v-tabs show-arrows class="pl-4">
      <v-tab
        v-if="projectType === 'premium'"
        key="install"
        :to="`/project/${projectId}/install`"
      >Install
      </v-tab>

      <v-tab
        v-if="projectType === ''"
        key="history"
        :to="`/project/${projectId}/history`"
      >{{ $t('history') }}
      </v-tab>
      <v-tab key="activity" :to="`/project/${projectId}/activity`">{{ $t('activity') }}</v-tab>
      <v-tab key="settings" :to="`/project/${projectId}/settings`">{{ $t('settings') }}</v-tab>
      <v-tab
        key="billing"
        :to="`/project/${projectId}/billing`"
      >Billing
      </v-tab>
    </v-tabs>

    <div>
      Coming soon
    </div>
  </div>
</template>
<style lang="scss">
</style>
<script>
import EventBus from '@/event-bus';

export default {
  components: {},
  props: {
    projectId: Number,
    projectType: String,
  },

  data() {
    return {
      deleteProjectDialog: null,
    };
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    onError(e) {
      EventBus.$emit('i-snackbar', {
        color: 'error',
        text: e.message,
      });
    },

  },
};
</script>
