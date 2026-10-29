<template>
  <div
    class="task-log-view"
    :class="{'task-log-view--with-message': item.message || item.commit_message}"
    data-testid="taskLog-page"
  >
    <div class="px-5 task-log-view__message" data-testid="taskLog-messageBar">
      <span
        v-if="item.message"
        class="mr-3"
        data-testid="taskLog-message"
      >
        <v-icon size="small">mdi-message-outline</v-icon>
        {{ item.message }}
      </span>

      <span
        class="d-inline-block"
        v-if="item.commit_message"
        data-testid="taskLog-commitMessage"
      >
        <v-icon size="small">mdi-source-fork</v-icon>
        {{ item.commit_message }}
      </span>
    </div>

    <div
      class="overflow-auto text-no-wrap px-5 task-log-view__status"
      data-testid="taskLog-statusBar"
    >
      <TaskStatus :status="item.status" data-testid="task-status" />

      <span class="ml-3 hidden-xs task-log-view__status_part" data-testid="taskLog-started">

        Started <span v-if="user" data-testid="taskLog-author">by <b>{{ user.name }}</b></span>

        at <b>{{ $filters.formatDate(item.start) }}</b>
      </span>

      <span
        class="ml-3 hidden-sm-and-down task-log-view__status_part"
        data-testid="taskLog-duration"
      >
        <v-icon
          size="small" style="transform: translateY(-1px)">mdi-clock-outline</v-icon>
        {{ $filters.formatMilliseconds([item.start, item.end]) }}
      </span>
    </div>

    <v-tabs class="task-log-view__tabs" align-tabs="end" v-model="tab" data-testid="taskLog-tabs">
      <v-tab :value="0" data-testid="taskLog-tab-log">Log</v-tab>
      <v-tab :value="1" data-testid="taskLog-tab-details">Details</v-tab>
      <v-tab
        v-if="isPro"
        :value="2"
        :disabled="!isTaskStopped"
        data-testid="taskLog-tab-summary"
      >
        Summary
      </v-tab>
    </v-tabs>

    <div v-if="tab === 0" data-testid="taskLog-logPanel">
      <div
        class="task-log-records"
        data-testid="taskLog-output"
        ref="records"
      >
        <div
          role="group"
          :style="{ position: 'relative', height: `${recordsVirtualizer.getTotalSize()}px` }"
        >
          <div
            v-for="row in recordsVirtualizer.getVirtualItems()"
            :key="output[row.index].id"
            :ref="measureRecord"
            :data-index="row.index"
            role="listitem"
            :style="{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              transform: `translateY(${row.start}px)`,
            }"
          >
            <TaskLogViewRecord :index="row.index" :source="output[row.index]" />
          </div>
        </div>
      </div>

      <v-btn
        color="success"
        class="task-log-action-button"
        style="right: 260px; width: 70px;"
        v-if="item.status === 'waiting_confirmation'"
        @click="confirmTask()"
        data-testid="taskLog-confirm"
      >
        <v-icon>mdi-check</v-icon>
      </v-btn>

      <v-btn
        color="warning"
        class="task-log-action-button"
        style="right: 180px; width: 70px;"
        v-if="item.status === 'waiting_confirmation'"
        @click="rejectTask()"
        data-testid="taskLog-reject"
      >
        <v-icon>mdi-close</v-icon>
      </v-btn>

      <v-btn
        color="error"
        class="task-log-action-button"
        style="right: 20px; width: 150px;"
        v-if="canStop"
        @click="stopTask(item.status === 'stopping')"
        :data-testid="item.status === 'stopping' ? 'taskLog-forceStop' : 'taskLog-stop'"
      >
        {{ item.status === 'stopping' ? $t('forceStop') : $t('stop') }}
      </v-btn>

      <v-btn
        v-if="isTaskStopped"
        color="blue-grey"
        :href="rawLogURL"
        class="task-log-action-button"
        style="right: 20px; width: 150px;"
        target="_blank"
        data-testid="task-rawLog"
      >{{ $t('raw_log') }}
      </v-btn>
    </div>

    <div v-else-if="tab === 1" data-testid="taskLog-detailsPanel">
      <v-divider style="margin-top: -1px;" />

      <v-container fluid class="py-0 px-5 overflow-auto pt-4">
        <TaskDetails
          :item="item"
          :user="user"
          :schedule="schedule"
          :integration="integration"
          :project-id="projectId"
        />
      </v-container>
    </div>

    <div v-else-if="tab === 2" data-testid="taskLog-summaryPanel">
      <v-divider style="margin-top: -1px;" />

      <AnsibleStageView
        :features="systemInfo.features"
        :project-id="projectId"
        :task-id="itemId"
      />
    </div>

  </div>
</template>

<style lang="scss">

$card-title-height: 68px;

$task-log-message-offset: -18px;
$task-log-message-height: 40px;
$task-log-message-height-total: $task-log-message-height + $task-log-message-offset;

$task-log-status-height: 32px;
$task-log-status-offset: -40px;
$task-log-tabs-height: 48px;

$task-log-status-tab-height:
  $task-log-tabs-height +
  $task-log-status-offset +
  $task-log-status-height;

.task-log-view__message {
  display: none;
  margin-top: $task-log-message-offset;
  height: $task-log-message-height;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.task-log-view__status {
  height: $task-log-status-height;
  margin-bottom: $task-log-status-offset;
}

.task-log-view__status_part {
  padding: 6px 10px;
  border-radius: 6px;
  background-color: var(--highlighted-card-bg-color);
}

.task-log-view__tabs {
  height: $task-log-tabs-height;
}

.task-log-action-button {
  position: absolute;
  bottom: 10px;
}

.task-log-records {
  background: black;
  color: white;
  height: calc(90dvh - #{$card-title-height + $task-log-status-tab-height});
  overflow: auto;
  font-family: monospace;
  margin: 0;
  padding: 5px 10px 50px;
}

.task-log-view--with-message .task-log-view__message {
  display: block;
}

.task-log-view--with-message .task-log-records {
  height: calc(90dvh -
    #{$card-title-height + $task-log-message-height-total + $task-log-status-tab-height});
}

.v-dialog--fullscreen {

  .task-log-records {
    height: calc(100dvh - #{$card-title-height + $task-log-status-tab-height});
  }

  .task-log-view--with-message .task-log-records {
    height: calc(100dvh -
      #{$card-title-height + $task-log-message-height-total + $task-log-status-tab-height});
  }
}

.task-log-records__record {
  display: flex;
  flex-direction: row;
  justify-content: left;
}

.task-log-records__time {
  width: 120px;
  min-width: 120px;
  user-select: none;
  -webkit-user-select: none;
}

.task-log-records__output {
  width: 100%;
  white-space: pre-wrap;
}

</style>
<script>
import axios from 'axios';
import TaskStatus from '@/components/TaskStatus.vue';
import socket from '@/socket';
import { computed, ref } from 'vue';
import { useVirtualizer } from '@tanstack/vue-virtual';
import TaskLogViewRecord from '@/components/TaskLogViewRecord.vue';
import ProjectMixin from '@/components/ProjectMixin';
import AnsibleStageView from '@/components/AnsibleStageView.vue';
import TaskDetails from '@/components/TaskDetails.vue';

const LOG_KEEPS = 100;
const LOG_BUFFER = Math.round(LOG_KEEPS / 3);

export default {
  components: {
    TaskDetails, AnsibleStageView, TaskStatus, TaskLogViewRecord,
  },

  mixins: [ProjectMixin],

  props: {
    item: Object,
    projectId: Number,
    systemInfo: Object,
    features: null,
  },

  // The log is a virtual list: a window of LOG_KEEPS records around the view is rendered,
  // a third of it above the first visible record (as vue-virtual-scroll-list with keeps="100").
  setup() {
    const output = ref([]);
    const records = ref(null);
    const recordsVirtualizer = useVirtualizer(computed(() => ({
      count: output.value.length,
      getScrollElement: () => records.value,
      estimateSize: () => 22,
      rangeExtractor: (range) => {
        const last = range.count - 1;
        let start = Math.max(range.startIndex - LOG_BUFFER, 0);
        const end = Math.min(Math.max(start + LOG_KEEPS - 1, range.endIndex), last);
        start = Math.max(Math.min(start, end - LOG_KEEPS + 1), 0);
        return Array.from({ length: end - start + 1 }, (_, i) => start + i);
      },
    })));
    const measureRecord = (el) => {
      if (el) {
        recordsVirtualizer.value.measureElement(el);
      }
    };
    return {
      output, records, recordsVirtualizer, measureRecord,
    };
  },

  data() {
    return {
      tab: 0,
      outputBuffer: [],
      user: {},
      // The schedule or integration that started this task, resolved by id so
      // the details panel can name and link it. Null when the task was started
      // by a person, or when the origin has since been deleted.
      schedule: null,
      integration: null,
      autoScroll: true,
      // stages: null,
    };
  },

  watch: {
    async itemId() {
      this.reset();
      await this.loadData();
    },

    async projectId() {
      this.reset();
      await this.loadData();
    },

    // async tab() {
    //   if (this.tab === 1) {
    //     this.stages = await this.loadProjectEndpoint(`/tasks/${this.itemId}/stages`);
    //   }
    // },
  },

  computed: {
    itemId() {
      return this.item?.id;
    },

    isTaskStopped() {
      return [
        'stopped',
        'error',
        'success',
        'canceled',
        'rejected',
      ].includes(this.item.status);
    },

    rawLogURL() {
      return `${this.systemInfo?.web_host || ''}/api/project/${this.projectId}/tasks/${this.itemId}/raw_output`;
    },

    canStop() {
      return [
        'running',
        'stopping',
        'waiting',
        'starting',
        'waiting_confirmation',
        'confirmed',
        'rejected',
      ].includes(this.item.status);
    },

    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },
  },

  async created() {
    this.outputInterval = setInterval(() => {
      this.$nextTick(() => {
        const len = this.outputBuffer.length;
        if (len === 0) {
          return;
        }

        const scrollContainer = this.$refs.records;
        if (!scrollContainer) {
          return;
        }

        // Check if the current position is already at the bottom
        const currentScrollTop = scrollContainer.scrollTop;
        const maxScrollTop = scrollContainer.scrollHeight - scrollContainer.clientHeight;

        // Add a new item to the list
        this.output.push(...this.outputBuffer.splice(0, len));

        // If the user is already at the bottom, keep it scrolled to the bottom
        // Otherwise, maintain the current scroll position
        this.$nextTick(() => {
          if (Math.abs(currentScrollTop - maxScrollTop) <= 1) {
            // User is at the bottom, scroll to the bottom
            scrollContainer.scrollTop = scrollContainer.scrollHeight;
          } else {
            // User is not at the bottom, preserve current scroll position
            scrollContainer.scrollTop = currentScrollTop;
          }
        });
      });
    }, 1000);
    this.socketListenerId = socket.addListener((data) => this.onWebsocketDataReceived(data));
    await this.loadData();
  },

  beforeUnmount() {
    clearInterval(this.outputInterval);
    socket.removeListener(this.socketListenerId);
  },

  methods: {
    async confirmTask() {
      await axios({
        method: 'post',
        url: `/api/project/${this.projectId}/tasks/${this.itemId}/confirm`,
        responseType: 'json',
        data: {},
      });
    },

    async rejectTask() {
      await axios({
        method: 'post',
        url: `/api/project/${this.projectId}/tasks/${this.itemId}/reject`,
        responseType: 'json',
        data: {},
      });
    },

    async stopTask(force) {
      await axios({
        method: 'post',
        url: `/api/project/${this.projectId}/tasks/${this.itemId}/stop`,
        responseType: 'json',
        data: {
          force,
        },
      });
    },

    reset() {
      this.output = [];
      this.outputBuffer = [];
      this.outputInterval = null;
      this.user = {};
      this.schedule = null;
      this.integration = null;
    },

    onWebsocketDataReceived(data) {
      if (data.project_id !== this.projectId || data.task_id !== this.itemId) {
        return;
      }

      switch (data.type) {
        case 'update':
          // Updates the task object of the parent in place (reported by eslint-plugin-vue 10).
          // eslint-disable-next-line vue/no-mutating-props
          Object.assign(this.item, {
            ...data,
            type: undefined,
          });
          break;
        case 'log':
          this.outputBuffer.push({
            ...data,
            id: data.time + data.output,
          });
          break;
        default:
          break;
      }
    },

    // Only a 404 proves the origin is gone. A timeout or a 500 means we simply
    // do not know, and must not be reported to the user as "deleted".
    async loadOrigin(url) {
      try {
        return { status: 'ready', data: (await axios({ method: 'get', url, responseType: 'json' })).data };
      } catch (e) {
        return { status: e.response && e.response.status === 404 ? 'missing' : 'error', data: null };
      }
    },

    async loadData() {
      // Distinguishes "still loading" from "not there", so a valid origin is
      // never briefly labelled as deleted.
      if (this.item.schedule_id) {
        this.schedule = { status: 'loading', data: null };
      }
      if (this.item.integration_id) {
        this.integration = { status: 'loading', data: null };
      }

      // These requests are fired again on every task change; without this the
      // slower response of an earlier task can land after a later one.
      const requested = { project: this.projectId, task: this.itemId };

      const [output, user, schedule, integration] = await Promise.all([

        (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/tasks/${this.itemId}/output`,
          responseType: 'json',
        })).data.map((item) => ({
          ...item,
          id: item.time + item.output,
        })),

        this.item.user_id ? (await axios({
          method: 'get',
          url: `/api/users/${this.item.user_id}`,
          responseType: 'json',
        })).data : null,

        this.item.schedule_id
          ? this.loadOrigin(`/api/project/${this.projectId}/schedules/${this.item.schedule_id}`)
          : null,

        this.item.integration_id
          ? this.loadOrigin(`/api/project/${this.projectId}/integrations/${this.item.integration_id}`)
          : null,
      ]);

      if (requested.project !== this.projectId || requested.task !== this.itemId) {
        return;
      }

      this.output = output;
      this.user = user;
      this.schedule = schedule;
      this.integration = integration;
    },
  },
};
</script>
