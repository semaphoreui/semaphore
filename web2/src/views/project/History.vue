<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <ItemDialog
      v-model="taskLogDialog"
      save-button-text="Delete"
      title="Task Log"
      :max-width="800"
    >
      <template v-slot:form="{}">
        <TaskLogView :project-id="projectId" :item-id="taskId" />
      </template>
    </ItemDialog>

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
        {{ item.status }}
      </template>
    </v-data-table>
  </div>
</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import ItemDialog from '@/components/ItemDialog.vue';
import TaskLogView from '@/components/TaskLogView.vue';

export default {
  components: {
    ItemDialog, TaskLogView,
  },

  mixins: [ItemListPageBase],

  data() {
    return {
      taskLogDialog: null,
      taskId: null,
    };
  },

  watch: {
    async projectId() {
      await this.loadItems();
    },
  },

  methods: {
    showTaskLog(taskId) {
      this.taskId = taskId;
      this.taskLogDialog = true;
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
          value: 'start',
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
