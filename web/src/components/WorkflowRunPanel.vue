<template>
  <transition name="WorkflowRunPanel-slide">
    <v-sheet
      v-if="modelValue && workflowId != null && runId != null"
      class="WorkflowRunPanel"
      :class="{ 'WorkflowRunPanel--resizing': resizing }"
      :style="{ width: panelWidth }"
      data-testid="workflowRunPanel"
      :data-workflow-id="workflowId"
      :data-run-id="runId"
    >
      <div
        v-if="!fullWidth"
        class="WorkflowRunPanel__resizer"
        data-testid="workflowRunPanel-resize"
        @pointerdown.prevent="startResize($event)"
        @pointermove="resize($event)"
        @pointerup="endResize()"
        @pointercancel="endResize()"
        @dblclick="resetWidth()"
      ></div>

      <!-- Another run (or the same ids in another project) is another view: its graph, its
           polling and its state start anew. -->
      <WorkflowRunView
        :key="`${projectId}/${workflowId}/${runId}`"
        panel
        :project-id="projectId"
        :workflow-id="workflowId"
        :run-id="runId"
        :user-permissions="userPermissions"
        :is-admin="isAdmin"
        @close="close()"
      />
    </v-sheet>
  </transition>
</template>

<style lang="scss">
// Above the page, its app bar and the navigation drawer, below every overlay of Vuetify (2000 and
// up): the task dialog opened from a node of the run lies over the panel.
.WorkflowRunPanel.v-sheet {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 1900;
  min-width: min(400px, 100vw);
  max-width: 100vw;
  display: flex;
  flex-direction: column;
  border-left: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  box-shadow: -8px 0 24px rgba(0, 0, 0, 0.2);
}

// The left edge the panel is dragged by. It lies outside the panel but for 2 px, as the edges of
// a resizable dialog do.
.WorkflowRunPanel__resizer {
  position: absolute;
  top: 0;
  bottom: 0;
  left: -10px;
  width: 12px;
  cursor: ew-resize;
  touch-action: none;
  z-index: 1;
}

// No text of the page gets selected while the edge is dragged over it.
.WorkflowRunPanel--resizing {
  user-select: none;
}

.WorkflowRunPanel-slide-enter-active,
.WorkflowRunPanel-slide-leave-active {
  transition: transform 0.2s ease;
}

.WorkflowRunPanel-slide-enter-from,
.WorkflowRunPanel-slide-leave-to {
  transform: translateX(100%);
}
</style>

<script>
import WorkflowRunView from '@/components/WorkflowRunView.vue';

// The narrowest the panel can be dragged to: the default width of a dialog.
const MIN_WIDTH = 400;

const WIDTH_KEY = 'WorkflowRunPanel__width';

// A workflow run in a panel that slides in from the right, over the page the link to the run was
// on. The page stays in use: there is no scrim, and a click on another run swaps the run of the
// panel. Half of the window wide until the user drags its left edge; a double click on the edge
// brings the half back. The owner opens it (`modelValue`, `workflowId`, `runId`) and closes it on
// `close`.
export default {
  components: { WorkflowRunView },

  props: {
    modelValue: Boolean,
    projectId: Number,
    workflowId: Number,
    runId: Number,
    userPermissions: Number,
    isAdmin: Boolean,
  },

  emits: ['close'],

  data() {
    return {
      // The width the user dragged the panel to; null until then (half of the window).
      width: null,
      resizing: false,
    };
  },

  computed: {
    // A narrow screen has no room for a page and a panel side by side.
    fullWidth() {
      return this.$vuetify.display.smAndDown;
    },

    panelWidth() {
      if (this.fullWidth) {
        return '100%';
      }
      return this.width == null ? '50%' : `${this.width}px`;
    },
  },

  watch: {
    modelValue(val) {
      this.listenToKeys(val);
    },
  },

  created() {
    // The drag in progress: where the pointer went down and what the panel measured then.
    this.resizeOrigin = null;
    const width = parseInt(localStorage.getItem(WIDTH_KEY), 10);
    this.width = width >= MIN_WIDTH ? width : null;
  },

  mounted() {
    this.listenToKeys(this.modelValue);
  },

  beforeUnmount() {
    this.listenToKeys(false);
  },

  methods: {
    close() {
      this.$emit('close');
    },

    listenToKeys(on) {
      window.removeEventListener('keydown', this.onKeyDown);
      if (on) {
        window.addEventListener('keydown', this.onKeyDown);
      }
    },

    // Escape closes the panel, unless it belongs to something above it: a dialog or a menu.
    onKeyDown(ev) {
      if (ev.key !== 'Escape' || ev.defaultPrevented) {
        return;
      }
      if (document.querySelector('.v-overlay--active:not(.v-tooltip)')) {
        return;
      }
      this.close();
    },

    startResize(ev) {
      this.resizeOrigin = {
        x: ev.clientX,
        width: ev.currentTarget.parentElement.getBoundingClientRect().width,
        maxWidth: window.innerWidth,
      };
      this.resizing = true;

      // The edge moves away from under the pointer when the panel cannot follow it (the limits
      // below); the capture keeps the events coming to the handle.
      ev.currentTarget.setPointerCapture(ev.pointerId);
    },

    resize(ev) {
      if (!this.resizeOrigin) {
        return;
      }

      const { x, width, maxWidth } = this.resizeOrigin;

      // The right edge is fixed to the window: the panel grows by what the pointer went left.
      const dragged = width + (x - ev.clientX);
      this.width = Math.round(Math.max(Math.min(dragged, maxWidth), MIN_WIDTH));
    },

    endResize() {
      if (!this.resizeOrigin) {
        return;
      }

      this.resizeOrigin = null;
      this.resizing = false;
      if (this.width != null) {
        localStorage.setItem(WIDTH_KEY, String(this.width));
      }
    },

    resetWidth() {
      this.width = null;
      localStorage.removeItem(WIDTH_KEY);
    },
  },
};
</script>
