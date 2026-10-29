<template>
  <div class="WorkflowNodeProperties" data-testid="workflowNodeProps-panel">
    <div class="WorkflowNodeProperties__header">
      <span
        class="WorkflowNodeProperties__tile"
        :class="`WorkflowNodeProperties__tile--${kind}`"
        :data-testid="`workflowNodeProps-tile-${kind}`"
      >
        <v-icon :color="tileColor">{{ icon }}</v-icon>
      </span>
      <div class="WorkflowNodeProperties__heading">
        <div
          class="text-label-large text-truncate"
          data-testid="workflowNodeProps-heading"
        >{{ heading }}</div>
        <div
          class="text-body-small text-medium-emphasis"
          data-testid="workflowNodeProps-nodeId"
        >{{ $t('workflowNodeId') }} #{{ node.id }}</div>
      </div>
      <v-btn
        icon
        variant="text"
        size="small"
        :title="$t('close')"
        @click="$emit('close')"
        data-testid="workflowNodeProps-close"
      >
        <v-icon size="small">mdi-close</v-icon>
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
        variant="outlined"
        density="compact"
        auto-grow
        rows="4"
        hide-details="auto"
        @update:model-value="emitChange"
        data-testid="workflowNodeProps-note"
      />

      <template v-else>
        <v-select
          v-model="node.kind"
          :items="kindOptions"
          item-value="value"
          item-title="text"
          :label="$t('workflowNodeKind')"
          :disabled="!canManage"
          variant="outlined"
          density="compact"
          hide-details="auto"
          class="mb-5"
          @update:model-value="onKindChanged"
          data-testid="workflowNodeProps-kind"
        />

        <v-autocomplete
          v-if="kind === 'task'"
          v-model="node.template_id"
          :items="templates"
          item-value="id"
          item-title="name"
          :label="$t('taskTemplate')"
          :disabled="!canManage"
          variant="outlined"
          density="compact"
          hide-details="auto"
          class="mb-5"
          @update:model-value="onTemplateChanged"
          data-testid="workflowNodeProps-templateId"
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
              :workflow-vars="workflowVars"
              :survey-inputs="bindings"
              class="mt-2"
              @update:model-value="emitChange"
              @bind="setBinding"
              @unbind="setBinding($event, null)"
              data-testid="workflowNodeProps-taskParams"
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
            variant="outlined"
            density="compact"
            hide-details="auto"
            class="mb-5"
            @update:model-value="emitChange"
            data-testid="workflowNodeProps-approvalTimeout"
          />
          <v-text-field
            v-model="node.approval_message"
            :label="$t('workflowApprovalMessage')"
            :disabled="!canManage"
            variant="outlined"
            density="compact"
            hide-details="auto"
            class="mb-5"
            @update:model-value="emitChange"
            data-testid="workflowNodeProps-approvalMessage"
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
          variant="outlined"
          density="compact"
          hide-details="auto"
          class="mb-5"
          @update:model-value="emitChange"
          data-testid="workflowNodeProps-delaySeconds"
        />

        <v-select
          v-model="node.convergence_mode"
          :items="convergenceOptions"
          item-value="value"
          item-title="text"
          :label="$t('workflowConvergence')"
          :disabled="!canManage"
          variant="outlined"
          density="compact"
          hide-details="auto"
          @update:model-value="emitChange"
          data-testid="workflowNodeProps-convergenceMode"
        />
      </template>
    </div>

    <v-divider />

    <div class="pa-4">
      <v-btn
        variant="outlined"
        size="small"
        block
        color="error"
        :disabled="!canManage"
        @click="$emit('delete')"
        data-testid="workflowNodeProps-delete"
      >
        <v-icon start size="small">mdi-delete-outline</v-icon>
        {{ $t('workflowToolbarDelete') }}
      </v-btn>
    </div>
  </div>
</template>

<script>
import TaskParamsForm from '@/components/TaskParamsForm.vue';
import { APP_ICONS } from '@/lib/constants';

// Right-hand panel of the editor. `modelValue` is the editor's working copy of the
// node; the fields edit it in place and every change emits `update:modelValue` so the
// canvas can repaint the card.
export default {
  components: { TaskParamsForm },

  props: {
    modelValue: { type: Object, required: true },
    templates: { type: Array, default: () => [] },
    canManage: { type: Boolean, default: false },
    // The workflow's own survey (WorkflowTemplate.survey_vars): the values a
    // task node may be bound to through node.survey_inputs.
    surveyVars: { type: Array, default: () => [] },
  },

  emits: ['close', 'delete', 'update:modelValue'],

  data() {
    return { node: this.modelValue };
  },

  watch: {
    modelValue(v) {
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
    // Workflow variables a survey field of the node may be bound to. A secret
    // is never offered: the run keeps its values in clear.
    workflowVars() {
      return (this.surveyVars || []).filter((v) => v.type !== 'secret');
    },
    bindings() {
      return Array.isArray(this.node.survey_inputs) ? this.node.survey_inputs : [];
    },
  },

  methods: {
    emitChange() {
      this.$emit('update:modelValue', this.node);
    },
    boundVar(name) {
      const b = this.bindings.find((x) => x.var === name);
      return b ? b.workflow_var : null;
    },
    // Writes node.survey_inputs as [{ var, workflow_var }]; an empty list
    // leaves the node without the field, the shape the API returns. Called
    // by the form's `bind` (name, workflow var) and `unbind` (name, null).
    setBinding(name, workflowVar) {
      const rest = this.bindings.filter((b) => b.var !== name);
      const list = workflowVar ? [...rest, { var: name, workflow_var: workflowVar }] : rest;
      if (list.length > 0) {
        this.node.survey_inputs = list;
      } else {
        delete this.node.survey_inputs;
      }
      this.emitChange();
    },
    // Bindings name variables of the old template; none of them is known to
    // carry over, so a new template starts without bindings.
    onTemplateChanged() {
      delete this.node.survey_inputs;
      this.emitChange();
    },
    onKindChanged() {
      const { node } = this;
      const { kind } = node;
      if (kind === 'approval' || kind === 'delay') {
        node.template_id = null;
        node.task_params = null;
        // Only a task node may carry bindings (backend validation).
        delete node.survey_inputs;
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
