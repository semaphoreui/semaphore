<template>
  <div class="TaskParamsAnsibleForm">

    <ArgsPicker
      v-if="templateParams.allow_override_limit"
      test-id="taskParams-limit"
      :vars="params.limit"
      @change="setLimit"
      :title="$t('limit')"
      :arg-title="$t('limit')"
      :add-arg-title="$t('addLimit')"
    />

    <ArgsPicker
      v-if="templateParams.allow_override_tags"
      test-id="taskParams-tags"
      :vars="params.tags"
      @change="setTags"
      :title="$t('tags')"
      :arg-title="$t('tags')"
      :add-arg-title="$t('addTag')"
    />

    <ArgsPicker
      v-if="templateParams.allow_override_skip_tags"
      test-id="taskParams-skipTags"
      :vars="params.skip_tags"
      @change="setSkipTags"
      :title="$t('skipTags')"
      :arg-title="$t('tag')"
      :add-arg-title="$t('addSkippedTag')"
    />

    <v-row no-gutters>
      <v-col v-if="templateParams.allow_debug">
        <v-checkbox
          class="mt-0"
          v-model="params.debug"
          @update:model-value="updateValue('debug', $event)"
          hide-details
          data-testid="taskParams-debug"
        >
          <template v-slot:label>
            <div class="text-no-wrap">
              {{ $t('debug') }} <code>-{{ 'v'.repeat(params.debug_level || 4) }}</code>
            </div>
          </template>
        </v-checkbox>
        <v-slider
          :disabled="!params.debug"
          class="ml-7 mb-2"
          style="max-width: 100px;"
          v-model="params.debug_level"
          @update:model-value="updateValue('debug_level', $event)"
          data-testid="taskParams-debugLevel"
          step="1"
          min="1"
          max="6"
          hide-details
        ></v-slider>
      </v-col>
      <v-col v-if="!templateParams.hide_dry_run">
        <v-checkbox
          class="mt-0"
          :model-value="params.dry_run"
          @update:model-value="updateValue('dry_run', $event)"
          data-testid="taskParams-dryRun"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('dryRun') }} <code>--check</code></div>
          </template>
        </v-checkbox>
      </v-col>
      <v-col v-if="!templateParams.hide_diff">
        <v-checkbox
          class="mt-0"
          :model-value="params.diff"
          @update:model-value="updateValue('diff', $event)"
          data-testid="taskParams-diff"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('diff') }} <code>--diff</code></div>
          </template>
        </v-checkbox>
      </v-col>
      <v-col v-if="templateParams.allow_override_skip_galaxy_install">
        <v-checkbox
          :model-value="params.skip_galaxy_install"
          class="mt-0"
          @update:model-value="updateValue('skip_galaxy_install', $event)"
          data-testid="taskParams-skipGalaxyInstall"
        >
          <template v-slot:label>
            <div class="text-no-wrap">{{ $t('skipGalaxyInstall') }}</div>
          </template>
        </v-checkbox>
      </v-col>
    </v-row>
  </div>
</template>

<style lang="scss">
// Vuetify 3 gives a .v-input min-width 0, so equal v-col columns squeezed the no-wrap labels
// over each other; as in Vuetify 2, a checkbox keeps its content width and the row wraps.
.TaskParamsAnsibleForm .v-checkbox {
  min-width: max-content;
}
</style>

<script>

import ArgsPicker from '@/components/ArgsPicker.vue';

const APP_PARAMS = {
  ansible: [
    'diff',
    'debug',
    'debug_level',
    'dry_run',
    'tags',
    'skip_tags',
    'limit',
    'skip_galaxy_install',
  ],
};

export default {
  components: { ArgsPicker },
  props: {
    modelValue: Object,
    app: String,
    templateParams: Object,
  },

  emits: ['update:modelValue'],

  computed: {
    debugLevelMissing() {
      return !!this.templateParams?.allow_debug && this.params.debug_level == null;
    },
  },

  watch: {
    modelValue(val) {
      this.params = val;
    },

    // The debug level slider defaults to its min, 1. Vuetify 2 v-slider clamped a missing value
    // to min and wrote it back through v-model, so New Task (its params come without debug_level)
    // showed «-v» and sent debug_level 1; Vuetify 3 leaves the model as it is.
    debugLevelMissing: {
      immediate: true,
      handler(missing) {
        if (missing) {
          this.params.debug_level = 1;
        }
      },
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
      ...this.modelValue,
      debug_level: this.modelValue.debug_level || 4,
    };
  },

  methods: {

    setSkipTags(tags) {
      this.updateValue('skip_tags', tags);
    },

    setTags(tags) {
      this.updateValue('tags', tags);
    },

    setLimit(limit) {
      this.updateValue('limit', limit);
    },

    updateValue(prop, value) {
      this.params[prop] = value;

      let input = { ...this.params, [prop]: value };

      if (APP_PARAMS[this.app]) {
        input = (APP_PARAMS[this.app] || []).reduce((res, param) => ({
          ...res,
          [param]: input[param],
        }), {});
      }

      this.$emit('update:modelValue', input);
    },
  },
};
</script>
