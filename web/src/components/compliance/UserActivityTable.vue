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
      <template v-slot:item.active="{ item }">
        <v-chip
          :color="item.active ? 'success' : 'grey'"
          :text-color="item.active ? 'white' : 'white'"
          small
        >
          <v-icon small left>
            {{ item.active ? 'mdi-check-circle' : 'mdi-close-circle' }}
          </v-icon>
          {{ item.active ? 'Active' : 'Inactive' }}
        </v-chip>
      </template>

      <template v-slot:item.admin="{ item }">
        <v-chip
          :color="item.admin ? 'primary' : 'grey'"
          :text-color="item.admin ? 'white' : 'white'"
          x-small
        >
          {{ item.admin ? 'Admin' : 'User' }}
        </v-chip>
      </template>

      <template v-slot:item.external="{ item }">
        <v-chip
          :color="item.external ? 'warning' : 'success'"
          :text-color="item.external ? 'white' : 'white'"
          x-small
        >
          {{ item.external ? 'External' : 'Internal' }}
        </v-chip>
      </template>

      <template v-slot:item.last_activity="{ item }">
        <div v-if="item.last_activity">
          <div class="text-body-2">{{ formatDate(item.last_activity) }}</div>
          <div class="text-caption grey--text">
            {{ formatRelativeTime(item.last_activity) }}
          </div>
        </div>
        <span v-else class="grey--text">Never</span>
      </template>

      <template v-slot:item.total_tasks="{ item }">
        <v-chip
          :color="getTaskCountColor(item.total_tasks)"
          :text-color="'white'"
          small
        >
          {{ item.total_tasks }}
        </v-chip>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn
          icon
          small
          @click="viewUserDetails(item)"
        >
          <v-icon small>mdi-eye</v-icon>
        </v-btn>
      </template>

      <template v-slot:no-data>
        <div class="text-center py-4">
          <v-icon size="48" color="grey">mdi-account-multiple</v-icon>
          <div class="text-h6 grey--text mt-2">No user activity data available</div>
        </div>
      </template>
    </v-data-table>
  </div>
</template>

<script>
export default {
  name: 'UserActivityTable',

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
          text: 'Username',
          value: 'username',
          sortable: true,
        },
        {
          text: 'Name',
          value: 'name',
          sortable: true,
        },
        {
          text: 'Email',
          value: 'email',
          sortable: true,
        },
        {
          text: 'Status',
          value: 'active',
          sortable: true,
          width: '100px',
        },
        {
          text: 'Role',
          value: 'admin',
          sortable: true,
          width: '80px',
        },
        {
          text: 'Type',
          value: 'external',
          sortable: true,
          width: '80px',
        },
        {
          text: 'Tasks',
          value: 'total_tasks',
          sortable: true,
          width: '80px',
        },
        {
          text: 'Last Activity',
          value: 'last_activity',
          sortable: true,
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
    getTaskCountColor(count) {
      if (count >= 50) return 'success';
      if (count >= 10) return 'info';
      if (count > 0) return 'warning';
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

      if (diffDays > 0) {
        return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
      }
      if (diffHours > 0) {
        return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
      }
      if (diffMinutes > 0) {
        return `${diffMinutes} minute${diffMinutes > 1 ? 's' : ''} ago`;
      }
      return 'Just now';
    },

    viewUserDetails(user) {
      // Emit event to parent component to handle user details view
      this.$emit('view-user', user);
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
