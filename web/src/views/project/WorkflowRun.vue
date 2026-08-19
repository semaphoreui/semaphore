<template>
  <div>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        <router-link :to="`/project/${projectId}/workflows`">
          {{ $t('workflows') }}
        </router-link>
        <span class="ml-2">
          / {{ workflow ? workflow.name : $t('workflowRun') }} #{{ runId }}
          <span v-if="details && details.run.version" class="text--secondary">
            · {{ details.run.version }}
          </span>
        </span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-chip
        v-if="details"
        :color="statusColor(details.run.status)"
        small
        class="mr-3"
      >{{ details.run.status }}</v-chip>

      <v-btn
        v-if="canStopRun"
        color="error"
        small
        outlined
        class="mr-3"
        :loading="stopping"
        @click="stopRun()"
      >
        <v-icon left small>mdi-stop</v-icon>
        {{ $t('stop') }}
      </v-btn>

      <v-btn icon :title="$t('workflowToolbarZoomOut')" @click="zoomOut()">
        <v-icon>mdi-magnify-minus-outline</v-icon>
      </v-btn>
      <v-btn icon :title="$t('workflowToolbarZoomIn')" @click="zoomIn()">
        <v-icon>mdi-magnify-plus-outline</v-icon>
      </v-btn>
      <v-btn icon :title="$t('workflowToolbarFit')" @click="zoomReset()">
        <v-icon>mdi-fit-to-page-outline</v-icon>
      </v-btn>
      <v-btn icon @click="loadData()" :title="$t('refresh')">
        <v-icon>mdi-refresh</v-icon>
      </v-btn>
    </v-toolbar>

    <v-divider />

    <div class="WorkflowRun__body">
      <template v-if="details != null">
        <v-alert
          v-if="hasRemoteRunnerNodes"
          type="warning"
          dense
          text
          tile
          class="ma-0"
          icon="mdi-alert-outline"
        >{{ $t('workflowArtifactsRemoteRunnerWarning') }}</v-alert>

        <div class="WorkflowRun__graph">
          <WorkflowGraph
            v-if="workflow"
            ref="graph"
            :nodes="workflow.nodes || []"
            :edges="workflow.edges || []"
            :templates="templates"
            :node-statuses="nodeStatuses"
            :node-delays="nodeDelays"
            :editable="false"
            @node-selected="onNodeClicked"
          />

          <div
            v-if="canResolveApprovals && pendingApprovals.length"
            class="WorkflowRun__approvals"
          >
            <v-card
              v-for="a in pendingApprovals"
              :key="`approval-${a.nodeId}`"
              class="WorkflowRun__approval px-3 py-2"
              outlined
            >
              <div class="WorkflowRun__approvalText">
                <strong>#{{ a.nodeId }}</strong>
                <span class="ml-2">{{ a.message || $t('workflowApprovalPending') }}</span>
              </div>
              <v-spacer />
              <v-btn
                small
                color="success"
                class="ml-3"
                @click="resolveApproval(a.nodeId, 'approved')"
              >{{ $t('workflowApprove') }}</v-btn>
              <v-btn
                small
                color="error"
                class="ml-2"
                @click="resolveApproval(a.nodeId, 'rejected')"
              >{{ $t('workflowReject') }}</v-btn>
            </v-card>
          </div>
        </div>
      </template>

      <div v-else class="pa-4 text-center">
        <v-progress-circular indeterminate color="primary" />
      </div>
    </div>
  </div>
</template>

<style lang="scss">
.WorkflowRun {

  &__body {
    height: calc(100vh - 65px);
    flex: 1 1 auto;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  &__graph {
    position: relative;
    flex: 1 1 auto;
    min-height: 0;
  }

  &__approvals {
    position: absolute;
    left: 50%;
    bottom: 16px;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-width: 92%;
    z-index: 5;
  }

  &__approval {
    display: flex;
    align-items: center;
    border-left: 3px solid #ff9800 !important;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.25) !important;
  }

  &__approvalText {
    font-size: 13px;
    word-break: break-word;
  }
}
</style>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import PermissionsCheck from '@/components/PermissionsCheck';
import WorkflowGraph from '@/components/WorkflowGraph.vue';
import { USER_PERMISSIONS } from '@/lib/constants';

export default {
  components: { WorkflowGraph },
  mixins: [PermissionsCheck],
  props: {
    projectId: Number,
  },
  data() {
    return {
      details: null,
      workflow: null,
      templates: [],
      pollHandle: null,
      stopping: false,
      USER_PERMISSIONS,
    };
  },
  computed: {
    workflowId() {
      return parseInt(this.$route.params.workflowId, 10);
    },
    runId() {
      return parseInt(this.$route.params.runId, 10);
    },
    canResolveApprovals() {
      return this.can(USER_PERMISSIONS.runProjectTasks);
    },
    // A run can be stopped while it is still in progress (executing tasks or
    // blocked on an approval) and the user may run project tasks.
    canStopRun() {
      if (!this.details) return false;
      const status = this.details.run.status;
      return (status === 'running' || status === 'approval')
        && this.can(USER_PERMISSIONS.runProjectTasks);
    },
    hasRemoteRunnerNodes() {
      if (!this.details) return false;
      return (this.details.nodes || []).some(
        (n) => n.task && n.task.used_runner_id != null,
      );
    },
    // node.id -> raw run status, used by the graph for color + active animation.
    nodeStatuses() {
      const map = {};
      (this.details?.nodes || []).forEach((n) => {
        if (n.task) map[n.node.id] = n.task.status;
        else if (n.approval) map[n.node.id] = n.approval.status;
        else if (n.delay) map[n.node.id] = n.delay.status;
      });
      return map;
    },
    // node.id -> resume_at, for delay nodes currently waiting — lets the graph
    // render a live countdown between polls instead of a static duration.
    nodeDelays() {
      const map = {};
      (this.details?.nodes || []).forEach((n) => {
        if (n.delay && n.delay.status === 'waiting') map[n.node.id] = n.delay.resume_at;
      });
      return map;
    },
    pendingApprovals() {
      return (this.details?.nodes || [])
        .filter((n) => n.approval && n.approval.status === 'pending')
        .map((n) => ({ nodeId: n.node.id, message: n.node.approval_message }));
    },
  },
  async created() {
    await this.loadData();
    this.pollHandle = setInterval(() => {
      const status = this.details && this.details.run.status;
      if (status === 'running' || status === 'approval') {
        this.loadData();
      } else if (this.pollHandle) {
        clearInterval(this.pollHandle);
        this.pollHandle = null;
      }
    }, 5000);
  },
  beforeDestroy() {
    if (this.pollHandle) clearInterval(this.pollHandle);
  },
  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },
    // Clicking a node that has a task (running or finished) opens its task log.
    onNodeClicked(nodeId) {
      if (nodeId == null) return;
      const entry = (this.details?.nodes || []).find((n) => n.node.id === nodeId);
      if (entry && entry.task) {
        EventBus.$emit('i-show-task', { taskId: entry.task.id });
      }
    },
    statusColor(status) {
      switch (status) {
        case 'success':
        case 'approved':
          return 'success';
        case 'failed':
        case 'error':
        case 'stopped':
        case 'rejected':
          return 'error';
        case 'running':
        case 'pending':
          return 'primary';
        case 'approval':
          return 'warning';
        default:
          return 'grey';
      }
    },
    zoomIn() {
      if (this.$refs.graph) this.$refs.graph.zoomIn();
    },
    zoomOut() {
      if (this.$refs.graph) this.$refs.graph.zoomOut();
    },
    zoomReset() {
      if (this.$refs.graph) this.$refs.graph.zoomReset();
    },
    async stopRun() {
      this.stopping = true;
      try {
        await axios.post(
          `/api/project/${this.projectId}/workflows/${this.workflowId}/runs/${this.runId}/stop`,
        );
        await this.loadData();
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.stopping = false;
      }
    },
    async resolveApproval(nodeId, status) {
      try {
        await axios.post(
          `/api/project/${this.projectId}/workflows/${this.workflowId}/runs/${this.runId}/approvals/${nodeId}`,
          { status },
        );
        await this.loadData();
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },
    async loadData() {
      try {
        const [details, workflow, templates] = await Promise.all([
          axios.get(
            `/api/project/${this.projectId}/workflows/${this.workflowId}/runs/${this.runId}`,
          ),
          axios.get(`/api/project/${this.projectId}/workflows/${this.workflowId}`),
          axios.get(`/api/project/${this.projectId}/templates`),
        ]);
        this.details = details.data;
        this.workflow = workflow.data;
        this.templates = templates.data || [];
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },
  },
};
</script>
