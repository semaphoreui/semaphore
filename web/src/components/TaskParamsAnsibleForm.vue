<template>
  <v-row no-gutters>
    <v-col v-if="templateParams.allow_debug">
      <v-checkbox
        class="mt-0"
        :input-value="params.debug"
        v-model="params.debug"
        @change="updateValue('debug', $event)"
        hide-details
      >
        <template v-slot:label>
          <div class="text-no-wrap">
            {{ $t('debug') }} <code>-{{ "v".repeat(params.debug_level || 4) }}</code>
          </div>
        </template>
      </v-checkbox>
      <v-slider
        :disabled="!params.debug"
        class="ml-7 mb-2"
        style="max-width: 100px;"
        v-model="params.debug_level"
        @change="updateValue('debug_level', $event)"
        step="1"
        min="1"
        max="6"
        hide-details
      ></v-slider>
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
</template>

<style lang="scss">

</style>

<script>

const APP_PARAMS = {
  ansible: [
    'diff',
    'debug',
    'debug_level',
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
      params: {
        debug_level: 4,
      },
    };
  },

  created() {
    this.params = {
      ...this.value,
      debug_level: this.value.debug_level || 4,
    };
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
