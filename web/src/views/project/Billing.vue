<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <YesNoDialog
      v-model="deleteProjectDialog"
      :title="$t('deleteProject')"
      :text="$t('askDeleteProj')"
      @yes="deleteProject()"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <v-tabs show-arrows class="pl-4">
      <v-tab key="history" :to="`/project/${projectId}/history`">{{ $t('history') }}</v-tab>
      <v-tab key="activity" :to="`/project/${projectId}/activity`">{{ $t('activity') }}</v-tab>
      <v-tab key="settings" :to="`/project/${projectId}/settings`">{{ $t('settings') }}</v-tab>
      <v-tab
        key="billing"
        :to="`/project/${projectId}/billing`"
      >Billing <v-chip color="red" x-small dark class="ml-1">Soon</v-chip></v-tab>
    </v-tabs>

    <v-alert
      text
      color="info"
      class="ma-4"
    >
      <h3 class="text-h5">
        Coming soon
      </h3>
      <div>The billing will be available soon.</div>
    </v-alert>
  </div>
</template>
<style lang="scss">
</style>
<script>
import EventBus from '@/event-bus';
import YesNoDialog from '@/components/YesNoDialog.vue';

export default {
  components: { YesNoDialog },
  props: {
    projectId: Number,
  },

  data() {
    return {
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
