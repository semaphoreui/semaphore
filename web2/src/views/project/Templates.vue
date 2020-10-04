<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <TemplateEditDialog
      :template-id="itemId"
      v-model="editDialog"
      @saved="onSaved"
    />

    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Task Templates</v-toolbar-title>
      <v-spacer></v-spacer>
      <!--    :to="`/project/${projectId}/templates/new/edit`"-->
      <v-btn
        color="primary"
        @click="editItem()"
      >New template</v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
    >
      <template v-slot:item.alias="{ item }">
        <router-link :to="`/project/${projectId}/templates/${item.id}`">
          {{ item.alias }}
        </router-link>
      </template>

      <template v-slot:item.actions="{}">
        <v-btn icon>
          <v-icon>mdi-history</v-icon>
        </v-btn>
        <v-btn icon>
          <v-icon>mdi-play</v-icon>
        </v-btn>
      </template>
    </v-data-table>
  </div>

</template>
<style lang="scss">

</style>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import TemplateEditDialog from '@/components/TemplateEditDialog.vue';

export default {
  components: {
    TemplateEditDialog,
  },
  props: {
    projectId: Number,
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
          text: 'Actions',
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

    onSaved(e) {
      EventBus.$emit('i-snackbar', {
        color: 'success',
        text: e.action === 'new' ? `Template "${e.item.alias}" created` : `Template "${e.item.alias}" saved`,
      });
    },

    async editItem(itemId) {
      this.editItemId = itemId;
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
