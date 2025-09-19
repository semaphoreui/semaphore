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
      <template v-slot:item.severity="{ item }">
        <v-chip
          :color="getSeverityColor(item.severity)"
          :text-color="'white'"
          small
        >
          <v-icon small left>
            {{ getSeverityIcon(item.severity) }}
          </v-icon>
          {{ item.severity.toUpperCase() }}
        </v-chip>
      </template>

      <template v-slot:item.resolved="{ item }">
        <v-chip
          :color="item.resolved ? success: 'error'"
          :text-color="'white'"
          x-small
        >
          <v-icon small left>
            {{ item.resolved ? 'mdi-check' : 'mdi-alert' }}
          </v-icon>
          {{ item.resolved ? Resolved: 'Open' }}
        </v-chip>
      </template>

      <template v-slot:item.created="{ item }">
        <div>
          <div class="text-body-2">{{ formatDate(item.created) }}</div>
          <div class="text-caption grey--text">
            {{ formatRelativeTime(item.created) }}
          </div>
        </div>
      </template>

      <template v-slot:item.description="{ item }">
        <div class="text-truncate" style="max-width: 200px;" :title="item.description">
          {{ item.description }}
        </div>
      </template>

      <template v-slot:item.actions="{ item }">
        <div class="d-flex">
          <v-btn
            icon
            small
            @click="viewEventDetails(item)"
            class="mr-1"
          >
            <v-icon small>mdi-eye</v-icon>
          </v-btn>
          <v-btn
            v-if="!item.resolved"
            icon
            small
            color="success"
            @click="resolveEvent(item)"
          >
            <v-icon small>mdi-check</v-icon>
          </v-btn>
        </div>
      </template>

      <template v-slot:no-data>
        <div class="text-center py-4">
          <v-icon size="48" color="grey">mdi-security</v-icon>
          <div class="text-h6 grey--text mt-2">No security events available</div>
        </div>
      </template>
    </v-data-table>
  </div>
</template>

<script>
export default {
  name: 'SecurityEventsTable',

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
          text: 'Event ID',
          value: 'event_id',
          sortable: true,
          width: '80px',
        },
        {
          text: 'Type',
          value: 'event_type',
          sortable: true,
          width: '100px',
        },
        {
          text: 'Severity',
          value: 'severity',
          sortable: true,
          width: '100px',
        },
        {
          text: 'Description',
          value: 'description',
          sortable: false,
        },
        {
          text: 'User',
          value: 'username',
          sortable: true,
          width: '120px',
        },
        {
          text: 'Project',
          value: 'project_name',
          sortable: true,
          width: '120px',
        },
        {
          text: 'Status',
          value: 'resolved',
          sortable: true,
          width: '100px',
        },
        {
          text: 'Created',
          value: 'created',
          sortable: true,
          width: '150px',
        },
        {
          text: 'Actions',
          value: 'actions',
          sortable: false,
          width: '120px',
        },
      ],
    };
  },

  methods: {
    getSeverityColor(severity) {
      const colorMap = {
        high: 'error',
        medium: 'warning',
        low: 'info',
      };
      return colorMap[severity] || 'grey';
    },

    getSeverityIcon(severity) {
      const iconMap = {
        high: 'mdi-alert-circle',
        medium: 'mdi-alert',
        low: 'mdi-information',
      };
      return iconMap[severity] || 'mdi-help-circle';
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

    viewEventDetails(event) {
      // Emit event to parent component to handle event details view
      this.$emit('view-event', event);
    },

    resolveEvent(event) {
      // Emit event to parent component to handle event resolution
      this.$emit('resolve-event', event);
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

.text-truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>