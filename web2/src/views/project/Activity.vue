<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items">
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
      <template v-slot:item.created="{ item }">
        {{ item.created | formatDate }}
      </template>
    </v-data-table>
  </div>
</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';

export default {
  mixins: [ItemListPageBase],

  methods: {
    getHeaders() {
      return [
        {
          text: 'Time',
          value: 'created',
          sortable: false,
          width: '20%',
        },
        {
          text: 'Description',
          value: 'description',
          sortable: false,
          width: '80%',
        },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/events/last`;
    },
  },
};
</script>
