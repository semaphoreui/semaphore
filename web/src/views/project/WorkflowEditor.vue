<template>
  <div class="workflow-editor">
    <v-toolbar flat>
      <v-btn icon @click="$router.back()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ workflow.name }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn text @click="showRuns">
        <v-icon left>mdi-history</v-icon>
        Runs
      </v-btn>
      <v-btn color="success" @click="runWorkflow">
        <v-icon left>mdi-play</v-icon>
        Run
      </v-btn>
      <v-btn color="primary" @click="saveWorkflow">
        <v-icon left>mdi-content-save</v-icon>
        Save
      </v-btn>
    </v-toolbar>

    <div class="editor-container">
      <!-- Toolbar -->
      <div class="node-toolbar">
        <v-card>
          <v-card-title class="text-subtitle-2">Add Nodes</v-card-title>
          <v-list dense>
            <v-list-item @click="addNode('task')">
              <v-list-item-icon>
                <v-icon>mdi-play-box</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>Task</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="addNode('pause')">
              <v-list-item-icon>
                <v-icon>mdi-pause</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>Pause</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="addNode('approval')">
              <v-list-item-icon>
                <v-icon>mdi-check-circle</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>Approval</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-card>
      </div>

      <!-- Canvas -->
      <div class="workflow-canvas" ref="canvas" @click="canvasClick">
        <svg
          :width="canvasWidth"
          :height="canvasHeight"
          @mousedown="startConnection"
          @mousemove="updateConnection"
          @mouseup="endConnection"
        >
          <!-- Links -->
          <g v-for="link in workflow.links" :key="`link-${link.id || link.temp_id}`">
            <line
              :x1="getLinkStart(link).x"
              :y1="getLinkStart(link).y"
              :x2="getLinkEnd(link).x"
              :y2="getLinkEnd(link).y"
              :stroke="getLinkColor(link.condition)"
              stroke-width="2"
              marker-end="url(#arrowhead)"
            />
            <circle
              :cx="(getLinkStart(link).x + getLinkEnd(link).x) / 2"
              :cy="(getLinkStart(link).y + getLinkEnd(link).y) / 2"
              r="12"
              :fill="getLinkColor(link.condition)"
              @click.stop="editLink(link)"
              class="link-edit-handle"
            />
          </g>

          <!-- Temporary connection line -->
          <line
            v-if="connectingFrom"
            :x1="connectingFrom.x + 50"
            :y1="connectingFrom.y + 25"
            :x2="mouseX"
            :y2="mouseY"
            stroke="#999"
            stroke-width="2"
            stroke-dasharray="5,5"
          />

          <!-- Arrow marker -->
          <defs>
            <marker
              id="arrowhead"
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

        <!-- Nodes -->
        <div
          v-for="node in workflow.nodes"
          :key="node.id || node.temp_id"
          class="workflow-node"
          :class="{ selected: selectedNode === node }"
          :style="{
            left: node.position_x + 'px',
            top: node.position_y + 'px',
          }"
          @mousedown.stop="startDrag(node, $event)"
          @click.stop="selectNode(node)"
        >
          <div class="node-header" :class="`node-type-${node.type}`">
            <v-icon small color="white">{{ getNodeIcon(node.type) }}</v-icon>
            <span class="node-name">{{ node.name }}</span>
            <v-btn
              icon
              x-small
              color="white"
              @click.stop="deleteNode(node)"
            >
              <v-icon x-small>mdi-close</v-icon>
            </v-btn>
          </div>
          <div class="node-content">
            <div v-if="node.type === 'task' && node.task_template_id">
              <small>Template: {{ getTemplateName(node.task_template_id) }}</small>
            </div>
            <div v-else-if="node.type === 'task'">
              <small class="error--text">No template selected</small>
            </div>
          </div>
          <div class="node-connectors">
            <div
              class="connector input"
              @mouseup.stop="connectNode(node, 'input')"
            ></div>
            <div
              class="connector output"
              @mousedown.stop="startConnectionFrom(node)"
            ></div>
          </div>
        </div>
      </div>

      <!-- Properties Panel -->
      <div class="properties-panel" v-if="selectedNode">
        <v-card>
          <v-card-title class="text-subtitle-2">Node Properties</v-card-title>
          <v-card-text>
            <v-text-field
              v-model="selectedNode.name"
              label="Name"
              dense
            ></v-text-field>

            <v-select
              v-if="selectedNode.type === 'task'"
              v-model="selectedNode.task_template_id"
              :items="templates"
              item-text="name"
              item-value="id"
              label="Template"
              dense
            ></v-select>

            <v-text-field
              v-if="selectedNode.type === 'pause'"
              v-model.number="pauseDuration"
              label="Duration (seconds)"
              type="number"
              dense
            ></v-text-field>
          </v-card-text>
        </v-card>
      </div>
    </div>

    <!-- Link Edit Dialog -->
    <v-dialog v-model="linkDialog" max-width="400">
      <v-card v-if="editingLink">
        <v-card-title>Edit Connection</v-card-title>
        <v-card-text>
          <v-select
            v-model="editingLink.condition"
            :items="linkConditions"
            label="Condition"
          ></v-select>
        </v-card-text>
        <v-card-actions>
          <v-btn text color="error" @click="deleteLink">Delete</v-btn>
          <v-spacer></v-spacer>
          <v-btn text @click="linkDialog = false">Close</v-btn>
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
        id: null,
        name: '',
        description: '',
        nodes: [],
        links: [],
      },
      templates: [],
      selectedNode: null,
      draggingNode: null,
      dragOffset: { x: 0, y: 0 },
      connectingFrom: null,
      mouseX: 0,
      mouseY: 0,
      canvasWidth: 2000,
      canvasHeight: 1500,
      nodeCounter: 0,
      linkCounter: 0,
      linkDialog: false,
      editingLink: null,
      linkConditions: [
        { text: 'On Success', value: 'success' },
        { text: 'On Failure', value: 'failure' },
        { text: 'Always', value: 'always' },
      ],
    };
  },
  mounted() {
    this.loadWorkflow();
    this.loadTemplates();
  },
  methods: {
    async loadWorkflow() {
      try {
        const { data } = await axios.get(
          `/api/project/${this.projectId}/workflows/${this.workflowId}`,
        );
        this.workflow = data;
        if (!this.workflow.nodes) this.workflow.nodes = [];
        if (!this.workflow.links) this.workflow.links = [];
      } catch (error) {
        console.error('Failed to load workflow:', error);
      }
    },
    async loadTemplates() {
      try {
        const { data } = await axios.get(`/api/project/${this.projectId}/templates`);
        this.templates = data || [];
      } catch (error) {
        console.error('Failed to load templates:', error);
      }
    },
    addNode(type) {
      const node = {
        temp_id: `temp-${++this.nodeCounter}`,
        type,
        name: `${type} ${this.workflow.nodes.length + 1}`,
        position_x: 100 + this.workflow.nodes.length * 50,
        position_y: 100 + this.workflow.nodes.length * 50,
        task_template_id: null,
        config: null,
      };
      this.workflow.nodes.push(node);
      this.selectedNode = node;
    },
    selectNode(node) {
      this.selectedNode = node;
    },
    startDrag(node, event) {
      this.draggingNode = node;
      const rect = event.target.getBoundingClientRect();
      this.dragOffset = {
        x: event.clientX - rect.left,
        y: event.clientY - rect.top,
      };
      document.addEventListener('mousemove', this.drag);
      document.addEventListener('mouseup', this.stopDrag);
    },
    drag(event) {
      if (!this.draggingNode) return;
      const canvas = this.$refs.canvas.getBoundingClientRect();
      this.draggingNode.position_x = event.clientX - canvas.left - this.dragOffset.x;
      this.draggingNode.position_y = event.clientY - canvas.top - this.dragOffset.y;
    },
    stopDrag() {
      this.draggingNode = null;
      document.removeEventListener('mousemove', this.drag);
      document.removeEventListener('mouseup', this.stopDrag);
    },
    startConnectionFrom(node) {
      this.connectingFrom = {
        node,
        x: node.position_x,
        y: node.position_y,
      };
    },
    updateConnection(event) {
      if (!this.connectingFrom) return;
      const canvas = this.$refs.canvas.getBoundingClientRect();
      this.mouseX = event.clientX - canvas.left;
      this.mouseY = event.clientY - canvas.top;
    },
    connectNode(toNode, type) {
      if (!this.connectingFrom || this.connectingFrom.node === toNode) return;
      
      const link = {
        temp_id: `temp-link-${++this.linkCounter}`,
        from_node_id: this.getNodeId(this.connectingFrom.node),
        to_node_id: this.getNodeId(toNode),
        condition: 'success',
      };
      
      this.workflow.links.push(link);
      this.connectingFrom = null;
    },
    endConnection() {
      this.connectingFrom = null;
    },
    deleteNode(node) {
      const index = this.workflow.nodes.indexOf(node);
      if (index > -1) {
        this.workflow.nodes.splice(index, 1);
        // Remove connected links
        const nodeId = this.getNodeId(node);
        this.workflow.links = this.workflow.links.filter(
          (l) => l.from_node_id !== nodeId && l.to_node_id !== nodeId,
        );
        if (this.selectedNode === node) {
          this.selectedNode = null;
        }
      }
    },
    editLink(link) {
      this.editingLink = link;
      this.linkDialog = true;
    },
    deleteLink() {
      const index = this.workflow.links.indexOf(this.editingLink);
      if (index > -1) {
        this.workflow.links.splice(index, 1);
      }
      this.linkDialog = false;
      this.editingLink = null;
    },
    getLinkStart(link) {
      const node = this.workflow.nodes.find((n) => this.getNodeId(n) === link.from_node_id);
      if (!node) return { x: 0, y: 0 };
      return { x: node.position_x + 100, y: node.position_y + 40 };
    },
    getLinkEnd(link) {
      const node = this.workflow.nodes.find((n) => this.getNodeId(n) === link.to_node_id);
      if (!node) return { x: 0, y: 0 };
      return { x: node.position_x, y: node.position_y + 40 };
    },
    getLinkColor(condition) {
      switch (condition) {
        case 'success': return '#4caf50';
        case 'failure': return '#f44336';
        case 'always': return '#2196f3';
        default: return '#666';
      }
    },
    getNodeId(node) {
      return node.id || node.temp_id;
    },
    getNodeIcon(type) {
      switch (type) {
        case 'task': return 'mdi-play-box';
        case 'pause': return 'mdi-pause';
        case 'approval': return 'mdi-check-circle';
        default: return 'mdi-help';
      }
    },
    getTemplateName(templateId) {
      const template = this.templates.find((t) => t.id === templateId);
      return template ? template.name : 'Unknown';
    },
    canvasClick() {
      this.selectedNode = null;
    },
    async saveWorkflow() {
      try {
        // Prepare nodes and links for saving
        const nodesToSave = this.workflow.nodes.map((node) => ({
          ...node,
          config: node.type === 'pause' ? JSON.stringify({ duration: this.pauseDuration || 5 }) : node.config,
        }));

        const linksToSave = this.workflow.links.map((link) => ({
          ...link,
          // Remove temp IDs if they exist
          id: link.id || undefined,
        }));

        await axios.put(
          `/api/project/${this.projectId}/workflows/${this.workflowId}`,
          {
            ...this.workflow,
            nodes: nodesToSave,
            links: linksToSave,
          },
        );
        
        this.$emit('snackbar', 'Workflow saved successfully');
        this.loadWorkflow(); // Reload to get proper IDs
      } catch (error) {
        console.error('Failed to save workflow:', error);
        this.$emit('snackbar', 'Failed to save workflow');
      }
    },
    async runWorkflow() {
      try {
        await axios.post(`/api/project/${this.projectId}/workflows/${this.workflowId}/run`);
        this.$router.push(`/project/${this.projectId}/workflows/${this.workflowId}/runs`);
      } catch (error) {
        console.error('Failed to run workflow:', error);
      }
    },
    showRuns() {
      this.$router.push(`/project/${this.projectId}/workflows/${this.workflowId}/runs`);
    },
  },
  computed: {
    projectId() {
      return this.$route.params.projectId;
    },
    workflowId() {
      return this.$route.params.workflowId;
    },
    pauseDuration: {
      get() {
        if (!this.selectedNode || !this.selectedNode.config) return 5;
        try {
          const config = JSON.parse(this.selectedNode.config);
          return config.duration || 5;
        } catch {
          return 5;
        }
      },
      set(value) {
        if (this.selectedNode) {
          this.selectedNode.config = JSON.stringify({ duration: value });
        }
      },
    },
  },
};
</script>

<style scoped>
.workflow-editor {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.editor-container {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.node-toolbar {
  width: 200px;
  border-right: 1px solid #ddd;
  overflow-y: auto;
  background: #fafafa;
}

.workflow-canvas {
  flex: 1;
  position: relative;
  overflow: auto;
  background: #f5f5f5;
  background-image: 
    linear-gradient(0deg, #e0e0e0 1px, transparent 1px),
    linear-gradient(90deg, #e0e0e0 1px, transparent 1px);
  background-size: 20px 20px;
}

.workflow-node {
  position: absolute;
  width: 180px;
  background: white;
  border: 2px solid #ddd;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  cursor: move;
  user-select: none;
}

.workflow-node.selected {
  border-color: #2196f3;
  box-shadow: 0 4px 8px rgba(33, 150, 243, 0.3);
}

.node-header {
  padding: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: white;
  border-radius: 6px 6px 0 0;
}

.node-type-task {
  background: #4caf50;
}

.node-type-pause {
  background: #ff9800;
}

.node-type-approval {
  background: #2196f3;
}

.node-name {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.node-content {
  padding: 8px;
  min-height: 40px;
}

.node-connectors {
  position: relative;
}

.connector {
  position: absolute;
  width: 12px;
  height: 12px;
  background: white;
  border: 2px solid #666;
  border-radius: 50%;
  cursor: pointer;
}

.connector.input {
  left: -6px;
  top: -46px;
}

.connector.output {
  right: -6px;
  top: -46px;
}

.connector:hover {
  background: #2196f3;
  border-color: #2196f3;
}

.properties-panel {
  width: 300px;
  border-left: 1px solid #ddd;
  overflow-y: auto;
  background: #fafafa;
}

.link-edit-handle {
  cursor: pointer;
  opacity: 0.7;
}

.link-edit-handle:hover {
  opacity: 1;
}
</style>
