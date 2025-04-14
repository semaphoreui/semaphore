<template>
  <div v-if="app === 'ansible'">
    <v-row no-gutters>
      <v-col v-if="templateParams.allow_debug">
        <v-checkbox
          class="mt-0"
          :input-value="params.debug"
          @change="updateValue('debug', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('debug') }} <code>--vvvv</code></div>
          </template>
        </v-checkbox>
      </v-col>
      <v-col>
        <v-checkbox
          class="mt-0"
          :input-value="params.dry_run"
          @change="updateValue('dry_run', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('dryRun') }} <code>--check</code></div>
          </template>
        </v-checkbox>
      </v-col>
      <v-col>
        <v-checkbox
          class="mt-0"
          :input-value="params.diff"
          @change="updateValue('diff', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('diff') }} <code>--diff</code></div>
          </template>
        </v-checkbox>
      </v-col>
    </v-row>
  </div>
  <div v-else-if="app === 'terraform' || app === 'tofu'">
    <v-row no-gutters>
      <v-col>
        <v-checkbox
          class="mt-0"
          :input-value="params.plan"
          @change="updateValue('plan', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('Plan') }}</div>
          </template>
        </v-checkbox>
      </v-col>

      <v-col>
        <v-checkbox
          class="mt-0"
          :input-value="params.destroy"
          @change="updateValue('destroy', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('Destroy') }} <code>-destroy</code></div>
          </template>
        </v-checkbox>
      </v-col>

      <v-col>
        <v-checkbox
          class="mt-0"
          :input-value="params.auto_approve"
          @change="updateValue('auto_approve', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('Auto Approve') }} <code>-auto-approve</code></div>
          </template>
        </v-checkbox>
      </v-col>

      <v-col>
        <v-checkbox
          class="mt-0"
          :input-value="params.upgrade"
          @change="updateValue('upgrade', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('Upgrade') }} <code>-upgrade</code></div>
          </template>
        </v-checkbox>
      </v-col>

      <v-col>
        <v-checkbox
          class="mt-0"
          :input-value="params.reconfigure"
          @change="updateValue('reconfigure', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('Reconfigure') }} <code>-reconfigure</code></div>
          </template>
        </v-checkbox>
      </v-col>
    </v-row>
  </div>
  <div v-else></div>
</template>

<style lang="scss">

</style>

<script>
const TERRAFORM_APP_PARAMS = [
  'plan',
  'auto_approve',
  'destroy',
  'reconfigure',
  'upgrade',
];

const APP_PARAMS = {
  terraform: TERRAFORM_APP_PARAMS,
  tofu: TERRAFORM_APP_PARAMS,
  ansible: [
    'diff',
    'debug',
    'dry_run',
    'tags',
    'skip_tags',
    'limit',
  ],
};

export default {
  props: {
    value: Object,
    app: String,
    templateParams: Object,
  },

  watch: {
    value(val) {
      this.params = val;
    },
  },

  data() {
    return {
      params: {},
    };
  },

  created() {
    this.params = this.value;
  },

  methods: {
    updateValue(prop, value) {
      this.params[prop] = value;

      let input = { ...this.params, [prop]: value };

      if (APP_PARAMS[this.app]) {
        input = (APP_PARAMS[this.app] || []).reduce((res, param) => ({
          ...res,
          [param]: input[param],
        }), {});
      }

      this.$emit('input', input);
    },
  },
};
</script>
