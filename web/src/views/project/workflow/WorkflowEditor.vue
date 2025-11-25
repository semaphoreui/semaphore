<template>
  <div v-if="!isLoaded">
    <v-progress-linear
      indeterminate
      color="primary darken-2"
    ></v-progress-linear>
  </div>
  <div v-else class="workflow-editor">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        {{ workflow.name }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        class="mr-1"
        @click="saveWorkflow()"
        :disabled="saving"
      >
        {{ $t('save') }}
      </v-btn>
      <v-btn
        color="success"
        class="mr-1"
        @click="runWorkflow()"
        :disabled="running"
      >
        {{ $t('run') }}
      </v-btn>
    </v-toolbar>

    <div class="workflow-canvas-container">
      <div class="workflow-canvas" ref="canvas" @click="handleCanvasClick">
        <!-- SVG overlay for connections -->
        <svg class="connections-layer" :width="canvasWidth" :height="canvasHeight">
          <line
            v-for="link in links"
            :key="link.id"
            :x1="getNodeX(link.from_node_id)"
            :y1="getNodeY(link.from_node_id)"
            :x2="getNodeX(link.to_node_id)"
            :y2="getNodeY(link.to_node_id)"
            stroke="#1976d2"
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
          :class="{ 'node-linking': linking && linkFromNode && linkFromNode.id === node.id }"
          :style="{
            left: node.position_x + 'px',
            top: node.position_y + 'px',
          }"
          @mousedown="startDrag(node, $event)"
          @click.stop="handleNodeClick(node)"
        >
          <div class="node-header" :class="getNodeTypeClass(node.type)">
            <v-icon small>{{ getNodeIcon(node.type) }}</v-icon>
            <span class="node-title">{{ getNodeTitle(node) }}</span>
            <v-btn
              icon
              x-small
              @click.stop="deleteNode(node)"
              class="ml-1"
            >
              <v-icon x-small>mdi-close</v-icon>
            </v-btn>
          </div>
        </div>
      </div>

      <!-- Sidebar -->
      <v-navigation-drawer
        v-model="sidebarOpen"
        right
        temporary
        width="300"
      >
        <v-toolbar flat>
          <v-toolbar-title>{{ $t('nodeConfiguration') }}</v-toolbar-title>
          <v-spacer></v-spacer>
          <v-btn icon @click="sidebarOpen = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar>
        <v-card-text v-if="selectedNode">
          <v-select
            v-model="selectedNode.type"
            :items="nodeTypes"
            :label="$t('nodeType')"
            @change="updateNodeType()"
          ></v-select>
          <v-select
            v-if="selectedNode.type === 'task'"
            v-model="selectedNode.task_id"
            :items="templates"
            item-text="name"
            item-value="id"
            :label="$t('template')"
            @change="updateNode()"
          ></v-select>
          <v-divider class="my-4"></v-divider>
          <v-btn
            color="primary"
            block
            @click="startLinking(selectedNode)"
          >
            {{ $t('createLink') }}
          </v-btn>
          <div v-if="links.filter(l => l.from_node_id === selectedNode.id || l.to_node_id === selectedNode.id).length > 0" class="mt-4">
            <div class="mb-2"><strong>{{ $t('links') }}:</strong></div>
            <div
              v-for="link in links.filter(l => l.from_node_id === selectedNode.id || l.to_node_id === selectedNode.id)"
              :key="link.id"
              class="mb-2"
            >
              <v-chip
                x-small
                @click="deleteLink(link)"
                close
              >
                {{ link.from_node_id === selectedNode.id ? $t('to') : $t('from') }}:
                {{ getNodeTitle(nodes.find(n => n.id === (link.from_node_id === selectedNode.id ? link.to_node_id : link.from_node_id))) }}
                ({{ link.condition }})
              </v-chip>
            </div>
          </div>
        </v-card-text>
      </v-navigation-drawer>
    </div>

    <!-- Add Node Menu -->
    <v-menu
      v-model="addNodeMenu"
      :position-x="menuX"
      :position-y="menuY"
      absolute
      offset-y
    >
      <v-list>
        <v-list-item @click="addNode('task')">
          <v-list-item-icon>
            <v-icon>mdi-play</v-icon>
          </v-list-item-icon>
          <v-list-item-title>{{ $t('taskNode') }}</v-list-item-title>
        </v-list-item>
        <v-list-item @click="addNode('pause')">
          <v-list-item-icon>
            <v-icon>mdi-pause</v-icon>
          </v-list-item-icon>
          <v-list-item-title>{{ $t('pauseNode') }}</v-list-item-title>
        </v-list-item>
        <v-list-item @click="addNode('approval')">
          <v-list-item-icon>
            <v-icon>mdi-check-circle</v-icon>
          </v-list-item-icon>
          <v-list-item-title>{{ $t('approvalNode') }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </div>
</template>

<script>
import { mapGetters } from 'vuex';
import api from '@/lib/api';

export default {
  name: 'WorkflowEditor',
  mixins: [
    require('@/components/mixins/DrawerMixin').default,
  ],
  data() {
    return {
      isLoaded: false,
      workflow: {},
      nodes: [],
      links: [],
      templates: [],
      selectedNode: null,
      sidebarOpen: false,
      saving: false,
      running: false,
      canvasWidth: 2000,
      canvasHeight: 2000,
      dragging: false,
      dragNode: null,
      dragOffset: { x: 0, y: 0 },
      addNodeMenu: false,
      menuX: 0,
      menuY: 0,
      nextNodeId: 1,
      linking: false,
      linkFromNode: null,
      nodeTypes: [
        { text: this.$t('task'), value: 'task' },
        { text: this.$t('pause'), value: 'pause' },
        { text: this.$t('approval'), value: 'approval' },
      ],
    };
  },
  computed: {
    ...mapGetters(['projectId']),
    workflowId() {
      return parseInt(this.$route.params.workflowId);
    },
  },
  mounted() {
    this.loadWorkflow();
    this.loadTemplates();
    document.addEventListener('mousemove', this.handleMouseMove);
    document.addEventListener('mouseup', this.handleMouseUp);
  },
  beforeDestroy() {
    document.removeEventListener('mousemove', this.handleMouseMove);
    document.removeEventListener('mouseup', this.handleMouseUp);
  },
  methods: {
    async loadWorkflow() {
      try {
        this.isLoaded = false;
        const { data } = await api.get(`/project/${this.projectId}/workflows/${this.workflowId}`);
        this.workflow = data;
        this.nodes = data.nodes || [];
        this.links = data.links || [];
        // Assign temporary IDs to new nodes
        this.nodes.forEach((node) => {
          if (!node.id || node.id < 0) {
            node.id = -this.nextNodeId++;
          }
        });
      } catch (err) {
        this.$store.dispatch('setError', err);
      } finally {
        this.isLoaded = true;
      }
    },
    async loadTemplates() {
      try {
        const { data } = await api.get(`/project/${this.projectId}/templates`);
        this.templates = data;
      } catch (err) {
        this.$store.dispatch('setError', err);
      }
    },
    handleCanvasClick(event) {
      if (event.target.classList.contains('workflow-canvas')) {
        this.menuX = event.clientX;
        this.menuY = event.clientY;
        this.addNodeMenu = true;
      }
    },
    addNode(type) {
      const rect = this.$refs.canvas.getBoundingClientRect();
      const newNode = {
        id: -this.nextNodeId++,
        workflow_id: this.workflowId,
        type,
        position_x: this.menuX - rect.left - 75,
        position_y: this.menuY - rect.top - 25,
        task_id: type === 'task' ? null : undefined,
      };
      this.nodes.push(newNode);
      this.addNodeMenu = false;
    },
    deleteNode(node) {
      this.nodes = this.nodes.filter((n) => n.id !== node.id);
      this.links = this.links.filter(
        (l) => l.from_node_id !== node.id && l.to_node_id !== node.id,
      );
      if (this.selectedNode && this.selectedNode.id === node.id) {
        this.selectedNode = null;
        this.sidebarOpen = false;
      }
      if (this.linkFromNode && this.linkFromNode.id === node.id) {
        this.linking = false;
        this.linkFromNode = null;
      }
    },
    deleteLink(link) {
      this.links = this.links.filter((l) => l.id !== link.id);
    },
    handleNodeClick(node) {
      if (this.linking && this.linkFromNode) {
        // Create link
        if (this.linkFromNode.id !== node.id) {
          const newLink = {
            id: -this.nextNodeId++,
            workflow_id: this.workflowId,
            from_node_id: this.linkFromNode.id,
            to_node_id: node.id,
            condition: 'success',
          };
          this.links.push(newLink);
        }
        this.linking = false;
        this.linkFromNode = null;
      } else {
        this.selectNode(node);
      }
    },
    selectNode(node) {
      this.selectedNode = node;
      this.sidebarOpen = true;
    },
    startLinking(node) {
      this.linking = true;
      this.linkFromNode = node;
      this.sidebarOpen = false;
    },
    startDrag(node, event) {
      this.dragging = true;
      this.dragNode = node;
      const rect = event.currentTarget.getBoundingClientRect();
      this.dragOffset = {
        x: event.clientX - rect.left,
        y: event.clientY - rect.top,
      };
    },
    handleMouseMove(event) {
      if (this.dragging && this.dragNode) {
        const rect = this.$refs.canvas.getBoundingClientRect();
        this.dragNode.position_x = event.clientX - rect.left - this.dragOffset.x;
        this.dragNode.position_y = event.clientY - rect.top - this.dragOffset.y;
      }
    },
    handleMouseUp() {
      this.dragging = false;
      this.dragNode = null;
    },
    getNodeX(nodeId) {
      const node = this.nodes.find((n) => n.id === nodeId);
      return node ? node.position_x + 75 : 0;
    },
    getNodeY(nodeId) {
      const node = this.nodes.find((n) => n.id === nodeId);
      return node ? node.position_y + 25 : 0;
    },
    getNodeTypeClass(type) {
      return {
        'node-type-task': type === 'task',
        'node-type-pause': type === 'pause',
        'node-type-approval': type === 'approval',
      };
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
    updateNodeType() {
      if (this.selectedNode.type !== 'task') {
        this.selectedNode.task_id = null;
      }
      this.updateNode();
    },
    updateNode() {
      // Node updated, will be saved on save
    },
    async saveWorkflow() {
      try {
        this.saving = true;
        await api.put(`/project/${this.projectId}/workflows/${this.workflowId}/nodes`, {
          nodes: this.nodes,
          links: this.links,
        });
        this.$store.dispatch('setSuccess', this.$t('workflowSaved'));
        this.loadWorkflow();
      } catch (err) {
        this.$store.dispatch('setError', err);
      } finally {
        this.saving = false;
      }
    },
    async runWorkflow() {
      try {
        this.running = true;
        const { data } = await api.post(`/project/${this.projectId}/workflows/${this.workflowId}/run`);
        this.$router.push(`/project/${this.projectId}/workflows/${this.workflowId}/runs/${data.id}`);
      } catch (err) {
        this.$store.dispatch('setError', err);
      } finally {
        this.running = false;
      }
    },
  },
};
</script>

<style scoped>
.workflow-editor {
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
  cursor: move;
  z-index: 10;
}

.node-header {
  padding: 8px;
  display: flex;
  align-items: center;
  font-weight: 500;
}

.node-type-task {
  background: #e3f2fd;
}

.node-type-pause {
  background: #fff3e0;
}

.node-type-approval {
  background: #e8f5e9;
}

.node-title {
  flex: 1;
  margin-left: 4px;
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-linking {
  border-color: #ff9800;
  box-shadow: 0 0 10px rgba(255, 152, 0, 0.5);
}
</style>
