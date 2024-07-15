<template>
  <div v-if="app === 'ansible'">
    <v-row no-gutters class="mt-6">
      <v-col cols="12" sm="6">
        <v-checkbox
          class="mt-0"
          :input-value="value.debug"
          @change="updateValue('debug', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('debug') }} <code>--vvvv</code></div>
          </template>
        </v-checkbox>
      </v-col>
      <v-col cols="12" sm="6">
        <v-checkbox
          class="mt-0"
          :input-value="value.dry_run"
          @change="updateValue('dry_run', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('dryRun') }} <code>--check</code></div>
          </template>
        </v-checkbox>
      </v-col>
      <v-col cols="12" sm="6">
        <v-checkbox
          class="mt-0"
          :input-value="value.diff"
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
    <v-checkbox
      class="mt-0"
      :input-value="value.plan"
      @change="updateValue('plan', $event)"
    >
      <template v-slot:label>
        <div class="text-no-wrap">{{ $t('Plan') }}</div>
      </template>
    </v-checkbox>
    <v-checkbox
      class="mt-0"
      :input-value="value.auto_approve"
      @change="updateValue('auto_approve', $event)"
    >
      <template v-slot:label>
        <div class="text-no-wrap">{{ $t('Auto Approve') }} <code>-auto-approve</code></div>
      </template>
    </v-checkbox>
  </div>
  <div v-else></div>
</template>

<style lang="scss">

</style>

<script>

const APP_PARAMS = {
  terraform: ['plan', 'auto_approve'],
  tofu: ['plan', 'auto_approve'],
};

export default {
  props: {
    value: Object,
    app: String,
  },

  methods: {
    updateValue(prop, value) {
      let input = { ...this.value, [prop]: value };

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
