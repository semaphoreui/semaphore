<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        {{ $t('dashboard2') }}
      </v-toolbar-title>
    </v-toolbar>

    <DashboardMenu
      :project-id="projectId"
      :project-type="projectType"
      :can-update-project="can(USER_PERMISSIONS.updateProject)"
    />

    <v-data-table
      :headers="headers"
      :items="items"
      :footer-props="{ itemsPerPageOptions: [20] }"
      class="mt-4 HistoryTable"
    >
      <template v-slot:item.tpl_alias="{ item }">
        <div class="d-flex align-center">
          <v-icon
            class="mr-3"
            small
          >
            {{ getAppIcon(item.tpl_app) }}
          </v-icon>

          <!--          <v-icon class="mr-3" small>-->
          <!--            {{ TEMPLATE_TYPE_ICONS[item.tpl_type] }}-->
          <!--          </v-icon>-->

          <TaskLink
            :task-id="item.id"
            :label="'#' + item.id"
          />

          <v-icon small class="ml-1 mr-1">mdi-arrow-left</v-icon>

          <router-link :to="
            '/project/' + item.project_id +
            '/templates/' + item.template_id"
          >{{ item.tpl_alias }}
          </router-link>
        </div>

        <div style="font-size: 14px;" class="ml-7">
            <span v-if="item.message">
              <v-icon x-small>mdi-message-outline</v-icon> {{ item.message }}
            </span>
          <span v-else-if="item.commit_hash">
              <v-icon x-small>mdi-source-fork</v-icon> {{ item.commit_message }}
            </span>
        </div>
      </template>

      <template v-slot:item.version="{ item }">
        <TaskLink
          :disabled="item.tpl_type === 'build'"
          class="ml-2"
          v-if="item.tpl_type !== ''"
          :status="item.status"

          :task-id="item.tpl_type === 'build'
              ? item.id
              : (item.build_task || {}).id"

          :label="item.tpl_type === 'build'
              ? item.version
              : (item.build_task || {}).version"

          :tooltip="item.tpl_type === 'build'
              ? item.message
              : (item.build_task || {}).message"
        />
        <div class="ml-2" v-else>&mdash;</div>
      </template>

      <template v-slot:item.status="{ item }">
        <TaskStatus :status="item.status"/>
      </template>

      <template v-slot:item.start="{ item }">
        {{ item.start | formatDate }}
      </template>

      <template v-slot:item.end="{ item }">
        {{ [item.start, item.end] | formatMilliseconds }}
      </template>
    </v-data-table>
  </div>
</template>

<style lang="scss">
.HistoryTable td {
  height: 60px !important;
}
</style>

<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import EventBus from '@/event-bus';
import TaskStatus from '@/components/TaskStatus.vue';
import TaskLink from '@/components/TaskLink.vue';
import socket from '@/socket';
import { TEMPLATE_TYPE_ICONS } from '@/lib/constants';
import AppsMixin from '@/components/AppsMixin';
import DashboardMenu from '@/components/DashboardMenu.vue';

export default {
  mixins: [ItemListPageBase, AppsMixin],

  data() {
    return { TEMPLATE_TYPE_ICONS };
  },

  components: { DashboardMenu, TaskStatus, TaskLink },

  watch: {
    async projectId() {
      await this.loadItems();
    },
  },

  created() {
    socket.addListener((data) => this.onWebsocketDataReceived(data));
  },

  methods: {
    showTaskLog(taskId) {
      EventBus.$emit('i-show-task', {
        taskId,
      });
    },

    async onWebsocketDataReceived(data) {
      if (data.project_id !== this.projectId || data.type !== 'update') {
        return;
      }

      if (!this.items.some((item) => item.id === data.task_id)) {
        await this.loadItems();
      }

      const task = this.items.find((item) => item.id === data.task_id);

      if (task) {
        Object.assign(task, {
          ...data,
          type: undefined,
        });
      }
    },

    getHeaders() {
      return [
        {
          text: this.$i18n.t('task2'),
          value: 'tpl_alias',
          sortable: false,
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
          text: this.$i18n.t('user'),
          value: 'user_name',
          sortable: false,
        },
        {
          text: this.$i18n.t('start'),
          value: 'start',
          sortable: false,
        },
        {
          text: this.$i18n.t('duration'),
          value: 'end',
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
