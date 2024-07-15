<template>
  <div v-if="template.app === 'ansible'">
    <v-row no-gutters class="mt-6">
      <v-col cols="12" sm="6">
        <v-checkbox
          class="mt-0"
          :value="value.debug"
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
          :value="value.dry_run"
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
          :value="value.diff"
          @change="updateValue('diff', $event)"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('diff') }} <code>--diff</code></div>
          </template>
        </v-checkbox>
      </v-col>
    </v-row>
  </div>
  <div v-else-if="template.app === 'terraform' || template.app === 'tofu'">
    <v-checkbox
      class="mt-0"
      :value="value.plan"
      @change="updateValue('plan', $event)"
    >
      <template v-slot:label>
        <div class="text-no-wrap">{{ $t('Plan') }}</div>
      </template>
    </v-checkbox>
  </div>
  <div v-else></div>
</template>

<style lang="scss">

</style>

<script>
export default {
  props: {
    value: Object,
    template: Object,
  },
  methods: {
    updateValue(prop, value) {
      this.$emit('input', { ...this.value, [prop]: value });
    },
  },
};
</script>
