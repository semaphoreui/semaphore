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
      <template v-slot:item.status="{ item }">
        <v-chip
          :color="getStatusColor(item.status)"
          :text-color="getStatusTextColor(item.status)"
          small
        >
          {{ item.status.toUpperCase() }}
        </v-chip>
      </template>

      <template v-slot:item.compliance_score="{ item }">
        <div class="d-flex align-center">
          <v-progress-circular
            :value="item.compliance_score"
            :color="getScoreColor(item.compliance_score)"
            size="24"
            width="3"
            class="mr-2"
          ></v-progress-circular>
          <span :class="getScoreTextColor(item.compliance_score)">
            {{ item.compliance_score }}%
          </span>
        </div>
      </template>

      <template v-slot:item.duration_seconds="{ item }">
        <span v-if="item.duration_seconds">
          {{ formatDuration(item.duration_seconds) }}
        </span>
        <span v-else class="grey--text">-</span>
      </template>

      <template v-slot:item.created="{ item }">
        {{ formatDate(item.created) }}
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
          No issues
        </span>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn
          icon
          small
          @click="viewTaskDetails(item)"
        >
          <v-icon small>mdi-eye</v-icon>
        </v-btn>
      </template>

      <template v-slot:no-data>
        <div class="text-center py-4">
          <v-icon size="48" color="grey">mdi-format-list-bulleted</v-icon>
          <div class="text-h6 grey--text mt-2">No task compliance data available</div>
        </div>
      </template>
    </v-data-table>
  </div>
</template>

<script>
export default {
  name: 'TaskComplianceTable',

  props: {
    data: {
      type: Array,
      default: () => [],
    },
    loading: {
      type: Boolean,
      default: false,
    }
  },

  data() {
    return {
      headers: [
        {
          text: 'Task ID',
          value: 'task_id',
          sortable: true,
          width: '80px',
        },
        {
          text: 'Project',
          value: 'project_name',
          sortable: true,
        },
        {
          text: 'Template',
          value: 'template_name',
          sortable: true,
        },
        {
          text: 'Status',
          value: 'status',
          sortable: true,
          width: '100px',
        },
        {
          text: 'Compliance',
          value: 'compliance_score',
          sortable: true,
          width: '120px',
        },
        {
          text: 'Duration',
          value: 'duration_seconds',
          sortable: true,
          width: '100px',
        },
        {
          text: 'User',
          value: 'username',
          sortable: true,
          width: '120px',
        },
        {
          text: 'Created',
          value: 'created',
          sortable: true,
          width: '120px',
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
    getStatusColor(status) {
      const colorMap = {
        success: 'success',
        failed: 'error',
        stopped: 'warning',
        running: 'info',
        waiting: 'grey',
      };
      return colorMap[status] || 'grey';
    },

    getStatusTextColor(status) {
      const colorMap = {
        success: 'white',
        failed: 'white',
        stopped: 'white',
        running: 'white',
        waiting: 'white',
      };
      return colorMap[status] || 'white';
    },

    getScoreColor(score) {
      if (score >= 90) return 'success';
      if (score >= 70) return 'warning';
      return 'error';
    },

    getScoreTextColor(score) {
      if (score >= 90) return 'success--text';
      if (score >= 70) return 'warning--text';
      return 'error--text';
    },

    formatDuration(seconds) {
      if (!seconds) return '-';

      const hours = Math.floor(seconds / 3600);
      const minutes = Math.floor((seconds % 3600) / 60);
      const secs = seconds % 60;

      if (hours > 0) {
        return `${hours}h ${minutes}m`;
      } else if (minutes > 0) {
        return `${minutes}m ${secs}s`;
      }
      return `${secs}s`;
    },

    formatDate(dateString) {
      if (!dateString) return '-';
      return new Date(dateString).toLocaleString();
    },

    viewTaskDetails(task) {
      // Emit event to parent component to handle task details view
      this.$emit('view-task', task);
    },
  }
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