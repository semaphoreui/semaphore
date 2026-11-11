<template>
  <div data-testid="hostTasks" :data-state="state">
    <v-data-table
      :headers="headers"
      :items="tasks || []"
      :items-per-page="-1"
      :loading="loading"
      hide-default-footer
      data-testid="hostTasks-table"
    >
      <template v-slot:item.task_id="{ item }">
        <router-link
          v-if="!projectId"
          :to="`/project/${item.project_id}?t=${item.task_id}`"
          data-testid="hostTasks-rowTask"
        >#{{ item.task_id }}</router-link>
        <TaskLink
          v-else
          :task-id="item.task_id"
          :label="'#' + item.task_id"
          data-testid="hostTasks-rowTask"
        />
      </template>

      <template v-slot:item.project="{ item }">
        <span data-testid="hostTasks-rowProject">
          {{ item.project_name }}
        </span>
      </template>

      <template v-slot:item.template_name="{ item }">
        <router-link
          :to="`/project/${item.project_id}/templates/${item.template_id}`"
          data-testid="hostTasks-rowTemplate"
        >{{ item.template_name }}</router-link>
      </template>

      <template v-slot:item.status="{ item }">
        <TaskStatus :status="item.status"/>
      </template>

      <template v-slot:item.result="{ item }">
        <HostStatus
          v-if="item.unreachable > 0"
          status="unreachable"
          data-testid="hostTasks-rowResult"
        />
        <span v-else data-testid="hostTasks-rowResult">{{ resultText(item) }}</span>
      </template>

      <template v-slot:item.start="{ item }">
        {{ $filters.formatDate(item.start) }}
      </template>

      <template v-slot:item.user_name="{ item }">
        {{ item.user_name || '—' }}
      </template>
    </v-data-table>

    <div class="d-flex align-center justify-end mt-2 mr-2 mb-4" data-testid="hostTasks-pagination">
      <v-btn
        :disabled="loading || pageIndex === 0"
        icon
        variant="text"
        data-testid="hostTasks-prevPage"
        @click="goPrev()"
      >
        <v-icon>mdi-chevron-left</v-icon>
      </v-btn>

      <span class="mx-2 text-medium-emphasis" data-testid="hostTasks-range">
        {{ rangeStart }} - {{ rangeEnd }}
      </span>

      <v-btn
        :disabled="loading || !hasNext"
        icon
        variant="text"
        data-testid="hostTasks-nextPage"
        @click="goNext()"
      >
        <v-icon>mdi-chevron-right</v-icon>
      </v-btn>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import TaskLink from '@/components/TaskLink.vue';
import TaskStatus from '@/components/TaskStatus.vue';
import HostStatus from '@/components/HostStatus.vue';

const COUNT = 20;
const COUNTERS = ['ok', 'changed', 'failed', 'skipped', 'rescued', 'ignored'];

export default {
  components: { TaskLink, TaskStatus, HostStatus },

  props: {
    projectId: Number,
    hostId: Number,
  },

  data() {
    return {
      tasks: null,
      loading: false,
      cursors: [null],
      pageIndex: 0,
      hasNext: false,
      loadSeq: 0,
    };
  },

  computed: {
    baseUrl() {
      return this.projectId
        ? `/api/project/${this.projectId}/hosts/${this.hostId}/tasks`
        : `/api/hosts/${this.hostId}/tasks`;
    },

    headers() {
      return [
        { title: this.$t('taskId'), key: 'task_id', sortable: false },
        ...(this.projectId ? [] : [{ title: this.$t('project'), key: 'project', sortable: false }]),
        { title: this.$t('template'), key: 'template_name', sortable: false },
        { title: this.$t('status'), key: 'status', sortable: false },
        { title: this.$t('host_col_result'), key: 'result', sortable: false },
        { title: this.$t('started'), key: 'start', sortable: false },
        { title: this.$t('user'), key: 'user_name', sortable: false },
      ];
    },

    state() {
      if (this.loading || this.tasks == null) return 'loading';
      return this.tasks.length === 0 ? 'empty' : 'ready';
    },

    rangeStart() {
      return this.pageIndex * COUNT;
    },

    rangeEnd() {
      return this.rangeStart + (this.tasks ? this.tasks.length : 0);
    },
  },

  watch: {
    async hostId() {
      await this.resetAndLoad();
    },
    async projectId() {
      await this.resetAndLoad();
    },
  },

  async created() {
    await this.loadData();
  },

  methods: {
    resultText(item) {
      return COUNTERS
        .filter((k) => item[k] > 0)
        .map((k) => `${k} ${item[k]}`)
        .join(' · ');
    },

    async loadData() {
      this.loadSeq += 1;
      const seq = this.loadSeq;
      const before = this.cursors[this.pageIndex];
      this.loading = true;
      try {
        const response = await axios.get(this.baseUrl, {
          params: { count: COUNT, before: before == null ? undefined : before },
        });
        if (seq !== this.loadSeq) return;
        this.tasks = response.data;
        this.hasNext = response.headers['x-has-next'] === 'true';
      } catch (err) {
        if (seq !== this.loadSeq) return;
        this.tasks = [];
        this.hasNext = false;
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      } finally {
        if (seq === this.loadSeq) this.loading = false;
      }
    },

    async resetAndLoad() {
      this.cursors = [null];
      this.pageIndex = 0;
      await this.loadData();
    },

    async goNext() {
      if (!this.hasNext || this.loading || this.tasks.length === 0) return;
      const before = this.tasks[this.tasks.length - 1].task_id;
      this.cursors = this.cursors.slice(0, this.pageIndex + 1);
      this.cursors.push(before);
      this.pageIndex += 1;
      await this.loadData();
    },

    async goPrev() {
      if (this.pageIndex === 0 || this.loading) return;
      this.pageIndex -= 1;
      await this.loadData();
    },
  },
};
</script>
