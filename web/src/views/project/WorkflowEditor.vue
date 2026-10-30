<template>
  <div
    class="WorkflowEditor"
    data-testid="workflowEditor-page"
    :data-state="item == null || templates == null ? 'loading' : 'ready'"
  >
    <YesNoDialog
      v-model="leaveDialog"
      :title="$t('workflowUnsavedChanges')"
      :text="$t('workflowUnsavedLeave')"
      :yes-button-title="$t('workflowLeaveWithoutSaving')"
      @yes="confirmLeave()"
      @no="cancelLeave()"
      test-id="workflowEditorLeaveDialog"
    />

    <v-toolbar flat class="WorkflowEditor__toolbar">
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="workflowEditor-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title
        class="WorkflowEditor__title d-flex align-center"
        data-testid="workflowEditor-title"
      >
        <router-link
          :to="`/project/${projectId}/workflows`"
          data-testid="workflowEditor-backToList"
        >
          {{ $t('workflows') }}
        </router-link>
        <span class="mx-2">/</span>
        <span
          v-if="item != null"
          class="WorkflowEditor__nameWrap"
          :class="{ 'WorkflowEditor__nameWrap--disabled': !canManage }"
        >
          <v-icon size="small" class="WorkflowEditor__nameIcon">mdi-pencil</v-icon>
          <span class="WorkflowEditor__nameSizer" :data-value="item.name || $t('newWorkflow')">
            <input
              v-model="item.name"
              :placeholder="$t('newWorkflow')"
              :disabled="!canManage"
              :aria-label="$t('name')"
              size="1"
              class="WorkflowEditor__nameField"
              data-testid="workflowEditor-name"
            />
          </span>
        </span>
        <v-chip
          v-if="item != null && item.revision"
          size="x-small"
          variant="outlined"
          class="WorkflowEditor__chip ml-3"
          :title="$t('workflowRevisionHint')"
          data-testid="workflowEditor-revision"
        >{{ $t('workflowRevisionLabel', { number: item.revision }) }}</v-chip>
        <v-tooltip v-if="activeRuns > 0" location="bottom">
          <template v-slot:activator="{ props }">
            <v-chip
              size="x-small"
              variant="outlined"
              color="warning"
              class="WorkflowEditor__chip ml-2"
              v-bind="props"
              data-testid="workflowEditor-activeRuns"
            >
              <v-icon size="x-small" start>mdi-play-circle-outline</v-icon>
              {{ $t('workflowActiveRunsCount', { count: activeRuns }, activeRuns) }}
            </v-chip>
          </template>
          <span
            data-testid="workflowEditor-activeRunsHint"
          >{{ $t('workflowActiveRunsHint') }}</span>
        </v-tooltip>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn
        icon
        variant="text"
        :disabled="!canUndo"
        :title="$t('workflowUndo')"
        @click="undo()"
        data-testid="workflowEditor-undo"
      >
        <v-icon>mdi-undo</v-icon>
      </v-btn>
      <v-btn
        icon
        variant="text"
        :disabled="!canRedo"
        :title="$t('workflowRedo')"
        class="mr-2"
        @click="redo()"
        data-testid="workflowEditor-redo"
      >
        <v-icon>mdi-redo</v-icon>
      </v-btn>

      <!-- Vuetify 2 opened the menu under the chip and pushed it in to 12 px from the right edge
           of the window (always so next to Save); Vuetify 4 flips it to end at the chip instead.
           Placed at the end of Save (16 px from the edge) 4 px further right and 6 px up, at the
           bottom of the chip. At most 80 % of the window wide (`.v-menu__content` of
           Vuetify 2). -->
      <v-menu
        v-if="problems.length > 0"
        location="bottom end"
        target=".WorkflowEditor__save"
        :offset="[-6, 4]"
        max-width="80%"
        :close-on-content-click="true"
      >
        <template v-slot:activator="{ props }">
          <v-chip
            size="small"
            variant="outlined"
            color="warning"
            class="WorkflowEditor__chip mr-4"
            v-bind="props"
            data-testid="workflowEditor-problemsMenu"
          >
            <v-icon start size="small">mdi-alert</v-icon>
            {{ $t('workflowProblemsCount', { count: problems.length }, problems.length) }}
          </v-chip>
        </template>
        <v-list
          density="compact"
          class="WorkflowEditor__problems"
          data-testid="workflowEditor-problemsMenu-list"
        >
          <v-list-item
            v-for="(p, i) in problems"
            :key="`problem-${i}`"
            @click="focusProblem(p)"
            data-testid="workflowEditor-problemsMenu-item"
          >
            <template v-slot:prepend>
              <v-icon size="small" color="warning">mdi-alert</v-icon>
            </template>
            <v-list-item-title>{{ p.text }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
      <v-chip
        v-else
        size="small"
        variant="outlined"
        color="success"
        class="WorkflowEditor__chip mr-4"
        data-testid="workflowEditor-validationPassed"
      >
        <v-icon start size="small">mdi-check</v-icon>
        {{ $t('workflowValidationPassed') }}
      </v-chip>

      <v-badge
        :model-value="dirty"
        dot
        color="warning"
        offset-x="0"
        offset-y="0"
        data-testid="workflowEditor-unsavedBadge"
      >
        <v-btn
          color="primary"
          variant="flat"
          class="WorkflowEditor__save"
          :disabled="!canManage || saving || problems.length > 0"
          :loading="saving"
          @click="save()"
          data-testid="workflowEditor-save"
        >{{ $t('save') }}</v-btn>
      </v-badge>
    </v-toolbar>

    <v-divider />

    <div
      class="WorkflowEditor__body"
      v-if="item != null && templates != null"
      data-testid="workflowEditor-body"
    >
      <!-- Palette + meta -->
      <div
        class="WorkflowEditor__side WorkflowEditor__side--left"
        :class="{ 'WorkflowEditor__side--collapsed': sideCollapsed }"
        data-testid="workflowEditor-side"
      >
        <button
          type="button"
          class="WorkflowEditor__sideToggle"
          :title="$t(sideCollapsed ? 'workflowSidebarExpand' : 'workflowSidebarCollapse')"
          @click="toggleSide()"
          data-testid="workflowEditor-sideToggle"
        >
          <v-icon size="small">
            {{ sideCollapsed ? 'mdi-chevron-right' : 'mdi-chevron-left' }}
          </v-icon>
        </button>

        <div class="WorkflowEditor__sideScroll">
          <template v-if="!sideCollapsed">
            <div class="pa-3">
              <v-text-field
                v-model="item.start_version"
                :label="$t('startVersion')"
                persistent-hint
                :disabled="!canManage"
                variant="outlined"
                density="compact"
                data-testid="workflowEditor-startVersion"
              />

              <SurveyVars
                :key="`survey-${graphKey}`"
                :vars="item.survey_vars"
                :allow-secret="false"
                @change="setSurveyVars"
              />
            </div>

            <v-divider />
          </template>

          <div class="pa-3">
            <template v-if="!sideCollapsed">
              <div
                class="text-label-large mb-1"
                data-testid="workflowEditor-paletteTitle"
              >{{ $t('workflowEditorPalette') }}</div>
              <div
                class="text-body-small text-medium-emphasis mb-3"
                data-testid="workflowEditor-paletteHint"
              >
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
              :data-testid="`workflowEditor-palette-${p.kind}`"
            >
              <span
                class="WorkflowEditor__paletteTile"
                :class="`WorkflowEditor__paletteTile--${p.kind}`"
              >
                <v-icon size="small" :color="p.color">{{ p.icon }}</v-icon>
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
          data-testid="workflowEditor-graph"
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
        data-testid="workflowEditor-props"
      >
        <WorkflowNodeProperties
          v-if="editingNode"
          :key="`node-${editingNode.id}`"
          :model-value="editingNode"
          :templates="templates"
          :survey-vars="item.survey_vars"
          :can-manage="canManage"
          @update:model-value="applyNodeEdit"
          @delete="deleteSelectedNode()"
          @close="closePanel()"
        />
        <WorkflowEdgeProperties
          v-else
          :key="edgeKey(editingEdge)"
          :model-value="editingEdge"
          :nodes="item.nodes"
          :templates="templates"
          :output-hints="outputHints"
          :can-manage="canManage"
          @update:model-value="applyEdgeEdit"
          @delete="deleteSelectedEdge()"
          @close="closePanel()"
        />
      </div>
    </div>

    <div v-else class="pa-4 text-center" data-testid="workflowEditor-loading">
      <!-- No track under the spinning arc: Vuetify 2 drew none for an indeterminate one -->
      <v-progress-circular indeterminate color="primary" bg-color="transparent" />
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import WorkflowGraph from '@/components/WorkflowGraph.vue';
import WorkflowNodeProperties from '@/components/workflow/WorkflowNodeProperties.vue';
import SurveyVars from '@/components/SurveyVars.vue';
import WorkflowEdgeProperties from '@/components/workflow/WorkflowEdgeProperties.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ProjectMixin from '@/components/ProjectMixin';
import PermissionsCheck from '@/components/PermissionsCheck';
import { USER_PERMISSIONS } from '@/lib/constants';
import { layoutWorkflowNodes, needsAutoLayout } from '@/lib/workflowLayout';
import WorkflowHistory from '@/lib/workflowHistory';
import { readSideCollapsed, writeSideCollapsed } from '@/lib/workflowEditorPrefs';

function isTypingTarget(target) {
  if (!target) return false;
  const tag = (target.tagName || '').toLowerCase();
  return tag === 'input' || tag === 'textarea' || tag === 'select' || target.isContentEditable;
}

export default {
  components: {
    WorkflowGraph, WorkflowNodeProperties, WorkflowEdgeProperties, YesNoDialog, SurveyVars,
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
      // Collapses the palette and, through App.vue, the main navigation to
      // icon-only strips. Remembered per browser; App restores the navigation
      // when the editor is left (see beforeUnmount).
      sideCollapsed: readSideCollapsed(),
      canUndo: false,
      canRedo: false,
      // Runs still in progress: they keep executing the revision they started
      // from, so saving is safe — the chip in the toolbar just says so.
      activeRuns: 0,
      // Outputs of the latest finished run of the loaded revision, keyed by
      // node id, for the key suggestions of the edge panel. Null when no such
      // run exists — node ids change with every save, so a run of an older
      // revision can not be matched to the nodes on the canvas.
      outputHints: null,
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
        survey_vars: this.item.survey_vars || [],
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

      // The workflow survey: unique names (db.ValidateWorkflowSurveyVars).
      const workflowVarNames = new Set();
      (this.item.survey_vars || []).forEach((v) => {
        if (workflowVarNames.has(v.name)) {
          out.push({ text: this.$t('workflowErrorSurveyVarDuplicate', { name: v.name }) });
        }
        workflowVarNames.add(v.name);
      });

      nodes.forEach((n) => {
        const kind = n.kind || 'task';
        if (kind === 'task' && !n.template_id) {
          out.push({ text: this.$t('workflowErrorTaskNeedsTemplate'), nodeId: n.id });
        }
        if (kind === 'task' && Array.isArray(n.survey_inputs) && n.survey_inputs.length > 0) {
          // A binding refers to names: the template's survey variable it feeds
          // and the workflow variable it reads. Either may have been renamed or
          // deleted since (validateWorkflowNodeSurveyInputs on the server).
          const tpl = (this.templates || []).find((t) => t.id === n.template_id);
          const templateVars = tpl
            ? new Set((tpl.survey_vars || []).filter((v) => v.type !== 'secret').map((v) => v.name))
            : null;
          const seen = new Set();
          n.survey_inputs.forEach((b) => {
            if (seen.has(b.var)) {
              out.push({ text: this.$t('workflowErrorSurveyInputDuplicate', { name: b.var, node: n.id }), nodeId: n.id });
            }
            seen.add(b.var);
            if (templateVars && !templateVars.has(b.var)) {
              out.push({ text: this.$t('workflowErrorSurveyInputUnknownVar', { name: b.var, node: n.id }), nodeId: n.id });
            }
            if (!workflowVarNames.has(b.workflow_var)) {
              out.push({ text: this.$t('workflowErrorSurveyInputUnknownWorkflowVar', { name: b.workflow_var, node: n.id }), nodeId: n.id });
            }
          });
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
    sideCollapsed(val) {
      writeSideCollapsed(val);
      this.setNavMini(val);
    },
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
    // The editor fills the viewport exactly; drop the always-on page
    // scrollbar Vuetify puts on <html>, its empty track shows as a strip.
    document.documentElement.classList.add('WorkflowEditor-html');
    this.setNavMini(this.sideCollapsed);
    this.templates = await this.loadProjectResources('templates');
    await this.loadData();
  },
  beforeUnmount() {
    window.removeEventListener('keydown', this.onWindowKeyDown);
    window.removeEventListener('beforeunload', this.onBeforeUnload);
    document.documentElement.classList.remove('WorkflowEditor-html');
    // The collapsed navigation is an editor-only state; other pages get the
    // full drawer back regardless of what is stored.
    this.setNavMini(false);
  },
  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },
    toggleSide() {
      this.sideCollapsed = !this.sideCollapsed;
    },
    setNavMini(mini) {
      EventBus.$emit('i-nav-mini', { mini });
    },
    // Informational only (see activeRuns, outputHints); a failure must not
    // block editing.
    async loadActiveRuns() {
      try {
        const runs = await this.loadEndpoint(
          `/api/project/${this.projectId}/workflows/${this.workflowId}/runs`,
        );
        this.activeRuns = (runs || [])
          .filter((r) => r.status === 'running' || r.status === 'approval').length;
        await this.loadOutputHints(runs || []);
      } catch (err) {
        this.activeRuns = 0;
        this.outputHints = null;
      }
    },
    // The latest finished run of the current revision tells which outputs each
    // task node produced; the edge panel offers them as keys.
    async loadOutputHints(runs) {
      this.outputHints = null;
      const revisionId = this.item?.revision_id;
      if (!revisionId) return;
      const candidates = runs
        .filter((r) => r.revision_id === revisionId && r.status !== 'running' && r.status !== 'approval')
        .sort((a, b) => b.id - a.id);
      if (candidates.length === 0) return;
      const run = candidates[0];
      const details = await this.loadEndpoint(
        `/api/project/${this.projectId}/workflows/${this.workflowId}/runs/${run.id}`,
      );
      const byNode = {};
      (details?.nodes || []).forEach((n) => {
        const raw = n.task?.artifacts;
        if (!raw) return;
        let doc;
        try {
          doc = typeof raw === 'string' ? JSON.parse(raw) : raw;
        } catch (err) {
          return;
        }
        const values = doc && typeof doc === 'object' && doc.values && typeof doc.values === 'object'
          ? doc.values : {};
        const skipped = doc && typeof doc === 'object' && doc.skipped && typeof doc.skipped === 'object'
          ? doc.skipped : {};
        const keys = Object.keys(values).sort();
        if (keys.length === 0 && Object.keys(skipped).length === 0) return;
        byNode[n.node.id] = { keys, skipped };
      });
      this.outputHints = { runId: run.id, byNode };
    },
    getNewItem() {
      return {
        name: '',
        description: '',
        survey_vars: [],
        nodes: [],
        edges: [],
      };
    },
    setSurveyVars(vars) {
      this.item.survey_vars = [...vars];
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
          if (!Array.isArray(this.item.survey_vars)) this.item.survey_vars = [];
          this.item.nodes = this.item.nodes.map((node) => ({
            kind: 'task',
            convergence_mode: 'all',
            position_x: 0,
            position_y: 0,
            ...node,
          }));
          // The same shape the canvas emits (see WorkflowGraph.liveModel), so
          // the unsaved-changes snapshot only differs after a real edit: the
          // API also returns row ids and the default input mode.
          this.item.edges = this.item.edges.map((edge) => ({
            source_node_id: edge.source_node_id,
            destination_node_id: edge.destination_node_id,
            condition: edge.condition || 'on_success',
            ...(edge.input_mode === 'explicit'
              ? { input_mode: 'explicit', input_mappings: edge.input_mappings || [] }
              : {}),
          }));
          this.autoLayout();
          this.loadActiveRuns();
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
    applyEdgeEdit(edge) {
      if (!this.editingEdge || !this.$refs.graph) return;
      if (edge) this.editingEdge = { ...edge };
      const { source_node_id: s, destination_node_id: d } = this.editingEdge;
      // Coalesce keystrokes in the mapping table into a single undo step.
      this.pendingHistoryKey = `edge-${s}-${d}`;
      this.$refs.graph.setEdge(s, d, this.editingEdge);
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
          // No reload: keep the canvas and the selected element intact. Every
          // save appends a revision, so bump the label without a round trip.
          this.item.revision = (this.item.revision || 0) + 1;
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
    // Vuetify 3 lets the title shrink to nothing (flex: 1 1, min-width: 0); as in
    // Vuetify 2 it keeps the width of its content.
    flex: 0 1 auto;
    min-width: auto;

    // Vuetify 3 wraps the title slot in a placeholder that clips (ellipsis) and is not
    // a flex box: keep the link, the name editor and the chips in one centred row.
    :deep(.v-toolbar-title__placeholder) {
      display: flex;
      align-items: center;
      overflow: visible;
    }
  }

  // The toolbar does not clip what does not fit (a phone: the chips and Save past the right edge),
  // as in Vuetify 2 (`.v-toolbar__content` without overflow); the toolbar has the layout
  // containment of Vuetify 2 (`.v-toolbar { contain: layout }`), so what sticks out is painted but
  // does not widen the page.
  &__toolbar {
    contain: layout;

    :deep(.v-toolbar__content) {
      overflow: visible;
    }
  }

  // The chips of the toolbar as in Vuetify 2 (`.v-chip`): x-small 16 px high, small 24 px, a
  // 20 px line, the weight of the text around them (the revision chip sits in the bold title); the
  // icon at its own size (x-small 12 px, small 16 px), 6 px out to the left and 6 px to the text
  // (`$chip-icon-margin-*`); an outlined chip without a colour has the thin border of the theme
  // (`.theme--light.v-chip`: dividers), not one in the text colour.
  &__chip.v-chip {
    line-height: 20px;
    font-weight: inherit;

    &.v-chip--size-x-small {
      --v-chip-height: 16px;
    }

    &.v-chip--size-small {
      --v-chip-height: 24px;
    }

    :deep(.v-icon--start) {
      --v-icon-size-multiplier: 1;
      margin-inline: -6px 6px;
    }

    &.v-chip--variant-outlined:not([class*="text-"]) {
      border-color: rgba(var(--v-border-color), var(--v-border-opacity));
    }
  }

  // The problems menu: the text 24 px after the icon (`v-list-item-icon` of Vuetify 2), not 16.
  &__problems {
    --v-list-prepend-gap: 24px;
  }

  // Icons without a colour of their own in the icon colour of Vuetify 2 (`.theme--light.v-icon`:
  // .54, white in the dark theme), not in the text colour; a coloured one keeps its colour (an
  // inline style).
  &__nameIcon,
  &__paletteTile > .v-icon,
  &__sideToggle > .v-icon {
    color: rgba(0, 0, 0, .54);
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
    background: white;

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
    padding: 0;
  }

  &__side--collapsed &__paletteTile {
    width: 33px;
    min-width: 33px;
    height: 33px;
    border-radius: 9px;
  }
}

.v-theme--dark {
  .WorkflowEditor__sideToggle {
    background: #1e1e1e;
  }

  .WorkflowEditor__nameIcon,
  .WorkflowEditor__paletteTile > .v-icon,
  .WorkflowEditor__sideToggle > .v-icon {
    color: #fff;
  }
}

</style>
