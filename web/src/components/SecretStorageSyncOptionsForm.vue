<template>
  <div>
    <v-alert
      variant="tonal"
      v-if="!paths || paths.length === 0"
      data-testid="secretStorageSync-empty"
    >
      {{ $t('No sync paths are defined.') }}
    </v-alert>
    <div
      v-for="(p, i) in paths"
      :key="i"
      class="d-flex align-start mb-4"
      data-testid="secretStorageSync-row"
    >
      <div style="flex: 1; min-width: 0" class="mb-2">
        <v-text-field
          class="mb-2"
          v-model="p.path"
          :label="$t('path')"
          variant="outlined"
          density="compact"
          @update:model-value="emitPaths"
          hide-details
          data-testid="secretStorageSync-path"
        />

        <div class="d-flex" style="gap: 8px">
          <v-text-field
            v-model="p.separator"
            :label="$t('separator')"
            variant="outlined"
            density="compact"
            style="flex: 1"
            @update:model-value="emitPaths"
            hide-details
            data-testid="secretStorageSync-separator"
          />
          <v-text-field
            v-model="p.prefix"
            :label="$t('prefix')"
            variant="outlined"
            density="compact"
            style="flex: 1"
            @update:model-value="emitPaths"
            hide-details
            data-testid="secretStorageSync-prefix"
          />
        </div>
      </div>

      <v-btn
        icon
        variant="text"
        class="ml-1 mt-1"
        @click="removePath(i)"
        data-testid="secretStorageSync-removePath"
      >
        <v-icon>mdi-delete</v-icon>
      </v-btn>
    </div>

    <v-btn
      variant="text"
      color="primary"
      @click="addPath"
      style="position: absolute; left: 12px; bottom: 12px"
      data-testid="secretStorageSync-addPath"
    >
      <v-icon start>mdi-plus</v-icon>
      Add path
    </v-btn>
  </div>
</template>

<script>
export default {
  props: {
    modelValue: {
      type: Array,
      default: () => [],
    },
  },

  emits: ['update:modelValue'],

  data() {
    return {
      paths: [],
    };
  },

  created() {
    this.paths = this.modelValue && this.modelValue.length
      ? this.modelValue.map((p) => ({ ...p }))
      : [];
  },

  watch: {
    modelValue(val) {
      if (JSON.stringify(val) !== JSON.stringify(this.paths)) {
        this.paths = (val || []).map((p) => ({ ...p }));
      }
    },
  },

  methods: {
    emitPaths() {
      this.$emit(
        'update:modelValue',
        this.paths.map((p) => ({ ...p })),
      );
    },

    addPath() {
      this.paths.push({ path: '', separator: '', prefix: '' });
      this.emitPaths();
    },

    removePath(index) {
      this.paths.splice(index, 1);
      this.emitPaths();
    },
  },
};
</script>
