<template>
  <div>
    <v-dialog v-model="clearDialog" max-width="450" persistent>
      <v-card>
        <v-card-title>{{ $t('clearTasksTitle') }}</v-card-title>
        <v-card-text>
          <v-alert type="warning" dense text>
            {{ $t('clearTasksWarning') }}
          </v-alert>
          <div class="mb-2">{{ $t('clearTasksGroups') }}</div>
          <v-checkbox
            v-model="clearScope.queue"
            :label="$t('queue')"
            hide-details dense
          />
          <v-checkbox
            v-model="clearScope.running"
            :label="$t('running')"
            hide-details dense
          />
          <v-checkbox
            v-model="clearScope.active"
            :label="$t('activeByProject')"
            hide-details dense
          />
          <v-checkbox
            v-model="clearScope.aliases"
            :label="$t('aliases')"
            hide-details dense
          />
          <v-checkbox
            v-model="clearScope.claims"
            :label="$t('claims')"
            hide-details dense
          />
          <v-checkbox
            v-model="clearScope.runtime_fields"
            :label="$t('runtimeFields')"
            hide-details dense
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text @click="clearDialog = false">{{ $t('cancel') }}</v-btn>
          <v-btn
            color="error"
            :disabled="!isAnyGroupSelected || clearing"
            :loading="clearing"
            @click="clearTasks()"
          >
            {{ $t('clearTasksFromRedis') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-toolbar flat>
      <v-btn icon class="mr-4" @click="returnToProjects()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ $t('clusterDashboard') }}</v-toolbar-title>
      <v-spacer />
      <v-btn color="error" @click="openClearDialog()">
        <v-icon left>mdi-delete-sweep</v-icon>
        {{ $t('clearTasksFromRedis') }}
      </v-btn>
    </v-toolbar>

    <v-divider />

    <div class="pa-4">
      <v-alert v-if="status && !status.ha_enabled" type="info" dense text>
        {{ $t('haDisabledBanner') }}
      </v-alert>

      <!-- Overview -->
      <v-card class="mb-4" outlined>
        <v-card-title class="subtitle-1">{{ $t('clusterOverview') }}</v-card-title>
        <v-card-text v-if="status">
          <v-row dense>
            <v-col cols="12" sm="4">
              <div class="text-caption grey--text">{{ $t('haEnabled') }}</div>
              <v-chip small :color="status.ha_enabled ? 'success' : 'grey'" dark>
                {{ status.ha_enabled ? 'ON' : 'OFF' }}
              </v-chip>
            </v-col>
            <v-col cols="12" sm="4">
              <div class="text-caption grey--text">{{ $t('nodeId') }}</div>
              <code v-if="status.node_id">{{ status.node_id }}</code>
              <span v-else>—</span>
            </v-col>
            <v-col cols="12" sm="4">
              <div class="text-caption grey--text">{{ $t('nodes') }}</div>
              <span>{{ (status.nodes || []).length || 1 }}</span>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- Nodes -->
      <v-card v-if="status && status.nodes" class="mb-4" outlined>
        <v-card-title class="subtitle-1">{{ $t('nodes') }}</v-card-title>
        <v-data-table
          :headers="nodeHeaders"
          :items="status.nodes"
          :items-per-page="20"
          dense
        >
          <template v-slot:item.node_id="{ item }">
            <code>{{ item.node_id }}</code>
            <v-chip v-if="item.is_self" x-small class="ml-2">{{ $t('thisNode') }}</v-chip>
          </template>
          <template v-slot:item.alive="{ item }">
            <v-chip x-small :color="item.alive ? 'success' : 'error'" dark>
              {{ item.alive ? $t('alive') : $t('dead') }}
            </v-chip>
          </template>
          <template v-slot:item.last_heartbeat="{ item }">
            {{ formatTime(item.last_heartbeat) }}
          </template>
          <template v-slot:item.started_at="{ item }">
            {{ formatTime(item.started_at) }}
          </template>
        </v-data-table>
      </v-card>

      <!-- Redis -->
      <v-card v-if="status && status.redis" class="mb-4" outlined>
        <v-card-title class="subtitle-1">{{ $t('redisStatus') }}</v-card-title>
        <v-card-text>
          <v-row dense>
            <v-col cols="6" sm="3">
              <div class="text-caption grey--text">Addr</div>
              <code>{{ status.redis.addr }}</code>
            </v-col>
            <v-col cols="6" sm="2">
              <div class="text-caption grey--text">{{ $t('connected') }}</div>
              <v-chip x-small :color="status.redis.connected ? 'success' : 'error'" dark>
                {{ status.redis.connected ? 'YES' : 'NO' }}
              </v-chip>
            </v-col>
            <v-col cols="6" sm="2">
              <div class="text-caption grey--text">{{ $t('version') }}</div>
              {{ status.redis.version || '—' }}
            </v-col>
            <v-col cols="6" sm="2">
              <div class="text-caption grey--text">{{ $t('usedMemory') }}</div>
              {{ status.redis.used_memory || '—' }}
            </v-col>
            <v-col cols="6" sm="3">
              <div class="text-caption grey--text">{{ $t('totalKeys') }}</div>
              {{ status.redis.total_keys }}
            </v-col>
          </v-row>
          <div class="text-caption grey--text mt-3">{{ $t('keyGroups') }}</div>
          <v-chip
            v-for="(count, group) in (status.redis.key_groups || {})"
            :key="group"
            small
            class="mr-2 mt-1"
          >
            {{ group }}: {{ count }}
          </v-chip>
        </v-card-text>
      </v-card>

      <!-- Task records -->
      <v-card outlined>
        <v-card-title class="subtitle-1">{{ $t('taskRecords') }}</v-card-title>
        <v-tabs v-model="tab">
          <v-tab>{{ $t('queue') }} ({{ queueRows.length }})</v-tab>
          <v-tab>{{ $t('running') }} ({{ runningRows.length }})</v-tab>
          <v-tab>{{ $t('activeByProject') }} ({{ activeRows.length }})</v-tab>
          <v-tab>{{ $t('aliases') }} ({{ aliasRows.length }})</v-tab>
          <v-tab>{{ $t('claims') }} ({{ claimRows.length }})</v-tab>
        </v-tabs>
        <v-tabs-items v-model="tab">
          <v-tab-item>
            <v-data-table :headers="taskHeaders" :items="queueRows" dense />
          </v-tab-item>
          <v-tab-item>
            <v-data-table :headers="taskHeaders" :items="runningRows" dense />
          </v-tab-item>
          <v-tab-item>
            <v-data-table :headers="taskHeaders" :items="activeRows" dense />
          </v-tab-item>
          <v-tab-item>
            <v-data-table :headers="aliasHeaders" :items="aliasRows" dense />
          </v-tab-item>
          <v-tab-item>
            <v-data-table :headers="claimHeaders" :items="claimRows" dense />
          </v-tab-item>
        </v-tabs-items>
      </v-card>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';

export default {
  data() {
    return {
      status: null,
      tasks: null,
      tab: 0,
      clearDialog: false,
      clearing: false,
      clearScope: {
        queue: true,
        running: false,
        active: false,
        aliases: false,
        claims: false,
        runtime_fields: false,
      },
      updateTimer: null,
    };
  },

  computed: {
    isAnyGroupSelected() {
      return Object.values(this.clearScope).some((v) => v);
    },

    nodeHeaders() {
      return [
        { text: this.$t('nodeId'), value: 'node_id' },
        { text: this.$t('status'), value: 'alive' },
        { text: this.$t('lastHeartbeat'), value: 'last_heartbeat' },
        { text: this.$t('startedAt'), value: 'started_at' },
        { text: this.$t('version'), value: 'version' },
      ];
    },

    taskHeaders() {
      return [
        { text: 'Task', value: 'task_id' },
        { text: this.$t('project'), value: 'project_id' },
        { text: 'Template', value: 'template_id' },
        { text: this.$t('status'), value: 'status' },
        { text: 'Runner', value: 'runner_id' },
        { text: this.$t('username'), value: 'username' },
        { text: this.$t('owningNode'), value: 'node_id' },
      ];
    },

    aliasHeaders() {
      return [
        { text: this.$t('aliases'), value: 'alias' },
        { text: 'Task', value: 'task_id' },
      ];
    },

    claimHeaders() {
      return [{ text: 'Task', value: 'task_id' }];
    },

    queueRows() {
      return (this.tasks && this.tasks.queue) || [];
    },

    runningRows() {
      return (this.tasks && this.tasks.running) || [];
    },

    activeRows() {
      const byProj = (this.tasks && this.tasks.active_by_project) || {};
      return Object.values(byProj).reduce((acc, list) => acc.concat(list || []), []);
    },

    aliasRows() {
      const aliases = (this.tasks && this.tasks.aliases) || {};
      return Object.keys(aliases).map((alias) => ({ alias, task_id: aliases[alias] }));
    },

    claimRows() {
      return ((this.tasks && this.tasks.claims) || []).map((id) => ({ task_id: id }));
    },
  },

  created() {
    this.reload();
    this.updateTimer = setInterval(() => {
      if (!document.hidden) {
        this.reload();
      }
    }, 10000);
  },

  destroyed() {
    clearInterval(this.updateTimer);
  },

  methods: {
    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    formatTime(value) {
      if (!value || value.startsWith('0001-01-01')) {
        return '—';
      }
      return new Date(value).toLocaleString();
    },

    async reload() {
      try {
        this.status = (await axios({
          method: 'get',
          url: '/api/cluster',
          responseType: 'json',
        })).data;

        this.tasks = (await axios({
          method: 'get',
          url: '/api/cluster/tasks',
          responseType: 'json',
        })).data;
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: `Failed to load cluster status: ${err.message}`,
        });
      }
    },

    openClearDialog() {
      this.clearScope = {
        queue: true,
        running: false,
        active: false,
        aliases: false,
        claims: false,
        runtime_fields: false,
      };
      this.clearDialog = true;
    },

    async clearTasks() {
      this.clearing = true;
      try {
        const res = (await axios({
          method: 'delete',
          url: '/api/cluster/tasks',
          responseType: 'json',
          data: { scope: this.clearScope },
        })).data;

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: `Cleared ${res.deleted_keys} record(s) from the backend.`,
        });
        this.clearDialog = false;
        await this.reload();
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: `Failed to clear tasks: ${err.message}`,
        });
      } finally {
        this.clearing = false;
      }
    },
  },
};
</script>
