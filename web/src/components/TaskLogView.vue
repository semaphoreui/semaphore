<template>
  <div
    class="task-log-view"
    :class="{'task-log-view--with-message': item.message || item.commit_message}"
  >
    <div class="px-5 task-log-view__message">
      <span
        v-if="item.message"
        class="mr-3"
      >
        <v-icon small>mdi-message-outline</v-icon>
        {{ item.message }}
      </span>

      <span
        class="d-inline-block"
        v-if="item.commit_message"
      >
        <v-icon small>mdi-source-fork</v-icon>
        {{ item.commit_message }}
      </span>
    </div>

    <div
      class="overflow-auto text-no-wrap px-5 task-log-view__status"
    >
      <TaskStatus :status="item.status" data-testid="task-status" />

      <span class="ml-3 hidden-xs-only task-log-view__status_part">

        Started <span v-if="user">by <b>{{ user.name }}</b></span>

        at <b>{{ item.start | formatDate }}</b>
      </span>

      <span class="ml-3 hidden-sm-and-down task-log-view__status_part">
        <v-icon
          small style="transform: translateY(-1px)">mdi-clock-outline</v-icon>
        {{ [item.start, item.end] | formatMilliseconds }}
      </span>
    </div>

    <v-tabs class="task-log-view__tabs" right v-model="tab">
      <v-tab>Log</v-tab>
      <v-tab :disabled="!isTaskStopped">Details</v-tab>
      <v-tab v-if="isPro" :disabled="!isTaskStopped">Summary</v-tab>
    </v-tabs>

    <div v-if="tab === 0">
      <VirtualList
        class="task-log-records"
        :data-key="'id'"
        :data-sources="output"
        :data-component="itemComponent"
        :estimate-size="22"
        :keeps="100"
        ref="records"
      >
        <div class="task-log-records__record" v-for="record in output" :key="record.id">
          <div class="task-log-records__time">
            {{ record.time | formatTime }}
          </div>
          <div class="task-log-records__output" v-html="$options.filters.formatLog(record.output)">
          </div>
        </div>
      </VirtualList>

      <v-btn
        color="success"
        class="task-log-action-button"
        style="right: 260px; width: 70px;"
        v-if="item.status === 'waiting_confirmation'"
        @click="confirmTask()"
      >
        <v-icon>mdi-check</v-icon>
      </v-btn>

      <v-btn
        color="warning"
        class="task-log-action-button"
        style="right: 180px; width: 70px;"
        v-if="item.status === 'waiting_confirmation'"
        @click="rejectTask()"
      >
        <v-icon>mdi-close</v-icon>
      </v-btn>

      <v-btn
        color="error"
        class="task-log-action-button"
        style="right: 20px; width: 150px;"
        v-if="canStop"
        @click="stopTask(item.status === 'stopping')"
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

    <div v-else-if="tab === 1">
      <v-divider style="margin-top: -1px;" />

      <v-container fluid class="py-0 px-5 overflow-auto pt-4">
        <TaskDetails :item="item" :user="user" :project-id="projectId" />
      </v-container>
    </div>

    <div v-else-if="tab === 2">
      <v-divider style="margin-top: -1px;" />

      <AnsibleStageView
        :premium-features="systemInfo.premium_features"
        :project-id="projectId"
        :task-id="itemId"
      />
    </div>

  </div>
</template>

<style lang="scss">

@import '~vuetify/src/styles/settings/_variables';

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
  height: calc(90vh - #{$card-title-height + $task-log-status-tab-height});
  overflow: auto;
  font-family: monospace;
  margin: 0;
  padding: 5px 10px 50px;
}

.task-log-view--with-message .task-log-view__message {
  display: block;
}

.task-log-view--with-message .task-log-records {
  height: calc(90vh -
    #{$card-title-height + $task-log-message-height-total + $task-log-status-tab-height});
}

.v-dialog--fullscreen {

  .task-log-records {
    height: calc(100vh - #{$card-title-height + $task-log-status-tab-height});
  }

  .task-log-view--with-message .task-log-records {
    height: calc(100vh -
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
import VirtualList from 'vue-virtual-scroll-list';
import TaskLogViewRecord from '@/components/TaskLogViewRecord.vue';
import ProjectMixin from '@/components/ProjectMixin';
import AnsibleStageView from '@/components/AnsibleStageView.vue';
import TaskDetails from '@/components/TaskDetails.vue';

export default {
  components: {
    TaskDetails, AnsibleStageView, TaskStatus, VirtualList,
  },

  mixins: [ProjectMixin],

  props: {
    item: Object,
    projectId: Number,
    systemInfo: Object,
    premiumFeatures: null,
  },

  data() {
    return {
      tab: 0,
      itemComponent: TaskLogViewRecord,
      output: [],
      outputBuffer: [],
      user: {},
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

        const scrollContainer = this.$refs.records?.$el;
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
    socket.addListener((data) => this.onWebsocketDataReceived(data));
    await this.loadData();
  },

  beforeDestroy() {
    clearInterval(this.outputInterval);
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
    },

    onWebsocketDataReceived(data) {
      if (data.project_id !== this.projectId || data.task_id !== this.itemId) {
        return;
      }

      switch (data.type) {
        case 'update':
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

    async loadData() {
      [
        this.output,
        this.user,
      ] = await Promise.all([

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
      ]);
    },
  },
};
</script>
