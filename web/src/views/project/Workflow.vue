<template>
  <div style="height: 100%; display: flex; flex-direction: column;">
    <v-toolbar flat dense>
      <v-toolbar-title>{{ workflow.name }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn @click="runWorkflow" color="success" :disabled="running">Run</v-btn>
      <v-btn @click="saveWorkflow" color="primary">Save</v-btn>
    </v-toolbar>

    <div style="flex: 1; display: flex; overflow: hidden;">
      <!-- Sidebar -->
      <div style="width: 250px; border-right: 1px solid #ddd; padding: 10px; overflow-y: auto;">
        <v-btn block @click="addNode('task')">Add Task Node</v-btn>
        <!-- <v-btn block @click="addNode('pause')">Add Pause Node</v-btn> -->
        
        <v-divider class="my-4"></v-divider>
        
        <div v-if="selectedNode">
          <h3>Edit Node</h3>
          <v-select
            v-if="selectedNode.type === 'task'"
            v-model="selectedNode.task_id"
            :items="templates"
            item-text="name"
            item-value="id"
            label="Template"
          ></v-select>
          
          <v-text-field v-model.number="selectedNode.position_x" label="X" type="number"></v-text-field>
          <v-text-field v-model.number="selectedNode.position_y" label="Y" type="number"></v-text-field>
          
          <v-btn color="error" block @click="deleteSelectedNode">Delete Node</v-btn>
          
          <h4 class="mt-4">Links</h4>
          <div v-for="(link, i) in getNodeLinks(selectedNode.id)" :key="i">
            Link to {{ link.to_node_id }} ({{ link.condition }})
            <v-btn icon small @click="removeLink(link)"><v-icon>mdi-close</v-icon></v-btn>
          </div>
          
          <h4 class="mt-4">Add Link</h4>
          <v-select
            v-model="newLinkTarget"
            :items="availableLinkTargets"
            item-text="id"
            item-value="id"
            label="Target Node ID"
          ></v-select>
          <v-select
            v-model="newLinkCondition"
            :items="['always', 'success', 'failure']"
            label="Condition"
          ></v-select>
          <v-btn block @click="addLink">Add Link</v-btn>
        </div>
      </div>

      <!-- Canvas -->
      <div 
        ref="canvas"
        style="flex: 1; position: relative; background-color: #f5f5f5; overflow: auto;"
        @mousedown="stopDrag"
        @mousemove="doDrag"
        @mouseup="stopDrag"
      >
        <svg style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; z-index: 1;">
           <line v-for="(link, i) in links" :key="i"
             :x1="getNodeX(link.from_node_id) + 50" 
             :y1="getNodeY(link.from_node_id) + 25"
             :x2="getNodeX(link.to_node_id) + 50"
             :y2="getNodeY(link.to_node_id) + 25"
             stroke="black"
             stroke-width="2"
             marker-end="url(#arrowhead)"
           />
           <defs>
             <marker id="arrowhead" markerWidth="10" markerHeight="7" 
             refX="0" refY="3.5" orient="auto">
               <polygon points="0 0, 10 3.5, 0 7" />
             </marker>
           </defs>
        </svg>

        <div
          v-for="node in nodes"
          :key="node.id"
          :style="{
            position: 'absolute',
            left: node.position_x + 'px',
            top: node.position_y + 'px',
            width: '100px',
            height: '50px',
            border: '1px solid #333',
            backgroundColor: selectedNode === node ? '#b3d4fc' : 'white',
            cursor: 'move',
            zIndex: 10
          }"
          @mousedown.stop="selectNode(node, $event)"
        >
          <div style="padding: 5px; text-align: center; font-size: 12px;">
            {{ getNodeName(node) }}
            <br>
            ID: {{ node.id }}
          </div>
        </div>
      </div>
    </div>
    
    <!-- Run History -->
    <div style="height: 200px; border-top: 1px solid #ddd; overflow-y: auto; padding: 10px;">
        <h3>Runs</h3>
        <v-list dense>
            <v-list-item v-for="run in runs" :key="run.id">
                <v-list-item-content>
                    <v-list-item-title>
                        Run #{{ run.id }} - {{ run.status }} - {{ run.created_at }}
                    </v-list-item-title>
                </v-list-item-content>
            </v-list-item>
        </v-list>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  data: () => ({
    workflow: {},
    nodes: [],
    links: [],
    templates: [],
    selectedNode: null,
    newLinkTarget: null,
    newLinkCondition: 'success',
    dragNode: null,
    dragOffsetX: 0,
    dragOffsetY: 0,
    projectId: null,
    workflowId: null,
    runs: [],
    running: false,
    
    // Temp ID counter for new nodes
    tempIdCounter: -1,
  }),
  
  computed: {
    availableLinkTargets() {
        if (!this.selectedNode) return [];
        return this.nodes.filter(n => n.id !== this.selectedNode.id);
    }
  },

  async created() {
    this.projectId = this.$route.params.projectId;
    this.workflowId = this.$route.params.workflowId;
    await this.loadTemplates();
    await this.loadWorkflow();
    await this.loadRuns();
  },

  methods: {
    async loadTemplates() {
        const res = await axios.get(`/api/project/${this.projectId}/templates`);
        this.templates = res.data;
    },
    async loadWorkflow() {
        const res = await axios.get(`/api/project/${this.projectId}/workflows/${this.workflowId}`);
        this.workflow = res.data;
        this.nodes = res.data.nodes || [];
        this.links = res.data.links || [];
    },
    async loadRuns() {
        const res = await axios.get(`/api/project/${this.projectId}/workflows/${this.workflowId}/runs`);
        this.runs = res.data;
    },
    
    async saveWorkflow() {
        const payload = {
            ...this.workflow,
            nodes: this.nodes,
            links: this.links
        };
        await axios.put(`/api/project/${this.projectId}/workflows/${this.workflowId}`, payload);
        await this.loadWorkflow(); // Reload to get real IDs
    },
    
    async runWorkflow() {
        this.running = true;
        try {
            await axios.post(`/api/project/${this.projectId}/workflows/${this.workflowId}/run`);
            this.loadRuns();
        } finally {
            this.running = false;
        }
    },
    
    addNode(type) {
        this.nodes.push({
            id: this.tempIdCounter--, // Negative ID for new nodes
            type: type,
            position_x: 50,
            position_y: 50,
            task_id: null
        });
    },
    
    selectNode(node, event) {
        this.selectedNode = node;
        this.dragNode = node;
        this.dragOffsetX = event.clientX - node.position_x;
        this.dragOffsetY = event.clientY - node.position_y;
    },
    
    doDrag(event) {
        if (this.dragNode) {
            const canvasRect = this.$refs.canvas.getBoundingClientRect();
            this.dragNode.position_x = event.clientX - canvasRect.left - this.dragOffsetX;
            this.dragNode.position_y = event.clientY - canvasRect.top - this.dragOffsetY;
        }
    },
    
    stopDrag() {
        this.dragNode = null;
    },
    
    getNodeName(node) {
        if (node.type === 'task') {
            const tpl = this.templates.find(t => t.id === node.task_id);
            return tpl ? tpl.name : 'Task Node';
        }
        return node.type;
    },
    
    getNodeX(id) {
        const n = this.nodes.find(n => n.id === id);
        return n ? n.position_x : 0;
    },
    
    getNodeY(id) {
        const n = this.nodes.find(n => n.id === id);
        return n ? n.position_y : 0;
    },
    
    getNodeLinks(nodeId) {
        return this.links.filter(l => l.from_node_id === nodeId);
    },
    
    addLink() {
        if (this.selectedNode && this.newLinkTarget) {
            this.links.push({
                from_node_id: this.selectedNode.id,
                to_node_id: this.newLinkTarget,
                condition: this.newLinkCondition
            });
        }
    },
    
    removeLink(link) {
        const idx = this.links.indexOf(link);
        if (idx > -1) this.links.splice(idx, 1);
    },
    
    deleteSelectedNode() {
        if (this.selectedNode) {
            // Remove links
            this.links = this.links.filter(l => l.from_node_id !== this.selectedNode.id && l.to_node_id !== this.selectedNode.id);
            // Remove node
            const idx = this.nodes.indexOf(this.selectedNode);
            if (idx > -1) this.nodes.splice(idx, 1);
            this.selectedNode = null;
        }
    }
  }
};
</script>
