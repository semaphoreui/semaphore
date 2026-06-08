<template>
  <div>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        <router-link :to="`/project/${projectId}/workflows`">
          {{ $t('workflows') }}
        </router-link>
        <span class="ml-2">/ {{ titleText }}</span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon :title="$t('workflowToolbarZoomOut')" @click="zoomOut()">
        <v-icon>mdi-magnify-minus-outline</v-icon>
      </v-btn>
      <v-btn icon :title="$t('workflowToolbarZoomIn')" @click="zoomIn()">
        <v-icon>mdi-magnify-plus-outline</v-icon>
      </v-btn>
      <v-btn icon :title="$t('workflowToolbarFit')" @click="zoomReset()" class="mr-4">
        <v-icon>mdi-fit-to-page-outline</v-icon>
      </v-btn>

      <v-btn
        color="primary"
        :disabled="!canManage || saving || problems.length > 0"
        :loading="saving"
        @click="save()"
        >{{ $t('save') }}</v-btn
      >
    </v-toolbar>

    <v-divider />

    <div class="WorkflowEditor__body" v-if="item != null && templates != null">
      <!-- Palette + meta -->
      <div class="WorkflowEditor__side WorkflowEditor__side--left">

        <div class="pa-3">
          <div class="text-subtitle-2 mb-2">{{ $t('workflowEditorPalette') }}</div>
          <div class="text-caption text--secondary mb-2">
            {{ $t('workflowDragToCanvasHint') }}
          </div>
          <div
            class="WorkflowEditor__paletteItem WorkflowEditor__paletteItem--task"
            draggable="true"
            @dragstart="onDragStart($event, 'task')"
          >
            <v-icon small left>mdi-cog</v-icon>{{ $t('workflowPaletteTaskNode') }}
          </div>
          <div
            class="WorkflowEditor__paletteItem WorkflowEditor__paletteItem--approval"
            draggable="true"
            @dragstart="onDragStart($event, 'approval')"
          >
            <v-icon small left>mdi-account-check</v-icon>{{ $t('workflowPaletteApprovalNode') }}
          </div>
        </div>

        <v-divider />

        <div class="pa-3">
          <div class="text-subtitle-2 mb-1">{{ $t('workflowProblemsPanelTitle') }}</div>
          <v-alert v-if="problems.length === 0" type="success" text dense class="mb-0">
            {{ $t('workflowValidationPassed') }}
          </v-alert>
          <v-alert
            v-for="(p, i) in problems"
            :key="`problem-${i}`"
            type="warning"
            text
            dense
            class="mb-1"
            >{{ p }}</v-alert
          >
        </div>
      </div>

      <!-- Canvas -->
      <div class="WorkflowEditor__canvas">
        <WorkflowGraph
          ref="graph"
          :key="graphKey"
          :nodes="item.nodes"
          :edges="item.edges"
          :templates="templates"
          editable
          @change="onGraphChange"
          @node-selected="onNodeSelected"
          @connection-selected="onConnectionSelected"
          @blocked="onBlocked"
        />
      </div>

      <!-- Properties panel -->
      <div class="WorkflowEditor__side WorkflowEditor__side--right">
        <div v-if="editingNode" class="pa-3">
          <div class="d-flex align-center mb-2">
            <span class="text-subtitle-2"> {{ $t('workflowNodeId') }} #{{ editingNode.id }} </span>
            <v-spacer />
            <v-btn icon small :disabled="!canManage" @click="deleteSelectedNode()">
              <v-icon small>mdi-delete</v-icon>
            </v-btn>
          </div>

          <v-select
            v-model="editingNode.kind"
            :items="kindOptions"
            item-value="value"
            item-text="text"
            :label="$t('workflowNodeKind')"
            :disabled="!canManage"
            outlined
            dense
            hide-details="auto"
            class="mb-2"
            @change="onKindChanged"
          />

          <v-select
            v-model="editingNode.convergence_mode"
            :items="convergenceOptions"
            item-value="value"
            item-text="text"
            :label="$t('workflowConvergence')"
            :disabled="!canManage"
            outlined
            dense
            hide-details="auto"
            class="mb-2"
            @change="applyNodeEdit"
          />

          <v-autocomplete
            v-if="editingNode.kind !== 'approval'"
            v-model="editingNode.template_id"
            :items="templates"
            item-value="id"
            item-text="name"
            :label="$t('taskTemplate')"
            :disabled="!canManage"
            outlined
            dense
            hide-details="auto"
            class="mb-2"
            @change="applyNodeEdit"
          />

          <ArgsPicker
            v-if="editingNode.kind !== 'approval'"
            :vars="editingNode.limit || []"
            @change="setNodeLimit"
            :title="$t('workflowNodeLimit')"
            :arg-title="$t('limit')"
            :add-arg-title="$t('addLimit')"
            class="mt-2"
          />

          <template v-else>
            <v-text-field
              v-model.number="editingNode.approval_timeout"
              type="number"
              min="1"
              :label="$t('workflowApprovalTimeout')"
              :disabled="!canManage"
              outlined
              dense
              hide-details="auto"
              class="mb-2"
              @change="applyNodeEdit"
            />
            <v-text-field
              v-model="editingNode.approval_message"
              :label="$t('workflowApprovalMessage')"
              :disabled="!canManage"
              outlined
              dense
              hide-details="auto"
              @change="applyNodeEdit"
            />
          </template>
        </div>

        <div v-else-if="editingEdge" class="pa-3">
          <div class="text-subtitle-2 mb-2">{{ $t('workflowEdgeCondition') }}</div>
          <div class="text-caption text--secondary mb-2">
            #{{ editingEdge.source_node_id }} → #{{ editingEdge.destination_node_id }}
          </div>
          <v-select
            v-model="editingEdge.condition"
            :items="conditionOptions"
            item-value="value"
            item-text="text"
            :disabled="!canManage"
            outlined
            dense
            hide-details="auto"
            @change="applyEdgeEdit"
          />
        </div>

        <div v-else class="pa-4 text-caption text--secondary">
          {{ $t('workflowConnectHint') }}
        </div>
      </div>
    </div>

    <div v-else class="pa-4 text-center">
      <v-progress-circular indeterminate color="primary" />
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import ArgsPicker from '@/components/ArgsPicker.vue';
import WorkflowGraph from '@/components/WorkflowGraph.vue';
import ProjectMixin from '@/components/ProjectMixin';
import PermissionsCheck from '@/components/PermissionsCheck';
import { USER_PERMISSIONS } from '@/lib/constants';

export default {
  components: { ArgsPicker, WorkflowGraph },
  mixins: [ProjectMixin, PermissionsCheck],
  props: {
    projectId: Number,
  },
  data() {
    return {
      item: null,
      templates: null,
      saving: false,
      formValid: false,
      graphKey: 0,
      selectedNodeId: null,
      editingNode: null,
      editingEdge: null,
      USER_PERMISSIONS,
    };
  },
  computed: {
    workflowId() {
      const raw = this.$route.params.workflowId;
      return raw ? parseInt(raw, 10) : null;
    },
    isNew() {
      return this.workflowId == null;
    },
    titleText() {
      if (this.isNew) return this.$t('newWorkflow');
      return (this.item && this.item.name) || this.$t('editWorkflow');
    },
    canManage() {
      return this.can(USER_PERMISSIONS.manageProjectResources);
    },
    kindOptions() {
      return [
        { value: 'task', text: this.$t('workflowNodeKindTask') },
        { value: 'approval', text: this.$t('workflowNodeKindApproval') },
      ];
    },
    convergenceOptions() {
      return [
        { value: 'all', text: this.$t('workflowConvergenceAll') },
        { value: 'any', text: this.$t('workflowConvergenceAny') },
      ];
    },
    conditionOptions() {
      return [
        { value: 'on_success', text: this.$t('workflowConditionOnSuccess') },
        { value: 'on_failure', text: this.$t('workflowConditionOnFailure') },
        { value: 'always', text: this.$t('workflowConditionAlways') },
      ];
    },
    // Client-side mirror of db.ValidateWorkflowTemplate (the structural subset).
    problems() {
      const out = [];
      const nodes = this.item?.nodes || [];
      const edges = this.item?.edges || [];
      if (!this.item?.name) out.push(this.$t('name_required'));
      if (nodes.length === 0) {
        out.push(this.$t('workflowErrorNoNodes'));
        return out;
      }
      const incoming = {};
      edges.forEach((e) => {
        incoming[e.destination_node_id] = (incoming[e.destination_node_id] || 0) + 1;
      });
      const roots = nodes.filter((n) => !incoming[n.id]).length;
      if (roots === 0) out.push(this.$t('workflowErrorNoRoot'));
      else if (roots > 1) out.push(this.$t('workflowErrorMultipleRoots', { count: roots }));

      const incompleteTask = nodes.some((n) => (n.kind || 'task') !== 'approval' && !n.template_id);
      if (incompleteTask) out.push(this.$t('workflowErrorTaskNeedsTemplate'));

      const badTimeout = nodes.some(
        (n) => n.kind === 'approval' && n.approval_timeout != null && n.approval_timeout <= 0,
      );
      if (badTimeout) out.push(this.$t('workflowErrorApprovalTimeoutPositive'));
      return out;
    },
  },
  watch: {
    '$route.params.workflowId': function reloadOnRoute() {
      this.loadData();
    },
  },
  async created() {
    this.templates = await this.loadProjectResources('templates');
    await this.loadData();
  },
  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },
    getNewItem() {
      return {
        name: '',
        description: '',
        nodes: [],
        edges: [],
      };
    },
    async loadData() {
      this.selectedNodeId = null;
      this.editingNode = null;
      this.editingEdge = null;
      try {
        if (this.isNew) {
          this.item = this.getNewItem();
        } else {
          this.item = await this.loadEndpoint(
            `/api/project/${this.projectId}/workflows/${this.workflowId}`,
          );
          if (!Array.isArray(this.item.nodes)) this.item.nodes = [];
          if (!Array.isArray(this.item.edges)) this.item.edges = [];
          this.item.nodes = this.item.nodes.map((node) => ({
            kind: 'task',
            convergence_mode: 'all',
            limit: [],
            position_x: 0,
            position_y: 0,
            ...node,
          }));
          this.autoLayout();
        }
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
        return;
      }
      // Force a clean canvas rebuild matching the freshly loaded model.
      this.graphKey += 1;
    },
    // Seed positions for legacy workflows whose nodes were created before the
    // graphical editor (all coordinates 0). Lay them out in topological columns.
    autoLayout() {
      const nodes = this.item.nodes;
      const allZero = nodes.every((n) => !n.position_x && !n.position_y);
      if (!allZero || nodes.length === 0) return;

      const incoming = {};
      const adjacency = {};
      nodes.forEach((n) => {
        incoming[n.id] = 0;
        adjacency[n.id] = [];
      });
      this.item.edges.forEach((e) => {
        if (incoming[e.destination_node_id] != null) incoming[e.destination_node_id] += 1;
        if (adjacency[e.source_node_id]) adjacency[e.source_node_id].push(e.destination_node_id);
      });

      const depth = {};
      const queue = nodes.filter((n) => !incoming[n.id]).map((n) => n.id);
      queue.forEach((id) => {
        depth[id] = 0;
      });
      const pending = { ...incoming };
      while (queue.length) {
        const id = queue.shift();
        (adjacency[id] || []).forEach((next) => {
          depth[next] = Math.max(depth[next] || 0, (depth[id] || 0) + 1);
          pending[next] -= 1;
          if (pending[next] <= 0) queue.push(next);
        });
      }
      const perColumn = {};
      nodes.forEach((n, i) => {
        const col = depth[n.id] || 0;
        const row = perColumn[col] || 0;
        // Assign through the array (not the forEach param) to satisfy no-param-reassign.
        nodes[i].position_x = 80 + col * 240;
        nodes[i].position_y = 60 + row * 130;
        perColumn[col] = row + 1;
      });
    },

    // ---- palette / canvas glue ------------------------------------------------
    onDragStart(ev, kind) {
      ev.dataTransfer.setData('node-kind', kind);
    },
    onGraphChange({ nodes, edges }) {
      this.item.nodes = nodes;
      this.item.edges = edges;
      // Keep the open property panel in sync with the latest model snapshot.
      if (this.selectedNodeId != null) {
        const found = nodes.find((n) => n.id === this.selectedNodeId);
        if (!found) {
          this.selectedNodeId = null;
          this.editingNode = null;
        }
      }
    },
    onNodeSelected(nodeId) {
      this.editingEdge = null;
      this.selectedNodeId = nodeId;
      if (nodeId == null) {
        this.editingNode = null;
        return;
      }
      const node = this.item.nodes.find((n) => n.id === nodeId);
      this.editingNode = node ? JSON.parse(JSON.stringify(node)) : null;
    },
    onConnectionSelected(edge) {
      this.selectedNodeId = null;
      this.editingNode = null;
      this.editingEdge = { ...edge };
    },
    onBlocked(reason) {
      const key = reason === 'cycle' ? 'workflowCycleBlocked' : 'workflowSelfEdgeBlocked';
      EventBus.$emit('i-snackbar', { color: 'warning', text: this.$t(key) });
    },
    onKindChanged() {
      if (this.editingNode.kind === 'approval') {
        this.editingNode.template_id = null;
        this.editingNode.limit = [];
      } else {
        this.editingNode.approval_timeout = null;
        this.editingNode.approval_message = null;
      }
      this.applyNodeEdit();
    },
    setNodeLimit(limit) {
      this.editingNode.limit = limit;
      this.applyNodeEdit();
    },
    applyNodeEdit() {
      if (!this.editingNode || !this.$refs.graph) return;
      this.$refs.graph.syncNode(this.editingNode.id, { ...this.editingNode });
    },
    applyEdgeEdit() {
      if (!this.editingEdge || !this.$refs.graph) return;
      this.$refs.graph.setCondition(
        this.editingEdge.source_node_id,
        this.editingEdge.destination_node_id,
        this.editingEdge.condition,
      );
    },
    deleteSelectedNode() {
      if (this.editingNode == null || !this.$refs.graph) return;
      this.$refs.graph.removeSelectedNode(this.editingNode.id);
      this.editingNode = null;
      this.selectedNodeId = null;
    },
    zoomIn() {
      this.$refs.graph?.zoomIn();
    },
    zoomOut() {
      this.$refs.graph?.zoomOut();
    },
    zoomReset() {
      this.$refs.graph?.zoomReset();
    },

    // ---- save -----------------------------------------------------------------
    async save() {
      if (!this.$refs.form.validate()) return;
      if (this.problems.length > 0) return;
      this.saving = true;
      try {
        const payload = { ...this.item, project_id: this.projectId };
        if (this.isNew) {
          const created = (await axios.post(`/api/project/${this.projectId}/workflows`, payload))
            .data;
          EventBus.$emit('i-snackbar', { color: 'success', text: this.$t('workflowSaved') });
          // Navigate to the edit route; the route watcher reloads with the
          // server-assigned node ids (the backend reassigns ids on every save).
          this.$router.replace(`/project/${this.projectId}/workflows/${created.id}/edit`);
        } else {
          await axios.put(`/api/project/${this.projectId}/workflows/${this.workflowId}`, payload);
          EventBus.$emit('i-snackbar', { color: 'success', text: this.$t('workflowSaved') });
          // Re-fetch to rebind reassigned node ids onto the canvas.
          await this.loadData();
        }
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      } finally {
        this.saving = false;
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.WorkflowEditor {

  &__body {
    display: flex;
    flex: 1 1 auto;
    min-height: 0;
    height: calc(100vh - 64px);
  }

  &__side {
    width: 280px;
    flex: 0 0 280px;
    overflow-y: auto;
    border-right: 1px solid rgba(127, 127, 127, 0.2);

    &--right {
      border-right: none;
      border-left: 1px solid rgba(127, 127, 127, 0.2);
    }
  }

  &__canvas {
    flex: 1 1 auto;
    min-width: 0;
    position: relative;
  }

  &__paletteItem {
    display: flex;
    align-items: center;
    padding: 8px 10px;
    margin-bottom: 8px;
    border: 1px dashed rgba(127, 127, 127, 0.5);
    border-radius: 6px;
    cursor: grab;
    user-select: none;
    font-size: 13px;

    &--task {
      border-left: 3px solid #2196f3;
    }
    &--approval {
      border-left: 3px solid #ab47bc;
    }
  }
}
</style>
