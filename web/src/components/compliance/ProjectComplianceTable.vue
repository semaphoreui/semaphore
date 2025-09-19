<template>
  <div>
    <v-data-table
      :headers="headers"
      :items="data"
      :items-per-page="10"
      :loading="loading"
      class="elevation-1"
      dense
    >
      <template v-slot:item.success_rate="{ item }">
        <div class="d-flex align-center">
          <v-progress-circular
            :value="item.success_rate"
            :color="getSuccessRateColor(item.success_rate)"
            size="24"
            width="3"
            class="mr-2"
          ></v-progress-circular>
          <span :class="getSuccessRateTextColor(item.success_rate)">
            {{ item.success_rate.toFixed(1) }}%
          </span>
        </div>
      </template>

      <template v-slot:item.compliance_score="{ item }">
        <div class="d-flex align-center">
          <v-progress-circular
            :value="item.compliance_score"
            :color="getComplianceScoreColor(item.compliance_score)"
            size="24"
            width="3"
            class="mr-2"
          ></v-progress-circular>
          <span :class="getComplianceScoreTextColor(item.compliance_score)">
            {{ item.compliance_score }}%
          </span>
        </div>
      </template>

      <template v-slot:item.last_activity="{ item }">
        <div v-if="item.last_activity">
          <div class="text-body-2">{{ formatDate(item.last_activity) }}</div>
          <div class="text-caption grey--text">
            {{ formatRelativeTime(item.last_activity) }}
          </div>
        </div>
        <span v-else class="grey--text">No activity</span>
      </template>

      <template v-slot:item.team_size="{ item }">
        <v-chip
          :color="getTeamSizeColor(item.team_size)"
          :text-color="'white'"
          small
        >
          {{ item.team_size }} member{{ item.team_size !== 1 ? s: '' }}
        </v-chip>
      </template>

      <template v-slot:item.issues="{ item }">
        <div v-if="item.issues && item.issues.length > 0">
          <v-chip
            v-for="(issue, index) in item.issues.slice(0, 2)"
            :key="index"
            color="error"
            text-color="white"
            x-small
            class="mr-1 mb-1"
          >
            {{ issue }}
          </v-chip>
          <v-chip
            v-if="item.issues.length > 2"
            color="grey"
            text-color="white"
            x-small
            class="mr-1 mb-1"
          >
            +{{ item.issues.length - 2 }} more
          </v-chip>
        </div>
        <span v-else class="success--text">
          <v-icon small>mdi-check</v-icon>
          Compliant
        </span>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn
          icon
          small
          @click="viewProjectDetails(item)"
        >
          <v-icon small>mdi-eye</v-icon>
        </v-btn>
      </template>

      <template v-slot:no-data>
        <div class="text-center py-4">
          <v-icon size="48" color="grey">mdi-folder-multiple</v-icon>
          <div class="text-h6 grey--text mt-2">No project compliance data available</div>
        </div>
      </template>
    </v-data-table>
  </div>
</template>

<script>
export default {
  name: 'ProjectComplianceTable',

  props: {
    data: {
      type: Array,
      default: () => [],
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },

  data() {
    return {
      headers: [
        {
          text: 'Project Name',
          value: 'project_name',
          sortable: true,
        },
        {
          text: 'Success Rate',
          value: 'success_rate',
          sortable: true,
          width: '120px',
        },
        {
          text: 'Compliance',
          value: 'compliance_score',
          sortable: true,
          width: '120px',
        },
        {
          text: 'Tasks',
          value: 'total_tasks',
          sortable: true,
          width: '80px',
        },
        {
          text: 'Team Size',
          value: 'team_size',
          sortable: true,
          width: '100px',
        },
        {
          text: 'Last Activity',
          value: 'last_activity',
          sortable: true,
          width: '150px',
        },
        {
          text: 'Issues',
          value: 'issues',
          sortable: false,
          width: '150px',
        },
        {
          text: 'Actions',
          value: 'actions',
          sortable: false,
          width: '80px',
        },
      ],
    };
  },

  methods: {
    getSuccessRateColor(rate) {
      if (rate >= 90) return 'success';
      if (rate >= 70) return 'warning';
      return 'error';
    },

    getSuccessRateTextColor(rate) {
      if (rate >= 90) return 'success--text';
      if (rate >= 70) return 'warning--text';
      return 'error--text';
    },

    getComplianceScoreColor(score) {
      if (score >= 90) return 'success';
      if (score >= 70) return 'warning';
      return 'error';
    },

    getComplianceScoreTextColor(score) {
      if (score >= 90) return 'success--text';
      if (score >= 70) return 'warning--text';
      return 'error--text';
    },

    getTeamSizeColor(size) {
      if (size >= 10) return 'success';
      if (size >= 5) return 'info';
      if (size > 0) return 'warning';
      return 'grey';
    },

    formatDate(dateString) {
      if (!dateString) return '-';
      return new Date(dateString).toLocaleString();
    },

    formatRelativeTime(dateString) {
      if (!dateString) return '';

      const now = new Date();
      const date = new Date(dateString);
      const diffMs = now - date;
      const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
      const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
      const diffMinutes = Math.floor(diffMs / (1000 * 60));

      if (diffDays > 0) { return `${diffDays} day${diffDays > 1 ? s: ''} ago`;
      } else if (diffHours > 0) {
        return `${diffHours} hour${diffHours > 1 ? s: ''} ago`;
      } else if (diffMinutes > 0) {
        return ; }
      return 'Just now';
    },

    viewProjectDetails(project) {
      // Emit event to parent component to handle project details view
      this.$emit('view-project', project);
    },
  },
};
</script>

<style scoped>
.v-data-table {
  font-size: 0.875rem;
}

.v-chip {
  font-size: 0.75rem;
}
</style>