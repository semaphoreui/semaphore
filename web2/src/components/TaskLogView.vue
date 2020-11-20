<template>
  <div>
    <v-container class="pa-0 mb-2">
      <v-row no-gutters>
        <v-col>
          <v-list two-line subheader class="pa-0">
            <v-list-item class="pa-0">
              <v-list-item-content>
                <div>
                  <TaskStatus :status="item.status" />
                </div>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
        <v-list two-line subheader class="pa-0">
          <v-list-item class="pa-0">
            <v-list-item-content>
              <v-list-item-title>Author</v-list-item-title>
              <v-list-item-subtitle>{{ user.name }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>
        </v-col>
        <v-col>
          <v-list two-line subheader class="pa-0">
            <v-list-item class="pa-0">
              <v-list-item-content>
                <v-list-item-title>Started</v-list-item-title>
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
              <v-list-item-title>Ended</v-list-item-title>
              <v-list-item-subtitle>{{ item.end | formatDate }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-col>
      </v-row>
    </v-container>
    <div class="task-log-view">
      <div class="task-log-view__record" v-for="record in output" :key="record.id">
        <div class="task-log-view__time">{{ record.time }}</div>
        <div class="task-log-view__output">{{ record.output }}</div>
      </div>
    </div>
  </div>
</template>
<style lang="scss">
  .task-log-view {
    height: 400px;
    overflow: auto;
    border: 1px solid gray;
    border-radius: 4px;
    font-family: monospace;
  }

  .task-log-view__record {
    display: flex;
    flex-direction: row;
    justify-content: left;
  }

  .task-log-view__time {
    width: 250px;
  }

  .task-log-view__output {
    width: calc(100% - 250px);
  }
</style>
<script>
import axios from 'axios';
import TaskStatus from '@/components/TaskStatus.vue';

export default {
  components: { TaskStatus },
  props: {
    itemId: Number,
    projectId: Number,
  },
  data() {
    return {
      item: {},
      output: [],
      user: {},
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
  async created() {
    await this.loadData();
  },
  methods: {
    reset() {
      this.item = {};
      this.output = [];
      this.user = {};
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
      })).data;

      this.user = (await axios({
        method: 'get',
        url: `/api/users/${this.item.user_id}`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
