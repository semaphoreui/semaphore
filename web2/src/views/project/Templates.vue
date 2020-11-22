<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="!isLoaded">
    <v-progress-linear
      indeterminate
      color="primary darken-2"
    ></v-progress-linear>
  </div>
  <div v-else>
    <EditDialog
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
    </EditDialog>

    <EditDialog
      v-model="newTaskDialog"
      save-button-text="Run"
      title="New Task"
      @save="onTaskCreated"
    >
      <template v-slot:title={}>
        <span class="breadcrumbs__item">{{ templateAlias }}</span>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">New Task</span>
      </template>

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
    </EditDialog>

    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Task Templates</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
        class="mr-1"
      >New template</v-btn>

      <v-btn icon @click="settingsSheet = true"><v-icon>mdi-cog</v-icon></v-btn>
    </v-toolbar>

    <v-data-table
      :headers="filteredHeaders"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
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

    <TableSettingsSheet
      v-model="settingsSheet"
      table-name="project__template"
      :headers="headers"
      @change="onTableSettingsChange"
    />
  </div>
</template>

<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import TemplateForm from '@/components/TemplateForm.vue';
import axios from 'axios';
import TaskForm from '@/components/TaskForm.vue';
import TableSettingsSheet from '@/components/TableSettingsSheet.vue';
import EventBus from '@/event-bus';

export default {
  components: { TemplateForm, TaskForm, TableSettingsSheet },
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
      taskId: null,
      settingsSheet: null,
      filteredHeaders: [],
    };
  },

  computed: {
    templateAlias() {
      if (this.itemId == null || this.itemId === 'new') {
        return '';
      }
      return this.items.find((x) => x.id === this.itemId).alias;
    },

    isLoaded() {
      return this.items
        && this.keys
        && this.inventory
        && this.environment
        && this.repositories;
    },
  },

  methods: {
    onTaskCreated(e) {
      EventBus.$emit('i-show-task', {
        taskId: e.item.id,
      });
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
          width: '0%',
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

    onTableSettingsChange({ headers }) {
      this.filteredHeaders = headers;
    },
  },
};
</script>
