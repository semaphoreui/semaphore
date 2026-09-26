<template>
  <div class="WorkflowNodeProperties">
    <div class="WorkflowNodeProperties__header">
      <span class="WorkflowNodeProperties__tile" :class="`WorkflowNodeProperties__tile--${kind}`">
        <v-icon :color="tileColor">{{ icon }}</v-icon>
      </span>
      <div class="WorkflowNodeProperties__heading">
        <div class="text-subtitle-2 text-truncate">{{ heading }}</div>
        <div class="text-caption text--secondary">{{ $t('workflowNodeId') }} #{{ node.id }}</div>
      </div>
      <v-btn icon small :title="$t('close')" @click="$emit('close')">
        <v-icon small>mdi-close</v-icon>
      </v-btn>
    </div>

    <v-divider />

    <div class="WorkflowNodeProperties__fields pa-4">
      <v-textarea
        v-if="kind === 'note'"
        v-model="node.note"
        :label="$t('workflowNoteText')"
        :placeholder="$t('workflowNotePlaceholder')"
        :disabled="!canManage"
        outlined
        dense
        auto-grow
        rows="4"
        hide-details="auto"
        @input="emitChange"
      />

      <template v-else>
        <v-select
          v-model="node.kind"
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

        <v-autocomplete
          v-if="kind === 'task'"
          v-model="node.template_id"
          :items="templates"
          item-value="id"
          item-text="name"
          :label="$t('taskTemplate')"
          :disabled="!canManage"
          outlined
          dense
          hide-details="auto"
          class="mb-5"
          @change="emitChange"
        />

        <v-card
          v-if="kind === 'task' && template"
          :key="`task-params-${node.id}-${node.template_id}`"
          class="WorkflowNodeProperties__params mb-5 pt-3"
          flat
        >
          <v-card-text class="py-0">
            <TaskParamsForm
              :template="template"
              v-model="node.task_params"
              class="mt-2"
              @input="emitChange"
            />
          </v-card-text>
        </v-card>

        <template v-if="kind === 'approval'">
          <v-text-field
            v-model.number="node.approval_timeout"
            type="number"
            min="1"
            :label="$t('workflowApprovalTimeout')"
            :disabled="!canManage"
            outlined
            dense
            hide-details="auto"
            class="mb-5"
            @input="emitChange"
          />
          <v-text-field
            v-model="node.approval_message"
            :label="$t('workflowApprovalMessage')"
            :disabled="!canManage"
            outlined
            dense
            hide-details="auto"
            class="mb-5"
            @input="emitChange"
          />
        </template>

        <v-text-field
          v-if="kind === 'delay'"
          v-model.number="node.delay_seconds"
          type="number"
          min="1"
          :label="$t('workflowDelaySeconds')"
          :hint="$t('workflowDelayHint')"
          persistent-hint
          :disabled="!canManage"
          outlined
          dense
          hide-details="auto"
          class="mb-5"
          @input="emitChange"
        />

        <v-select
          v-model="node.convergence_mode"
          :items="convergenceOptions"
          item-value="value"
          item-text="text"
          :label="$t('workflowConvergence')"
          :disabled="!canManage"
          outlined
          dense
          hide-details="auto"
          @change="emitChange"
        />
      </template>
    </div>

    <v-divider />

    <div class="pa-4">
      <v-btn
        outlined
        small
        block
        color="error"
        :disabled="!canManage"
        @click="$emit('delete')"
      >
        <v-icon left small>mdi-delete-outline</v-icon>
        {{ $t('workflowToolbarDelete') }}
      </v-btn>
    </div>
  </div>
</template>

<script>
import TaskParamsForm from '@/components/TaskParamsForm.vue';
import { APP_ICONS } from '@/lib/constants';

// Right-hand panel of the editor. `value` is the editor's working copy of the
// node; the fields edit it in place and every change emits `input` so the
// canvas can repaint the card.
export default {
  components: { TaskParamsForm },

  props: {
    value: { type: Object, required: true },
    templates: { type: Array, default: () => [] },
    canManage: { type: Boolean, default: false },
  },

  data() {
    return { node: this.value };
  },

  watch: {
    value(v) {
      this.node = v;
    },
  },

  computed: {
    kind() {
      return this.node.kind || 'task';
    },
    template() {
      if (!this.node.template_id) return null;
      return this.templates.find((t) => t.id === this.node.template_id) || null;
    },
    icon() {
      if (this.kind === 'approval') return 'mdi-account-check';
      if (this.kind === 'delay') return 'mdi-timer-outline';
      if (this.kind === 'note') return 'mdi-note-text-outline';
      const app = this.template ? this.template.app : null;
      return (app && APP_ICONS[app] && APP_ICONS[app].icon) || 'mdi-cog';
    },
    tileColor() {
      if (this.kind === 'approval') return '#ab47bc';
      if (this.kind === 'delay') return '#ff9800';
      return null;
    },
    heading() {
      if (this.kind === 'task') {
        return this.template ? this.template.name : this.$t('workflowNodeKindTask');
      }
      return this.kindOptions.concat(this.noteOption).find((o) => o.value === this.kind).text;
    },
    kindOptions() {
      return [
        { value: 'task', text: this.$t('workflowNodeKindTask') },
        { value: 'approval', text: this.$t('workflowNodeKindApproval') },
        { value: 'delay', text: this.$t('workflowNodeKindDelay') },
      ];
    },
    noteOption() {
      return [{ value: 'note', text: this.$t('workflowNodeKindNote') }];
    },
    convergenceOptions() {
      return [
        { value: 'all', text: this.$t('workflowConvergenceAll') },
        { value: 'any', text: this.$t('workflowConvergenceAny') },
      ];
    },
  },

  methods: {
    emitChange() {
      this.$emit('input', this.node);
    },
    onKindChanged() {
      const { node } = this;
      const { kind } = node;
      if (kind === 'approval' || kind === 'delay') {
        node.template_id = null;
        node.task_params = null;
      }
      if (kind !== 'approval') {
        node.approval_timeout = null;
        node.approval_message = null;
      }
      if (kind !== 'delay') {
        node.delay_seconds = null;
      } else if (node.delay_seconds == null) {
        node.delay_seconds = 60;
      }
      if (kind === 'task' && !node.task_params) {
        node.task_params = {};
      }
      this.emitChange();
    },
  },
};
</script>

<style lang="scss">
.WorkflowNodeProperties {
  display: flex;
  flex-direction: column;
  height: 100%;

  &__header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 12px 12px 16px;
  }

  &__tile {
    flex: 0 0 36px;
    width: 36px;
    height: 36px;
    border-radius: 9px;
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

  &__heading {
    flex: 1 1 auto;
    min-width: 0;
  }

  &__fields {
    flex: 1 1 auto;
    overflow-y: auto;
  }

  &__params {
    background: rgba(133, 133, 133, 0.06) !important;
  }
}
</style>
