<template>
  <v-dialog v-model="dialog" max-width="520" :persistent="running">
    <v-card>
      <v-card-title>{{ title }}</v-card-title>
      <v-card-text class="pb-0">
        <div class="text-caption text--secondary mb-4">{{ $t('workflowRunDialogHint') }}</div>
        <v-form ref="form" lazy-validation @submit.prevent="run()">
          <SurveyFields
            v-if="environment"
            :vars="vars"
            :value="environment"
            :secret-value="secretEnvironment"
          />
        </v-form>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text :disabled="running" @click="dialog = false">{{ $t('cancel') }}</v-btn>
        <v-btn
          color="primary"
          depressed
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
    value: Boolean,
    workflow: Object,
    running: Boolean,
  },

  data() {
    return {
      dialog: this.value,
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
    value(v) {
      this.dialog = v;
      if (v) this.reset();
    },
    dialog(v) {
      if (v !== this.value) this.$emit('input', v);
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

    run() {
      if (this.$refs.form && !this.$refs.form.validate()) return;
      this.$emit('run', { ...this.environment });
    },
  },
};
</script>
