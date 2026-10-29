<template>
  <div data-testid="cluster-page" :data-state="status == null ? 'loading' : 'ready'">
    <v-dialog v-model="clearDialog" max-width="450" persistent>
      <v-card
        data-testid="clusterClearDialog"
        :data-state="clearing ? 'loading' : 'ready'"
      >
        <v-card-title
          data-testid="clusterClearDialog-title"
        >{{ $t('clearTasksTitle') }}</v-card-title>
        <v-card-text>
          <v-alert
            type="warning"
            density="compact"
            variant="tonal"
            data-testid="clusterClearDialog-warning"
          >
            {{ $t('clearTasksWarning') }}
          </v-alert>
          <div class="mb-2">{{ $t('clearTasksGroups') }}</div>
          <v-checkbox
            v-model="clearScope.queue"
            :label="$t('queue')"
            hide-details
            density="compact"
            data-testid="clusterClearDialog-queue"
          />
          <v-checkbox
            v-model="clearScope.running"
            :label="$t('running')"
            hide-details
            density="compact"
            data-testid="clusterClearDialog-running"
          />
          <v-checkbox
            v-model="clearScope.active"
            :label="$t('activeByProject')"
            hide-details
            density="compact"
            data-testid="clusterClearDialog-active"
          />
          <v-checkbox
            v-model="clearScope.aliases"
            :label="$t('aliases')"
            hide-details
            density="compact"
            data-testid="clusterClearDialog-aliases"
          />
          <v-checkbox
            v-model="clearScope.claims"
            :label="$t('claims')"
            hide-details
            density="compact"
            data-testid="clusterClearDialog-claims"
          />
          <v-checkbox
            v-model="clearScope.runtime_fields"
            :label="$t('runtimeFields')"
            hide-details
            density="compact"
            data-testid="clusterClearDialog-runtimeFields"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            @click="clearDialog = false"
            data-testid="clusterClearDialog-cancel"
          >{{ $t('cancel') }}</v-btn>
          <v-btn
            color="error"
            variant="elevated"
            :disabled="!isAnyGroupSelected || clearing"
            :loading="clearing"
            @click="clearTasks()"
            data-testid="clusterClearDialog-clear"
          >
            {{ $t('clearTasksFromRedis') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-toolbar flat>
      <v-btn icon class="mr-4" @click="returnToProjects()" data-testid="cluster-back">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title data-testid="cluster-title">{{ $t('clusterDashboard') }}</v-toolbar-title>
      <v-spacer />
      <v-btn
        color="error"
        variant="elevated"
        @click="openClearDialog()"
        :disabled="!features.high_availability || !status.ha_enabled"
        data-testid="cluster-clear"
      >
        <v-icon start>mdi-delete-sweep</v-icon>
        {{ $t('clearTasksFromRedis') }}
      </v-btn>
    </v-toolbar>

    <v-divider />

    <v-alert
      v-if="!features.high_availability"
      variant="tonal"
      color="amber-darken-3"
      class="PageAlert"
      data-testid="cluster-licenseAlert"
    >
      <span class="mr-1" v-html="$t('ha_only_enterprise')"></span>

      <v-btn
        variant="flat"
        v-if="canManageSystem"
        color="amber-darken-3"
        href="https://semaphoreui.com/enterprise?utm_source=app&utm_content=feature_cluster"
        target="_blank"
        data-testid="cluster-upgrade"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>

      <span v-else style="font-weight: bold" data-testid="cluster-contactAdmin">
        {{ $t('contact_admin_to_upgrade_enterprise') }}
      </span>
    </v-alert>

    <v-alert
      v-if="status && !status.ha_enabled"
      type="info"
      density="compact"
      variant="tonal"
      data-testid="cluster-haDisabled"
    >
      {{ $t('haDisabledBanner') }}
    </v-alert>

    <div class="pa-4">
      <!-- Nodes -->
      <v-card v-if="status && status.nodes" class="mb-4" border data-testid="cluster-nodes">
        <v-card-title class="text-subtitle-1">{{ $t('nodes') }}</v-card-title>
        <v-data-table
          :headers="nodeHeaders"
          :items="status.nodes"
          :items-per-page="20"
          density="compact"
          :mobile-breakpoint="600"
          data-testid="cluster-nodesTable"
          :data-state="status.nodes.length === 0 ? 'empty' : 'ready'"
        >
          <template v-slot:item.node_id="{ item }">
            <code data-testid="cluster-rowNodeId">{{ item.node_id }}</code>
            <v-chip
              v-if="item.is_self"
              size="x-small"
              class="ml-2"
              data-testid="cluster-rowThisNode"
            >{{ $t('thisNode') }}</v-chip>
          </template>
          <template v-slot:item.alive="{ item }">
            <v-chip
              size="x-small"
              :color="item.alive ? 'success' : 'error'"
              variant="flat"
              :data-testid="`cluster-rowStatus-${item.alive ? 'alive' : 'dead'}`"
            >
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
      <v-card v-if="status && status.redis" class="mb-4" border data-testid="cluster-redis">
        <v-card-title class="text-subtitle-1">{{ $t('redisStatus') }}</v-card-title>
        <v-card-text>
          <v-row dense>
            <v-col cols="12" md="6">
              <v-row dense>
                <v-col>
                  <div class="text-caption text-grey">Addr</div>
                  <code data-testid="cluster-redisAddr">{{ status.redis.addr }}</code>
                </v-col>
                <v-col>
                  <div class="text-caption text-grey">{{ $t('connected') }}</div>
                  <v-chip
                    size="x-small"
                    :color="status.redis.connected ? 'success' : 'error'"
                    variant="flat"
                    data-testid="cluster-redisConnected"
                  >
                    {{ status.redis.connected ? 'YES' : 'NO' }}
                  </v-chip>
                </v-col>
                <v-col data-testid="cluster-redisVersion">
                  <div class="text-caption text-grey">{{ $t('version') }}</div>
                  {{ status.redis.version || '—' }}
                </v-col>
                <v-col data-testid="cluster-redisUsedMemory">
                  <div class="text-caption text-grey">{{ $t('usedMemory') }}</div>
                  {{ status.redis.used_memory || '—' }}
                </v-col>
                <v-col data-testid="cluster-redisTotalKeys">
                  <div class="text-caption text-grey">{{ $t('totalKeys') }}</div>
                  {{ status.redis.total_keys }}
                </v-col>
              </v-row>
              <div class="text-caption text-grey mt-3">{{ $t('keyGroups') }}</div>
              <v-chip
                v-for="(count, group) in status.redis.key_groups || {}"
                :key="group"
                size="small"
                class="mr-2 mt-1"
                :data-testid="`cluster-redisKeyGroup-${group}`"
              >
                {{ group }}: {{ count }}
              </v-chip>
            </v-col>
            <v-col cols="12" md="6">
              <div class="text-caption text-grey">{{ $t('redisMemoryChart') }}</div>
              <RedisMemoryChart
                data-testid="cluster-redisChart"
                :history="memoryHistory"
                :window-ms="memoryWindowMs"
                :now-ts="memoryNow"
              />
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- Task records -->
      <v-card border data-testid="cluster-taskRecords">
        <v-card-title class="text-subtitle-1">{{ $t('taskRecords') }}</v-card-title>
        <v-tabs v-model="tab" data-testid="cluster-tabs">
          <v-tab
            :value="0"
            data-testid="cluster-tab-queue"
          >{{ $t('queue') }} ({{ queueRows.length }})</v-tab>
          <v-tab
            :value="1"
            data-testid="cluster-tab-running"
          >{{ $t('running') }} ({{ runningRows.length }})</v-tab>
          <v-tab
            :value="2"
            data-testid="cluster-tab-activeByProject"
          >{{ $t('activeByProject') }} ({{ activeRows.length }})</v-tab>
          <v-tab
            :value="3"
            data-testid="cluster-tab-aliases"
          >{{ $t('aliases') }} ({{ aliasRows.length }})</v-tab>
          <v-tab
            :value="4"
            data-testid="cluster-tab-claims"
          >{{ $t('claims') }} ({{ claimRows.length }})</v-tab>
        </v-tabs>
        <v-tabs-window v-model="tab">
          <v-tabs-window-item :value="0">
            <v-data-table
              :headers="taskHeaders"
              :items="queueRows"
              density="compact"
              :mobile-breakpoint="600"
              data-testid="cluster-queueTable"
              :data-state="tasks == null
                ? 'loading' : (queueRows.length === 0 ? 'empty' : 'ready')"
            />
          </v-tabs-window-item>
          <v-tabs-window-item :value="1">
            <v-data-table
              :headers="taskHeaders"
              :items="runningRows"
              density="compact"
              :mobile-breakpoint="600"
              data-testid="cluster-runningTable"
              :data-state="tasks == null
                ? 'loading' : (runningRows.length === 0 ? 'empty' : 'ready')"
            />
          </v-tabs-window-item>
          <v-tabs-window-item :value="2">
            <v-data-table
              :headers="taskHeaders"
              :items="activeRows"
              density="compact"
              :mobile-breakpoint="600"
              data-testid="cluster-activeByProjectTable"
              :data-state="tasks == null
                ? 'loading' : (activeRows.length === 0 ? 'empty' : 'ready')"
            />
          </v-tabs-window-item>
          <v-tabs-window-item :value="3">
            <v-data-table
              :headers="aliasHeaders"
              :items="aliasRows"
              density="compact"
              :mobile-breakpoint="600"
              data-testid="cluster-aliasesTable"
              :data-state="tasks == null
                ? 'loading' : (aliasRows.length === 0 ? 'empty' : 'ready')"
            />
          </v-tabs-window-item>
          <v-tabs-window-item :value="4">
            <v-data-table
              :headers="claimHeaders"
              :items="claimRows"
              density="compact"
              :mobile-breakpoint="600"
              data-testid="cluster-claimsTable"
              :data-state="tasks == null
                ? 'loading' : (claimRows.length === 0 ? 'empty' : 'ready')"
            />
          </v-tabs-window-item>
        </v-tabs-window>
      </v-card>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import RedisMemoryChart from '@/components/RedisMemoryChart.vue';
import { GLOBAL_PERMISSIONS } from '@/lib/constants';
import { hasGlobalPermission } from '@/lib/role-permissions';

const MEMORY_WINDOW_MS = 60 * 60 * 1000;

export default {
  components: { RedisMemoryChart },

  props: {
    features: Object,
    isAdmin: Boolean,
    user: Object,
  },

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
      memoryHistory: [],
      memoryNow: Date.now(),
      memoryWindowMs: MEMORY_WINDOW_MS,
    };
  },

  computed: {
    canManageSystem() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageSystem);
    },

    isAnyGroupSelected() {
      return Object.values(this.clearScope).some((v) => v);
    },

    nodeHeaders() {
      return [
        { title: this.$t('nodeId'), key: 'node_id' },
        { title: this.$t('status'), key: 'alive' },
        { title: this.$t('lastHeartbeat'), key: 'last_heartbeat' },
        { title: this.$t('startedAt'), key: 'started_at' },
        { title: this.$t('version'), key: 'version' },
      ];
    },

    taskHeaders() {
      return [
        { title: 'Task', key: 'task_id' },
        { title: this.$t('project'), key: 'project_id' },
        { title: 'Template', key: 'template_id' },
        { title: this.$t('status'), key: 'status' },
        { title: 'Runner', key: 'runner_id' },
        { title: this.$t('username'), key: 'username' },
        { title: this.$t('owningNode'), key: 'node_id' },
      ];
    },

    aliasHeaders() {
      return [
        { title: this.$t('aliases'), key: 'alias' },
        { title: 'Task', key: 'task_id' },
      ];
    },

    claimHeaders() {
      return [{ title: 'Task', key: 'task_id' }];
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

  unmounted() {
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
        this.status = (
          await axios({
            method: 'get',
            url: '/api/cluster',
            responseType: 'json',
          })
        ).data;

        this.recordMemorySample();

        this.tasks = (
          await axios({
            method: 'get',
            url: '/api/cluster/tasks',
            responseType: 'json',
          })
        ).data;
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: `Failed to load cluster status: ${err.message}`,
        });
      }
    },

    recordMemorySample() {
      const now = Date.now();
      this.memoryNow = now;
      const redis = this.status && this.status.redis;
      if (!redis || !redis.connected) {
        return;
      }
      this.memoryHistory.push({
        time: now,
        bytes: redis.used_memory_bytes || 0,
      });
      const cutoff = now - this.memoryWindowMs;
      const firstInWindow = this.memoryHistory.findIndex((p) => p.time >= cutoff);
      if (firstInWindow > 0) {
        this.memoryHistory.splice(0, firstInWindow);
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
        const res = (
          await axios({
            method: 'delete',
            url: '/api/cluster/tasks',
            responseType: 'json',
            data: { scope: this.clearScope },
          })
        ).data;

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
