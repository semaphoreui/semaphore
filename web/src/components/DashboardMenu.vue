<template>
  <div class="DashboardMenu" data-testid="dashboard-menu">
    <v-tabs class="pl-4" data-testid="dashboard-tabs">
      <v-tab
        v-if="projectType === ''"
        key="history"
        :to="`/project/${projectId}/history`"
        data-testid="dashboard-history"
      >{{ $t('history') }}
      </v-tab>

      <v-tab
        v-if="projectType === ''"
        key="stats"
        :to="`/project/${projectId}/stats`"
        data-testid="dashboard-stats"
      >{{ $t('project_stats') }}</v-tab>

      <v-tab
        key="activity"
        :to="`/project/${projectId}/activity`"
        data-testid="dashboard-activity"
      >{{ $t('activity') }}</v-tab>

      <v-tab
        v-if="canUpdateProject"
        key="settings"
        :to="`/project/${projectId}/settings`"
        data-testid="dashboard-settings"
      >{{ $t('settings') }}
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;" />
  </div>
</template>
<style lang="scss">
// The slider of Vuetify 2 is drawn over the divider under the tabs
// (`.v-tabs-slider-wrapper { z-index: 1 }`); the divider of Vuetify 4 is translucent and so
// painted above a slider without a z-index, which turns the slider darker.
.DashboardMenu .v-tab__slider {
  z-index: 1;
}
</style>
<script>

export default {

  props: {
    projectId: Number,
    projectType: String,
    canUpdateProject: Boolean,
  },

  computed: {
    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },
  },

  data() {
    return {
      id: null,
    };
  },

};
</script>
