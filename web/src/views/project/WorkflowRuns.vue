<template>
  <div>
    <v-toolbar flat>
      <v-btn icon @click="$router.back()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ workflow.name }} - Runs</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn color="primary" @click="runWorkflow">
        <v-icon left>mdi-play</v-icon>
        New Run
      </v-btn>
    </v-toolbar>

    <v-container fluid>
      <v-data-table
        :headers="headers"
        :items="runs"
        :loading="loading"
        @click:row="viewRun"
        class="clickable-rows"
      >
        <template v-slot:item.status="{ item }">
          <v-chip
            small
            :color="getStatusColor(item.status)"
            dark
          >
            {{ item.status }}
          </v-chip>
        </template>

        <template v-slot:item.start="{ item }">
          {{ formatDateTime(item.start) }}
        </template>

        <template v-slot:item.end="{ item }">
          {{ item.end ? formatDateTime(item.end) : '-' }}
        </template>

        <template v-slot:item.duration="{ item }">
          {{ getDuration(item) }}
        </template>

        <template v-slot:item.actions="{ item }">
          <v-btn
            v-if="item.status === 'running' || item.status === 'pending'"
            icon
            @click.stop="stopRun(item)"
          >
            <v-icon>mdi-stop</v-icon>
          </v-btn>
        </template>
      </v-data-table>
    </v-container>

    <!-- Run Details Dialog -->
    <v-dialog v-model="runDialog" max-width="900">
      <v-card v-if="selectedRun">
        <v-card-title>
          Run #{{ selectedRun.run.id }}
          <v-spacer></v-spacer>
          <v-chip
            small
            :color="getStatusColor(selectedRun.run.status)"
            dark
          >
            {{ selectedRun.run.status }}
          </v-chip>
        </v-card-title>

        <v-card-text>
          <div class="mb-4">
            <strong>Started:</strong> {{ formatDateTime(selectedRun.run.start) }}<br>
            <strong>Ended:</strong>
            {{ selectedRun.run.end ? formatDateTime(selectedRun.run.end) : 'Running...' }}<br>
            <strong>Duration:</strong> {{ getDuration(selectedRun.run) }}
          </div>

          <v-divider class="mb-4"></v-divider>

          <h3 class="mb-3">Node Execution</h3>

          <!-- Visual workflow with status -->
          <div class="workflow-visualization">
            <svg width="100%" height="400" ref="runCanvas">
              <!-- Links -->
              <g v-for="link in workflow.links" :key="`link-${link.id}`">
                <line
                  :x1="getNodePosition(link.from_node_id).x + 90"
                  :y1="getNodePosition(link.from_node_id).y + 30"
                  :x2="getNodePosition(link.to_node_id).x + 10"
                  :y2="getNodePosition(link.to_node_id).y + 30"
                  :stroke="getLinkColor(link.condition)"
                  stroke-width="2"
                  marker-end="url(#arrowhead-dialog)"
                />
              </g>

              <!-- Nodes -->
              <g
                v-for="node in workflow.nodes"
                :key="`node-${node.id}`"
              >
                <rect
                  :x="node.position_x * 0.4"
                  :y="node.position_y * 0.4"
                  width="100"
                  height="60"
                  :fill="getNodeStatusColor(node)"
                  :stroke="getNodeStatusStroke(node)"
                  stroke-width="3"
                  rx="5"
                />
                <text
                  :x="node.position_x * 0.4 + 50"
                  :y="node.position_y * 0.4 + 35"
                  text-anchor="middle"
                  fill="white"
                  font-size="12"
                  font-weight="bold"
                >
                  {{ node.name }}
                </text>
              </g>

              <defs>
                <marker
                  id="arrowhead-dialog"
                  markerWidth="10"
                  markerHeight="10"
                  refX="9"
                  refY="3"
                  orient="auto"
                >
                  <polygon points="0 0, 10 3, 0 6" fill="#666" />
                </marker>
              </defs>
            </svg>
          </div>

          <v-divider class="my-4"></v-divider>

          <h3 class="mb-3">Node Details</h3>
          <v-simple-table>
            <thead>
              <tr>
                <th>Node</th>
                <th>Status</th>
                <th>Start</th>
                <th>End</th>
                <th>Task</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="nodeRun in selectedRun.node_runs" :key="nodeRun.id">
                <td>{{ getNodeName(nodeRun.node_id) }}</td>
                <td>
                  <v-chip
                    x-small
                    :color="getStatusColor(nodeRun.status)"
                    dark
                  >
                    {{ nodeRun.status }}
                  </v-chip>
                </td>
                <td>{{ formatDateTime(nodeRun.start) }}</td>
                <td>{{ nodeRun.end ? formatDateTime(nodeRun.end) : '-' }}</td>
                <td>
                  <v-btn
                    v-if="nodeRun.task_id"
                    text
                    small
                    :to="`/project/${projectId}/history?task=${nodeRun.task_id}`"
                  >
                    View Task #{{ nodeRun.task_id }}
                  </v-btn>
                  <span v-else>-</span>
                </td>
              </tr>
            </tbody>
          </v-simple-table>
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn text @click="runDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  data() {
    return {
      workflow: {
        name: '',
        nodes: [],
        links: [],
      },
      runs: [],
      loading: false,
      runDialog: false,
      selectedRun: null,
      refreshInterval: null,
      headers: [
        { text: 'ID', value: 'id' },
        { text: 'Status', value: 'status' },
        { text: 'Started', value: 'start' },
        { text: 'Ended', value: 'end' },
        { text: 'Duration', value: 'duration' },
        { text: 'Actions', value: 'actions', sortable: false },
      ],
    };
  },
  mounted() {
    this.loadWorkflow();
    this.loadRuns();
    this.startAutoRefresh();
  },
  beforeDestroy() {
    this.stopAutoRefresh();
  },
  methods: {
    async loadWorkflow() {
      try {
        const { data } = await axios.get(
          `/api/project/${this.projectId}/workflows/${this.workflowId}`,
        );
        this.workflow = data;
      } catch (error) {
        console.error('Failed to load workflow:', error);
      }
    },
    async loadRuns() {
      this.loading = true;
      try {
        const { data } = await axios.get(
          `/api/project/${this.projectId}/workflows/${this.workflowId}/runs`,
        );
        this.runs = data || [];
      } catch (error) {
        console.error('Failed to load runs:', error);
      } finally {
        this.loading = false;
      }
    },
    async viewRun(run) {
      try {
        const { data } = await axios.get(
          `/api/project/${this.projectId}/workflows/${this.workflowId}/runs/${run.id}`,
        );
        this.selectedRun = data;
        this.runDialog = true;
      } catch (error) {
        console.error('Failed to load run details:', error);
      }
    },
    async runWorkflow() {
      try {
        await axios.post(`/api/project/${this.projectId}/workflows/${this.workflowId}/run`);
        this.loadRuns();
      } catch (error) {
        console.error('Failed to run workflow:', error);
      }
    },
    async stopRun(run) {
      try {
        await axios.post(
          `/api/project/${this.projectId}/workflows/${this.workflowId}/runs/${run.id}/stop`,
        );
        this.loadRuns();
      } catch (error) {
        console.error('Failed to stop run:', error);
      }
    },
    getStatusColor(status) {
      switch (status) {
        case 'success': return 'green';
        case 'running': return 'blue';
        case 'pending': return 'orange';
        case 'failure': return 'red';
        case 'stopped': return 'grey';
        default: return 'grey';
      }
    },
    getNodeStatusColor(node) {
      if (!this.selectedRun) return '#cccccc';
      const nodeRun = this.selectedRun.node_runs.find((nr) => nr.node_id === node.id);
      if (!nodeRun) return '#cccccc';

      switch (nodeRun.status) {
        case 'success': return '#4caf50';
        case 'running': return '#2196f3';
        case 'pending': return '#ff9800';
        case 'failure': return '#f44336';
        case 'stopped': return '#9e9e9e';
        case 'skipped': return '#757575';
        default: return '#cccccc';
      }
    },
    getNodeStatusStroke(node) {
      if (!this.selectedRun) return '#999';
      const nodeRun = this.selectedRun.node_runs.find((nr) => nr.node_id === node.id);
      if (!nodeRun) return '#999';
      return nodeRun.status === 'running' ? '#fff' : '#333';
    },
    getNodePosition(nodeId) {
      const node = this.workflow.nodes.find((n) => n.id === nodeId);
      if (!node) return { x: 0, y: 0 };
      return { x: node.position_x * 0.4, y: node.position_y * 0.4 };
    },
    getLinkColor(condition) {
      switch (condition) {
        case 'success': return '#4caf50';
        case 'failure': return '#f44336';
        case 'always': return '#2196f3';
        default: return '#666';
      }
    },
    getNodeName(nodeId) {
      const node = this.workflow.nodes.find((n) => n.id === nodeId);
      return node ? node.name : `Node ${nodeId}`;
    },
    formatDateTime(datetime) {
      if (!datetime) return '-';
      return new Date(datetime).toLocaleString();
    },
    getDuration(run) {
      if (!run.start) return '-';
      const start = new Date(run.start);
      const end = run.end ? new Date(run.end) : new Date();
      const diff = Math.floor((end - start) / 1000);
      const minutes = Math.floor(diff / 60);
      const seconds = diff % 60;
      return `${minutes}m ${seconds}s`;
    },
    startAutoRefresh() {
      this.refreshInterval = setInterval(() => {
        this.loadRuns();
        if (this.runDialog && this.selectedRun) {
          this.viewRun({ id: this.selectedRun.run.id });
        }
      }, 3000);
    },
    stopAutoRefresh() {
      if (this.refreshInterval) {
        clearInterval(this.refreshInterval);
      }
    },
  },
  computed: {
    projectId() {
      return this.$route.params.projectId;
    },
    workflowId() {
      return this.$route.params.workflowId;
    },
  },
};
</script>

<style scoped>
.clickable-rows tbody tr {
  cursor: pointer;
}

.clickable-rows tbody tr:hover {
  background-color: #f5f5f5;
}

.workflow-visualization {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 16px;
  background: #fafafa;
}
</style>
