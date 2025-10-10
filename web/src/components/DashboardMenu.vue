<template>
  <div>
    <v-tabs class="pl-4">
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

      <v-tab key="activity" :to="`/project/${projectId}/activity`">{{ $t('activity') }}</v-tab>

      <v-tab
        v-if="canUpdateProject"
        key="settings"
        :to="`/project/${projectId}/settings`"
        data-testid="dashboard-settings"
      >{{ $t('settings') }}
      </v-tab>

      <v-tab
        v-if="canUpdateProject && projectType === ''"
        key="runners"
        :to="`/project/${projectId}/runners`"
        data-testid="dashboard-runners"
      >
        {{ $t('runners') }}
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;" />
  </div>
</template>
<script>

export default {

  props: {
    projectId: Number,
    projectType: String,
    canUpdateProject: Boolean,
  },

  data() {
    return {
      id: null,
    };
  },

};
</script>
