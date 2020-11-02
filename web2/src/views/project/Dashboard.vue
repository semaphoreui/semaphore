<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Dashboard</v-toolbar-title>
      <v-spacer></v-spacer>
      <template v-slot:extension>
        <v-tabs v-model="tab">
          <v-tab key="history">History</v-tab>
          <v-tab key="activity">Activity</v-tab>
          <v-tab key="settings">Settings</v-tab>
        </v-tabs>
      </template>
    </v-toolbar>
    <v-tabs-items v-model="tab">
      <v-tab-item key="history">
        History test
      </v-tab-item>
      <v-tab-item key="activity">
        Activity test
      </v-tab-item>
      <v-tab-item key="settings">
        Settings test
      </v-tab-item>
    </v-tabs-items>
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
      tab: null,
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
        url: `/api/project/${this.projectId}/templates`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
