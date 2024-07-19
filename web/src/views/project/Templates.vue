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
          {{ $t('editViews') }}
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

    <EditTemplateDialog
        v-model="editDialog"
        :project-id="projectId"
        :item-app="itemApp"
        item-id="new"
        @save="loadItems()"
    ></EditTemplateDialog>

    <NewTaskDialog
      v-model="newTaskDialog"
      @save="itemId = null"
      @close="itemId = null"
      :project-id="projectId"
      :template-id="itemId"
      :template-alias="templateAlias"
      :template-type="templateType"
      :template-app="templateApp"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        {{ $t('taskTemplates2') }}
      </v-toolbar-title>
      <v-spacer></v-spacer>

      <v-menu
        offset-y
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            v-bind="attrs"
            v-on="on"
            color="primary"
            class="mr-1 pr-2"
            v-if="can(USER_PERMISSIONS.manageProjectResources)"
            :disabled="!isAdmin && appsMixin.activeAppIds.length === 0"
          >
            {{ $t('newTemplate') }}
            <v-icon>mdi-chevron-down</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item
            v-for="appID in appsMixin.activeAppIds"
            :key="appID"
            link
            @click="editItem('new'); itemApp = appID;"
          >
            <v-list-item-icon>
              <v-icon
                :color="getAppColor(appID)"
              >
                {{ getAppIcon(appID) }}
              </v-icon>
            </v-list-item-icon>
            <v-list-item-title>{{ getAppTitle(appID) }}</v-list-item-title>
          </v-list-item>

          <v-divider v-if="isAdmin && appsMixin.activeAppIds.length > 0"/>

          <v-list-item
              v-if="isAdmin"
              key="other"
              link
              href="/apps"
          >
            <v-list-item-icon>
              <v-icon>mdi-cogs</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Applications</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>

      <v-btn icon @click="settingsSheet = true">
        <v-icon>mdi-cog</v-icon>
      </v-btn>
    </v-toolbar>

    <v-tabs show-arrows class="pl-4" v-model="viewTab">
      <v-tab :to="getViewUrl(null)" :disabled="viewItemsLoading">{{ $t('all') }}</v-tab>

      <v-tab
        v-for="(view) in views"
        :key="view.id"
        :to="getViewUrl(view.id)"
        :disabled="viewItemsLoading"
      >{{ view.title }}
      </v-tab>

      <v-btn
        icon
        class="mt-2 ml-4"
        @click="editViewsDialog = true"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >
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
      <template v-slot:item.name="{ item }">
        <v-icon
          class="mr-3"
          small
        >
          {{ getAppIcon(item.app) }}
        </v-icon>

        <v-icon class="mr-3" small>
          {{ TEMPLATE_TYPE_ICONS[item.type] }}
        </v-icon>

        <router-link
          :to="viewId
              ? `/project/${projectId}/views/${viewId}/templates/${item.id}`
              : `/project/${projectId}/templates/${item.id}`"
        >{{ item.name }}
        </router-link>
      </template>

      <template v-slot:item.version="{ item }">
        <TaskLink
          v-if="item.last_task && item.last_task.tpl_type !== ''"
          :disabled="true"
          :status="item.last_task.status"

          :task-id="item.last_task.tpl_type === 'build'
              ? item.last_task.id
              : (item.last_task.build_task || {}).id"

          :label="item.last_task.tpl_type === 'build'
              ? item.last_task.version
              : (item.last_task.build_task || {}).version"

          :tooltip="item.last_task.tpl_type === 'build'
              ? item.last_task.message
              : (item.last_task.build_task || {}).message"
        />
        <div v-else>&mdash;</div>
      </template>

      <template v-slot:item.status="{ item }">
        <div class="mt-2 mb-2 d-flex" v-if="item.last_task != null">
          <TaskStatus :status="item.last_task.status"/>
        </div>
        <div v-else class="mt-3 mb-2 d-flex" style="color: gray;">{{ $t('notLaunched') }}</div>
      </template>

      <template v-slot:item.last_task="{ item }">
        <div class="mt-2 mb-2" v-if="item.last_task != null" style="line-height: 1">
          <TaskLink
            :task-id="item.last_task.id"
            :label="'#' + item.last_task.id"
            :tooltip="item.last_task.message"
          />
          <div style="color: gray; font-size: 14px;">
            {{ $t('by', {user_name: item.last_task.user_name }) }}
          </div>
        </div>
      </template>

      <template v-slot:item.inventory_id="{ item }">
        {{ (inventory.find((x) => x.id === item.inventory_id) || {name: '—'}).name }}
      </template>

      <template v-slot:item.environment_id="{ item }">
        {{ (environment.find((x) => x.id === item.environment_id) || {name: '—'}).name }}
      </template>

      <template v-slot:item.repository_id="{ item }">
        {{ repositories.find((x) => x.id === item.repository_id).name }}
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn text class="pl-1 pr-2" @click="createTask(item.id)">
          <v-icon class="pr-1">mdi-play</v-icon>
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
@import '~vuetify/src/styles/settings/variables';

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
import TaskLink from '@/components/TaskLink.vue';
import axios from 'axios';
import EditViewsForm from '@/components/EditViewsForm.vue';
import TableSettingsSheet from '@/components/TableSettingsSheet.vue';
import TaskList from '@/components/TaskList.vue';
import EventBus from '@/event-bus';
import TaskStatus from '@/components/TaskStatus.vue';
import socket from '@/socket';
import NewTaskDialog from '@/components/NewTaskDialog.vue';

import {
  TEMPLATE_TYPE_ACTION_TITLES,
  TEMPLATE_TYPE_ICONS,
} from '@/lib/constants';
import EditTemplateDialog from '@/components/EditTemplateDialog.vue';
import AppsMixin from '@/components/AppsMixin';

export default {
  components: {
    EditTemplateDialog,
    TableSettingsSheet,
    TaskStatus,
    TaskLink,
    TaskList,
    EditViewsForm,
    NewTaskDialog,
  },
  mixins: [ItemListPageBase, AppsMixin],
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
      apps: null,
      itemApp: '',
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
      return this.items.find((x) => x.id === this.itemId).name;
    },

    templateApp() {
      if (this.itemId == null || this.itemId === 'new') {
        return '';
      }
      return this.items.find((x) => x.id === this.itemId).app;
    },

    isLoaded() {
      return this.items
        && this.inventory
        && this.environment
        && this.repositories
        && this.views
        && this.isAppsLoaded;
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

    allowActions() {
      return true;
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

      const template = (this.items || []).find((item) => item.id === data.template_id);

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
          text: this.$i18n.t('name'),
          value: 'name',
        },
        {
          text: this.$i18n.t('version'),
          value: 'version',
          sortable: false,
        },
        {
          text: this.$i18n.t('status'),
          value: 'status',
          sortable: false,
        },
        {
          text: this.$i18n.t('lastTask'),
          value: 'last_task',
          sortable: false,
        },
        {
          text: this.$i18n.t('playbook'),
          value: 'playbook',
          sortable: false,
        },
        {
          text: this.$i18n.t('inventory'),
          value: 'inventory_id',
          sortable: false,
        },
        {
          text: this.$i18n.t('environment'),
          value: 'environment_id',
          sortable: false,
        },
        {
          text: this.$i18n.t('repository2'),
          value: 'repository_id',
          sortable: false,
        },
        {
          text: this.$i18n.t('actions'),
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
