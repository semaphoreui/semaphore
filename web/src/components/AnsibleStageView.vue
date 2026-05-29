<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div style="overflow: hidden" class="pb-5">
    <v-alert text color="hsl(348deg, 86%, 61%)" class="PageAlert" v-if="!features.task_summary">
      <span class="mr-2">
        This is <b>DEMO</b> data. Task summary available only in <b>PRO</b> version.
      </span>

      <v-btn dark class="ml-2" color="hsl(348deg, 86%, 61%)" @click="upgradeToPro('task_summary')">
        {{ $t('upgrade_to_pro') }}
      </v-btn>
    </v-alert>

    <div class="pl-5 pt-5 d-flex" style="column-gap: 10px">
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
      <v-btn value="notOkServers"> Not ok servers </v-btn>
      <v-btn value="allServers"> All servers </v-btn>
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
        <div style="overflow: hidden; color: #ff5252; max-width: 400px; text-overflow: ellipsis">
          {{ item.error }}
        </div>
      </template>
      <template v-slot:expanded-item="{ headers, item }">
        <td :colspan="headers.length">
          <pre
            style="
              overflow: auto;
              background: gray;
              font-size: 14px;
              color: white;
              border-radius: 10px;
              white-space: pre-wrap;
              margin-top: 5px;
              margin-bottom: 5px;
            "
            class="pa-2"
            >{{ item.error.trim() }}</pre
          >
        </td>
      </template>
    </v-data-table>

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

            <td
              :style="{
                color: host.changed > 0 ? 'rgb(170,85,0)' : undefined,
                'font-weight': host.changed > 0 ? 'bold' : undefined,
              }"
            >
              {{ host.changed }}
            </td>

            <td
              :style="{
                color: host.failed > 0 ? 'red' : undefined,
                'font-weight': host.failed > 0 ? 'bold' : undefined,
              }"
            >
              {{ host.failed }}
            </td>

            <td
              :style="{
                color: host.ignored > 0 ? 'red' : undefined,
                'font-weight': host.ignored > 0 ? 'bold' : undefined,
              }"
            >
              {{ host.ignored }}
            </td>

            <td
              :style="{
                color: host.ok > 0 ? 'green' : undefined,
                'font-weight': host.ok > 0 ? 'bold' : undefined,
              }"
            >
              {{ host.ok }}
            </td>

            <td
              :style="{
                'font-weight': host.rescued > 0 ? 'bold' : undefined,
              }"
            >
              {{ host.rescued }}
            </td>

            <td
              :style="{
                color: host.skipped > 0 ? 'rgb(0,170,170)' : undefined,
                'font-weight': host.skipped > 0 ? 'bold' : undefined,
              }"
            >
              {{ host.skipped }}
            </td>

            <td
              :style="{
                color: host.unreachable > 0 ? 'red' : undefined,
                'font-weight': host.unreachable > 0 ? 'bold' : undefined,
              }"
            >
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
    features: Object,
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
      notOkServersHeaders: [
        {
          text: 'Server',
          value: 'host',
          sortable: false,
        },
        {
          text: 'Task',
          value: 'task',
          sortable: false,
        },
        {
          text: 'Error',
          value: 'error',
          sortable: false,
        },
      ],
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
      if (this.features.task_summary) {
        [this.failedTasks, this.hosts, this.stages] = await Promise.all([
          this.loadProjectEndpoint(`/tasks/${this.taskId}/ansible/errors`),
          this.loadProjectEndpoint(`/tasks/${this.taskId}/ansible/hosts`),
          this.loadProjectEndpoint(`/tasks/${this.taskId}/stages`),
        ]);
      } else {
        [this.failedTasks, this.hosts, this.stages] = this.getDemoData();
      }

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

    getDemoData() {
      const failedTasks = [
        {
          id: 1,
          host: 'web-01.prod.example.com',
          task: 'Install nginx package',
          error: 'fatal: [web-01.prod.example.com]: FAILED! => {"changed": false, "msg": "No package matching \'nginx\' is available"}',
        },
        {
          id: 2,
          host: 'web-02.prod.example.com',
          task: 'Start nginx service',
          error: 'fatal: [web-02.prod.example.com]: FAILED! => {"changed": false, "msg": "Could not find the requested service nginx: host"}',
        },
        {
          id: 3,
          host: 'db-01.prod.example.com',
          task: 'Apply database migrations',
          error: 'fatal: [db-01.prod.example.com]: FAILED! => {"changed": false, "msg": "Migration failed: relation \\"users\\" already exists", "rc": 1}',
        },
        {
          id: 4,
          host: 'cache-01.prod.example.com',
          task: 'Configure redis maxmemory',
          error: 'fatal: [cache-01.prod.example.com]: FAILED! => {"changed": false, "msg": "Destination /etc/redis/redis.conf does not exist!"}',
        },
      ];

      const hosts = [
        {
          host: 'web-01.prod.example.com', changed: 3, failed: 1, ignored: 0, ok: 12, rescued: 0, skipped: 2, unreachable: 0,
        },
        {
          host: 'web-02.prod.example.com', changed: 2, failed: 1, ignored: 0, ok: 11, rescued: 0, skipped: 2, unreachable: 0,
        },
        {
          host: 'web-03.prod.example.com', changed: 4, failed: 0, ignored: 0, ok: 15, rescued: 0, skipped: 1, unreachable: 0,
        },
        {
          host: 'db-01.prod.example.com', changed: 1, failed: 1, ignored: 0, ok: 8, rescued: 0, skipped: 3, unreachable: 0,
        },
        {
          host: 'db-02.prod.example.com', changed: 0, failed: 0, ignored: 0, ok: 9, rescued: 0, skipped: 3, unreachable: 0,
        },
        {
          host: 'cache-01.prod.example.com', changed: 0, failed: 1, ignored: 0, ok: 5, rescued: 0, skipped: 0, unreachable: 0,
        },
        {
          host: 'worker-01.prod.example.com', changed: 2, failed: 0, ignored: 1, ok: 10, rescued: 0, skipped: 1, unreachable: 0,
        },
        {
          host: 'worker-02.prod.example.com', changed: 0, failed: 0, ignored: 0, ok: 0, rescued: 0, skipped: 0, unreachable: 1,
        },
      ];

      const stages = [
        {
          name: 'Gathering Facts', ok: 7, failed: 0, changed: 0,
        },
        {
          name: 'Install packages', ok: 5, failed: 2, changed: 3,
        },
        {
          name: 'Configure services', ok: 6, failed: 1, changed: 4,
        },
        {
          name: 'Start services', ok: 6, failed: 1, changed: 2,
        },
        {
          name: 'Run migrations', ok: 5, failed: 1, changed: 1,
        },
      ];

      return [failedTasks, hosts, stages];
    },
  },
};
</script>
