<template>
  <v-dialog v-model="dialog" max-width="520" :persistent="running">
    <v-card data-testid="workflowRunDialog" :data-state="running ? 'loading' : 'ready'">
      <v-card-title data-testid="workflowRunDialog-title">{{ title }}</v-card-title>
      <v-card-text class="pb-0">
        <div
          class="text-caption text-medium-emphasis mb-4"
          data-testid="workflowRunDialog-hint"
        >{{ $t('workflowRunDialogHint') }}</div>
        <v-form
          ref="form"
          lazy-validation
          @submit.prevent="run()"
          data-testid="workflowRunDialog-form"
        >
          <SurveyFields
            data-testid="workflowRunDialog-survey"
            v-if="environment"
            :vars="vars"
            :value="environment"
            :secret-value="secretEnvironment"
          />
        </v-form>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          variant="text"
          :disabled="running"
          @click="dialog = false"
          data-testid="workflowRunDialog-cancel"
        >{{ $t('cancel') }}</v-btn>
        <v-btn
          color="primary"
          variant="flat"
          :loading="running"
          data-testid="workflow-run-dialog-run"
          @click="run()"
        >{{ $t('run') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
import SurveyFields from '@/components/SurveyFields.vue';

// Asks the workflow's survey before a run starts. Opens with the survey
// defaults filled in; "Run" validates the form and emits `run` with the
// answers as an object (variable name → value), the body of
// POST .../workflows/{id}/runs. The owner performs the request and reports
// the outcome: it sets `running` while the request is in flight and closes
// the dialog (v-model) on success, so a server-side validation error keeps
// the answers on screen.
export default {
  components: { SurveyFields },

  props: {
    modelValue: Boolean,
    workflow: Object,
    running: Boolean,
  },

  emits: ['update:modelValue', 'run'],

  data() {
    return {
      dialog: this.modelValue,
      environment: null,
      secretEnvironment: {},
    };
  },

  computed: {
    title() {
      return this.$t('workflowRunDialogTitle', { name: this.workflow ? this.workflow.name : '' });
    },
    vars() {
      return (this.workflow?.survey_vars || []).filter((v) => v.type !== 'secret');
    },
  },

  watch: {
    modelValue(v) {
      this.dialog = v;
      if (v) this.reset();
    },
    dialog(v) {
      if (v !== this.modelValue) this.$emit('update:modelValue', v);
    },
  },

  created() {
    if (this.dialog) this.reset();
  },

  methods: {
    // The same seeding TaskParamsForm.afterLoadData does for a task: every
    // default applied, a select always an array.
    reset() {
      const env = {};
      this.vars.forEach((v) => {
        const def = v.default_value;
        if (v.type === 'select') {
          if (Array.isArray(def)) env[v.name] = [...def];
          else env[v.name] = def ? [def] : [];
        } else if (def != null && def !== '') {
          env[v.name] = Array.isArray(def) ? def[0] : def;
        } else {
          env[v.name] = v.type === 'enum' ? null : '';
        }
      });
      this.environment = env;
      this.secretEnvironment = {};
      if (this.$refs.form) this.$refs.form.resetValidation();
    },

    async run() {
      if (this.running) return;
      if (this.$refs.form && !(await this.$refs.form.validate()).valid) return;
      this.$emit('run', { ...this.environment });
    },
  },
};
</script>
