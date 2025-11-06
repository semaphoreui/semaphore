<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <DashboardMenu
      :project-id="projectId"
      :project-type="projectType"
      :can-update-project="can(USER_PERMISSIONS.updateProject)"
    />

    <TaskStats class="mt-4" :project-id="projectId"  />

  </div>
</template>
<script>
import DashboardMenu from '@/components/DashboardMenu.vue';
import {
  TEMPLATE_TYPE_ACTION_TITLES,
  TEMPLATE_TYPE_ICONS,
  TEMPLATE_TYPE_TITLES,
  USER_PERMISSIONS,
} from '@/lib/constants';
import TaskStats from '@/components/TaskStats.vue';
import PermissionsCheck from '@/components/PermissionsCheck';

export default {
  computed: {
    USER_PERMISSIONS() {
      return USER_PERMISSIONS;
    },
  },

  components: { TaskStats, DashboardMenu },

  mixins: [PermissionsCheck],

  props: {
    projectId: Number,
    projectType: String,
    userId: Number,
    userRole: String,
    user: Object,
  },

  data() {
    return {
      dateRanges: [{
        text: 'Past week',
        value: 'last_week',
      }, {
        text: 'Past month',
        value: 'last_month',
      }, {
        text: 'Past year',
        value: 'last_year',
      }],
      users: [{
        text: 'All users',
        value: null,
      }],

      TEMPLATE_TYPE_ICONS,
      TEMPLATE_TYPE_TITLES,
      TEMPLATE_TYPE_ACTION_TITLES,
      stats: null,
      dateRange: 'last_week',
    };
  },

  methods: {
  },
};
</script>
