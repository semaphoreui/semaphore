<template>
  <div
    class="WorkflowGraph"
    :class="{
      'WorkflowGraph--editable': editable,
      'WorkflowGraph--dark': $vuetify.theme.dark,
    }"
  >
    <div
      ref="canvas"
      class="WorkflowGraph__canvas"
      @drop="onDrop"
      @dragover.prevent
    ></div>

    <!--
      Edge decorations. This element is moved into Drawflow's canvas after
      start() (and again after every clear()), so the condition pills pan and
      zoom together with the nodes. Coordinates are canvas-space.
    -->
    <div ref="overlay" class="WorkflowGraph__overlay">
      <svg class="WorkflowGraph__defs" width="0" height="0" aria-hidden="true">
        <defs>
          <marker
            v-for="c in conditionNames"
            :key="c"
            :id="`wf-arrow-${c}`"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="5"
            markerHeight="5"
            orient="auto-start-reverse"
          >
            <path d="M0 1 L9 5 L0 9 z" :class="`WorkflowGraph__arrow WorkflowGraph__arrow--${c}`" />
          </marker>
        </defs>
      </svg>
      <WorkflowEdgeLabel
        v-for="label in edgeLabels"
        :key="label.key"
        :label="label"
        :editable="editable"
        @change="setCondition(label.source, label.dest, $event)"
        @remove="removeEdge(label.source, label.dest)"
      />
    </div>

    <WorkflowCanvasControls
      :zoom="zoom"
      :editable="editable"
      @zoom-in="zoomIn"
      @zoom-out="zoomOut"
      @fit="fitView"
      @tidy="tidyUp"
    />

    <v-btn
      v-if="editable"
      fab
      small
      depressed
      color="primary"
      class="WorkflowGraph__fab"
      :title="$t('workflowQuickAdd')"
      @click="openQuickAddAtCenter()"
    >
      <v-icon>mdi-plus</v-icon>
    </v-btn>

    <div v-if="editable && isEmpty" class="WorkflowGraph__empty">
      <v-icon size="40" class="WorkflowGraph__emptyIcon">mdi-sitemap-outline</v-icon>
      <div class="text-subtitle-1">{{ $t('workflowEmptyTitle') }}</div>
      <div class="text-body-2 text--secondary mb-3">{{ $t('workflowEmptyHint') }}</div>
      <v-btn color="primary" depressed small @click="openQuickAddAtCenter()">
        <v-icon left small>mdi-plus</v-icon>
        {{ $t('workflowQuickAdd') }}
      </v-btn>
    </div>

    <WorkflowQuickAddMenu
      v-model="quickAdd.open"
      :x="quickAdd.x"
      :y="quickAdd.y"
      :templates="templates"
      :hint="quickAdd.sourceNodeId != null ? $t('workflowQuickAddConnectHint') : null"
      @add="onQuickAdd"
    />
  </div>
</template>

<script>
import Vue from 'vue';
import Drawflow from 'drawflow';
import 'drawflow/dist/drawflow.min.css';
import WorkflowNodeCard from '@/components/workflow/WorkflowNodeCard.vue';
import WorkflowEdgeLabel from '@/components/workflow/WorkflowEdgeLabel.vue';
import WorkflowCanvasControls from '@/components/workflow/WorkflowCanvasControls.vue';
import WorkflowQuickAddMenu from '@/components/workflow/WorkflowQuickAddMenu.vue';
import AppsMixin from '@/components/AppsMixin';
import { layoutWorkflowNodes, needsAutoLayout } from '@/lib/workflowLayout';
import { wouldCreateCycle, nextNodeId, edgeRunState } from '@/lib/workflowGraph';
import {
  graphBounds, fitViewport, zoomAt, snapToGrid, clampZoom, wheelZoomFactor,
  ZOOM_MIN, ZOOM_MAX, ZOOM_STEP,
} from '@/lib/workflowViewport';

const CONDITIONS = ['on_success', 'on_failure', 'always'];
const CONDITION_DEFAULT = 'on_success';
const EDGE_STATES = ['active', 'passed', 'dim'];
const QUICK_ADD_OFFSET_X = 320;
const QUICK_ADD_ROW = 120;
const NODE_MOUNT_HTML = '<div class="WorkflowGraph__mount"></div>';
// After a build, layout shifts (app drawer, fonts) re-fit the graph for this long.
const AUTO_FIT_WINDOW_MS = 1500;

const NodeCard = Vue.extend(WorkflowNodeCard);

export default {
  components: { WorkflowEdgeLabel, WorkflowCanvasControls, WorkflowQuickAddMenu },
  mixins: [AppsMixin],

  props: {
    // Workflow model — the editor's source of truth on load.
    nodes: { type: Array, default: () => [] },
    edges: { type: Array, default: () => [] },
    // Templates list to resolve template names / app icons for node cards.
    templates: { type: Array, default: () => [] },
    // editable=false renders a read-only run view (pan/zoom only).
    editable: { type: Boolean, default: false },
    // Run view: nodeId -> { status, start, end, resumeAt, taskId, message }.
    nodeRuns: { type: Object, default: () => ({}) },
    // Editor: nodeId -> validation problem text (drawn as a badge).
    nodeProblems: { type: Object, default: () => ({}) },
    // Run view: show Approve / Reject on pending approval cards.
    canResolveApprovals: { type: Boolean, default: false },
  },

  data() {
    return {
      editor: null,
      // condition keyed by `${sourceNodeId}->${destNodeId}`
      conditions: {},
      // guards re-entrancy while we mutate Drawflow programmatically
      syncing: false,
      built: false,
      zoom: 1,
      // [{ key, source, dest, condition, x, y, state }] — condition pills
      edgeLabels: [],
      quickAdd: {
        open: false, x: 0, y: 0, sourceNodeId: null,
      },
      // Shared with every mounted node card; cards react to changes in place.
      store: {
        nodes: {},
        templates: [],
        runs: {},
        problems: {},
        apps: null,
        canResolveApprovals: false,
      },
      conditionNames: CONDITIONS,
    };
  },

  computed: {
    isEmpty() {
      return (this.nodes || []).length === 0;
    },
  },

  watch: {
    // The edit case loads asynchronously: nodes arrive after mount. Build once
    // the model is first populated. Subsequent edits are driven by Drawflow
    // events (canvas is the source of truth), so we do not rebuild on change.
    nodes() {
      if (!this.built && this.editor && (this.nodes.length > 0 || this.edges.length > 0)) {
        this.buildCanvas();
      }
    },
    templates: {
      immediate: true,
      handler(list) { this.store.templates = list || []; },
    },
    nodeRuns: {
      immediate: true,
      handler(runs) {
        this.store.runs = runs || {};
        this.$nextTick(() => this.refreshEdges());
      },
    },
    nodeProblems: {
      immediate: true,
      handler(problems) { this.store.problems = problems || {}; },
    },
    canResolveApprovals: {
      immediate: true,
      handler(v) { this.store.canResolveApprovals = v; },
    },
    'appsMixin.apps': function syncApps(apps) {
      this.store.apps = apps;
    },
  },

  created() {
    // Non-reactive bookkeeping: mounted card instances and id maps.
    this.cards = {};
    this.nodeIdByDf = {};
    this.labelRaf = null;
    this.resizeObserver = null;
    // Set once the user touches the canvas; stops the auto-fit on resize.
    this.userMovedViewport = false;
    this.autoFitUntil = 0;
  },

  mounted() {
    const editor = new Drawflow(this.$refs.canvas);
    editor.reroute = false;
    editor.curvature = 0.35;
    editor.zoom_min = ZOOM_MIN;
    editor.zoom_max = ZOOM_MAX;
    editor.editor_mode = this.editable ? 'edit' : 'fixed';
    editor.start();
    // Our viewport math assumes the canvas scales from its top-left corner.
    editor.precanvas.style.transformOrigin = '0 0';
    this.editor = editor;
    this.attachOverlay();

    editor.on('zoom', (zoom) => { this.zoom = zoom; });
    editor.on('translate', () => { this.userMovedViewport = true; });
    editor.on('nodeMoved', (dfId) => this.onNodeMoved(dfId));
    editor.on('mouseMove', () => {
      if (this.editor && (this.editor.drag || this.editor.connection)) {
        this.scheduleEdgeRefresh();
      }
    });

    if (this.editable) {
      editor.on('nodeRemoved', (dfId) => this.onNodeRemoved(dfId));
      editor.on('connectionCreated', (e) => this.onConnectionCreated(e));
      editor.on('connectionRemoved', (e) => this.onConnectionRemoved(e));
      editor.on('connectionSelected', (e) => this.onConnectionSelected(e));
      editor.on('nodeSelected', (dfId) => this.$emit('node-selected', this.nodeIdOf(dfId)));
      editor.on('nodeUnselected', () => this.$emit('node-selected', null));
      editor.on('click', (ev) => this.onCanvasClick(ev));
    }

    // Own wheel handling: plain wheel pans, Ctrl/Cmd (and trackpad pinch)
    // zooms towards the cursor. Registered in the capture phase so it runs
    // before Drawflow's ctrl-only zoom and can stop it.
    const canvas = this.$refs.canvas;
    canvas.addEventListener('wheel', this.onWheel, { passive: false, capture: true });
    canvas.addEventListener('contextmenu', this.onContextMenu, { capture: true });
    canvas.addEventListener('keydown', this.onKeyDown);

    // Vuetify lays the app drawer out a frame after mount, which changes the
    // canvas width; keep the graph fitted while the page settles, unless the
    // user has already touched the canvas. Later resizes (the properties
    // panel opening) leave the viewport alone.
    if (typeof ResizeObserver !== 'undefined') {
      this.resizeObserver = new ResizeObserver(() => {
        if (this.built && !this.userMovedViewport && Date.now() < this.autoFitUntil) {
          this.fitView();
        }
      });
      this.resizeObserver.observe(canvas);
    }
    canvas.addEventListener('pointerdown', this.onPointerDown);

    this.buildCanvas();
  },

  beforeDestroy() {
    const canvas = this.$refs.canvas;
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
      this.resizeObserver = null;
    }
    if (canvas) canvas.removeEventListener('pointerdown', this.onPointerDown);
    if (canvas) {
      canvas.removeEventListener('wheel', this.onWheel, { capture: true });
      canvas.removeEventListener('contextmenu', this.onContextMenu, { capture: true });
      canvas.removeEventListener('keydown', this.onKeyDown);
    }
    if (this.labelRaf) cancelAnimationFrame(this.labelRaf);
    this.destroyCards();
    if (this.editor) {
      this.editor.clear();
      this.editor = null;
    }
  },

  methods: {
    // ---- building the canvas from the model ----------------------------------

    buildCanvas(options = {}) {
      if (!this.editor) return;
      const { fit = true } = options;
      this.syncing = true;
      try {
        this.destroyCards();
        this.editor.clear();
        this.attachOverlay();
        this.conditions = {};
        this.edgeLabels = [];
        this.store.nodes = {};
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
          dfIdByNodeId[node.id] = this.createDrawflowNode(node, pos.x, pos.y);
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
      this.$nextTick(() => {
        this.refreshEdges();
        if (fit) {
          this.fitView();
          this.autoFitUntil = Date.now() + AUTO_FIT_WINDOW_MS;
        }
      });
    },

    // Rebuilds the canvas from the current props without touching the
    // viewport (undo/redo).
    reload() {
      this.buildCanvas({ fit: false });
    },

    // Drawflow's clear() wipes its canvas, overlay included; put it back.
    attachOverlay() {
      const overlay = this.$refs.overlay;
      if (this.editor && overlay && overlay.parentNode !== this.editor.precanvas) {
        this.editor.precanvas.appendChild(overlay);
      }
    },

    createDrawflowNode(node, x, y) {
      // Note nodes have no ports so they can not be connected on the canvas.
      const ports = node.kind === 'note' ? 0 : 1;
      const dfId = this.editor.addNode(
        `wf-${node.id}`,
        ports,
        ports,
        x,
        y,
        this.nodeClass(node),
        { nodeId: node.id, node: this.stripPosition(node) },
        NODE_MOUNT_HTML,
        false,
      );
      this.nodeIdByDf[dfId] = node.id;
      this.$set(this.store.nodes, node.id, this.stripPosition(node));
      this.mountCard(dfId, node.id);
      return dfId;
    },

    mountCard(dfId, nodeId) {
      const el = this.$refs.canvas.querySelector(`#node-${dfId} .WorkflowGraph__mount`);
      if (!el) return;
      const vm = new NodeCard({
        parent: this,
        propsData: { store: this.store, nodeId, editable: this.editable },
      });
      vm.$on('quick-add', () => this.openQuickAddFromNode(nodeId));
      vm.$on('select', () => this.$emit('node-click', nodeId));
      vm.$on('resolve-approval', (status) => this.$emit('resolve-approval', { nodeId, status }));
      vm.$mount(el);
      this.cards[dfId] = vm;
    },

    destroyCards() {
      Object.keys(this.cards).forEach((dfId) => this.cards[dfId].$destroy());
      this.cards = {};
      this.nodeIdByDf = {};
    },

    // ---- model <-> Drawflow translation --------------------------------------

    liveData() {
      return this.editor.drawflow.drawflow.Home.data;
    },

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
      this.$nextTick(() => this.refreshEdges());
    },

    // ---- public API used by the editor page ----------------------------------

    addNode(kind, posX, posY, extra = {}) {
      const nodeId = this.nextNodeId();
      const node = kind === 'note'
        ? { id: nodeId, kind, note: '' }
        : {
          id: nodeId,
          kind,
          convergence_mode: 'all',
          template_id: extra.template_id || null,
          // Approval and delay nodes must not carry task params (backend validation).
          task_params: kind === 'task' ? {} : null,
          delay_seconds: kind === 'delay' ? 60 : undefined,
        };
      this.createDrawflowNode(node, snapToGrid(posX), snapToGrid(posY));
      this.emitChange();
      return nodeId;
    },

    // Adds a node in the middle of the visible canvas (palette click).
    addNodeAtCenter(kind) {
      const rect = this.$refs.canvas.getBoundingClientRect();
      const { x, y } = this.canvasCoords(rect.left + rect.width / 2, rect.top + rect.height / 2);
      const nodeId = this.addNode(kind, x - 110, y - 32);
      this.selectNode(nodeId);
      return nodeId;
    },

    // Re-render a node after its properties were edited in the side panel.
    syncNode(nodeId, node) {
      const dfId = this.dfIdOf(nodeId);
      if (dfId == null) return;
      this.editor.updateNodeDataFromId(dfId, { nodeId, node: this.stripPosition(node) });
      this.$set(this.store.nodes, nodeId, this.stripPosition(node));
      const wrapper = this.$refs.canvas.querySelector(`#node-${dfId}`);
      if (wrapper) {
        const selected = wrapper.classList.contains('selected');
        wrapper.className = `drawflow-node ${this.nodeClass(node)}${selected ? ' selected' : ''}`;
      }
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

    removeEdge(sourceNodeId, destNodeId) {
      const out = this.dfIdOf(sourceNodeId);
      const inp = this.dfIdOf(destNodeId);
      if (out == null || inp == null) return;
      // Dispatches connectionRemoved, which drops the condition and emits change.
      this.editor.removeSingleConnection(String(out), String(inp), 'output_1', 'input_1');
      this.$emit('connection-selected', null);
    },

    // Programmatic selection, mirrored into Drawflow so Delete works.
    selectNode(nodeId) {
      const dfId = this.dfIdOf(nodeId);
      const el = dfId != null ? this.$refs.canvas.querySelector(`#node-${dfId}`) : null;
      if (!el) return;
      this.clearDrawflowSelection();
      this.editor.node_selected = el;
      el.classList.add('selected');
      this.$emit('node-selected', nodeId);
    },

    clearSelection() {
      this.clearDrawflowSelection();
      this.$emit('node-selected', null);
      this.$emit('connection-selected', null);
    },

    clearDrawflowSelection() {
      const { editor } = this;
      if (!editor) return;
      if (editor.node_selected) {
        editor.node_selected.classList.remove('selected');
        editor.node_selected = null;
      }
      if (editor.connection_selected) {
        editor.connection_selected.classList.remove('selected');
        editor.connection_selected = null;
      }
    },

    // ---- viewport -------------------------------------------------------------

    getViewport() {
      const { editor } = this;
      return { x: editor.canvas_x, y: editor.canvas_y, zoom: editor.zoom };
    },

    applyViewport(viewport) {
      const { editor } = this;
      if (!editor) return;
      const zoom = clampZoom(viewport.zoom);
      editor.zoom = zoom;
      editor.zoom_last_value = zoom;
      editor.canvas_x = viewport.x;
      editor.canvas_y = viewport.y;
      editor.precanvas.style.transform = `translate(${viewport.x}px, ${viewport.y}px) scale(${zoom})`;
      this.zoom = zoom;
    },

    containerSize() {
      const rect = this.$refs.canvas.getBoundingClientRect();
      return { width: rect.width, height: rect.height };
    },

    fitView() {
      if (!this.editor) return;
      const data = this.liveData();
      const nodes = [];
      const sizes = {};
      Object.keys(data).forEach((dfId) => {
        const id = data[dfId].data.nodeId;
        nodes.push({ id, x: data[dfId].pos_x, y: data[dfId].pos_y });
        const el = this.$refs.canvas.querySelector(`#node-${dfId}`);
        if (el) sizes[id] = { width: el.offsetWidth, height: el.offsetHeight };
      });
      const bounds = graphBounds(nodes, sizes);
      if (!bounds) {
        this.applyViewport({ x: 0, y: 0, zoom: 1 });
        return;
      }
      this.applyViewport(fitViewport(bounds, this.containerSize()));
    },

    zoomBy(factor) {
      const size = this.containerSize();
      const center = { x: size.width / 2, y: size.height / 2 };
      const vp = this.getViewport();
      this.applyViewport(zoomAt(vp, center, vp.zoom * factor));
    },

    zoomIn() {
      this.userMovedViewport = true;
      this.zoomBy(ZOOM_STEP);
    },
    zoomOut() {
      this.userMovedViewport = true;
      this.zoomBy(1 / ZOOM_STEP);
    },
    // Kept for callers of the previous API.
    zoomReset() { this.fitView(); },

    onWheel(ev) {
      if (!this.editor) return;
      ev.preventDefault();
      ev.stopImmediatePropagation();
      this.userMovedViewport = true;
      const vp = this.getViewport();
      if (ev.ctrlKey || ev.metaKey) {
        const rect = this.$refs.canvas.getBoundingClientRect();
        const point = { x: ev.clientX - rect.left, y: ev.clientY - rect.top };
        this.applyViewport(zoomAt(vp, point, vp.zoom * wheelZoomFactor(ev.deltaY)));
      } else {
        this.applyViewport({ ...vp, x: vp.x - ev.deltaX, y: vp.y - ev.deltaY });
      }
    },

    onPointerDown() {
      this.userMovedViewport = true;
    },

    onKeyDown(ev) {
      if (ev.key === 'Escape') {
        this.quickAdd.open = false;
        if (this.editable) this.clearSelection();
      }
    },

    onContextMenu(ev) {
      // Drawflow draws its own delete button on right click; replace it with
      // the quick-add menu on empty canvas.
      ev.preventDefault();
      ev.stopImmediatePropagation();
      if (!this.editable) return;
      if (ev.target.closest('.drawflow-node') || ev.target.closest('.connection')) return;
      this.openQuickAddAt(ev.clientX, ev.clientY);
    },

    // ---- node movement, snapping, tidy up ------------------------------------

    onNodeMoved(dfId) {
      const data = this.liveData()[dfId];
      if (!data) return;
      const x = snapToGrid(data.pos_x);
      const y = snapToGrid(data.pos_y);
      if (x !== data.pos_x || y !== data.pos_y) this.moveNode(dfId, x, y);
      this.emitChange();
    },

    moveNode(dfId, x, y) {
      const data = this.liveData()[dfId];
      const el = this.$refs.canvas.querySelector(`#node-${dfId}`);
      if (!data || !el) return;
      data.pos_x = x;
      data.pos_y = y;
      el.style.left = `${x}px`;
      el.style.top = `${y}px`;
      this.editor.updateConnectionNodes(`node-${dfId}`);
    },

    tidyUp() {
      const model = this.exportModel();
      if (model.nodes.length === 0) return;
      const layout = layoutWorkflowNodes(model.nodes, model.edges);
      model.nodes.forEach((node) => {
        const dfId = this.dfIdOf(node.id);
        if (dfId != null) this.moveNode(dfId, layout[node.id].x, layout[node.id].y);
      });
      this.emitChange();
      this.$nextTick(() => this.fitView());
    },

    // ---- quick add ------------------------------------------------------------

    openQuickAddAt(clientX, clientY, sourceNodeId = null) {
      this.quickAdd = {
        open: true, x: clientX, y: clientY, sourceNodeId,
      };
    },

    openQuickAddFromNode(nodeId) {
      const dfId = this.dfIdOf(nodeId);
      const el = dfId != null ? this.$refs.canvas.querySelector(`#node-${dfId}`) : null;
      if (!el) return;
      const rect = el.getBoundingClientRect();
      this.openQuickAddAt(rect.right + 12, rect.top, nodeId);
    },

    openQuickAddAtCenter() {
      const rect = this.$refs.canvas.getBoundingClientRect();
      this.openQuickAddAt(rect.left + rect.width / 2 - 130, rect.top + rect.height / 2 - 160);
    },

    onQuickAdd({ kind, template_id: templateId }) {
      const source = this.quickAdd.sourceNodeId;
      let x;
      let y;
      if (source != null && this.dfIdOf(source) != null) {
        const from = this.liveData()[this.dfIdOf(source)];
        x = from.pos_x + QUICK_ADD_OFFSET_X;
        y = from.pos_y;
        // Step down while the slot to the right is already taken.
        const taken = (px, py) => Object.values(this.liveData())
          .some((d) => Math.abs(d.pos_x - px) < 200 && Math.abs(d.pos_y - py) < 60);
        while (taken(x, y)) y += QUICK_ADD_ROW;
      } else {
        const c = this.canvasCoords(this.quickAdd.x, this.quickAdd.y);
        x = c.x;
        y = c.y;
      }
      const nodeId = this.addNode(kind, x, y, { template_id: templateId });
      if (source != null && kind !== 'note') this.connectNodes(source, nodeId);
      this.quickAdd.open = false;
      this.selectNode(nodeId);
    },

    connectNodes(sourceNodeId, destNodeId) {
      if (wouldCreateCycle(this.exportModel().edges, sourceNodeId, destNodeId)) {
        this.$emit('blocked', 'cycle');
        return;
      }
      const out = this.dfIdOf(sourceNodeId);
      const inp = this.dfIdOf(destNodeId);
      if (out == null || inp == null) return;
      this.syncing = true;
      try {
        this.editor.addConnection(out, inp, 'output_1', 'input_1');
      } finally {
        this.syncing = false;
      }
      this.conditions[this.condKey(sourceNodeId, destNodeId)] = CONDITION_DEFAULT;
      this.emitChange();
    },

    // ---- Drawflow event handlers ---------------------------------------------

    onNodeRemoved(dfId) {
      const vm = this.cards[dfId];
      if (vm) {
        vm.$destroy();
        delete this.cards[dfId];
      }
      const nodeId = this.nodeIdByDf[dfId];
      delete this.nodeIdByDf[dfId];
      if (nodeId != null) {
        this.$delete(this.store.nodes, nodeId);
        Object.keys(this.conditions).forEach((key) => {
          const [s, d] = key.split('->').map(Number);
          if (s === nodeId || d === nodeId) delete this.conditions[key];
        });
      }
      this.emitChange();
    },

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
        this.$emit('connection-selected', null);
      }
    },

    onDrop(ev) {
      if (!this.editable) return;
      ev.preventDefault();
      const kind = ev.dataTransfer.getData('node-kind');
      if (!kind) return;
      const { x, y } = this.canvasCoords(ev.clientX, ev.clientY);
      const nodeId = this.addNode(kind, x, y);
      this.selectNode(nodeId);
    },

    // ---- edge decoration -------------------------------------------------------

    scheduleEdgeRefresh() {
      if (this.labelRaf) return;
      this.labelRaf = requestAnimationFrame(() => {
        this.labelRaf = null;
        this.refreshEdges();
      });
    },

    // Colors each connection by its condition, adds the arrow head, marks the
    // run state, and positions the condition pill on the path's midpoint.
    refreshEdges() {
      if (!this.editor) return;
      const data = this.liveData();
      const labels = [];
      Object.keys(data).forEach((dfOut) => {
        const out = data[dfOut];
        const source = out.data.nodeId;
        const conns = out.outputs.output_1 ? out.outputs.output_1.connections : [];
        conns.forEach((conn) => {
          const dfIn = conn.node;
          const dest = data[dfIn] ? data[dfIn].data.nodeId : null;
          if (dest == null) return;
          const key = this.condKey(source, dest);
          const condition = this.conditions[key] || CONDITION_DEFAULT;
          const svg = this.$refs.canvas.querySelector(
            `.connection.node_in_node-${dfIn}.node_out_node-${dfOut}`,
          );
          if (!svg) return;

          const edge = { source_node_id: source, destination_node_id: dest, condition };
          const state = this.editable ? null : edgeRunState(edge, this.store.runs);

          CONDITIONS.forEach((c) => svg.classList.remove(`WorkflowGraph__conn--${c}`));
          EDGE_STATES.forEach((s) => svg.classList.remove(`WorkflowGraph__conn--${s}`));
          svg.classList.add(`WorkflowGraph__conn--${condition}`);
          if (state) svg.classList.add(`WorkflowGraph__conn--${state}`);

          const path = svg.querySelector('path.main-path');
          if (!path) return;
          path.setAttribute('marker-end', `url(#wf-arrow-${condition})`);

          let point;
          try {
            point = path.getPointAtLength(path.getTotalLength() / 2);
          } catch (err) {
            return;
          }
          labels.push({
            key, source, dest, condition, x: point.x, y: point.y, state,
          });
        });
      });
      this.edgeLabels = labels;
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
      const { zoom } = this.editor;
      const rect = precanvas.getBoundingClientRect();
      return { x: (clientX - rect.x) / zoom, y: (clientY - rect.y) / zoom };
    },

    wouldCreateCycle(source, dest) {
      return wouldCreateCycle(this.exportModel().edges, source, dest);
    },

    nextNodeId() {
      const data = this.liveData();
      return nextNodeId(Object.keys(data).map((dfId) => data[dfId].data.nodeId || 0));
    },

    nodeIdOf(dfId) {
      const node = this.editor.getNodeFromId(dfId);
      return node ? node.data.nodeId : null;
    },

    dfIdOf(nodeId) {
      const found = Object.keys(this.nodeIdByDf).find((dfId) => this.nodeIdByDf[dfId] === nodeId);
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

    nodeClass(node) {
      return `WorkflowGraph__nodeWrap--${node.kind || 'task'}`;
    },
  },
};
</script>

<style lang="scss">
@import '~@/components/workflow/palette';

.WorkflowGraph {
  --wf-surface: #{$wf-surface};
  --wf-text: #{$wf-text};
  --wf-text2: #{$wf-text2};
  --wf-divider: #{$wf-divider};
  --wf-divider-strong: #{$wf-divider-strong};
  --wf-dot: #{$wf-dot};
  --wf-tile-bg: #{$wf-tile-bg};
  --wf-primary: #{$wf-primary};
  --wf-primary-ring: #{$wf-primary-ring};
  --wf-success: #{$wf-success};
  --wf-error: #{$wf-error};
  --wf-warning: #{$wf-warning};
  --wf-always: #{$wf-always};
  --wf-note-bg: #{$wf-note-bg};
  --wf-note-border: #{$wf-note-border};
  --wf-note-text: #{$wf-note-text};
  --wf-shadow: #{$wf-shadow};
  --wf-shadow-hover: #{$wf-shadow-hover};

  position: relative;
  width: 100%;
  height: 100%;
  min-height: 480px;
  overflow: hidden;

  &--dark {
    --wf-surface: #{$wf-surface-dark};
    --wf-text: #{$wf-text-dark};
    --wf-text2: #{$wf-text2-dark};
    --wf-divider: #{$wf-divider-dark};
    --wf-divider-strong: #{$wf-divider-strong-dark};
    --wf-dot: #{$wf-dot-dark};
    --wf-tile-bg: #{$wf-tile-bg-dark};
    --wf-primary: #{$wf-primary-dark};
    --wf-primary-ring: #{$wf-primary-ring-dark};
    --wf-note-bg: #{$wf-note-bg-dark};
    --wf-note-border: #{$wf-note-border-dark};
    --wf-note-text: #{$wf-note-text-dark};
    --wf-shadow: #{$wf-shadow-dark};
    --wf-shadow-hover: #{$wf-shadow-hover-dark};
  }

  &__canvas {
    width: 100%;
    height: 100%;
    outline: none;
    background-image: radial-gradient(circle, var(--wf-dot) 1px, transparent 1px);
    background-size: $wf-grid $wf-grid;
  }

  &__overlay {
    position: absolute;
    left: 0;
    top: 0;
    width: 0;
    height: 0;
    overflow: visible;
    z-index: 1;
    pointer-events: none;
  }

  &__defs {
    position: absolute;
  }

  &__arrow {
    &--on_success { fill: var(--wf-success); }
    &--on_failure { fill: var(--wf-error); }
    &--always { fill: var(--wf-always); }
  }

  &__fab {
    position: absolute;
    top: 12px;
    right: 12px;
    z-index: 4;
  }

  &__empty {
    position: absolute;
    inset: 0;
    z-index: 3;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    pointer-events: none;

    .v-btn {
      pointer-events: auto;
    }
  }

  &__emptyIcon {
    opacity: 0.5;
    margin-bottom: 4px;
  }

  // ---- Drawflow overrides ----------------------------------------------------

  .drawflow {
    // Drawflow's zoom origin (we scale from the top-left corner).
    transform-origin: 0 0;
  }

  .drawflow .drawflow-node {
    display: block;
    position: absolute;
    width: auto;
    min-height: 0;
    padding: 0;
    border: 0;
    border-radius: $wf-card-radius;
    background: transparent;
    box-shadow: none;
    color: inherit;
    z-index: 2;

    &.selected {
      background: transparent;
    }

    // Selection ring lives on the card, driven by Drawflow's `selected` class.
    &.selected .WorkflowNodeCard {
      border-color: var(--wf-primary);
      box-shadow: 0 0 0 4px var(--wf-primary-ring), var(--wf-shadow-hover);
    }

    &.selected .WorkflowNodeCard__plus {
      opacity: 1;
    }

    .drawflow_content_node {
      display: block;
      width: auto;
    }

    // Ports: 10 px circles centered on the card's left / right edge.
    .inputs,
    .outputs {
      position: absolute;
      top: 50%;
      width: 0;
      height: 0;
    }

    .inputs { left: 0; }
    .outputs { right: 0; }

    .input,
    .output {
      position: absolute;
      left: -5px;
      top: -5px;
      width: 10px;
      height: 10px;
      margin: 0;
      border-radius: 50%;
      background: var(--wf-surface);
      border: 2px solid var(--wf-divider-strong);
      cursor: crosshair;
      z-index: 3;
      transition: border-color 0.12s ease, transform 0.12s ease;

      // Larger hit area for starting / dropping a connection.
      &::before {
        content: '';
        position: absolute;
        inset: -7px;
        border-radius: 50%;
      }

      &:hover {
        border-color: var(--wf-primary);
        transform: scale(1.3);
      }
    }

    &:hover .input,
    &:hover .output,
    &.selected .output {
      border-color: var(--wf-primary);
    }
  }

  // Run view: cards with a task open its log on click.
  &:not(.WorkflowGraph--editable) .drawflow .drawflow-node {
    cursor: default;
  }

  .drawflow .connection {
    z-index: 0;
  }

  .drawflow .connection .main-path {
    fill: none;
    stroke: var(--wf-always);
    stroke-width: 2px;
    stroke-linecap: round;
    transition: stroke-width 0.1s ease;

    &:hover,
    &.selected {
      stroke-width: 3px;
      cursor: pointer;
    }
  }

  // Connection being dragged out of a port (has no node_in_ class yet).
  .drawflow .connection:not([class*="node_in_"]) .main-path {
    stroke: var(--wf-primary);
    stroke-dasharray: 6 4;
  }

  .drawflow .connection.WorkflowGraph__conn--on_success .main-path { stroke: var(--wf-success); }
  .drawflow .connection.WorkflowGraph__conn--on_failure .main-path { stroke: var(--wf-error); }
  .drawflow .connection.WorkflowGraph__conn--always .main-path { stroke: var(--wf-always); }

  // Run view edge states.
  .drawflow .connection.WorkflowGraph__conn--dim {
    opacity: 0.4;
  }

  .drawflow .connection.WorkflowGraph__conn--active .main-path {
    stroke-dasharray: 6 5;
    animation: WorkflowGraphDash 1s linear infinite;
  }

  @keyframes WorkflowGraphDash {
    to { stroke-dashoffset: -22; }
  }
}
</style>
