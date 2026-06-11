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

      :options.sync="options"
      :server-items-length="totalItems"
      :loading="loading"
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
import axios from 'axios';
import ItemListPageBase from '@/components/ItemListPageBase';
import EventBus from '@/event-bus';
import TaskStatus from '@/components/TaskStatus.vue';
import TaskLink from '@/components/TaskLink.vue';
import socket from '@/socket';
import { TEMPLATE_TYPE_ICONS } from '@/lib/constants';
import AppsMixin from '@/components/AppsMixin';
import DashboardMenu from '@/components/DashboardMenu.vue';

const PER_PAGE = 20;

export default {
  mixins: [ItemListPageBase, AppsMixin],

  data() {
    return {
      TEMPLATE_TYPE_ICONS,
      totalItems: 0,
      loading: false,
      options: {},
    };
  },

  components: { DashboardMenu, TaskStatus, TaskLink },

  watch: {
    async projectId() {
      this.options = { ...this.options, page: 1 };
      await this.loadItems(true);
    },

    options: {
      deep: true,
      handler() {
        this.loadItems();
      },
    },
  },

  created() {
    this.itemsLoading = null;
    this.itemsReloadRequested = false;
    this.socketListenerId = socket.addListener((data) => this.onWebsocketDataReceived(data));
  },

  beforeDestroy() {
    socket.removeListener(this.socketListenerId);
  },

  methods: {
    showTaskLog(taskId) {
      EventBus.$emit('i-show-task', {
        taskId,
      });
    },

    async reloadItems() {
      // Coalesce concurrent reloads: share the in-flight request and
      // do at most one trailing reload for events received meanwhile.
      if (this.itemsLoading) {
        this.itemsReloadRequested = true;
        await this.itemsLoading;
        return;
      }

      this.itemsLoading = (async () => {
        try {
          do {
            this.itemsReloadRequested = false;
            // eslint-disable-next-line no-await-in-loop
            await this.loadItems(true);
          } while (this.itemsReloadRequested);
        } finally {
          this.itemsLoading = null;
        }
      })();

      await this.itemsLoading;
    },

    async onWebsocketDataReceived(data) {
      if (data.project_id !== this.projectId || data.type !== 'update') {
        return;
      }

      if (!this.items.some((item) => item.id === data.task_id)) {
        await this.reloadItems();
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

    // Override ItemListPageBase.loadItems to fetch a single page from the
    // backend and read the total count from the X-Total-Count header.
    async loadItems(force = false) {
      const page = this.options.page || 1;
      const count = this.options.itemsPerPage > 0 ? this.options.itemsPerPage : PER_PAGE;
      const offset = (page - 1) * count;

      const key = `${this.projectId}:${offset}:${count}`;
      if (!force && key === this.loadedKey) {
        return;
      }
      this.loadedKey = key;

      this.loading = true;
      try {
        const response = await axios({
          method: 'get',
          url: `${this.getItemsUrl()}?count=${count}&offset=${offset}`,
          responseType: 'json',
        });

        this.items = response.data;

        const total = parseInt(response.headers['x-total-count'], 10);
        this.totalItems = Number.isNaN(total) ? this.items.length : total;
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
