<template>
  <div class="SurveyFields">
    <div v-for="v in vars" :key="v.name">
      <v-text-field
        v-if="isBound(v.name)"
        :label="v.title + (v.required ? ' *' : '')"
        :model-value="boundLabel(v.name)"
        :hint="$t('workflowNodeSurveyInputsBoundHint')"
        :error-messages="boundErrors(v.name)"
        persistent-hint
        readonly
        prepend-inner-icon="mdi-link-variant"
        variant="outlined"
        density="compact"
        class="SurveyFields__bound"
        :data-testid="`surveyFields-bound-${v.name}`"
      >
        <template v-slot:append-outer>
          <v-btn
            icon
            size="small"
            :title="$t('workflowNodeSurveyInputsUnbind')"
            class="SurveyFields__unbind"
            :data-testid="`surveyFields-unbind-${v.name}`"
            @click="$emit('unbind', v.name)"
          >
            <v-icon size="small">mdi-close</v-icon>
          </v-btn>
        </template>
      </v-text-field>

      <v-text-field
        v-else-if="v.type === 'secret'"
        :label="v.title"
        :hint="v.description"
        v-model="secret[v.name]"
        :required="v.required"
        class="masked-secret-input"
        :data-testid="`surveyFields-var-${v.name}`"
        :rules="[(val) => !v.required || !!val || v.title + $t('isRequired')]"
        variant="outlined"
        density="compact"
      />

      <v-select
        clearable
        v-else-if="v.type === 'enum' || v.type === 'select'"
        :label="v.title + (v.required ? ' *' : '')"
        :hint="v.description"
        v-model="env[v.name]"
        :required="v.required"
        :rules="[(val) => !v.required || (Array.isArray(val) ? val.length > 0 : val != null) ||
         v.title + ' ' + $t('isRequired')]"
        :items="v.values"
        item-title="name"
        item-value="value"
        :multiple="v.type === 'select'"
        :chips="v.type === 'select'"
        variant="outlined"
        density="compact"
        :data-testid="`surveyFields-var-${v.name}`"
      >
        <template v-if="v.type === 'select'" v-slot:chip="{ item, index }">
          <v-chip
            size="small"
            closable
            @click:close="removeSelectedItem(v.name, index)"
            :data-testid="`surveyFields-chip-${v.name}`"
          >
            {{ item && item.name ? item.name : String(item) }}
          </v-chip>
        </template>
        <template v-if="canBind(v)" v-slot:append-outer>
          <WorkflowVarAssign
            :vars="workflowVars"
            :test-id="`surveyFields-assign-${v.name}`"
            @select="$emit('bind', v.name, $event)"
          />
        </template>
      </v-select>

      <v-textarea
        v-else-if="v.type === 'text'"
        :label="v.title + (v.required ? ' *' : '')"
        :hint="v.description"
        v-model="env[v.name]"
        :required="v.required"
        :rules="[(val) => !v.required || !!val || v.title + ' ' + $t('isRequired')]"
        rows="3"
        auto-grow
        variant="outlined"
        density="compact"
        :data-testid="`surveyFields-var-${v.name}`"
      >
        <template v-if="canBind(v)" v-slot:append-outer>
          <WorkflowVarAssign
            :vars="workflowVars"
            :test-id="`surveyFields-assign-${v.name}`"
            @select="$emit('bind', v.name, $event)"
          />
        </template>
      </v-textarea>

      <v-text-field
        v-else
        :label="v.title + (v.required ? ' *' : '')"
        :hint="v.description"
        v-model="env[v.name]"
        :required="v.required"
        :rules="[
          (val) => !v.required || !!val || v.title + ' ' + $t('isRequired'),
          (val) =>
            !val || v.type !== 'int' || /^\d+$/.test(val) || v.title + ' ' + $t('mustBeInteger'),
        ]"
        variant="outlined"
        density="compact"
        :data-testid="`surveyFields-var-${v.name}`"
      >
        <template v-if="canBind(v)" v-slot:append-outer>
          <WorkflowVarAssign
            :vars="workflowVars"
            :test-id="`surveyFields-assign-${v.name}`"
            @select="$emit('bind', v.name, $event)"
          />
        </template>
      </v-text-field>
    </div>
  </div>
</template>

<script>
import WorkflowVarAssign from '@/components/workflow/WorkflowVarAssign.vue';

// The survey of a template (or of a workflow) as form fields, one per
// variable by its type. The fields edit `value` (non-secret answers) and
// `secretValue` (secrets) in place — the same objects the owner holds, so an
// owner's deep watcher sees every keystroke; they are not replaced.
//
// Workflow task node only: `workflowVars` (null = no binding UI) and
// `surveyInputs` ([{ var, workflow_var }]) show which fields read a workflow
// variable instead of a typed value: a bound field is read-only and shows the
// variable, an unbound one has an "assign" button. The component emits
// `bind(var, workflow_var)` and `unbind(var)`; the owner stores them.
export default {
  components: { WorkflowVarAssign },

  props: {
    vars: { type: Array, default: () => [] },
    value: { type: Object, required: true },
    secretValue: { type: Object, default: () => ({}) },
    workflowVars: { type: Array, default: null },
    surveyInputs: { type: Array, default: () => [] },
  },

  emits: ['unbind', 'bind'],

  data() {
    return { env: this.value, secret: this.secretValue };
  },

  watch: {
    value(v) {
      this.env = v;
    },
    secretValue(v) {
      this.secret = v;
    },
  },

  methods: {
    canBind(v) {
      return this.workflowVars != null && v.type !== 'secret';
    },

    binding(varName) {
      return (this.surveyInputs || []).find((b) => b.var === varName) || null;
    },

    isBound(varName) {
      return this.binding(varName) != null;
    },

    boundWorkflowVar(varName) {
      const b = this.binding(varName);
      if (!b) return null;
      return (this.workflowVars || []).find((w) => w.name === b.workflow_var) || null;
    },

    boundLabel(varName) {
      const w = this.boundWorkflowVar(varName);
      if (w) return w.title && w.title !== w.name ? `${w.title} (${w.name})` : w.name;
      const b = this.binding(varName);
      return b ? b.workflow_var : '';
    },

    // A binding may outlive its workflow variable (renamed or deleted in the
    // editor's left panel): say what it points at instead of a blank field.
    boundErrors(varName) {
      const b = this.binding(varName);
      if (!b || this.boundWorkflowVar(varName)) return [];
      return [this.$t('workflowNodeSurveyInputsUnknown', { name: b.workflow_var })];
    },

    removeSelectedItem(varName, index) {
      const env = this.env;
      if (!Object.prototype.hasOwnProperty.call(env, varName) || !Array.isArray(env[varName])) {
        return;
      }

      if (index < env[varName].length) {
        env[varName].splice(index, 1);
      }
    },
  },
};
</script>

<style lang="scss">
.SurveyFields__bound input {
  cursor: default;
}
</style>
