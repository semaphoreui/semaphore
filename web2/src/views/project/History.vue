<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Dashboard</v-toolbar-title>
      <v-spacer></v-spacer>
      <div>
        <v-tabs centered>
          <v-tab key="history" :to="`/project/${projectId}/history`">History</v-tab>
          <v-tab key="activity" :to="`/project/${projectId}/activity`">Activity</v-tab>
          <v-tab key="settings" :to="`/project/${projectId}/settings`">Settings</v-tab>
        </v-tabs>
      </div>
    </v-toolbar>
    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
    >
      <template v-slot:item.tpl_alias="{ item }">
        <a @click="showTaskLog(item.id)">{{ item.tpl_alias }}</a>
        <span style="color: gray; margin-left: 10px;">#{{ item.id }}</span>
      </template>

      <template v-slot:item.status="{ item }">
        <v-chip style="font-weight: bold;" :color="getStatusColor(item.status)">
          <v-icon left>{{ getStatusIcon(item.status) }}</v-icon>
          {{ humanizeStatus(item.status) }}
        </v-chip>
      </template>

      <template v-slot:item.start="{ item }">
        <span v-if="item.start">{{ item.start | formatDate }}</span>
        <v-chip v-else>Not started</v-chip>
      </template>

      <template v-slot:item.end="{ item }">
        <span v-if="item.end">{{ (item.end - item.start) | formatMinutes }}</span>
        <v-chip v-else>Not ended</v-chip>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import EventBus from '@/event-bus';

export default {
  mixins: [ItemListPageBase],

  watch: {
    async projectId() {
      await this.loadItems();
    },
  },

  methods: {
    getStatusIcon(status) {
      switch (status) {
        case 'error':
          return 'mdi-check-circle';
        default:
          return status;
      }
    },

    humanizeStatus(status) {
      switch (status) {
        case 'error':
          return 'Failed';
        default:
          return status;
      }
    },

    getStatusColor(status) {
      return status;
    },

    showTaskLog(taskId) {
      EventBus.$emit('i-show-task', {
        taskId,
      });
    },

    getHeaders() {
      return [
        {
          text: 'Task',
          value: 'tpl_alias',
          sortable: false,
        },
        {
          text: 'Status',
          value: 'status',
          sortable: false,
        },
        {
          text: 'User',
          value: 'user_name',
          sortable: false,
        },
        {
          text: 'Start',
          value: 'start',
          sortable: false,
        },
        {
          text: 'Duration',
          value: 'end',
          sortable: false,
        },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/tasks/last`;
    },
  },
};
</script>
