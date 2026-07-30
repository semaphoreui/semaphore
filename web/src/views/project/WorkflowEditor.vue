<template>
  <div>
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
        v-if="!isNew && canRun"
        color="primary"
        outlined
        class="mr-3"
        :disabled="saving"
        @click="runWorkflow()"
        data-testid="workflow-run"
      >{{ $t('run') }}
      </v-btn
      >

      <v-btn
        color="primary"
        :disabled="!canManage || saving || problems.length > 0"
        :loading="saving"
        @click="save()"
      >{{ $t('save') }}
      </v-btn
      >
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
              <div class="text-subtitle-2 mb-2">{{ $t('workflowEditorPalette') }}</div>
              <div class="text-caption text--secondary mb-2">
                {{ $t('workflowDragToCanvasHint') }}
              </div>
            </template>
            <div
              class="WorkflowEditor__paletteItem WorkflowEditor__paletteItem--task"
              draggable="true"
              :title="sideCollapsed ? $t('workflowPaletteTaskNode') : null"
              @dragstart="onDragStart($event, 'task')"
            >
              <v-icon small :left="!sideCollapsed">mdi-cog</v-icon>
              <template v-if="!sideCollapsed">{{ $t('workflowPaletteTaskNode') }}</template>
            </div>
            <div
              class="WorkflowEditor__paletteItem WorkflowEditor__paletteItem--approval"
              draggable="true"
              :title="sideCollapsed ? $t('workflowPaletteApprovalNode') : null"
              @dragstart="onDragStart($event, 'approval')"
            >
              <v-icon small :left="!sideCollapsed">mdi-account-check</v-icon>
              <template v-if="!sideCollapsed">{{ $t('workflowPaletteApprovalNode') }}</template>
            </div>
            <div
              class="WorkflowEditor__paletteItem WorkflowEditor__paletteItem--note"
              draggable="true"
              :title="sideCollapsed ? $t('workflowPaletteNoteNode') : null"
              @dragstart="onDragStart($event, 'note')"
            >
              <v-icon small :left="!sideCollapsed">mdi-note-text-outline</v-icon>
              <template v-if="!sideCollapsed">{{ $t('workflowPaletteNoteNode') }}</template>
            </div>
          </div>

          <v-divider />

          <div class="pa-3">
            <template v-if="!sideCollapsed">
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
              >{{ p }}
              </v-alert
              >
            </template>
            <div v-else class="d-flex flex-column align-center">
              <v-tooltip
                v-if="problems.length === 0"
                right
                max-width="320"
                transition="fade-transition"
              >
                <template v-slot:activator="{ on, attrs }">
                  <v-icon color="success" v-bind="attrs" v-on="on">mdi-check-circle</v-icon>
                </template>
                <span>{{ $t('workflowValidationPassed') }}</span>
              </v-tooltip>
              <template v-else>
                <v-tooltip
                  v-for="(p, i) in problems"
                  :key="`problem-icon-${i}`"
                  right
                  max-width="320"
                  transition="fade-transition"
                >
                  <template v-slot:activator="{ on, attrs }">
                    <v-icon color="warning" class="mb-1" v-bind="attrs" v-on="on">
                      mdi-alert
                    </v-icon>
                  </template>
                  <span>{{ p }}</span>
                </v-tooltip>
              </template>
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
        <div v-if="editingNode" class="pa-3">
          <div class="d-flex align-center mb-2">
            <span class="text-subtitle-2"> {{ $t('workflowNodeId') }} #{{ editingNode.id }} </span>
            <v-spacer />
            <v-btn icon small :disabled="!canManage" @click="deleteSelectedNode()">
              <v-icon small>mdi-delete</v-icon>
            </v-btn>
          </div>

          <v-textarea
            v-if="editingNode.kind === 'note'"
            v-model="editingNode.note"
            :label="$t('workflowNoteText')"
            :placeholder="$t('workflowNotePlaceholder')"
            :disabled="!canManage"
            outlined
            dense
            auto-grow
            rows="4"
            hide-details="auto"
            @input="applyNodeEdit"
          />

          <template v-if="editingNode.kind !== 'note'">
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
              class="mb-5"
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
              class="mb-5"
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
              class="mb-5"
              @change="applyNodeEdit"
            />

            <v-card
              v-if="editingNode.kind !== 'approval' && editingNodeTemplate"
              :key="`task-params-${editingNode.id}-${editingNode.template_id}`"
              style="background: rgba(133, 133, 133, 0.06)"
              class="mb-6 pt-3"
            >

              <div style="
                position: absolute;
                background: var(--highlighted-card-bg-color);
                width: 28px;
                height: 28px;
                transform: rotate(45deg);
                left: calc(50% - 14px);
                top: -14px;
                border-radius: 0;
              "></div>

              <v-card-text class="py-0">
                <TaskParamsForm
                  :template="editingNodeTemplate"
                  v-model="editingNode.task_params"
                  class="mt-2"
                  @input="applyNodeEdit"
                />
              </v-card-text>
            </v-card>

            <template v-if="editingNode.kind === 'approval'">
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
import TaskParamsForm from '@/components/TaskParamsForm.vue';
import WorkflowGraph from '@/components/WorkflowGraph.vue';
import ProjectMixin from '@/components/ProjectMixin';
import PermissionsCheck from '@/components/PermissionsCheck';
import { USER_PERMISSIONS } from '@/lib/constants';
import { layoutWorkflowNodes, needsAutoLayout } from '@/lib/workflowLayout';

export default {
  components: { TaskParamsForm, WorkflowGraph },
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
    canRun() {
      return this.can(USER_PERMISSIONS.runProjectTasks);
    },
    editingNodeTemplate() {
      if (!this.editingNode || !this.editingNode.template_id) return null;
      return this.templates.find((t) => t.id === this.editingNode.template_id) || null;
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
      // Note nodes are annotations: excluded from the run graph (root count) and
      // from the task-completeness check.
      const roots = nodes.filter((n) => n.kind !== 'note' && !incoming[n.id]).length;
      if (roots === 0) out.push(this.$t('workflowErrorNoRoot'));
      else if (roots > 1) out.push(this.$t('workflowErrorMultipleRoots', { count: roots }));

      const incompleteTask = nodes.some((n) => (n.kind || 'task') === 'task' && !n.template_id);
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
      if (this.skipNextRouteReload) {
        this.skipNextRouteReload = false;
        return;
      }
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
    // graphical editor (all coordinates 0) so the computed layout persists on
    // the next save. Uses the same algorithm as the read-only run view.
    autoLayout() {
      const nodes = this.item.nodes;
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
      const clone = node ? JSON.parse(JSON.stringify(node)) : null;
      // Define task_params up front so later assignments stay reactive.
      // Approval/note nodes must not carry task params (backend validation).
      if (clone && clone.task_params === undefined) {
        clone.task_params = (clone.kind || 'task') === 'task' ? {} : null;
      }
      this.editingNode = clone;
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
        this.editingNode.task_params = null;
      } else {
        this.editingNode.approval_timeout = null;
        this.editingNode.approval_message = null;
        if (!this.editingNode.task_params) this.editingNode.task_params = {};
      }
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

    // ---- run ------------------------------------------------------------------
    // Starts the last-saved version of the workflow (same endpoint as the
    // Run buttons on the list and view pages) and opens the run page.
    async runWorkflow() {
      try {
        const run = (await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/workflows/${this.workflowId}/run`,
          responseType: 'json',
        })).data;

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.$t('workflowRunStarted'),
        });

        await this.$router.push(
          `/project/${this.projectId}/workflows/${this.workflowId}/runs/${run.id}`,
        );
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
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
          this.skipNextRouteReload = true;
          this.$router.replace(`/project/${this.projectId}/workflows/${created.id}/edit`);
        } else {
          await axios.put(`/api/project/${this.projectId}/workflows/${this.workflowId}`, payload);
          EventBus.$emit('i-snackbar', { color: 'success', text: this.$t('workflowSaved') });
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

  // Inline name editor: looks like a title, but the pencil + dashed underline +
  // hover/focus affordances make it clear it is editable. The field auto-sizes
  // to the width of its text via the grid-sizer trick below.
  &__nameWrap {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    max-width: 60vw;
    padding: 2px 8px;
    border-radius: 4px;
    background: rgba(127, 127, 127, 0.12);
    //border-bottom: 2px solid transparent;
    transition: background-color 0.15s ease,
    border-color 0.15s ease;

    &:hover {
      background: rgba(127, 127, 127, 0.12);
      //border-bottom-color: rgba(127, 127, 127, 0.9);
    }

    &:focus-within {
      background: rgba(127, 127, 127, 0.12);
      //border-bottom: 2px solid var(--v-primary-base, #1976d2);
    }

    &--disabled {
      pointer-events: none;
      //border-bottom-style: solid;
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
    background: var(--v-background-base, #fff);
    cursor: pointer;
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

    &--note {
      border-left: 3px solid #e6d873;
    }
  }

  &__side--collapsed &__paletteItem {
    padding-left: 8px;
  }
}
</style>
