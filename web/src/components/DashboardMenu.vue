<template>
  <div>
    <v-tabs show-arrows class="pl-4">
      <v-tab
        v-if="projectType === ''"
        key="history"
        :to="`/project/${projectId}/history`"
      >{{ $t('history') }}
      </v-tab>

      <v-tab key="activity" :to="`/project/${projectId}/activity`">{{ $t('activity') }}</v-tab>

      <v-tab
        v-if="canUpdateProject"
        key="settings"
        :to="`/project/${projectId}/settings`"
      >{{ $t('settings') }}
      </v-tab>

      <v-tab
        v-if="projectType === ''"
        key="runners"
        :to="`/project/${projectId}/runners`"
      >
        {{ $t('runners') }}
        <!-- <v-chip small class="ml-1" color="purple" style="color: white">Pro</v-chip> -->
        <v-icon class="ml-1" large color="hsl(348deg, 86%, 61%)">mdi-professional-hexagon</v-icon>
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;" />
  </div>
</template>
<script>
import PermissionsCheck from '@/components/PermissionsCheck';

export default {

  mixins: [PermissionsCheck],

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
