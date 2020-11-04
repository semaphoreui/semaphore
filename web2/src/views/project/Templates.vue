<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="isLoaded">
    <ItemDialog
      v-model="editDialog"
      save-button-text="Create"
      title="New template"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TemplateForm
          :project-id="projectId"
          item-id="new"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </ItemDialog>

    <ItemDialog
      v-model="newTaskDialog"
      save-button-text="Run"
      title="New Task"
      @save="onTaskCreate"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TaskForm
          :project-id="projectId"
          item-id="new"
          :template-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </ItemDialog>

    <ItemDialog
      v-model="taskLogDialog"
      save-button-text="Delete"
      title="Task Log"
      @save="onTaskAskDelete"
    >
      <template v-slot:form="{}">
        <TaskLogView :item-id="taskId" />
      </template>
    </ItemDialog>

    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Task Templates</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
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

      <template v-slot:item.ssh_key_id="{ item }">
        {{ keys.find((x) => x.id === item.ssh_key_id).name }}
      </template>

      <template v-slot:item.inventory_id="{ item }">
        {{ inventory.find((x) => x.id === item.inventory_id).name }}
      </template>

      <template v-slot:item.environment_id="{ item }">
        {{ environment.find((x) => x.id === item.environment_id).name }}
      </template>

      <template v-slot:item.repository_id="{ item }">
        {{ repositories.find((x) => x.id === item.repository_id).name }}
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn text color="black" class="pl-1 pr-2" @click="createTask(item.id)">
          <v-icon class="pr-1">mdi-play</v-icon>
          Run
        </v-btn>
      </template>
    </v-data-table>
  </div>

</template>
<style lang="scss">

</style>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import TemplateForm from '@/components/TemplateForm.vue';
import axios from 'axios';
import TaskForm from '@/components/TaskForm.vue';
import TaskLogView from '@/components/TaskLogView.vue';

export default {
  components: { TaskLogView, TemplateForm, TaskForm },
  mixins: [ItemListPageBase],
  async created() {
    await this.loadData();
  },
  data() {
    return {
      keys: null,
      inventory: null,
      environment: null,
      repositories: null,
      newTaskDialog: null,
      taskLogDialog: null,
      taskId: null,
    };
  },
  computed: {
    isLoaded() {
      return this.items && this.keys && this.inventory && this.environment && this.repositories;
    },
  },
  methods: {
    onTaskCreate(e) {
      this.taskId = e.item.id;
      this.taskLogDialog = true;
    },

    onTaskAskDelete() {
      // TODO: stop task
    },

    createTask(itemId) {
      this.itemId = itemId;
      this.newTaskDialog = true;
    },

    getHeaders() {
      return [
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
          value: 'ssh_key_id',
          sortable: false,
        },
        {
          text: 'Inventory',
          value: 'inventory_id',
          sortable: false,
        },
        {
          text: 'Environment',
          value: 'environment_id',
          sortable: false,
        },
        {
          text: 'Repository',
          value: 'repository_id',
          sortable: false,
        },
        {
          text: 'Actions',
          value: 'actions',
          sortable: false,
        },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/templates`;
    },

    async loadData() {
      this.inventory = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/inventory`,
        responseType: 'json',
      })).data;

      this.environment = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/environment`,
        responseType: 'json',
      })).data;

      this.keys = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/keys`,
        responseType: 'json',
      })).data;

      this.repositories = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/repositories`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
