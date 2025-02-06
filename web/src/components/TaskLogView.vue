<template>
  <div
    class="task-log-view"
    :class="{'task-log-view--with-message': item.message || item.commit_message}"
  >

    <div class="overflow-auto text-no-wrap">
      <v-alert
        dense
        class="d-inline-block mb-2 mr-2"
        text
        icon="mdi-message-outline"
        v-if="item.message"
      >
        {{ item.message }}
      </v-alert>

      <v-alert
        dense
        class="d-inline-block mb-2"
        text
        icon="mdi-source-fork"
        v-if="item.commit_message"
      >
        {{ item.commit_message }}
      </v-alert>
    </div>

    <v-container fluid class="pa-0 mb-2 overflow-auto">
      <v-row no-gutters class="flex-nowrap">
        <v-col>
          <v-list two-line subheader class="pa-0">
            <v-list-item class="pa-0">
              <v-list-item-content>
                <div class="pr-4">
                  <TaskStatus :status="item.status"/>
                </div>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col class="pr-4">
          <v-list two-line subheader class="pa-0">
            <v-list-item class="pa-0">
              <v-list-item-content v-if="item.user_id != null">
                <v-list-item-title>{{ $t('author') }}</v-list-item-title>
                <v-list-item-subtitle>{{ user.name || '-' }}</v-list-item-subtitle>
              </v-list-item-content>
              <v-list-item-content v-else-if="item.integration_id != null">
                <v-list-item-title>{{ $t('integration') }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col class="pr-4">
          <v-list two-line subheader class="pa-0">
            <v-list-item class="pa-0">
              <v-list-item-content>
                <v-list-item-title>{{ $t('started') || '-' }}</v-list-item-title>
                <v-list-item-subtitle>
                  {{ item.start | formatDate }}
                </v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list-item class="pa-0">
            <v-list-item-content>
              <v-list-item-title>{{ $t('duration') || '-' }}</v-list-item-title>
              <v-list-item-subtitle>
                {{ [item.start, item.end] | formatMilliseconds }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-col>
      </v-row>
    </v-container>

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
      style="position: absolute; bottom: 10px; right: 250px; width: 70px;"
      v-if="item.status === 'waiting_confirmation'"
      @click="confirmTask()"
    >
      <v-icon>mdi-check</v-icon>
    </v-btn>

    <v-btn
      color="warning"
      style="position: absolute; bottom: 10px; right: 170px; width: 70px;"
      v-if="item.status === 'waiting_confirmation'"
      @click="rejectTask()"
    >
      <v-icon>mdi-close</v-icon>
    </v-btn>

    <v-btn
      color="error"
      style="position: absolute; bottom: 10px; right: 10px; width: 150px;"
      v-if="canStop"
      @click="stopTask(item.status === 'stopping')"
    >
      {{ item.status === 'stopping' ? $t('forceStop') : $t('stop') }}
    </v-btn>

  </div>
</template>

<style lang="scss">

@import '~vuetify/src/styles/settings/_variables';

$task-log-header-height: 62px + 64px + 8px;
$task-log-message-height: 48px;

.task-log-records {
  background: black;
  color: white;
  height: calc(100vh - 280px);
  overflow: auto;
  font-family: monospace;
  margin: 0 -24px;
  padding: 5px 10px 50px;
}

.task-log-view--with-message .task-log-records {
  height: calc(100vh - #{280px + $task-log-message-height});
}

.v-dialog--fullscreen {

  .task-log-records {
    height: calc(100vh - #{$task-log-header-height});
  }

  .task-log-view--with-message .task-log-records {
    height: calc(100vh - #{$task-log-header-height + $task-log-message-height});
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

@media #{map-get($display-breakpoints, 'sm-and-down')} {
  .task-log-records {
    height: calc(100vh - 340px);
  }

  .task-log-view--with-message .task-log-records {
    height: calc(100vh - 370px);
  }
}
</style>
<script>
import axios from 'axios';
import TaskStatus from '@/components/TaskStatus.vue';
import socket from '@/socket';
import VirtualList from 'vue-virtual-scroll-list';
import TaskLogViewRecord from '@/components/TaskLogViewRecord.vue';

export default {
  components: { TaskStatus, VirtualList },
  props: {
    itemId: Number,
    projectId: Number,
  },
  data() {
    return {
      itemComponent: TaskLogViewRecord,
      item: {},
      output: [],
      outputBuffer: [],
      user: {},
      autoScroll: true,
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
  },

  computed: {
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
  },

  async created() {
    this.outputInterval = setInterval(() => {
      this.$nextTick(() => {
        const len = this.outputBuffer.length;
        if (len === 0) {
          return;
        }

        const scrollContainer = this.$refs.records.$el;

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
    }, 500);
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
      this.item = {};
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

          // this.$nextTick(() => {
          //   if (this.$refs.records) {
          //     this.$refs.records.scrollToBottom();
          //   }
          // });

          break;
        default:
          break;
      }
    },

    async loadData() {
      this.item = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/tasks/${this.itemId}`,
        responseType: 'json',
      })).data;

      this.output = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/tasks/${this.itemId}/output`,
        responseType: 'json',
      })).data.map((item) => ({
        ...item,
        id: item.time + item.output,
      }));

      this.user = this.item.user_id ? (await axios({
        method: 'get',
        url: `/api/users/${this.item.user_id}`,
        responseType: 'json',
      })).data : null;
    },
  },
};
</script>
