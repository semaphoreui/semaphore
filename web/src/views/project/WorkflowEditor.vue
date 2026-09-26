<template>
  <div class="WorkflowEditor">
    <YesNoDialog
      v-model="leaveDialog"
      :title="$t('workflowUnsavedChanges')"
      :text="$t('workflowUnsavedLeave')"
      :yes-button-title="$t('workflowLeaveWithoutSaving')"
      @yes="confirmLeave()"
      @no="cancelLeave()"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title class="WorkflowEditor__title d-flex align-center">
        <router-link :to="`/project/${projectId}/workflows`">
          {{ $t('workflows') }}
        </router-link>
        <span class="mx-2">/</span>
        <span
          v-if="item != null"
          class="WorkflowEditor__nameWrap"
          :class="{ 'WorkflowEditor__nameWrap--disabled': !canManage }"
        >
          <v-icon small class="WorkflowEditor__nameIcon">mdi-pencil</v-icon>
          <span class="WorkflowEditor__nameSizer" :data-value="item.name || $t('newWorkflow')">
            <input
              v-model="item.name"
              :placeholder="$t('newWorkflow')"
              :disabled="!canManage"
              :aria-label="$t('name')"
              size="1"
              class="WorkflowEditor__nameField"
            />
          </span>
        </span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon :disabled="!canUndo" :title="$t('workflowUndo')" @click="undo()">
        <v-icon>mdi-undo</v-icon>
      </v-btn>
      <v-btn icon :disabled="!canRedo" :title="$t('workflowRedo')" class="mr-2" @click="redo()">
        <v-icon>mdi-redo</v-icon>
      </v-btn>

      <v-menu v-if="problems.length > 0" offset-y :close-on-content-click="true">
        <template v-slot:activator="{ on, attrs }">
          <v-chip small outlined color="warning" class="mr-4" v-bind="attrs" v-on="on">
            <v-icon left small>mdi-alert</v-icon>
            {{ $tc('workflowProblemsCount', problems.length, { count: problems.length }) }}
          </v-chip>
        </template>
        <v-list dense>
          <v-list-item
            v-for="(p, i) in problems"
            :key="`problem-${i}`"
            @click="focusProblem(p)"
          >
            <v-list-item-icon class="mr-2">
              <v-icon small color="warning">mdi-alert</v-icon>
            </v-list-item-icon>
            <v-list-item-title>{{ p.text }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
      <v-chip v-else small outlined color="success" class="mr-4">
        <v-icon left small>mdi-check</v-icon>
        {{ $t('workflowValidationPassed') }}
      </v-chip>

      <v-badge :value="dirty" dot color="warning" overlap offset-x="8" offset-y="8">
        <v-btn
          color="primary"
          depressed
          :disabled="!canManage || saving || problems.length > 0"
          :loading="saving"
          @click="save()"
        >{{ $t('save') }}</v-btn>
      </v-badge>
    </v-toolbar>

    <v-divider />

    <div class="WorkflowEditor__body" v-if="item != null && templates != null">
      <!-- Palette + meta -->
      <div
        class="WorkflowEditor__side WorkflowEditor__side--left"
        :class="{ 'WorkflowEditor__side--collapsed': sideCollapsed }"
      >
        <button
          type="button"
          class="WorkflowEditor__sideToggle"
          :title="$t(sideCollapsed ? 'workflowSidebarExpand' : 'workflowSidebarCollapse')"
          @click="sideCollapsed = !sideCollapsed"
        >
          <v-icon small>
            {{ sideCollapsed ? 'mdi-chevron-right' : 'mdi-chevron-left' }}
          </v-icon>
        </button>

        <div class="WorkflowEditor__sideScroll">
          <template v-if="!sideCollapsed">
            <div class="pa-3">
              <v-text-field
                v-model="item.start_version"
                :label="$t('startVersion')"
                :hint="$t('workflowStartVersionHint')"
                persistent-hint
                :disabled="!canManage"
                outlined
                dense
              />
            </div>

            <v-divider />
          </template>

          <div class="pa-3">
            <template v-if="!sideCollapsed">
              <div class="text-subtitle-2 mb-1">{{ $t('workflowEditorPalette') }}</div>
              <div class="text-caption text--secondary mb-3">
                {{ $t('workflowPaletteClickHint') }}
              </div>
            </template>
            <div
              v-for="p in palette"
              :key="p.kind"
              class="WorkflowEditor__paletteItem"
              :class="`WorkflowEditor__paletteItem--${p.kind}`"
              draggable="true"
              :title="sideCollapsed ? p.text : null"
              @dragstart="onDragStart($event, p.kind)"
              @click="addFromPalette(p.kind)"
            >
              <span
                class="WorkflowEditor__paletteTile"
                :class="`WorkflowEditor__paletteTile--${p.kind}`"
              >
                <v-icon small :color="p.color">{{ p.icon }}</v-icon>
              </span>
              <span v-if="!sideCollapsed" class="WorkflowEditor__paletteText">{{ p.text }}</span>
            </div>
          </div>
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
          :node-problems="problemsByNode"
          editable
          @change="onGraphChange"
          @node-selected="onNodeSelected"
          @connection-selected="onConnectionSelected"
          @blocked="onBlocked"
        />
      </div>

      <!-- Properties panel -->
      <div
        v-if="editingNode || editingEdge"
        class="WorkflowEditor__side WorkflowEditor__side--right"
      >
        <WorkflowNodeProperties
          v-if="editingNode"
          :key="`node-${editingNode.id}`"
          :value="editingNode"
          :templates="templates"
          :can-manage="canManage"
          @input="applyNodeEdit"
          @delete="deleteSelectedNode()"
          @close="closePanel()"
        />
        <WorkflowEdgeProperties
          v-else
          :key="edgeKey(editingEdge)"
          :value="editingEdge"
          :can-manage="canManage"
          @input="applyEdgeEdit"
          @delete="deleteSelectedEdge()"
          @close="closePanel()"
        />
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
import WorkflowGraph from '@/components/WorkflowGraph.vue';
import WorkflowNodeProperties from '@/components/workflow/WorkflowNodeProperties.vue';
import WorkflowEdgeProperties from '@/components/workflow/WorkflowEdgeProperties.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ProjectMixin from '@/components/ProjectMixin';
import PermissionsCheck from '@/components/PermissionsCheck';
import { USER_PERMISSIONS } from '@/lib/constants';
import { layoutWorkflowNodes, needsAutoLayout } from '@/lib/workflowLayout';
import WorkflowHistory from '@/lib/workflowHistory';

function isTypingTarget(target) {
  if (!target) return false;
  const tag = (target.tagName || '').toLowerCase();
  return tag === 'input' || tag === 'textarea' || tag === 'select' || target.isContentEditable;
}

export default {
  components: {
    WorkflowGraph, WorkflowNodeProperties, WorkflowEdgeProperties, YesNoDialog,
  },
  mixins: [ProjectMixin, PermissionsCheck],
  props: {
    projectId: Number,
  },
  data() {
    return {
      item: null,
      templates: null,
      saving: false,
      graphKey: 0,
      // Set before navigating new -> /edit after a create, so the route watcher
      // does not reload (which would reset the selection and rebuild the canvas).
      skipNextRouteReload: false,
      selectedNodeId: null,
      editingNode: null,
      editingEdge: null,
      sideCollapsed: false,
      canUndo: false,
      canRedo: false,
      // JSON of the last loaded / saved model, for the unsaved-changes guard.
      savedSnapshot: null,
      leaveDialog: false,
      pendingLeave: null,
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
    canManage() {
      return this.can(USER_PERMISSIONS.manageProjectResources);
    },
    palette() {
      return [
        {
          kind: 'task', icon: 'mdi-cog', color: null, text: this.$t('workflowPaletteTaskNode'),
        },
        {
          kind: 'approval', icon: 'mdi-account-check', color: '#ab47bc', text: this.$t('workflowPaletteApprovalNode'),
        },
        {
          kind: 'delay', icon: 'mdi-timer-outline', color: '#ff9800', text: this.$t('workflowPaletteDelayNode'),
        },
        {
          kind: 'note', icon: 'mdi-note-text-outline', color: null, text: this.$t('workflowPaletteNoteNode'),
        },
      ];
    },
    modelSnapshot() {
      if (!this.item) return null;
      return JSON.stringify({
        name: this.item.name,
        start_version: this.item.start_version || '',
        nodes: this.item.nodes,
        edges: this.item.edges,
      });
    },
    dirty() {
      return this.savedSnapshot != null && this.modelSnapshot !== this.savedSnapshot;
    },
    // Client-side mirror of db.ValidateWorkflowTemplate (the structural subset).
    // Each problem may point at a node, which gets a badge on the canvas.
    problems() {
      const out = [];
      const nodes = this.item?.nodes || [];
      const edges = this.item?.edges || [];
      if (!this.item?.name) out.push({ text: this.$t('name_required') });
      if (nodes.length === 0) {
        out.push({ text: this.$t('workflowErrorNoNodes') });
        return out;
      }
      const incoming = {};
      edges.forEach((e) => {
        incoming[e.destination_node_id] = (incoming[e.destination_node_id] || 0) + 1;
      });
      // Note nodes are annotations: excluded from the run graph (root count) and
      // from the task-completeness check.
      const roots = nodes.filter((n) => n.kind !== 'note' && !incoming[n.id]);
      if (roots.length === 0) {
        out.push({ text: this.$t('workflowErrorNoRoot') });
      } else if (roots.length > 1) {
        const text = this.$t('workflowErrorMultipleRoots', { count: roots.length });
        roots.forEach((n) => out.push({ text, nodeId: n.id }));
      }

      nodes.forEach((n) => {
        const kind = n.kind || 'task';
        if (kind === 'task' && !n.template_id) {
          out.push({ text: this.$t('workflowErrorTaskNeedsTemplate'), nodeId: n.id });
        }
        if (kind === 'approval' && n.approval_timeout != null && n.approval_timeout <= 0) {
          out.push({ text: this.$t('workflowErrorApprovalTimeoutPositive'), nodeId: n.id });
        }
        if (kind === 'delay' && (n.delay_seconds == null || n.delay_seconds <= 0)) {
          out.push({ text: this.$t('workflowErrorDelayPositive'), nodeId: n.id });
        }
      });
      return out;
    },
    problemsByNode() {
      const map = {};
      this.problems.forEach((p) => {
        if (p.nodeId == null || map[p.nodeId]) return;
        map[p.nodeId] = p.text;
      });
      return map;
    },
  },
  watch: {
    '$route.params.workflowId': function reloadOnRoute() {
      if (this.skipNextRouteReload) {
        this.skipNextRouteReload = false;
        return;
      }
      this.loadData();
    },
  },
  beforeRouteLeave(to, from, next) {
    if (!this.dirty) {
      next();
      return;
    }
    this.pendingLeave = next;
    this.leaveDialog = true;
  },
  created() {
    this.history = new WorkflowHistory(50);
    this.pendingHistoryKey = null;
  },
  async mounted() {
    window.addEventListener('keydown', this.onWindowKeyDown);
    window.addEventListener('beforeunload', this.onBeforeUnload);
    this.templates = await this.loadProjectResources('templates');
    await this.loadData();
  },
  beforeDestroy() {
    window.removeEventListener('keydown', this.onWindowKeyDown);
    window.removeEventListener('beforeunload', this.onBeforeUnload);
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
      this.history.reset({ nodes: this.item.nodes, edges: this.item.edges });
      this.syncHistoryFlags();
      this.savedSnapshot = this.modelSnapshot;
      // Force a clean canvas rebuild matching the freshly loaded model.
      this.graphKey += 1;
    },
    // Seed positions for legacy workflows whose nodes were created before the
    // graphical editor (all coordinates 0) so the computed layout persists on
    // the next save. Uses the same algorithm as the read-only run view.
    autoLayout() {
      const { nodes } = this.item;
      if (!needsAutoLayout(nodes)) return;
      const layout = layoutWorkflowNodes(nodes, this.item.edges);
      nodes.forEach((n, i) => {
        // Assign through the array (not the forEach param) to satisfy no-param-reassign.
        nodes[i].position_x = layout[n.id].x;
        nodes[i].position_y = layout[n.id].y;
      });
    },

    // ---- palette / canvas glue ------------------------------------------------
    onDragStart(ev, kind) {
      ev.dataTransfer.setData('node-kind', kind);
    },
    addFromPalette(kind) {
      if (!this.canManage || !this.$refs.graph) return;
      this.$refs.graph.addNodeAtCenter(kind);
    },
    onGraphChange({ nodes, edges }) {
      this.item.nodes = nodes;
      this.item.edges = edges;
      this.history.push({ nodes, edges }, this.pendingHistoryKey);
      this.pendingHistoryKey = null;
      this.syncHistoryFlags();
      // Keep the open property panel in sync with the latest model snapshot.
      if (this.selectedNodeId != null) {
        const found = nodes.find((n) => n.id === this.selectedNodeId);
        if (!found) {
          this.selectedNodeId = null;
          this.editingNode = null;
        }
      }
      if (this.editingEdge) {
        const { source_node_id: s, destination_node_id: d } = this.editingEdge;
        if (!edges.some((e) => e.source_node_id === s && e.destination_node_id === d)) {
          this.editingEdge = null;
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
      const clone = node ? JSON.parse(JSON.stringify(node)) : null;
      // Define task_params up front so later assignments stay reactive.
      // Approval/note nodes must not carry task params (backend validation).
      if (clone && clone.task_params === undefined) {
        clone.task_params = (clone.kind || 'task') === 'task' ? {} : null;
      }
      this.editingNode = clone;
    },
    onConnectionSelected(edge) {
      if (edge == null) {
        this.editingEdge = null;
        return;
      }
      this.selectedNodeId = null;
      this.editingNode = null;
      this.editingEdge = { ...edge };
    },
    onBlocked(reason) {
      const key = reason === 'cycle' ? 'workflowCycleBlocked' : 'workflowSelfEdgeBlocked';
      EventBus.$emit('i-snackbar', { color: 'warning', text: this.$t(key) });
    },
    edgeKey(edge) {
      return `edge-${edge.source_node_id}-${edge.destination_node_id}`;
    },
    closePanel() {
      this.editingNode = null;
      this.editingEdge = null;
      this.selectedNodeId = null;
      if (this.$refs.graph) this.$refs.graph.clearSelection();
    },
    focusProblem(problem) {
      if (problem.nodeId == null || !this.$refs.graph) return;
      this.$refs.graph.selectNode(problem.nodeId);
    },
    applyNodeEdit() {
      if (!this.editingNode || !this.$refs.graph) return;
      // Coalesce keystrokes on one node into a single undo step.
      this.pendingHistoryKey = `node-${this.editingNode.id}`;
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
    deleteSelectedEdge() {
      if (this.editingEdge == null || !this.$refs.graph) return;
      this.$refs.graph.removeEdge(
        this.editingEdge.source_node_id,
        this.editingEdge.destination_node_id,
      );
      this.editingEdge = null;
    },

    // ---- history ----------------------------------------------------------------
    syncHistoryFlags() {
      this.canUndo = this.history.canUndo();
      this.canRedo = this.history.canRedo();
    },
    applySnapshot(snapshot) {
      if (!snapshot) return;
      this.item.nodes = snapshot.nodes;
      this.item.edges = snapshot.edges;
      this.editingNode = null;
      this.editingEdge = null;
      this.selectedNodeId = null;
      this.syncHistoryFlags();
      this.$nextTick(() => {
        if (this.$refs.graph) this.$refs.graph.reload();
      });
    },
    undo() {
      if (!this.canManage) return;
      this.applySnapshot(this.history.undo());
    },
    redo() {
      if (!this.canManage) return;
      this.applySnapshot(this.history.redo());
    },
    onWindowKeyDown(ev) {
      if (!(ev.ctrlKey || ev.metaKey) || isTypingTarget(ev.target)) return;
      const key = ev.key.toLowerCase();
      if (key === 'z' && ev.shiftKey) {
        ev.preventDefault();
        this.redo();
      } else if (key === 'z') {
        ev.preventDefault();
        this.undo();
      } else if (key === 'y') {
        ev.preventDefault();
        this.redo();
      }
    },

    // ---- unsaved changes guard --------------------------------------------------
    onBeforeUnload(ev) {
      if (!this.dirty) return undefined;
      ev.preventDefault();
      // eslint-disable-next-line no-param-reassign
      ev.returnValue = '';
      return '';
    },
    confirmLeave() {
      const next = this.pendingLeave;
      this.pendingLeave = null;
      if (next) next();
    },
    cancelLeave() {
      const next = this.pendingLeave;
      this.pendingLeave = null;
      if (next) next(false);
    },

    // ---- save -----------------------------------------------------------------
    async save() {
      if (this.problems.length > 0) return;
      this.saving = true;
      try {
        const payload = { ...this.item, project_id: this.projectId };
        // An empty start_version means "no run versioning" — send null, not "".
        if (!payload.start_version) delete payload.start_version;
        if (this.isNew) {
          const created = (await axios.post(`/api/project/${this.projectId}/workflows`, payload))
            .data;
          EventBus.$emit('i-snackbar', { color: 'success', text: this.$t('workflowSaved') });
          // Adopt the server id so subsequent saves PUT to the right workflow,
          // then switch the URL to the edit route — but keep the current canvas
          // and selection (skip the route-triggered reload). The backend remaps
          // client node ids on every save, so no reload is needed.
          this.item.id = created.id;
          this.savedSnapshot = this.modelSnapshot;
          this.skipNextRouteReload = true;
          this.$router.replace(`/project/${this.projectId}/workflows/${created.id}/edit`);
        } else {
          await axios.put(`/api/project/${this.projectId}/workflows/${this.workflowId}`, payload);
          EventBus.$emit('i-snackbar', { color: 'success', text: this.$t('workflowSaved') });
          this.savedSnapshot = this.modelSnapshot;
          // No reload: keep the canvas and the selected element intact.
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

$worklow_pallete_width_collapsed: 60px;

.WorkflowEditor {
  // Inline, borderless name editor in the toolbar title.
  &__title {
    overflow: visible;
  }

  // Inline name editor: looks like a title, but the pencil + hover/focus
  // affordances make it clear it is editable. The field auto-sizes to the
  // width of its text via the grid-sizer trick below.
  &__nameWrap {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    max-width: 60vw;
    padding: 2px 8px;
    border-radius: 4px;
    background: rgba(127, 127, 127, 0.12);
    transition: background-color 0.15s ease, border-color 0.15s ease;

    &--disabled {
      pointer-events: none;
      opacity: 0.7;
    }
  }

  &__nameIcon {
    opacity: 0.6;
    flex: 0 0 auto;
  }

  &__nameWrap:focus-within &__nameIcon {
    opacity: 1;
  }

  // The sizer ::after mirrors the value and gives the grid cell the text width;
  // the input shares the same cell, so it is exactly as wide as the text.
  &__nameSizer {
    display: inline-grid;
    align-items: center;
    min-width: 0;

    &::after,
    & > .WorkflowEditor__nameField {
      grid-area: 1 / 1;
      width: auto;
      min-width: 1ch;
      font: inherit;
      font-weight: 500;
      letter-spacing: inherit;
      padding: 0;
      margin: 0;
      border: 0;
      background: transparent;
      white-space: pre;
    }

    &::after {
      content: attr(data-value);
      visibility: hidden;
    }
  }

  &__nameField {
    color: inherit;
    outline: none;

    &::placeholder {
      color: inherit;
      opacity: 0.5;
    }
  }

  &__body {
    display: flex;
    flex: 1 1 auto;
    min-height: 0;
    height: calc(100vh - 65px);
  }

  &__side {
    width: 280px;
    flex: 0 0 280px;
    overflow-y: auto;
    border-right: 1px solid rgba(127, 127, 127, 0.2);

    // The left panel hosts the collapse tab protruding over the canvas, so it
    // must not clip overflow itself — scrolling moves to __sideScroll. z-index
    // lifts the tab above the (positioned) canvas that follows in the DOM.
    &--left {
      position: relative;
      overflow: visible;
      z-index: 1;
    }

    &--right {
      width: 360px;
      flex: 0 0 360px;
      border-right: none;
      border-left: 1px solid rgba(127, 127, 127, 0.2);
    }

    &--collapsed {
      width: $worklow_pallete_width_collapsed;
      flex: 0 0 $worklow_pallete_width_collapsed;
    }
  }

  &__sideScroll {
    height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
  }

  // Collapse/expand handle: a tab sticking out of the panel's right edge,
  // vertically centered.
  &__sideToggle {
    position: absolute;
    top: 50%;
    right: -20px;
    transform: translateY(-50%);
    width: 20px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    border: 1px solid rgba(127, 127, 127, 0.2);
    border-left: none;
    border-radius: 0 8px 8px 0;
    cursor: pointer;
    background: inherit;
  }

  &__canvas {
    flex: 1 1 auto;
    min-width: 0;
    position: relative;
  }

  &__paletteItem {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 8px;
    margin-bottom: 8px;
    border: 1px solid rgba(127, 127, 127, 0.3);
    border-radius: 10px;
    cursor: grab;
    user-select: none;
    font-size: 13px;
    transition: border-color 0.12s ease, box-shadow 0.12s ease;

    &:hover {
      border-color: rgba(127, 127, 127, 0.6);
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
    }

    &:active {
      cursor: grabbing;
    }
  }

  &__paletteTile {
    flex: 0 0 30px;
    width: 30px;
    height: 30px;
    border-radius: 8px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: rgba(127, 127, 127, 0.12);

    &--approval {
      background: rgba(171, 71, 188, 0.14);
    }

    &--delay {
      background: rgba(255, 152, 0, 0.16);
    }

    &--note {
      background: rgba(230, 216, 115, 0.35);
    }
  }

  &__paletteText {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  &__side--collapsed &__paletteItem {
    justify-content: center;
    padding: 6px;
  }
}
</style>
