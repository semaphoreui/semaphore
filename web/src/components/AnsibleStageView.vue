<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div style="overflow: hidden;" class="pb-5">
    <v-alert
      type="info"
      text
      color="hsl(348deg, 86%, 61%)"
      style="border-radius: 0;"
      v-if="!premiumFeatures.task_result"
    >
        <span class="mr-2">
          This is <b>DEMO</b> data.
          Task summary available only in <b>PRO</b> version.
        </span>
      <v-btn
        color="hsl(348deg, 86%, 61%)"
        target="_blank"
        href="https://semaphoreui.com/pro#task_result"
      >
        Learn more
        <v-icon>mdi-chevron-right</v-icon>
      </v-btn>
    </v-alert>

    <div class="pl-5 pt-5 d-flex" style="column-gap: 10px;">
      <div class="AnsibleServerStatus AnsibleServerStatus--ok">
        <div class="AnsibleServerStatus__count">{{ okServers }}</div>
        <div class="AnsibleServerStatus__title">OK SERVERS</div>
      </div>

      <div class="AnsibleServerStatus AnsibleServerStatus--bad">
        <div class="AnsibleServerStatus__count">{{ notOkServers }}</div>
        <div class="AnsibleServerStatus__title">NOT OK SERVERS</div>
      </div>
    </div>

    <v-btn-toggle class="pl-5 mt-8 mb-3" dense v-model="tab" mandatory>
      <v-btn value="notOkServers">
        Not ok servers
      </v-btn>
      <v-btn value="allServers">
        All servers
      </v-btn>
    </v-btn-toggle>

    <v-data-table
      v-if="tab === 'notOkServers'"
      hide-default-footer
      single-expand
      show-expand
      :headers="notOkServersHeaders"
      :items="failedTasks"
      :items-per-page="Number.MAX_VALUE"
      class="w-100"
    >
      <template v-slot:item.error="{ item }">
        <div
          style="overflow: hidden; color: #ff5252; max-width: 400px; text-overflow: ellipsis">
          {{ item.error }}
        </div>
      </template>
      <template v-slot:expanded-item="{ headers, item }">
        <td
          :colspan="headers.length"
        >
            <pre style="overflow: auto;
                  background: gray;
                  font-size: 14px;
                  color: white;
                  border-radius: 10px;
                  white-space: pre-wrap;
                  margin-top: 5px;
                  margin-bottom: 5px;"

                 class="pa-2"
            >{{ item.error.trim() }}</pre>
        </td>
      </template>
    </v-data-table>

<!--    <v-simple-table v-if="tab === 'notOkServers'">-->
<!--      <template v-slot:default>-->
<!--        <thead>-->
<!--        <tr>-->
<!--          <th style="width: 150px;">Server</th>-->
<!--          <th style="width: 200px;">Task</th>-->
<!--          <th style="width: calc(100% - 350px);">Error</th>-->
<!--        </tr>-->
<!--        </thead>-->
<!--        <tbody>-->
<!--        <tr v-if="!failedTasks || failedTasks.length === 0">-->
<!--          <td colspan="3" class="text-center">No failed tasks</td>-->
<!--        </tr>-->

<!--        <tr v-else v-for="(task, index) in failedTasks" :key="index">-->
<!--          <td style="width: 150px;">{{ task.host }}</td>-->
<!--          <td style="width: 200px;">{{ task.task }}</td>-->
<!--          <td>-->
<!--            <div-->
<!--              style="overflow: hidden; color: #ff5252;
max-width: 400px; text-overflow: ellipsis">-->
<!--              {{ task.error }}-->
<!--            </div>-->
<!--          </td>-->
<!--        </tr>-->

<!--        </tbody>-->
<!--      </template>-->
<!--    </v-simple-table>-->

    <v-simple-table v-else-if="tab === 'allServers'">
      <template v-slot:default>
        <thead>
        <tr>
          <th>Host</th>
          <th>Changed</th>
          <th>Failed</th>
          <th>Ignored</th>
          <th>Ok</th>
          <th>Rescued</th>
          <th>Skipped</th>
          <th>Unreachable</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="(host, index) in hosts" :key="index">
          <td>{{ host.host }}</td>

          <td :style="{
            color: (host.changed > 0 ? 'rgb(170,85,0)' : undefined),
            'font-weight': (host.changed > 0 ? 'bold' : undefined),
          }"
          >{{ host.changed }}</td>

          <td :style="{
            color: (host.failed > 0 ? 'red' : undefined),
            'font-weight': (host.failed > 0 ? 'bold' : undefined),
          }">{{ host.failed }}</td>

          <td :style="{
            color: (host.ignored > 0 ? 'red' : undefined),
            'font-weight': (host.ignored > 0 ? 'bold' : undefined),
          }"
          >{{ host.ignored }}</td>

          <td :style="{
            color: (host.ok > 0 ? 'green' : undefined),
            'font-weight': (host.ok > 0 ? 'bold' : undefined),
          }"
          >{{ host.ok }}</td>

          <td :style="{
            'font-weight': (host.rescued > 0 ? 'bold' : undefined),
          }"
          >{{ host.rescued }}</td>

          <td :style="{
            color: (host.skipped > 0 ? 'rgb(0,170,170)' : undefined),
            'font-weight': (host.skipped > 0 ? 'bold' : undefined),
          }"
          >{{ host.skipped }}</td>

          <td :style="{
            color: (host.unreachable > 0 ? 'red' : undefined),
            'font-weight': (host.unreachable > 0 ? 'bold' : undefined),
          }">
            {{ host.unreachable }}
          </td>
        </tr>
        </tbody>
      </template>
    </v-simple-table>
  </div>
</template>
<style lang="scss">
.AnsibleServerStatus {
  text-align: center;
  width: 250px;
  font-weight: bold;
  color: white;
  font-size: 24px;
  line-height: 1.2;
  border-radius: 8px;
}

.AnsibleServerStatus__count {
  padding-top: 10px;
  font-size: 80px;
  line-height: 1;
}

.AnsibleServerStatus--ok {
  background-color: #4caf50;
}

.AnsibleServerStatus--bad {
  background-color: #ff5252;
}

.AnsibleServerStatus__title {
  padding-bottom: 10px;
}
</style>

<script>

import ProjectMixin from '@/components/ProjectMixin';

export default {
  props: {
    projectId: Number,
    taskId: Number,
    premiumFeatures: Object,
  },

  mixins: [ProjectMixin],

  data() {
    return {
      stages: null,
      okServers: 0,
      notOkServers: 0,
      tab: 'notOkServers',
      failedTasks: [],
      hosts: null,
      notOkServersHeaders: [{
        text: 'Server',
        value: 'host',
        sortable: false,
      }, {
        text: 'Task',
        value: 'task',
        sortable: false,
      }, {
        text: 'Error',
        value: 'error',
        sortable: false,
      }],
    };
  },

  watch: {
    async taskId() {
      await this.loadData();
      this.calcStats();
    },
  },

  async created() {
    await this.loadData();
    this.calcStats();
  },

  methods: {
    async loadData() {
      [this.failedTasks, this.hosts, this.stages] = await Promise.all([
        this.loadProjectEndpoint(`/tasks/${this.taskId}/ansible/errors`),
        this.loadProjectEndpoint(`/tasks/${this.taskId}/ansible/hosts`),
        this.loadProjectEndpoint(`/tasks/${this.taskId}/stages`),
      ]);

      this.hosts.forEach((host) => {
        if (host.unreachable) {
          this.failedTasks.push({
            host: host.host,
            task: '—',
            error: 'Host is unreachable',
          });
        }
      });
    },

    calcStats() {
      this.hosts.forEach((host) => {
        if (host.failed > 0 || host.unreachable > 0) {
          this.notOkServers += 1;
        } else {
          this.okServers += 1;
        }
      });
    },
  },
};
</script>
