<template>
  <div class="WorkflowEdgeProperties">
    <div class="WorkflowEdgeProperties__header">
      <span class="WorkflowEdgeProperties__tile">
        <v-icon>mdi-ray-start-arrow</v-icon>
      </span>
      <div class="WorkflowEdgeProperties__heading">
        <div class="text-subtitle-2">{{ $t('workflowEdgeCondition') }}</div>
        <div class="text-caption text--secondary">
          #{{ edge.source_node_id }} → #{{ edge.destination_node_id }}
        </div>
      </div>
      <v-btn icon small :title="$t('close')" @click="$emit('close')">
        <v-icon small>mdi-close</v-icon>
      </v-btn>
    </div>

    <v-divider />

    <div class="pa-4">
      <v-select
        v-model="edge.condition"
        :items="conditionOptions"
        item-value="value"
        item-text="text"
        :disabled="!canManage"
        outlined
        dense
        hide-details="auto"
        @change="emitChange"
      >
        <template v-slot:item="{ item }">
          <v-icon small left :color="item.color">{{ item.icon }}</v-icon>
          {{ item.text }}
        </template>
        <template v-slot:selection="{ item }">
          <v-icon small left :color="item.color">{{ item.icon }}</v-icon>
          {{ item.text }}
        </template>
      </v-select>
    </div>

    <template v-if="destinationTemplate">
      <v-divider />

      <div class="pa-4 WorkflowEdgeProperties__inputs">
        <div class="text-subtitle-2 mb-1">{{ $t('workflowEdgeInputs') }}</div>

        <v-checkbox
          v-model="explicit"
          :label="$t('workflowEdgeInputsExplicit')"
          :disabled="!canManage"
          dense
          hide-details
          class="mt-0 mb-2"
          @change="onModeChanged"
        />

        <template v-if="!explicit">
          <div v-if="surveyVars.length === 0" class="text-caption text--secondary">
            {{ $t('workflowEdgeInputsByNameNone', { template: destinationTemplate.name }) }}
          </div>
          <template v-else>
            <div class="text-caption text--secondary mb-2">
              {{ $t('workflowEdgeInputsByNameHint', { template: destinationTemplate.name }) }}
            </div>
            <div class="WorkflowEdgeProperties__chips">
              <v-chip
                v-for="v in surveyVars"
                :key="v.name"
                x-small
                label
                outlined
                :color="hintHas(v.name) ? 'success' : null"
                :title="hintHas(v.name)
                  ? $t('workflowEdgeInputsMatched', { run: hints.runId }) : null"
              >
                <v-icon v-if="hintHas(v.name)" x-small left>mdi-check</v-icon>
                {{ v.name }}
              </v-chip>
            </div>
          </template>
        </template>

        <template v-else>
          <div v-if="surveyVars.length === 0" class="text-caption text--secondary">
            {{ $t('workflowEdgeInputsByNameNone', { template: destinationTemplate.name }) }}
          </div>
          <template v-else>
            <div class="text-caption text--secondary mb-3">
              {{ $t('workflowEdgeInputsExplicitHint') }}
            </div>
            <div
              v-for="v in surveyVars"
              :key="v.name"
              class="WorkflowEdgeProperties__row"
            >
              <div class="WorkflowEdgeProperties__var" :title="v.title || v.name">
                <span class="WorkflowEdgeProperties__varName">{{ v.name }}</span>
                <span class="text-caption text--secondary">{{ typeLabel(v) }}</span>
              </div>
              <v-combobox
                :value="mappingKey(v.name)"
                :items="hintKeys"
                :disabled="!canManage"
                :label="$t('workflowEdgeInputsKey')"
                :error-messages="keyErrors(v.name)"
                :hint="keyHint(v.name)"
                persistent-hint
                outlined
                dense
                clearable
                hide-details="auto"
                class="WorkflowEdgeProperties__key"
                @input="setMapping(v.name, $event)"
              />
            </div>
          </template>
        </template>

        <div class="text-caption text--secondary mt-2">{{ hintsCaption }}</div>
      </div>
    </template>

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
        {{ $t('workflowEdgeDelete') }}
      </v-btn>
    </div>
  </div>
</template>

<script>
// Output names: letters, digits, _ and -, not starting with a digit — the same
// pattern the server validates (db.IsValidWorkflowOutputName).
const OUTPUT_KEY_PATTERN = /^[A-Za-z_][A-Za-z0-9_-]*$/;

export default {
  props: {
    // { source_node_id, destination_node_id, condition, input_mode?, input_mappings? }
    value: { type: Object, required: true },
    canManage: { type: Boolean, default: false },
    // The workflow's nodes and the project's templates, to find the
    // destination's survey variables.
    nodes: { type: Array, default: () => [] },
    templates: { type: Array, default: () => [] },
    // Outputs of the latest run of this revision, keyed by node id:
    // { runId, byNode: { [nodeId]: { keys: [...], skipped: { key: reason } } } }.
    // Null when no finished run of the current revision exists.
    outputHints: { type: Object, default: null },
  },

  data() {
    return {
      edge: this.normalize(this.value),
      explicit: this.value.input_mode === 'explicit',
    };
  },

  watch: {
    value(v) {
      this.edge = this.normalize(v);
      this.explicit = v.input_mode === 'explicit';
    },
  },

  computed: {
    conditionOptions() {
      return [
        {
          value: 'on_success', icon: 'mdi-check', color: 'success', text: this.$t('workflowConditionOnSuccess'),
        },
        {
          value: 'on_failure', icon: 'mdi-close', color: 'error', text: this.$t('workflowConditionOnFailure'),
        },
        {
          value: 'always', icon: 'mdi-arrow-right', color: 'grey', text: this.$t('workflowConditionAlways'),
        },
      ];
    },
    destinationNode() {
      return this.nodes.find((n) => n.id === this.edge.destination_node_id) || null;
    },
    sourceNode() {
      return this.nodes.find((n) => n.id === this.edge.source_node_id) || null;
    },
    // Only a task node has survey variables to feed.
    destinationTemplate() {
      const node = this.destinationNode;
      if (!node || (node.kind || 'task') !== 'task' || !node.template_id) return null;
      return this.templates.find((t) => t.id === node.template_id) || null;
    },
    // Secret variables are never fed from outputs.
    surveyVars() {
      return (this.destinationTemplate?.survey_vars || []).filter((v) => v.type !== 'secret');
    },
    // Outputs the source task produced in the latest run; an approval or delay
    // source has none of its own, so no hints there.
    hints() {
      if (!this.outputHints || !this.sourceNode) return null;
      if ((this.sourceNode.kind || 'task') !== 'task') return null;
      const entry = this.outputHints.byNode[this.sourceNode.id];
      if (!entry) return null;
      return { runId: this.outputHints.runId, ...entry };
    },
    hintKeys() {
      return this.hints ? this.hints.keys : [];
    },
    // Where the suggested keys come from, or why there are none.
    hintsCaption() {
      if (this.hints) return this.$t('workflowEdgeInputsHintsFrom', { run: this.hints.runId });
      if (this.outputHints) {
        return this.$t('workflowEdgeInputsNoSourceOutputs', { run: this.outputHints.runId });
      }
      return this.$t('workflowEdgeInputsNoHints');
    },
  },

  methods: {
    normalize(edge) {
      return {
        ...edge,
        input_mode: edge.input_mode === 'explicit' ? 'explicit' : 'by_name',
        input_mappings: (edge.input_mappings || []).map((m) => ({ var: m.var, key: m.key })),
      };
    },
    typeLabel(v) {
      return v.type || 'string';
    },
    hintHas(name) {
      return !!this.hints && this.hints.keys.includes(name);
    },
    mappingKey(name) {
      const m = this.edge.input_mappings.find((x) => x.var === name);
      return m ? m.key : null;
    },
    setMapping(name, key) {
      const value = (key || '').trim();
      const rest = this.edge.input_mappings.filter((m) => m.var !== name);
      this.edge.input_mappings = value ? [...rest, { var: name, key: value }] : rest;
      this.emitChange();
    },
    keyErrors(name) {
      const key = this.mappingKey(name);
      if (!key || OUTPUT_KEY_PATTERN.test(key)) return [];
      return [this.$t('workflowEdgeInputsKeyInvalid')];
    },
    keyHint(name) {
      const key = this.mappingKey(name);
      if (!key || !this.hints) return '';
      const reason = this.hints.skipped[key];
      if (reason) {
        return this.$t('workflowEdgeInputsKeySkipped', { reason: this.skipReason(reason) });
      }
      return '';
    },
    skipReason(reason) {
      const key = `workflowOutputSkipped_${reason}`;
      return this.$te(key) ? this.$t(key) : reason;
    },
    onModeChanged(explicit) {
      this.edge.input_mode = explicit ? 'explicit' : 'by_name';
      if (!explicit) this.edge.input_mappings = [];
      this.emitChange();
    },
    emitChange() {
      // by_name carries no mappings on the wire; explicit sends its list (an
      // empty one means the edge passes nothing).
      const out = { ...this.edge };
      if (out.input_mode !== 'explicit') {
        delete out.input_mode;
        delete out.input_mappings;
      }
      this.$emit('input', out);
    },
  },
};
</script>

<style lang="scss">
.WorkflowEdgeProperties {
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
  }

  &__heading {
    flex: 1 1 auto;
    min-width: 0;
  }

  &__chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  &__row {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    margin-bottom: 10px;
  }

  &__var {
    flex: 0 0 40%;
    min-width: 0;
    padding-top: 9px;
    display: flex;
    flex-direction: column;
    line-height: 1.2;
  }

  &__varName {
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__key {
    flex: 1 1 auto;
    min-width: 0;
  }
}
</style>
