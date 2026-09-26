<template>
  <div
    v-if="isNote"
    class="WorkflowNodeCard WorkflowNodeCard--note"
    @pointerdown="onPointerDown"
    @pointerup="onPointerUp"
  >
    <v-icon x-small class="WorkflowNodeCard__noteIcon">mdi-note-text-outline</v-icon>
    <div
      class="WorkflowNodeCard__noteText"
      :class="{ 'WorkflowNodeCard__noteText--empty': !node.note }"
    >{{ node.note || $t('workflowNotePlaceholder') }}</div>
  </div>

  <div
    v-else
    class="WorkflowNodeCard"
    :class="stateClasses"
    :aria-label="title"
    @pointerdown="onPointerDown"
    @pointerup="onPointerUp"
  >
    <div class="WorkflowNodeCard__tile" :class="`WorkflowNodeCard__tile--${kind}`">
      <v-icon :color="tileColor">{{ icon }}</v-icon>
    </div>

    <div class="WorkflowNodeCard__body">
      <div
        class="WorkflowNodeCard__title"
        :class="{ 'WorkflowNodeCard__title--placeholder': titleIsPlaceholder }"
        :title="title"
      >{{ title }}</div>
      <div class="WorkflowNodeCard__sub" :title="subtitle">{{ subtitle }}</div>

      <div v-if="showApprovalActions" class="WorkflowNodeCard__actions">
        <v-btn
          x-small
          depressed
          color="success"
          @mousedown.native.stop
          @click.stop="resolve('approved')"
        >{{ $t('workflowApprove') }}</v-btn>
        <v-btn
          x-small
          outlined
          color="error"
          class="ml-1"
          @mousedown.native.stop
          @click.stop="resolve('rejected')"
        >{{ $t('workflowReject') }}</v-btn>
      </div>
    </div>

    <div v-if="statusIcon" class="WorkflowNodeCard__status">
      <v-progress-circular
        v-if="statusIcon === 'spinner'"
        indeterminate
        size="14"
        width="2"
        color="primary"
      />
      <v-icon v-else small :color="statusColor">{{ statusIcon }}</v-icon>
    </div>

    <v-tooltip v-if="problem" bottom max-width="280">
      <template v-slot:activator="{ on, attrs }">
        <span class="WorkflowNodeCard__badge" v-bind="attrs" v-on="on">
          <v-icon x-small color="white">mdi-alert</v-icon>
        </span>
      </template>
      <span>{{ problem }}</span>
    </v-tooltip>

    <button
      v-if="editable"
      type="button"
      class="WorkflowNodeCard__plus"
      :title="$t('workflowQuickAdd')"
      @mousedown.stop
      @click.stop="$emit('quick-add', nodeId)"
    >
      <v-icon small color="primary">mdi-plus</v-icon>
    </button>
  </div>
</template>

<script>
import { APP_ICONS, APP_SHORT_TITLE } from '@/lib/constants';
import { formatDurationLong, statusKind } from '@/lib/workflowGraph';

const TAP_THRESHOLD_PX = 4;

// Rendered inside a Drawflow node. All data comes through the graph's
// reactive `store` so the card repaints in place when statuses change.
export default {
  props: {
    store: { type: Object, required: true },
    nodeId: { type: Number, required: true },
    editable: { type: Boolean, default: false },
  },

  data() {
    return {
      now: Date.now(),
      timer: null,
      pointerDown: null,
    };
  },

  computed: {
    node() {
      return this.store.nodes[this.nodeId] || {};
    },
    kind() {
      return this.node.kind || 'task';
    },
    isNote() {
      return this.kind === 'note';
    },
    template() {
      if (!this.node.template_id) return null;
      return (this.store.templates || []).find((t) => t.id === this.node.template_id) || null;
    },
    run() {
      return (this.store.runs || {})[this.nodeId] || null;
    },
    status() {
      return this.run ? this.run.status : null;
    },
    statusKind() {
      return statusKind(this.status);
    },
    problem() {
      return (this.store.problems || {})[this.nodeId] || null;
    },
    app() {
      return this.template ? this.template.app : null;
    },
    icon() {
      if (this.kind === 'approval') return 'mdi-account-check';
      if (this.kind === 'delay') return 'mdi-timer-outline';
      const apps = this.store.apps || {};
      if (this.app && apps[this.app] && apps[this.app].icon) {
        return `mdi-${apps[this.app].icon}`;
      }
      if (this.app && APP_ICONS[this.app]) return APP_ICONS[this.app].icon;
      return 'mdi-cog';
    },
    tileColor() {
      if (this.kind === 'approval') return '#ab47bc';
      if (this.kind === 'delay') return '#ff9800';
      const apps = this.store.apps || {};
      if (this.app && apps[this.app] && apps[this.app].color) return apps[this.app].color;
      if (this.app && APP_ICONS[this.app]) {
        const dark = this.$vuetify && this.$vuetify.theme.dark;
        return (dark ? APP_ICONS[this.app].darkColor : APP_ICONS[this.app].color) || null;
      }
      return null;
    },
    appTitle() {
      const apps = this.store.apps || {};
      if (this.app && apps[this.app] && apps[this.app].title) return apps[this.app].title;
      if (this.app && APP_SHORT_TITLE[this.app]) return APP_SHORT_TITLE[this.app];
      return this.app || '';
    },
    titleIsPlaceholder() {
      return this.kind === 'task' && !this.template;
    },
    title() {
      if (this.kind === 'approval') return this.$t('workflowNodeKindApproval');
      if (this.kind === 'delay') return this.$t('workflowNodeKindDelay');
      if (this.template) return this.template.name;
      return this.node.template_id
        ? `#${this.node.template_id}`
        : this.$t('workflowNodeIncomplete');
    },
    subtitle() {
      // Read-only canvas = run view: describe what happened, not the config.
      if (!this.editable) return this.runSubtitle;
      return this.configSubtitle;
    },
    configSubtitle() {
      const parts = [];
      if (this.kind === 'task') {
        if (this.appTitle) parts.push(this.appTitle);
        if (this.hasParamOverrides) parts.push(this.$t('workflowParamsOverridden'));
      } else if (this.kind === 'approval') {
        parts.push(this.node.approval_timeout
          ? this.$t('workflowApprovalTimeoutShort', {
            time: formatDurationLong(this.node.approval_timeout * 1000),
          })
          : this.$t('workflowApprovalNoTimeout'));
      } else if (this.kind === 'delay') {
        parts.push(this.$t('workflowDelayWait', {
          time: formatDurationLong((this.node.delay_seconds || 0) * 1000),
        }));
      }
      if (this.node.convergence_mode === 'any') parts.push(this.$t('workflowConvergenceAny'));
      return parts.join(' · ');
    },
    runSubtitle() {
      if (!this.run || !this.status) return this.$t('workflowNodeNotStarted');
      if (this.kind === 'delay' && this.status === 'waiting' && this.run.resumeAt) {
        const remainingMs = new Date(this.run.resumeAt).getTime() - this.now;
        if (remainingMs <= 0) return this.$t('workflowDelayElapsed');
        return this.$t('workflowDelayRemaining', { time: formatDurationLong(remainingMs) });
      }
      if (this.kind === 'approval' && this.status === 'pending') {
        return this.run.message || this.$t('workflowApprovalPending');
      }
      const parts = [this.status];
      if (this.run.start) {
        const end = this.run.end ? new Date(this.run.end).getTime() : this.now;
        parts.push(formatDurationLong(end - new Date(this.run.start).getTime()));
      }
      return parts.join(' · ');
    },
    hasParamOverrides() {
      const params = this.node.task_params;
      if (!params || typeof params !== 'object') return false;
      return Object.keys(params).some((k) => {
        const v = params[k];
        if (v == null || v === '') return false;
        if (Array.isArray(v)) return v.length > 0;
        if (typeof v === 'object') return Object.keys(v).length > 0;
        return true;
      });
    },
    statusIcon() {
      if (!this.run) return null;
      if (this.kind === 'delay' && this.status === 'waiting') return 'mdi-timer-sand';
      switch (this.statusKind) {
        case 'success': return 'mdi-check-circle';
        case 'error': return 'mdi-close-circle';
        case 'running': return 'spinner';
        case 'waiting': return this.status === 'waiting_confirmation' ? 'mdi-pause-circle' : 'mdi-clock-outline';
        case 'pending': return 'mdi-account-clock';
        default: return null;
      }
    },
    statusColor() {
      if (this.kind === 'delay' && this.status === 'waiting') return '#ff9800';
      switch (this.statusKind) {
        case 'success': return 'success';
        case 'error': return 'error';
        case 'pending': return 'warning';
        case 'waiting': return this.status === 'waiting_confirmation' ? 'warning' : 'primary';
        default: return 'primary';
      }
    },
    isActive() {
      if (!this.run) return false;
      if (this.kind === 'delay' && this.status === 'waiting') return true;
      return this.statusKind === 'running';
    },
    showApprovalActions() {
      return this.kind === 'approval' && this.status === 'pending' && !!this.store.canResolveApprovals;
    },
    clickable() {
      return !this.editable && !!this.run && this.run.taskId != null;
    },
    ticking() {
      if (!this.run) return false;
      if (this.kind === 'delay' && this.status === 'waiting') return true;
      return this.statusKind === 'running' && !!this.run.start && !this.run.end;
    },
    stateClasses() {
      return {
        'WorkflowNodeCard--running': this.isActive,
        'WorkflowNodeCard--pending': this.statusKind === 'pending',
        'WorkflowNodeCard--dim': !!this.store.runs && Object.keys(this.store.runs).length > 0 && !this.run,
        'WorkflowNodeCard--clickable': this.clickable,
        'WorkflowNodeCard--problem': !!this.problem,
      };
    },
  },

  watch: {
    ticking: {
      immediate: true,
      handler(active) {
        if (active && !this.timer) {
          this.timer = setInterval(() => { this.now = Date.now(); }, 1000);
        } else if (!active && this.timer) {
          clearInterval(this.timer);
          this.timer = null;
        }
        this.now = Date.now();
      },
    },
  },

  beforeDestroy() {
    if (this.timer) clearInterval(this.timer);
  },

  methods: {
    resolve(status) {
      this.$emit('resolve-approval', status);
    },

    // Read-only canvas: Drawflow's pan handler swallows clicks on nodes, so a
    // tap is detected here from a press and release that barely moved.
    onPointerDown(ev) {
      if (this.editable) return;
      this.pointerDown = { x: ev.clientX, y: ev.clientY };
    },
    onPointerUp(ev) {
      const down = this.pointerDown;
      this.pointerDown = null;
      if (this.editable || !down) return;
      if (Math.abs(ev.clientX - down.x) > TAP_THRESHOLD_PX
        || Math.abs(ev.clientY - down.y) > TAP_THRESHOLD_PX) return;
      this.$emit('select', this.nodeId);
    },
  },
};
</script>

<style lang="scss">
@import '~@/components/workflow/palette';

.WorkflowNodeCard {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  width: $wf-card-width;
  min-height: 64px;
  padding: 10px 12px;
  border-radius: $wf-card-radius;
  border: 1px solid var(--wf-divider);
  background: var(--wf-surface);
  color: var(--wf-text);
  box-shadow: var(--wf-shadow);
  transition: box-shadow 0.15s ease, border-color 0.15s ease, opacity 0.2s ease;
  cursor: move;
  user-select: none;

  &:hover {
    box-shadow: var(--wf-shadow-hover);
    border-color: var(--wf-divider-strong);
  }

  &--clickable {
    cursor: pointer;
  }

  &--dim {
    opacity: 0.55;
  }

  &--running {
    border-color: var(--wf-primary);
    animation: WorkflowNodeCardPulse 1.6s ease-out infinite;
  }

  &--pending {
    border-color: var(--wf-warning);
    animation: WorkflowNodeCardPulseAmber 1.6s ease-out infinite;
  }

  &__tile {
    flex: 0 0 40px;
    width: 40px;
    height: 40px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--wf-tile-bg);

    &--approval {
      background: rgba(171, 71, 188, 0.14);
    }

    &--delay {
      background: rgba(255, 152, 0, 0.16);
    }
  }

  &__body {
    flex: 1 1 auto;
    min-width: 0;
  }

  &__title {
    font-size: 13px;
    font-weight: 600;
    line-height: 18px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    &--placeholder {
      font-style: italic;
      color: var(--wf-text2);
    }
  }

  &__sub {
    font-size: 11px;
    line-height: 15px;
    color: var(--wf-text2);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  &__actions {
    display: flex;
    margin-top: 6px;
  }

  &__status {
    position: absolute;
    top: 5px;
    right: 7px;
    line-height: 1;
    display: flex;
  }

  &__badge {
    position: absolute;
    top: -8px;
    right: -8px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--wf-warning);
    box-shadow: var(--wf-shadow);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: help;
  }

  // "+" quick-add handle: sticks out of the output port on hover.
  &__plus {
    position: absolute;
    top: 50%;
    right: -42px;
    margin-top: -11px;
    width: 22px;
    height: 22px;
    padding: 0;
    border-radius: 50%;
    border: 1.5px solid var(--wf-primary);
    background: var(--wf-surface);
    box-shadow: var(--wf-shadow);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.12s ease;

    &::before {
      content: '';
      position: absolute;
      left: -16px;
      top: 50%;
      width: 15px;
      height: 1.5px;
      background: var(--wf-primary);
    }
  }

  &:hover &__plus {
    opacity: 1;
  }

  // Sticky note: annotation only, no ports.
  &--note {
    display: block;
    width: 200px;
    padding: 8px 12px 10px;
    border-radius: 10px;
    background: var(--wf-note-bg);
    border-color: var(--wf-note-border);
    color: var(--wf-note-text);
  }

  &__noteIcon {
    opacity: 0.55;
  }

  &__noteText {
    margin-top: 2px;
    font-size: 12.5px;
    line-height: 17px;
    white-space: pre-wrap;
    word-break: break-word;

    &--empty {
      font-style: italic;
      opacity: 0.6;
    }
  }
}

@keyframes WorkflowNodeCardPulse {
  0% { box-shadow: 0 0 0 0 rgba(33, 150, 243, 0.45); }
  70% { box-shadow: 0 0 0 10px rgba(33, 150, 243, 0); }
  100% { box-shadow: 0 0 0 0 rgba(33, 150, 243, 0); }
}

@keyframes WorkflowNodeCardPulseAmber {
  0% { box-shadow: 0 0 0 0 rgba(251, 140, 0, 0.45); }
  70% { box-shadow: 0 0 0 10px rgba(251, 140, 0, 0); }
  100% { box-shadow: 0 0 0 0 rgba(251, 140, 0, 0); }
}
</style>
