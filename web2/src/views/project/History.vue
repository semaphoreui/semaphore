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
<!--      <template v-slot:extension>-->
<!--        <v-tabs centered>-->
<!--          <v-tab key="history" :to="`/project/${projectId}/history`">History</v-tab>-->
<!--          <v-tab key="activity" :to="`/project/${projectId}/activity`">Activity</v-tab>-->
<!--          <v-tab key="settings" :to="`/project/${projectId}/settings`">Settings</v-tab>-->
<!--        </v-tabs>-->
<!--      </template>-->
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
    >
    </v-data-table>
  </div>
</template>
<style lang="scss">

</style>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';

export default {
  components: {
  },
  props: {
    projectId: Number,
  },
  data() {
    return {
      headers: [
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
      ],
      items: null,
    };
  },

  async created() {
    await this.loadItems();
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async editItem(itemId = 'new') {
      this.itemId = itemId;
      this.editDialog = true;
    },

    async loadItems() {
      this.items = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/tasks/last`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
