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
      :loading="loading"
      :items-per-page="-1"
      hide-default-footer
      class="mt-4 HistoryTable"
    >
      <template v-slot:item.tpl_alias="{ item }">
        <div class="d-flex align-center">
          <v-icon class="mr-3" small>
            {{ getAppIcon(item.tpl_app) }}
          </v-icon>

          <!--          <v-icon class="mr-3" small>-->
          <!--            {{ TEMPLATE_TYPE_ICONS[item.tpl_type] }}-->
          <!--          </v-icon>-->

          <TaskLink :task-id="item.id" :label="'#' + item.id" />

          <v-icon small class="ml-1 mr-1">mdi-arrow-left</v-icon>

          <router-link :to="'/project/' + item.project_id + '/templates/' + item.template_id"
            >{{ item.tpl_alias }}
          </router-link>
        </div>

        <div style="font-size: 14px" class="ml-7">
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
          :task-id="item.tpl_type === 'build' ? item.id : (item.build_task || {}).id"
          :label="item.tpl_type === 'build' ? item.version : (item.build_task || {}).version"
          :tooltip="item.tpl_type === 'build' ? item.message : (item.build_task || {}).message"
        />
        <div class="ml-2" v-else>&mdash;</div>
      </template>

      <template v-slot:item.status="{ item }">
        <TaskStatus :status="item.status" />
      </template>

      <template v-slot:item.start="{ item }">
        {{ item.start | formatDate }}
      </template>

      <template v-slot:item.end="{ item }">
        {{ [item.start, item.end] | formatMilliseconds }}
      </template>
    </v-data-table>

    <div class="d-flex align-center justify-end mt-2 mr-2 mb-4 HistoryPagination">
      <span class="text--secondary mr-2">{{ $t('rowsPerPage') }}</span>

      <v-select
        v-model="perPage"
        :items="perPageOptions"
        :disabled="loading"
        dense
        hide-details
        class="HistoryPerPage mr-4"
        style="max-width: 72px"
      ></v-select>

      <v-btn icon :disabled="loading || pageIndex === 0" @click="goPrev()">
        <v-icon>mdi-chevron-left</v-icon>
      </v-btn>

      <span class="mx-2 text--secondary">{{ rangeStart }} - {{ rangeEnd }}</span>

      <v-btn icon :disabled="loading || !hasNext" @click="goNext()">
        <v-icon>mdi-chevron-right</v-icon>
      </v-btn>
    </div>
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

export default {
  mixins: [ItemListPageBase, AppsMixin],

  data() {
    return {
      TEMPLATE_TYPE_ICONS,
      loading: false,
      // Keyset (cursor) pagination state. `cursors[i]` is the `before` task id
      // used to load page i (null for the first page). `hasNext` comes from the
      // backend X-Has-Next header so we never need an expensive total count.
      cursors: [null],
      pageIndex: 0,
      hasNext: false,
      perPage: 20,
      perPageOptions: [10, 20, 50, 100],
    };
  },

  components: { DashboardMenu, TaskStatus, TaskLink },

  computed: {
    // Positional range of the current page relative to the start of navigation
    // (page 1 = newest). Keyset pagination has no real offset, so this is the
    // virtual offset `pageIndex * perPage` .. that plus the rows actually
    // loaded on the current page.
    rangeStart() {
      return this.pageIndex * this.perPage;
    },

    rangeEnd() {
      return this.rangeStart + (this.items ? this.items.length : 0);
    },
  },

  watch: {
    async projectId() {
      await this.resetAndLoad();
    },

    async perPage() {
      await this.resetAndLoad();
    },
  },

  created() {
    this.itemsLoading = null;
    this.itemsReloadRequested = false;
    this.socketListenerId = socket.addListener((data) => this.onWebsocketDataReceived(data));
  },

  beforeDestroy() {
    socket.removeListener(this.socketListenerId);
    if (this.reloadTimer) {
      clearTimeout(this.reloadTimer);
      this.reloadTimer = null;
    }
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
            await this.loadItems();
          } while (this.itemsReloadRequested);
        } finally {
          this.itemsLoading = null;
        }
      })();

      await this.itemsLoading;
    },

    // Throttle reloads: a burst of websocket updates triggers at most one
    // reload per 5 seconds (with a trailing reload for events received while
    // the interval is cooling down).
    reloadItemsThrottled() {
      const RELOAD_INTERVAL = 5000;

      if (this.reloadTimer) {
        // A trailing reload is already scheduled — it will pick up this event.
        return;
      }

      const elapsed = Date.now() - this.lastReloadTime;
      const delay = Math.max(0, RELOAD_INTERVAL - elapsed);

      this.reloadTimer = setTimeout(async () => {
        this.reloadTimer = null;
        this.lastReloadTime = Date.now();
        await this.reloadItems();
      }, delay);
    },

    async onWebsocketDataReceived(data) {
      if (data.project_id !== this.projectId || data.type !== 'update') {
        return;
      }

      if (!this.items.some((item) => item.id === data.task_id)) {
        this.reloadItemsThrottled();
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

    // Reset to the first page and reload. Used when the project or the page
    // size changes, since existing cursors become meaningless.
    async resetAndLoad() {
      this.cursors = [null];
      this.pageIndex = 0;
      await this.loadItems();
    },

    // Override ItemListPageBase.loadItems to fetch the current page using keyset
    // pagination. `hasNext` is read from the X-Has-Next header.
    async loadItems() {
      const before = this.cursors[this.pageIndex];

      let url = `${this.getItemsUrl()}?count=${this.perPage}`;
      if (before != null) {
        url += `&before=${before}`;
      }

      this.loading = true;
      try {
        const response = await axios({
          method: 'get',
          url,
          responseType: 'json',
        });

        this.items = response.data;
        this.hasNext = response.headers['x-has-next'] === 'true';
      } finally {
        this.loading = false;
      }
    },

    async goNext() {
      if (!this.hasNext || this.loading || this.items.length === 0) {
        return;
      }

      // Tasks are ordered by id desc, so the last row holds the smallest id on
      // the current page — that becomes the cursor for the next (older) page.
      const before = this.items[this.items.length - 1].id;

      // Drop any forward history before pushing the new cursor.
      this.cursors = this.cursors.slice(0, this.pageIndex + 1);
      this.cursors.push(before);
      this.pageIndex += 1;

      await this.loadItems();
    },

    async goPrev() {
      if (this.pageIndex === 0 || this.loading) {
        return;
      }

      this.pageIndex -= 1;
      await this.loadItems();
    },
  },
};
</script>
