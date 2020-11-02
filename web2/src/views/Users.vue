<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <v-toolbar flat color="white">
      <v-btn
        icon
        class="mr-4"
        @click="returnToProjects()"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>Users</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem()"
      >New User</v-btn>
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
  data() {
    return {
      headers: [
        {
          text: 'Alias',
          value: 'alias',
        },
        {
          text: 'Playbook',
          value: 'playbook',
          sortable: false,
        },
        {
          text: 'SSH key',
          value: 'email',
          sortable: false,
        },
        {
          text: 'Inventory',
          value: 'inventory',
          sortable: false,
        },
        {
          text: 'Environment',
          value: 'environment',
          sortable: false,
        },
        {
          text: 'Repository',
          value: 'repository',
          sortable: false,
        },
        {
          text: '',
          value: 'actions',
          sortable: false,
        },
      ],
      items: null,
      itemId: null,
      editDialog: null,
    };
  },

  async created() {
    await this.loadItems();
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    async editItem(itemId = 'new') {
      this.itemId = itemId;
      this.editDialog = true;
    },

    async loadItems() {
      this.items = (await axios({
        method: 'get',
        url: '/api/users',
        responseType: 'json',
      })).data;
    },
  },
};
</script>
