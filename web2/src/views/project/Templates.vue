<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="!isLoaded">
    <v-progress-linear
        indeterminate
        color="primary darken-2"
    ></v-progress-linear>
  </div>
  <div v-else>
    <v-dialog
        v-model="editViewsDialog"
        :max-width="400"
        persistent
        :transition="false"
    >
      <v-card>
        <v-card-title>
          Edit Views
          <v-spacer></v-spacer>
          <v-btn icon @click="closeEditViewDialog()">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text>
          <EditViewsForm :project-id="projectId"/>
        </v-card-text>
      </v-card>
    </v-dialog>

    <EditDialog
        :max-width="700"
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
        :save-button-text="TEMPLATE_TYPE_ACTION_TITLES[templateType]"
        title="New Task"
        @save="onTaskCreated"
        @close="this.itemId = null"
    >
      <template v-slot:title={}>
        <v-icon small class="mr-4">{{ TEMPLATE_TYPE_ICONS[templateType] }}</v-icon>
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
      >New template
      </v-btn>

      <v-btn icon @click="settingsSheet = true">
        <v-icon>mdi-cog</v-icon>
      </v-btn>
    </v-toolbar>

    <v-tabs show-arrows class="pl-4" v-model="viewTab">
      <v-tab :to="getViewUrl(null)" :disabled="viewItemsLoading">All</v-tab>

      <v-tab
          v-for="(view) in views"
          :key="view.id"
          :to="getViewUrl(view.id)"
          :disabled="viewItemsLoading"
      >{{ view.title }}
      </v-tab>

      <v-btn icon class="mt-2 ml-4" @click="editViewsDialog = true">
        <v-icon>mdi-pencil</v-icon>
      </v-btn>
    </v-tabs>

    <v-data-table
        hide-default-footer
        class="mt-4 templates-table"
        single-expand
        show-expand
        :headers="filteredHeaders"
        :items="items"
        :items-per-page="Number.MAX_VALUE"
        :expanded.sync="openedItems"
        :style="{
          opacity: viewItemsLoading ? 0.3 : 1,
        }"
    >
      <template v-slot:item.alias="{ item }">
        <v-icon class="mr-3" small>
          {{ TEMPLATE_TYPE_ICONS[item.type] }}
        </v-icon>
        <router-link
            :to="`/project/${projectId}/templates/${item.id}`">{{ item.alias }}
        </router-link>
      </template>

      <template v-slot:item.version="{ item }">
        <TaskLink
            v-if="item.last_task && item.last_task.tpl_type !== ''"
            :disabled="true"
            :status="item.last_task.status"

            :task-id="item.last_task.tpl_type === 'build'
              ? item.last_task.id
              : item.last_task.build_task.id"

            :label="item.last_task.tpl_type === 'build'
              ? item.last_task.version
              : item.last_task.build_task.version"

            :tooltip="item.last_task.tpl_type === 'build'
              ? item.last_task.message
              : item.last_task.build_task.message"
        />
        <div v-else>&mdash;</div>
      </template>

      <template v-slot:item.status="{ item }">
        <div class="mt-2 mb-2 d-flex" v-if="item.last_task != null">
          <TaskStatus :status="item.last_task.status"/>
        </div>
        <div v-else class="mt-3 mb-2 d-flex" style="color: gray;">Not launched</div>
      </template>

      <template v-slot:item.last_task="{ item }">
        <div class="mt-2 mb-2" v-if="item.last_task != null" style="line-height: 1">
          <TaskLink
              :task-id="item.last_task.id"
              :label="'#' + item.last_task.id"
              :tooltip="item.last_task.message"
          />
          <div style="color: gray; font-size: 14px;">
            by {{ item.last_task.user_name }} {{ item.last_task.created|formatDate }}
          </div>
        </div>
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
          <v-icon class="pr-1">mdi-replay</v-icon>
          {{ TEMPLATE_TYPE_ACTION_TITLES[item.type] }}
        </v-btn>
      </template>

      <template v-slot:expanded-item="{ headers, item }">
        <td
            :colspan="headers.length"
            v-if="openedItems.some((template) => template.id === item.id)"
        >
          <TaskList
              style="border: 1px solid lightgray; border-radius: 6px; margin: 10px 0;"
              :template="item"
              :limit="5"
              :hide-footer="true"
          />
        </td>
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
<style lang="scss">
@import '~vuetify/src/styles/settings/_variables';

.templates-table .text-start:first-child {
  padding-right: 0 !important;
}

@media #{map-get($display-breakpoints, 'sm-and-down')} {
  .templates-table .v-data-table__mobile-row:first-child {
    display: none !important;
  }
}
</style>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import TemplateForm from '@/components/TemplateForm.vue';
import TaskLink from '@/components/TaskLink.vue';
import axios from 'axios';
import TaskForm from '@/components/TaskForm.vue';
import EditViewsForm from '@/components/EditViewsForm.vue';
import TableSettingsSheet from '@/components/TableSettingsSheet.vue';
import TaskList from '@/components/TaskList.vue';
import EventBus from '@/event-bus';
import TaskStatus from '@/components/TaskStatus.vue';
import socket from '@/socket';

import { TEMPLATE_TYPE_ACTION_TITLES, TEMPLATE_TYPE_ICONS } from '../../lib/constants';

export default {
  components: {
    TemplateForm, TaskForm, TableSettingsSheet, TaskStatus, TaskLink, TaskList, EditViewsForm,
  },
  mixins: [ItemListPageBase],
  async created() {
    socket.addListener((data) => this.onWebsocketDataReceived(data));

    await this.loadData();
  },
  data() {
    return {
      TEMPLATE_TYPE_ICONS,
      TEMPLATE_TYPE_ACTION_TITLES,
      inventory: null,
      environment: null,
      repositories: null,
      newTaskDialog: null,
      settingsSheet: null,
      filteredHeaders: [],
      openedItems: [],
      views: null,
      editViewsDialog: null,
      viewItemsLoading: null,
      viewTab: null,
    };
  },
  computed: {
    viewId() {
      if (/^-?\d+$/.test(this.$route.params.viewId)) {
        return parseInt(this.$route.params.viewId, 10);
      }
      return this.$route.params.viewId;
    },

    templateType() {
      if (this.itemId == null || this.itemId === 'new') {
        return '';
      }
      return this.items.find((x) => x.id === this.itemId).type;
    },

    templateAlias() {
      if (this.itemId == null || this.itemId === 'new') {
        return '';
      }
      return this.items.find((x) => x.id === this.itemId).alias;
    },

    isLoaded() {
      return this.items
          && this.inventory
          && this.environment
          && this.repositories
          && this.views;
    },
  },
  watch: {
    async viewId() {
      try {
        this.viewItemsLoading = true;
        await this.loadItems();
        if (this.viewId) {
          localStorage.setItem(`project${this.projectId}__lastVisitedViewId`, this.viewId);
        } else {
          localStorage.removeItem(`project${this.projectId}__lastVisitedViewId`);
        }
      } finally {
        this.viewItemsLoading = false;
      }
    },
  },
  methods: {
    async beforeLoadItems() {
      await this.loadViews();
    },

    getViewUrl(viewId) {
      if (viewId == null) {
        return `/project/${this.projectId}/templates`;
      }
      return `/project/${this.projectId}/views/${viewId}/templates`;
    },

    async loadViews() {
      this.views = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/views`,
        responseType: 'json',
      })).data;
      this.views.sort((v1, v2) => v1.position - v2.position);

      if (this.viewId != null && !this.views.some((v) => v.id === this.viewId)) {
        await this.$router.push({ path: `/project/${this.projectId}/templates` });
      }
    },

    async closeEditViewDialog() {
      this.editViewsDialog = false;
      await this.loadViews();
    },

    async onWebsocketDataReceived(data) {
      if (data.project_id !== this.projectId || data.type !== 'update') {
        return;
      }

      const template = this.items.find((item) => item.id === data.template_id);

      if (template == null) {
        return;
      }

      if (data.task_id !== template.last_task_id) {
        Object.assign(template.last_task, (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/tasks/${data.task_id}`,
          responseType: 'json',
        })).data);
        template.last_task_id = data.task_id;
      }

      Object.assign(template.last_task, {
        ...data,
        type: undefined,
      });
    },

    onTaskCreated(e) {
      EventBus.$emit('i-show-task', {
        taskId: e.item.id,
      });
      this.itemId = null;
    },

    showTaskLog(taskId) {
      EventBus.$emit('i-show-task', {
        taskId,
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
          text: 'Version',
          value: 'version',
          sortable: false,
        },
        {
          text: 'Status',
          value: 'status',
          sortable: false,
        },
        {
          text: 'Last task',
          value: 'last_task',
          sortable: false,
        },
        {
          text: 'Playbook',
          value: 'playbook',
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
      return this.viewId == null
        ? `/api/project/${this.projectId}/templates`
        : `/api/project/${this.projectId}/views/${this.viewId}/templates`;
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
