<template>
  <div>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        <router-link :to="`/project/${projectId}/workflows`">
          {{ $t('workflows') }}
        </router-link>
        <span class="ml-2">/ {{ $t('workflowRun') }} #{{ runId }}</span>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn icon @click="loadData()" :title="$t('refresh')">
        <v-icon>mdi-refresh</v-icon>
      </v-btn>
    </v-toolbar>

    <v-divider />

    <div
      v-if="details != null"
      class="pa-4"
      style="max-width: calc(var(--breakpoint-lg) - var(--nav-drawer-width) - 200px); margin: auto"
    >
      <v-card class="mb-4 pa-4">
        <div class="d-flex align-center">
          <div>
            <div class="text-h6">{{ workflow ? workflow.name : '' }}</div>
            <div class="text-caption" v-if="workflow && workflow.description">
              {{ workflow.description }}
            </div>
          </div>
          <v-spacer />
          <v-chip :color="statusColor(details.run.status)" small>
            {{ details.run.status }}
          </v-chip>
        </div>
        <div class="text-caption mt-2">
          <span v-if="details.run.start">
            <strong>{{ $t('start') }}:</strong> {{ details.run.start }}
          </span>
          <span v-if="details.run.end" class="ml-3">
            <strong>{{ $t('end') }}:</strong> {{ details.run.end }}
          </span>
        </div>
      </v-card>

      <v-alert
        v-if="hasRemoteRunnerNodes"
        type="warning"
        dense
        text
        class="mb-4"
        icon="mdi-alert-outline"
      >
        {{ $t('workflowArtifactsRemoteRunnerWarning') }}
      </v-alert>

      <v-card class="mb-4" outlined>
        <WorkflowGraph
          v-if="workflow"
          :nodes="workflow.nodes || []"
          :edges="workflow.edges || []"
          :templates="templates"
          :node-statuses="nodeStatuses"
          :editable="false"
          style="height: 420px"
        />
      </v-card>

      <v-data-table
        :headers="headers"
        :items="details.nodes"
        item-key="node.id"
        hide-default-footer
        :items-per-page="Number.MAX_VALUE"
        :expanded.sync="expandedNodes"
        show-expand
        single-expand
      >
        <template v-slot:item.id="{ item }">
          #{{ item.node.id }}
        </template>
        <template v-slot:item.kind="{ item }">
          {{ nodeKindText(item.node.kind) }}
        </template>
        <template v-slot:item.template="{ item }">
          <router-link
            v-if="item.task"
            :to="`/project/${projectId}/templates/${item.task.template_id}`"
          >{{ item.task.tpl_alias || item.task.template_id }}</router-link>
          <span v-else-if="item.node.kind !== 'approval'">
            {{ templateName(item.node.template_id) }}
          </span>
          <span v-else>{{ item.node.approval_message || '—' }}</span>
        </template>
        <template v-slot:item.status="{ item }">
          <v-chip
            v-if="item.task || item.approval"
            :color="statusColor(nodeStatusRaw(item))"
            small
          >{{ nodeStatus(item) }}</v-chip>
          <span v-else class="text-caption text--secondary">{{ $t('workflowNodePending') }}</span>
        </template>
        <template v-slot:item.approval_actions="{ item }">
          <div
            v-if="item.approval && item.approval.status === 'pending' && canResolveApprovals"
            class="d-flex"
          >
            <v-btn
              x-small
              color="success"
              class="mr-2"
              @click="resolveApproval(item.node.id, 'approved')"
            >{{ $t('workflowApprove') }}</v-btn>
            <v-btn
              x-small
              color="error"
              @click="resolveApproval(item.node.id, 'rejected')"
            >{{ $t('workflowReject') }}</v-btn>
          </div>
          <span v-else>—</span>
        </template>
        <template v-slot:item.task="{ item }">
          <router-link
            v-if="item.task"
            :to="`/project/${projectId}/templates/${item.task.template_id}/tasks`"
          >#{{ item.task.id }}</router-link>
          <span v-else>—</span>
        </template>
        <template v-slot:item.data-table-expand="{ item, expand, isExpanded }">
          <v-btn
            v-if="taskArtifacts(item)"
            icon
            x-small
            :title="$t('workflowArtifacts')"
            @click="expand(!isExpanded)"
          >
            <v-icon>{{ isExpanded ? 'mdi-chevron-up' : 'mdi-chevron-down' }}</v-icon>
          </v-btn>
        </template>
        <template v-slot:expanded-item="{ headers: hs, item }">
          <td :colspan="hs.length" class="pa-3">
            <div class="text-caption mb-1">
              <strong>{{ $t('workflowArtifacts') }}</strong>
              <span class="ml-2 text--secondary">{{ $t('workflowArtifactsHint') }}</span>
            </div>
            <pre class="WorkflowRun__artifacts">{{ taskArtifacts(item) }}</pre>
          </td>
        </template>
      </v-data-table>

      <v-card
        v-if="formattedMergedArtifacts"
        class="mt-4 pa-4"
      >
        <div class="d-flex align-center mb-2">
          <div class="text-subtitle-1">
            <v-icon small class="mr-1">mdi-package-variant-closed</v-icon>
            {{ $t('workflowMergedArtifacts') }}
          </div>
          <v-spacer />
          <span class="text-caption text--secondary">
            {{ $t('workflowMergedArtifactsHint') }}
          </span>
        </div>
        <pre class="WorkflowRun__artifacts">{{ formattedMergedArtifacts }}</pre>
      </v-card>
    </div>

    <div v-else class="pa-4 text-center">
      <v-progress-circular indeterminate color="primary" />
    </div>
  </div>
</template>

<style lang="scss">
.WorkflowRun__artifacts {
  white-space: pre-wrap;
  word-break: break-word;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  margin: 0;
  background: rgba(133, 133, 133, 0.06);
  padding: 8px;
  border-radius: 4px;
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
      mergedArtifacts: null,
      expandedNodes: [],
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
    headers() {
      return [
        { text: this.$i18n.t('workflowNodeId'), value: 'id', sortable: false },
        { text: this.$i18n.t('workflowNodeKind'), value: 'kind', sortable: false },
        { text: this.$i18n.t('taskTemplate'), value: 'template', sortable: false },
        { text: this.$i18n.t('status'), value: 'status', sortable: false },
        { text: this.$i18n.t('workflowApprovalActions'), value: 'approval_actions', sortable: false },
        { text: this.$i18n.t('workflowTaskColumn'), value: 'task', sortable: false },
      ];
    },
    canResolveApprovals() {
      return this.can(USER_PERMISSIONS.runProjectTasks);
    },
    hasRemoteRunnerNodes() {
      if (!this.details) return false;
      return (this.details.nodes || []).some(
        (n) => n.task && n.task.used_runner_id != null,
      );
    },
    formattedMergedArtifacts() {
      if (!this.mergedArtifacts || typeof this.mergedArtifacts !== 'object'
          || Object.keys(this.mergedArtifacts).length === 0) {
        return '';
      }
      return JSON.stringify(this.mergedArtifacts, null, 2);
    },
    nodeStatuses() {
      const map = {};
      (this.details?.nodes || []).forEach((n) => {
        if (n.task) map[n.node.id] = n.task.status;
        else if (n.approval) map[n.node.id] = n.approval.status;
      });
      return map;
    },
  },
  async created() {
    await this.loadData();
    this.pollHandle = setInterval(() => {
      if (this.details && this.details.run.status === 'running') {
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
        default:
          return 'grey';
      }
    },
    templateName(id) {
      const t = this.templates.find((x) => x.id === id);
      return t ? t.name : `#${id}`;
    },
    nodeKindText(kind) {
      return kind === 'approval' ? this.$t('workflowNodeKindApproval') : this.$t('workflowNodeKindTask');
    },
    nodeStatus(item) {
      const rawStatus = this.nodeStatusRaw(item);
      if (rawStatus === 'pending') {
        return this.$t('workflowApprovalPending');
      }
      return rawStatus;
    },
    nodeStatusRaw(item) {
      if (item.task) {
        return item.task.status;
      }
      if (item.approval) {
        return item.approval.status;
      }
      return this.$t('workflowNodePending');
    },
    taskArtifacts(item) {
      const raw = item && item.task ? item.task.artifacts : null;
      if (raw == null || raw === '') return null;
      let parsed = raw;
      if (typeof raw === 'string') {
        try {
          parsed = JSON.parse(raw);
        } catch (e) {
          return null;
        }
      }
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)
          || Object.keys(parsed).length === 0) {
        return null;
      }
      return JSON.stringify(parsed, null, 2);
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
        const [details, workflow, templates, mergedArtifacts] = await Promise.all([
          axios.get(
            `/api/project/${this.projectId}/workflows/${this.workflowId}/runs/${this.runId}`,
          ),
          axios.get(`/api/project/${this.projectId}/workflows/${this.workflowId}`),
          axios.get(`/api/project/${this.projectId}/templates`),
          axios.get(
            `/api/project/${this.projectId}/workflows/${this.workflowId}/runs/${this.runId}/artifacts`,
          ).catch(() => ({ data: null })),
        ]);
        this.details = details.data;
        this.workflow = workflow.data;
        this.templates = templates.data || [];
        this.mergedArtifacts = mergedArtifacts.data || null;
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
