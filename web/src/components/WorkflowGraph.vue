<template>
  <div class="WorkflowGraph" :class="{ 'WorkflowGraph--editable': editable }">
    <div
      ref="canvas"
      class="WorkflowGraph__canvas"
      @drop="onDrop"
      @dragover.prevent
    ></div>

    <div class="WorkflowGraph__legend">
      <span class="WorkflowGraph__legendItem">
        <i class="WorkflowGraph__legendDot WorkflowGraph__legendDot--on_success"></i>
        {{ $t('workflowConditionOnSuccess') }}
      </span>
      <span class="WorkflowGraph__legendItem">
        <i class="WorkflowGraph__legendDot WorkflowGraph__legendDot--on_failure"></i>
        {{ $t('workflowConditionOnFailure') }}
      </span>
      <span class="WorkflowGraph__legendItem">
        <i class="WorkflowGraph__legendDot WorkflowGraph__legendDot--always"></i>
        {{ $t('workflowConditionAlways') }}
      </span>
    </div>
  </div>
</template>

<script>
import Drawflow from 'drawflow';
import 'drawflow/dist/drawflow.min.css';
import { layoutWorkflowNodes, needsAutoLayout } from '@/lib/workflowLayout';

const CONDITION_DEFAULT = 'on_success';

export default {
  props: {
    // Workflow model — the editor's source of truth on load.
    nodes: { type: Array, default: () => [] },
    edges: { type: Array, default: () => [] },
    // Templates list to resolve template names for node labels.
    templates: { type: Array, default: () => [] },
    // editable=false renders a read-only run view (pan/zoom only).
    editable: { type: Boolean, default: false },
    // Map of nodeId -> run status string ('success'|'failed'|'running'|...).
    nodeStatuses: { type: Object, default: () => ({}) },
  },

  data() {
    return {
      editor: null,
      // condition keyed by `${sourceNodeId}->${destNodeId}`
      conditions: {},
      // guards re-entrancy while we mutate Drawflow programmatically
      syncing: false,
      built: false,
    };
  },

  watch: {
    // The edit case loads asynchronously: nodes arrive after mount. Rebuild once
    // the model is first populated. Subsequent edits are driven by Drawflow
    // events (canvas is the source of truth), so we do not rebuild on every change.
    nodes() {
      if (!this.built && this.editor && (this.nodes.length > 0 || this.edges.length > 0)) {
        this.buildCanvas();
      }
    },
    // Run view polls for fresh statuses; repaint nodes in place (status color +
    // active animation) without rebuilding, so the user's pan/zoom is preserved.
    nodeStatuses() {
      if (this.built && !this.editable) this.refreshStatuses();
    },
  },

  mounted() {
    const editor = new Drawflow(this.$refs.canvas);
    editor.reroute = false;
    editor.editor_mode = this.editable ? 'edit' : 'fixed';
    editor.start();
    this.editor = editor;

    if (this.editable) {
      editor.on('nodeMoved', () => this.emitChange());
      editor.on('nodeRemoved', () => this.emitChange());
      editor.on('connectionCreated', (e) => this.onConnectionCreated(e));
      editor.on('connectionRemoved', (e) => this.onConnectionRemoved(e));
      editor.on('connectionSelected', (e) => this.onConnectionSelected(e));
      editor.on('nodeSelected', (id) => this.$emit('node-selected', this.nodeIdOf(id)));
      editor.on('click', (ev) => this.onCanvasClick(ev));
    } else {
      // Drawflow's "fixed" mode does not fire nodeSelected on node clicks, so
      // detect node clicks via DOM delegation and emit node-selected ourselves.
      this.$refs.canvas.addEventListener('click', this.onReadonlyClick);
    }

    this.buildCanvas();
  },

  beforeDestroy() {
    if (this.$refs.canvas) {
      this.$refs.canvas.removeEventListener('click', this.onReadonlyClick);
    }
    if (this.editor) {
      this.editor.clear();
      this.editor = null;
    }
  },

  methods: {
    // ---- building the canvas from the model ----------------------------------

    buildCanvas() {
      if (!this.editor) return;
      this.syncing = true;
      try {
        this.editor.clear();
        this.conditions = {};
        const dfIdByNodeId = {};

        // Lay out nodes when no coordinates are stored (legacy workflows, or runs
        // of workflows that were never positioned in the editor) so they don't
        // all stack at the origin.
        const layout = needsAutoLayout(this.nodes)
          ? layoutWorkflowNodes(this.nodes, this.edges)
          : null;

        this.nodes.forEach((node) => {
          const pos = layout
            ? layout[node.id]
            : { x: node.position_x || 0, y: node.position_y || 0 };
          // Note nodes have no ports so they can not be connected on the canvas.
          const ports = node.kind === 'note' ? 0 : 1;
          const dfId = this.editor.addNode(
            `wf-${node.id}`,
            ports,
            ports,
            pos.x,
            pos.y,
            this.nodeClass(node),
            { nodeId: node.id, node: this.stripPosition(node) },
            this.nodeHtml(node),
            false,
          );
          dfIdByNodeId[node.id] = dfId;
        });

        this.edges.forEach((edge) => {
          const src = dfIdByNodeId[edge.source_node_id];
          const dst = dfIdByNodeId[edge.destination_node_id];
          if (src == null || dst == null) return;
          this.editor.addConnection(src, dst, 'output_1', 'input_1');
          const key = this.condKey(edge.source_node_id, edge.destination_node_id);
          this.conditions[key] = edge.condition || CONDITION_DEFAULT;
        });
        this.built = true;
      } finally {
        this.syncing = false;
      }
      this.$nextTick(() => this.decorateConnections());
    },

    // ---- model <-> Drawflow translation --------------------------------------

    exportModel() {
      const data = this.editor.export().drawflow.Home.data;
      const nodes = [];
      const edges = [];
      Object.keys(data).forEach((dfId) => {
        const dfNode = data[dfId];
        const nodeId = dfNode.data.nodeId;
        nodes.push({
          ...dfNode.data.node,
          id: nodeId,
          position_x: Math.round(dfNode.pos_x),
          position_y: Math.round(dfNode.pos_y),
        });
        const outputs = dfNode.outputs.output_1 ? dfNode.outputs.output_1.connections : [];
        outputs.forEach((conn) => {
          const destNodeId = data[conn.node] ? data[conn.node].data.nodeId : null;
          if (destNodeId == null) return;
          edges.push({
            source_node_id: nodeId,
            destination_node_id: destNodeId,
            condition: this.conditions[this.condKey(nodeId, destNodeId)] || CONDITION_DEFAULT,
          });
        });
      });
      return { nodes, edges };
    },

    emitChange() {
      if (this.syncing) return;
      this.$emit('change', this.exportModel());
      this.$nextTick(() => this.decorateConnections());
    },

    // ---- public API used by the editor page ----------------------------------

    addNode(kind, posX, posY) {
      const nodeId = this.nextNodeId();
      const node = kind === 'note'
        ? { id: nodeId, kind, note: '' }
        : {
          id: nodeId,
          kind,
          convergence_mode: 'all',
          template_id: null,
          limit: [],
        };
      // Note nodes have no ports so they can not be connected on the canvas.
      const ports = kind === 'note' ? 0 : 1;
      this.editor.addNode(
        `wf-${nodeId}`,
        ports,
        ports,
        posX,
        posY,
        this.nodeClass(node),
        { nodeId, node },
        this.nodeHtml(node),
        false,
      );
      this.emitChange();
      return nodeId;
    },

    // Re-render a node after its properties were edited in the side panel.
    syncNode(nodeId, node) {
      const dfId = this.dfIdOf(nodeId);
      if (dfId == null) return;
      this.editor.updateNodeDataFromId(dfId, { nodeId, node: this.stripPosition(node) });
      const el = this.$refs.canvas.querySelector(`#node-${dfId} .WorkflowGraph__node`);
      if (el) {
        const fresh = document.createElement('div');
        fresh.innerHTML = this.nodeHtml(node);
        el.outerHTML = fresh.innerHTML;
      }
      const wrapper = this.$refs.canvas.querySelector(`#node-${dfId}`);
      if (wrapper) wrapper.className = `drawflow-node ${this.nodeClass(node)}`;
      this.emitChange();
    },

    setCondition(sourceNodeId, destNodeId, condition) {
      this.conditions[this.condKey(sourceNodeId, destNodeId)] = condition;
      this.emitChange();
    },

    removeSelectedNode(nodeId) {
      const dfId = this.dfIdOf(nodeId);
      if (dfId != null) this.editor.removeNodeId(`node-${dfId}`);
    },

    zoomIn() { this.editor.zoom_in(); },
    zoomOut() { this.editor.zoom_out(); },
    // Reset both zoom AND pan. Drawflow's zoom_reset() only restores zoom=1 and
    // keeps the canvas translation, so zero canvas_x/canvas_y as well.
    zoomReset() {
      if (!this.editor) return;
      this.editor.canvas_x = 0;
      this.editor.canvas_y = 0;
      this.editor.zoom = 1;
      this.editor.zoom_last_value = 1;
      this.editor.zoom_refresh();
    },

    // ---- Drawflow event handlers ---------------------------------------------

    onConnectionCreated(e) {
      // Ignore connections we add programmatically while (re)building the canvas;
      // their conditions are set explicitly in buildCanvas().
      if (this.syncing) return;
      const source = this.nodeIdOf(e.output_id);
      const dest = this.nodeIdOf(e.input_id);

      // Live guard: no self-edge.
      if (source === dest) {
        this.removeConnection(e);
        this.$emit('blocked', 'self');
        return;
      }
      // Live guard: no cycle.
      if (this.wouldCreateCycle(source, dest)) {
        this.removeConnection(e);
        this.$emit('blocked', 'cycle');
        return;
      }

      this.conditions[this.condKey(source, dest)] = CONDITION_DEFAULT;
      this.emitChange();
    },

    onConnectionRemoved(e) {
      if (this.syncing) return;
      const source = this.nodeIdOf(e.output_id);
      const dest = this.nodeIdOf(e.input_id);
      delete this.conditions[this.condKey(source, dest)];
      this.emitChange();
    },

    onConnectionSelected(e) {
      const source = this.nodeIdOf(e.output_id);
      const dest = this.nodeIdOf(e.input_id);
      this.$emit('connection-selected', {
        source_node_id: source,
        destination_node_id: dest,
        condition: this.conditions[this.condKey(source, dest)] || CONDITION_DEFAULT,
      });
    },

    onCanvasClick(ev) {
      // Clicking empty canvas clears the selection in the side panel.
      if (ev && ev.target && ev.target.closest && !ev.target.closest('.drawflow-node')
          && !ev.target.closest('.connection')) {
        this.$emit('node-selected', null);
      }
    },

    // Read-only mode: emit node-selected when a node is clicked (see mounted()).
    onReadonlyClick(ev) {
      const el = ev.target && ev.target.closest ? ev.target.closest('.drawflow-node') : null;
      if (!el) return;
      const dfId = el.id.replace('node-', '');
      this.$emit('node-selected', this.nodeIdOf(dfId));
    },

    onDrop(ev) {
      if (!this.editable) return;
      ev.preventDefault();
      const kind = ev.dataTransfer.getData('node-kind');
      if (!kind) return;
      const { x, y } = this.canvasCoords(ev.clientX, ev.clientY);
      this.addNode(kind, x, y);
    },

    // ---- helpers --------------------------------------------------------------

    removeConnection(e) {
      this.syncing = true;
      try {
        this.editor.removeSingleConnection(e.output_id, e.input_id, e.output_class, e.input_class);
      } finally {
        this.syncing = false;
      }
    },

    // Translate viewport coordinates to canvas coordinates accounting for zoom/pan.
    canvasCoords(clientX, clientY) {
      const { precanvas } = this.editor;
      const zoom = this.editor.zoom;
      const rect = precanvas.getBoundingClientRect();
      const x = (clientX - rect.x) / zoom;
      const y = (clientY - rect.y) / zoom;
      return { x, y };
    },

    wouldCreateCycle(source, dest) {
      // Walk forward from `dest` over existing edges; a path back to `source`
      // means the new edge source->dest closes a cycle.
      const adjacency = {};
      const { edges } = this.exportModel();
      edges.forEach((edge) => {
        if (!adjacency[edge.source_node_id]) adjacency[edge.source_node_id] = [];
        adjacency[edge.source_node_id].push(edge.destination_node_id);
      });
      const stack = [dest];
      const seen = new Set();
      while (stack.length) {
        const cur = stack.pop();
        if (cur === source) return true;
        if (!seen.has(cur)) {
          seen.add(cur);
          (adjacency[cur] || []).forEach((n) => stack.push(n));
        }
      }
      return false;
    },

    nextNodeId() {
      const ids = [];
      const data = this.editor.export().drawflow.Home.data;
      Object.keys(data).forEach((dfId) => ids.push(data[dfId].data.nodeId || 0));
      return (ids.length === 0 ? 0 : Math.max(...ids)) + 1;
    },

    nodeIdOf(dfId) {
      const node = this.editor.getNodeFromId(dfId);
      return node ? node.data.nodeId : null;
    },

    dfIdOf(nodeId) {
      const data = this.editor.export().drawflow.Home.data;
      const found = Object.keys(data).find((dfId) => data[dfId].data.nodeId === nodeId);
      return found != null ? Number(found) : null;
    },

    condKey(source, dest) {
      return `${source}->${dest}`;
    },

    stripPosition(node) {
      const copy = { ...node };
      delete copy.position_x;
      delete copy.position_y;
      return copy;
    },

    templateName(id) {
      const t = this.templates.find((x) => x.id === id);
      if (t) return t.name;
      return id ? `#${id}` : this.$t('workflowNodeIncomplete');
    },

    nodeClass(node) {
      const kind = node.kind || 'task';
      const classes = [`WorkflowGraph__nodeWrap--${kind}`];
      const status = this.nodeStatuses[node.id];
      if (status) classes.push(`WorkflowGraph__nodeWrap--status-${status}`);
      return classes.join(' ');
    },

    nodeHtml(node) {
      const kind = node.kind || 'task';

      if (kind === 'note') {
        const text = node.note ? this.escape(node.note) : this.escape(this.$t('workflowNotePlaceholder'));
        const empty = node.note ? '' : ' WorkflowGraph__noteText--empty';
        return `
          <div class="WorkflowGraph__node WorkflowGraph__node--note">
            <div class="WorkflowGraph__noteHeader">
              <i class="mdi mdi-note-text-outline"></i>
            </div>
            <div class="WorkflowGraph__noteText${empty}">${text}</div>
          </div>`;
      }

      const isApproval = kind === 'approval';
      const icon = isApproval ? 'mdi-account-check' : 'mdi-cog';
      const title = isApproval
        ? this.$t('workflowNodeKindApproval')
        : this.escape(this.templateName(node.template_id));
      const status = this.nodeStatuses[node.id];
      const statusHtml = status
        ? `<span class="WorkflowGraph__nodeStatus WorkflowGraph__nodeStatus--${status}">${this.escape(status)}</span>`
        : '';
      return `
        <div class="WorkflowGraph__node WorkflowGraph__node--${kind}">
          <div class="WorkflowGraph__nodeHeader">
            <i class="mdi ${icon}"></i>
            <span class="WorkflowGraph__nodeId">#${node.id}</span>
            ${statusHtml}
          </div>
          <div class="WorkflowGraph__nodeTitle">${title}</div>
        </div>`;
    },

    escape(value) {
      return String(value == null ? '' : value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
    },

    // Color each connection's path by its condition.
    decorateConnections() {
      if (!this.editor) return;
      const data = this.editor.export().drawflow.Home.data;
      const dfIdByNodeId = {};
      Object.keys(data).forEach((dfId) => { dfIdByNodeId[data[dfId].data.nodeId] = dfId; });
      Object.keys(this.conditions).forEach((key) => {
        const [source, dest] = key.split('->').map(Number);
        const outId = dfIdByNodeId[source];
        const inId = dfIdByNodeId[dest];
        if (outId == null || inId == null) return;
        const conn = this.$refs.canvas.querySelector(
          `.connection.node_in_node-${inId}.node_out_node-${outId}`,
        );
        if (!conn) return;
        conn.classList.remove(
          'WorkflowGraph__conn--on_success',
          'WorkflowGraph__conn--on_failure',
          'WorkflowGraph__conn--always',
        );
        conn.classList.add(`WorkflowGraph__conn--${this.conditions[key]}`);
      });
    },

    // Repaint node status (color + active animation) in place, without
    // rebuilding the canvas, so polling the run view preserves pan/zoom.
    refreshStatuses() {
      if (!this.editor) return;
      const data = this.editor.export().drawflow.Home.data;
      Object.keys(data).forEach((dfId) => {
        const nodeId = data[dfId].data.nodeId;
        const node = { ...data[dfId].data.node, id: nodeId };
        const wrapper = this.$refs.canvas.querySelector(`#node-${dfId}`);
        if (!wrapper) return;
        wrapper.className = `drawflow-node ${this.nodeClass(node)}`;
        const inner = wrapper.querySelector('.WorkflowGraph__node');
        if (inner) {
          const fresh = document.createElement('div');
          fresh.innerHTML = this.nodeHtml(node);
          inner.outerHTML = fresh.innerHTML;
        }
      });
    },
  },
};
</script>

<style lang="scss">
.WorkflowGraph {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 480px;

  &__canvas {
    width: 100%;
    height: 100%;
    background:
      linear-gradient(90deg, rgba(133, 133, 133, 0.08) 1px, transparent 1px) 0 0 / 24px 24px,
      linear-gradient(rgba(133, 133, 133, 0.08) 1px, transparent 1px) 0 0 / 24px 24px;
  }

  &__legend {
    position: absolute;
    bottom: 8px;
    right: 12px;
    display: flex;
    gap: 12px;
    font-size: 11px;
    padding: 4px 8px;
    border-radius: 4px;
    background: rgba(127, 127, 127, 0.1);
    pointer-events: none;
  }

  &__legendItem { display: inline-flex; align-items: center; gap: 4px; }

  &__legendDot {
    width: 10px; height: 10px; border-radius: 50%; display: inline-block;
    &--on_success { background: #4caf50; }
    &--on_failure { background: #f44336; }
    &--always { background: #9e9e9e; }
  }

  // Node card
  .drawflow .drawflow-node {
    padding: 0;
    border-radius: 6px;
    border: 1px solid rgba(127, 127, 127, 0.4);
    background: var(--v-background-base, #fff);
    min-width: 170px;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
  }

  &__node {
    padding: 8px 12px;
    position: relative;
    overflow: hidden;
    border-radius: 6px;
  }

  &__nodeHeader {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    opacity: 0.7;
  }

  &__nodeId { font-weight: 600; }

  &__nodeTitle {
    margin-top: 2px;
    font-size: 13px;
    font-weight: 600;
    word-break: break-word;
  }

  &__nodeStatus {
    margin-left: auto;
    text-transform: uppercase;
    font-size: 10px;
    padding: 0 6px;
    border-radius: 8px;
    color: #fff;
    &--success, &--approved { background: #4caf50; }
    &--failed, &--error, &--stopped, &--rejected { background: #f44336; }
    &--running, &--pending, &--waiting { background: #2196f3; }
  }

  .drawflow-node.WorkflowGraph__nodeWrap--approval { border-left: 3px solid #ab47bc; }
  .drawflow-node.WorkflowGraph__nodeWrap--task { border-left: 3px solid #2196f3; }

  // Note node — sticky-note look, no execution status.
  .drawflow-node.WorkflowGraph__nodeWrap--note {
    background: #fff8c4;
    border-color: #e6d873;
    color: #5b5320;
  }

  &__noteHeader {
    font-size: 12px;
    opacity: 0.6;
  }

  &__noteText {
    margin-top: 2px;
    font-size: 13px;
    white-space: pre-wrap;
    word-break: break-word;

    &--empty {
      font-style: italic;
      opacity: 0.6;
    }
  }
  .drawflow-node.WorkflowGraph__nodeWrap--status-success { border-color: #4caf50; }
  .drawflow-node.WorkflowGraph__nodeWrap--status-failed,
  .drawflow-node.WorkflowGraph__nodeWrap--status-error,
  .drawflow-node.WorkflowGraph__nodeWrap--status-rejected { border-color: #f44336; }

  // Active node — Concourse-style: glowing pulse + moving diagonal stripes.
  .drawflow-node.WorkflowGraph__nodeWrap--status-running,
  .drawflow-node.WorkflowGraph__nodeWrap--status-waiting {
    border-color: #2196f3;
    animation: WorkflowGraphPulse 1.6s ease-out infinite;
  }
  .drawflow-node.WorkflowGraph__nodeWrap--status-running .WorkflowGraph__node::after,
  .drawflow-node.WorkflowGraph__nodeWrap--status-waiting .WorkflowGraph__node::after {
    content: '';
    position: absolute;
    inset: 0;
    background: repeating-linear-gradient(
      45deg,
      rgba(33, 150, 243, 0.16) 0 8px,
      transparent 8px 16px
    );
    background-size: 22px 22px;
    animation: WorkflowGraphStripes 0.7s linear infinite;
    pointer-events: none;
  }

  // Pending approval — awaiting user action: amber pulse.
  .drawflow-node.WorkflowGraph__nodeWrap--status-pending {
    border-color: #ff9800;
    animation: WorkflowGraphPulseAmber 1.6s ease-out infinite;
  }

  @keyframes WorkflowGraphPulse {
    0% { box-shadow: 0 0 0 0 rgba(33, 150, 243, 0.5); }
    70% { box-shadow: 0 0 0 10px rgba(33, 150, 243, 0); }
    100% { box-shadow: 0 0 0 0 rgba(33, 150, 243, 0); }
  }
  @keyframes WorkflowGraphPulseAmber {
    0% { box-shadow: 0 0 0 0 rgba(255, 152, 0, 0.5); }
    70% { box-shadow: 0 0 0 10px rgba(255, 152, 0, 0); }
    100% { box-shadow: 0 0 0 0 rgba(255, 152, 0, 0); }
  }
  @keyframes WorkflowGraphStripes {
    from { background-position: 0 0; }
    to { background-position: 22px 0; }
  }

  // Connection condition colors
  .connection.WorkflowGraph__conn--on_success .main-path { stroke: #4caf50; }
  .connection.WorkflowGraph__conn--on_failure .main-path { stroke: #f44336; }
  .connection.WorkflowGraph__conn--always .main-path { stroke: #9e9e9e; }
}
</style>
