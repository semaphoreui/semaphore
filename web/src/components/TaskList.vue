<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="tasks != null">

    <NewTaskDialog
      v-model="newTaskDialog"
      :project-id="template.project_id"
      :template="template"
      :source-task="sourceTask"
    />

    <v-data-table
        :headers="headers"
        :items="tasks"
        :loading="loading"
        :items-per-page="-1"
        hide-default-footer
        class="mt-0 TaskListTable"
    >
      <template v-slot:item.id="{ item }">
        <TaskLink
            :task-id="item.id"
            :label="'#' + item.id"
        />
        <div style="font-size: 14px;">
          <span v-if="item.message">
            <v-icon x-small>mdi-message-outline</v-icon> {{ item.message }}
          </span>
          <span v-else-if="item.commit_hash">
            <v-icon x-small>mdi-source-fork</v-icon> {{ item.commit_message }}
          </span>
        </div>
      </template>

      <template v-slot:item.version="{ item }">
        <div v-if="item.tpl_type !== ''">
          <TaskLink
              :disabled="item.tpl_type === 'build'"
              :task-id="item.build_task_id"
              :tooltip="item.tpl_type === 'build' ? item.message : (item.build_task || {}).message"
              :label="item.tpl_type === 'build' ? item.version : (item.build_task || {}).version"
              :status="item.status"
          />
        </div>
        <div v-else>&mdash;</div>
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

      <template v-slot:item.actions="{ item }">
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn @click="createTask(item)">
            <v-icon>mdi-replay</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>

    <div
      v-if="!hideFooter"
      class="d-flex align-center justify-end mt-2 mr-2 mb-4 TaskListPagination"
    >
      <span class="text--secondary mr-2">{{ $t('rowsPerPage') }}</span>

      <v-select
        v-model="perPage"
        :items="perPageOptions"
        :disabled="loading"
        dense
        hide-details
        class="TaskListPerPage mr-4"
        style="max-width: 72px;"
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
.TaskListTable td {
  height: 60px !important;
}
</style>
<script>
import axios from 'axios';
import TaskStatus from '@/components/TaskStatus.vue';
import TaskLink from '@/components/TaskLink.vue';
import socket from '@/socket';
import { TEMPLATE_TYPE_ACTION_TITLES, TEMPLATE_TYPE_ICONS } from '@/lib/constants';
import NewTaskDialog from '@/components/NewTaskDialog.vue';

export default {
  components: {
    NewTaskDialog,
    TaskStatus,
    TaskLink,
  },
  props: {
    template: Object,
    limit: Number,
    hideFooter: Boolean,
    needUpdate: Boolean,
  },
  data() {
    return {
      headers: [
        {
          text: this.$i18n.t('taskId'),
          value: 'id',
          sortable: false,
        },
        {
          value: 'actions',
          sortable: false,
          width: '0%',
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
      ],
      tasks: null,
      taskId: null,
      newTaskDialog: null,
      sourceTask: null,
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
  computed: {
    // Effective page size: embedded lists pin it via the `limit` prop; the
    // full-page task list lets the user pick it with the perPage selector.
    pageSize() {
      return this.limit || this.perPage;
    },

    // Positional range of the current page relative to the start of navigation
    // (page 1 = newest). Keyset pagination has no real offset, so this is the
    // virtual offset `pageIndex * pageSize` .. that plus the rows actually
    // loaded on the current page.
    rangeStart() {
      return this.pageIndex * this.pageSize;
    },

    rangeEnd() {
      return this.rangeStart + (this.tasks ? this.tasks.length : 0);
    },
  },
  watch: {
    async template() {
      await this.resetAndLoad();
    },
    async needUpdate(val) {
      if (val) {
        await this.resetAndLoad();
      }
    },
    async perPage() {
      await this.resetAndLoad();
    },
  },
  created() {
    this.itemsLoading = null;
    this.itemsReloadRequested = false;
    // Throttle state: `reloadTimer` holds a pending trailing reload, and
    // `lastReloadTime` is when the last reload actually fired.
    this.reloadTimer = null;
    this.lastReloadTime = 0;
    this.socketListenerId = socket.addListener((data) => this.onWebsocketDataReceived(data));
  },
  async mounted() {
    await this.loadData();
  },
  beforeDestroy() {
    socket.removeListener(this.socketListenerId);
    if (this.reloadTimer) {
      clearTimeout(this.reloadTimer);
      this.reloadTimer = null;
    }
  },
  methods: {
    getItemsUrl() {
      return `/api/project/${this.template.project_id}/templates/${this.template.id}/tasks/last`;
    },

    // Load the current page using keyset pagination. `hasNext` is read from the
    // X-Has-Next header.
    async loadData() {
      const before = this.cursors[this.pageIndex];

      let url = `${this.getItemsUrl()}?count=${this.pageSize}`;
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

        this.tasks = response.data;
        this.hasNext = response.headers['x-has-next'] === 'true';
      } finally {
        this.loading = false;
      }
    },

    // Reset to the first page and reload. Used when the template or the page
    // size changes, since existing cursors become meaningless.
    async resetAndLoad() {
      this.cursors = [null];
      this.pageIndex = 0;
      await this.loadData();
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
            await this.loadData();
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
      if (data.template_id !== this.template.id || data.type !== 'update') {
        return;
      }

      if (this.tasks != null && !this.tasks.some((task) => task.id === data.task_id)) {
        this.reloadItemsThrottled();
      }

      const task = (this.tasks || []).find((t) => t.id === data.task_id);

      if (task) {
        Object.assign(task, {
          ...data,
          type: undefined,
        });
      }
    },

    async goNext() {
      if (!this.hasNext || this.loading || this.tasks.length === 0) {
        return;
      }

      // Tasks are ordered by id desc, so the last row holds the smallest id on
      // the current page — that becomes the cursor for the next (older) page.
      const before = this.tasks[this.tasks.length - 1].id;

      // Drop any forward history before pushing the new cursor.
      this.cursors = this.cursors.slice(0, this.pageIndex + 1);
      this.cursors.push(before);
      this.pageIndex += 1;

      await this.loadData();
    },

    async goPrev() {
      if (this.pageIndex === 0 || this.loading) {
        return;
      }

      this.pageIndex -= 1;
      await this.loadData();
    },

    getActionButtonTitle() {
      return this.$i18n.t(`Re${TEMPLATE_TYPE_ACTION_TITLES[this.template.type]}`);
    },

    createTask(task) {
      this.sourceTask = task;
      this.newTaskDialog = true;
    },
  },
};
</script>
