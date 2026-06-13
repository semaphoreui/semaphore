<template>
  <div v-if="!isLoaded">
    <v-progress-linear
      indeterminate
      color="primary darken-2"
    ></v-progress-linear>
  </div>
  <div v-else>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        {{ workflow.name }} - {{ $t('run') }} #{{ run.id }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-chip
        :color="getStatusColor(run.status)"
        text-color="white"
        class="mr-2"
      >
        {{ run.status }}
      </v-chip>
    </v-toolbar>

    <div class="workflow-run-view">
      <div class="workflow-canvas-container">
        <div class="workflow-canvas" ref="canvas">
          <!-- SVG overlay for connections -->
          <svg class="connections-layer" :width="canvasWidth" :height="canvasHeight">
            <line
              v-for="link in links"
              :key="link.id"
              :x1="getNodeX(link.from_node_id)"
              :y1="getNodeY(link.from_node_id)"
              :x2="getNodeX(link.to_node_id)"
              :y2="getNodeY(link.to_node_id)"
              :stroke="getLinkColor(link)"
              stroke-width="2"
              marker-end="url(#arrowhead)"
            />
            <defs>
              <marker
                id="arrowhead"
                markerWidth="10"
                markerHeight="10"
                refX="9"
                refY="3"
                orient="auto"
              >
                <polygon points="0 0, 10 3, 0 6" fill="#1976d2" />
              </marker>
            </defs>
          </svg>

          <!-- Nodes -->
          <div
            v-for="node in nodes"
            :key="node.id"
            class="workflow-node"
            :class="getNodeStatusClass(node)"
            :style="{
              left: node.position_x + 'px',
              top: node.position_y + 'px',
            }"
            @click="selectNode(node)"
          >
            <div class="node-header">
              <v-icon small>{{ getNodeIcon(node.type) }}</v-icon>
              <span class="node-title">{{ getNodeTitle(node) }}</span>
              <v-chip
                x-small
                :color="getNodeStatusColor(node)"
                text-color="white"
                class="ml-1"
              >
                {{ getNodeStatus(node) }}
              </v-chip>
            </div>
          </div>
        </div>
      </div>

      <!-- Node Details Sidebar -->
      <v-navigation-drawer
        v-model="sidebarOpen"
        right
        temporary
        width="400"
      >
        <v-toolbar flat>
          <v-toolbar-title>{{ $t('nodeDetails') }}</v-toolbar-title>
          <v-spacer></v-spacer>
          <v-btn icon @click="sidebarOpen = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar>
        <v-card-text v-if="selectedRunNode">
          <div class="mb-4">
            <strong>{{ $t('status') }}:</strong>
            <v-chip
              :color="getNodeStatusColor(selectedRunNode)"
              text-color="white"
              class="ml-2"
            >
              {{ selectedRunNode.status }}
            </v-chip>
          </div>
          <div v-if="selectedRunNode.task_id" class="mb-4">
            <strong>{{ $t('task') }}:</strong>
            <router-link
              :to="`/project/${projectId}/tasks/${selectedRunNode.task_id}`"
              class="ml-2"
            >
              #{{ selectedRunNode.task_id }}
            </router-link>
          </div>
          <div v-if="selectedRunNode.message" class="mb-4">
            <strong>{{ $t('message') }}:</strong>
            <div class="mt-1">{{ selectedRunNode.message }}</div>
          </div>
          <div v-if="selectedRunNode.start" class="mb-2">
            <strong>{{ $t('started') }}:</strong>
            <span class="ml-2">{{ formatDate(selectedRunNode.start) }}</span>
          </div>
          <div v-if="selectedRunNode.end" class="mb-2">
            <strong>{{ $t('ended') }}:</strong>
            <span class="ml-2">{{ formatDate(selectedRunNode.end) }}</span>
          </div>
        </v-card-text>
      </v-navigation-drawer>
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex';
import api from '@/lib/api';
import dayjs from 'dayjs';

export default {
  name: 'WorkflowRun',
  mixins: [
    require('@/components/mixins/DrawerMixin').default,
  ],
  data() {
    return {
      isLoaded: false,
      workflow: {},
      run: {},
      nodes: [],
      links: [],
      runNodes: [],
      templates: [],
      selectedRunNode: null,
      sidebarOpen: false,
      canvasWidth: 2000,
      canvasHeight: 2000,
      pollInterval: null,
    };
  },
  computed: {
    ...mapGetters(['projectId']),
    workflowId() {
      return parseInt(this.$route.params.workflowId);
    },
    runId() {
      return parseInt(this.$route.params.runId);
    },
  },
  mounted() {
    this.loadWorkflowRun();
    // Poll for updates if workflow is still running
    this.pollInterval = setInterval(() => {
      if (this.run.status === 'running' || this.run.status === 'pending') {
        this.loadWorkflowRun();
      }
    }, 2000);
  },
  beforeDestroy() {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
    }
  },
  methods: {
    async loadWorkflowRun() {
      try {
        this.isLoaded = false;
        const { data: workflowData } = await api.get(`/project/${this.projectId}/workflows/${this.workflowId}`);
        this.workflow = workflowData;
        this.nodes = workflowData.nodes || [];
        this.links = workflowData.links || [];

        const { data: runData } = await api.get(`/project/${this.projectId}/workflows/${this.workflowId}/runs/${this.runId}`);
        this.run = runData;
        this.runNodes = runData.nodes || [];

        // Load templates for node titles
        const { data: templatesData } = await api.get(`/project/${this.projectId}/templates`);
        this.templates = templatesData;
      } catch (err) {
        this.$store.dispatch('setError', err);
      } finally {
        this.isLoaded = true;
      }
    },
    selectNode(node) {
      const runNode = this.runNodes.find((rn) => rn.workflow_node_id === node.id);
      if (runNode) {
        this.selectedRunNode = runNode;
        this.sidebarOpen = true;
      }
    },
    getNodeStatus(node) {
      const runNode = this.runNodes.find((rn) => rn.workflow_node_id === node.id);
      return runNode ? runNode.status : 'pending';
    },
    getNodeStatusClass(node) {
      const status = this.getNodeStatus(node);
      return {
        'node-status-pending': status === 'pending',
        'node-status-running': status === 'running',
        'node-status-success': status === 'success',
        'node-status-error': status === 'error',
        'node-status-skipped': status === 'skipped',
      };
    },
    getNodeStatusColor(node) {
      const status = this.getNodeStatus(node);
      switch (status) {
        case 'running':
          return 'blue';
        case 'success':
          return 'green';
        case 'error':
          return 'red';
        case 'skipped':
          return 'grey';
        default:
          return 'grey';
      }
    },
    getStatusColor(status) {
      switch (status) {
        case 'running':
          return 'blue';
        case 'success':
          return 'green';
        case 'error':
          return 'red';
        case 'pending':
          return 'orange';
        default:
          return 'grey';
      }
    },
    getLinkColor(link) {
      const fromStatus = this.getNodeStatus(this.nodes.find((n) => n.id === link.from_node_id));
      if (fromStatus === 'success' || fromStatus === 'error') {
        return '#4caf50';
      }
      return '#1976d2';
    },
    getNodeX(nodeId) {
      const node = this.nodes.find((n) => n.id === nodeId);
      return node ? node.position_x + 75 : 0;
    },
    getNodeY(nodeId) {
      const node = this.nodes.find((n) => n.id === nodeId);
      return node ? node.position_y + 25 : 0;
    },
    getNodeIcon(type) {
      switch (type) {
        case 'task':
          return 'mdi-play';
        case 'pause':
          return 'mdi-pause';
        case 'approval':
          return 'mdi-check-circle';
        default:
          return 'mdi-circle';
      }
    },
    getNodeTitle(node) {
      if (node.type === 'task' && node.task_id) {
        const template = this.templates.find((t) => t.id === node.task_id);
        return template ? template.name : this.$t('task');
      }
      return this.$t(node.type);
    },
    formatDate(date) {
      return dayjs(date).format('YYYY-MM-DD HH:mm:ss');
    },
  },
};
</script>

<style scoped>
.workflow-run-view {
  height: calc(100vh - 64px);
  display: flex;
  flex-direction: column;
}

.workflow-canvas-container {
  flex: 1;
  position: relative;
  overflow: auto;
}

.workflow-canvas {
  position: relative;
  width: 2000px;
  height: 2000px;
  background: #f5f5f5;
}

.connections-layer {
  position: absolute;
  top: 0;
  left: 0;
  pointer-events: none;
}

.workflow-node {
  position: absolute;
  width: 150px;
  min-height: 50px;
  background: white;
  border: 2px solid #1976d2;
  border-radius: 4px;
  cursor: pointer;
  z-index: 10;
}

.node-status-pending {
  border-color: #9e9e9e;
  opacity: 0.6;
}

.node-status-running {
  border-color: #2196f3;
  animation: pulse 2s infinite;
}

.node-status-success {
  border-color: #4caf50;
}

.node-status-error {
  border-color: #f44336;
}

.node-status-skipped {
  border-color: #9e9e9e;
  opacity: 0.4;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

.node-header {
  padding: 8px;
  display: flex;
  align-items: center;
  font-weight: 500;
}

.node-title {
  flex: 1;
  margin-left: 4px;
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
