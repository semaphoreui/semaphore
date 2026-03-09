<template>
  <div>
    <v-alert text v-if="!paths || paths.length === 0">
      {{ $t('No sync paths are defined.') }}
    </v-alert>
    <div v-for="(p, i) in paths" :key="i" class="d-flex align-start mb-4">
      <div style="flex: 1; min-width: 0" class="mb-2">
        <v-text-field
          class="mb-2"
          v-model="p.path"
          :label="$t('path')"
          outlined
          dense
          @input="emitPaths"
          hide-details
        />

        <div class="d-flex" style="gap: 8px">
          <v-text-field
            v-model="p.separator"
            :label="$t('separator')"
            outlined
            dense
            style="flex: 1"
            @input="emitPaths"
            hide-details
          />
          <v-text-field
            v-model="p.prefix"
            :label="$t('prefix')"
            outlined
            dense
            style="flex: 1"
            @input="emitPaths"
            hide-details
          />
        </div>
      </div>

      <v-btn icon class="ml-1 mt-1" @click="removePath(i)">
        <v-icon>mdi-delete</v-icon>
      </v-btn>
    </div>

    <v-btn
      text
      color="primary"
      @click="addPath"
      style="position: absolute; left: 12px; bottom: 12px"
    >
      <v-icon left>mdi-plus</v-icon>
      Add path
    </v-btn>
  </div>
</template>

<script>
export default {
  props: {
    value: {
      type: Array,
      default: () => [],
    },
  },

  data() {
    return {
      paths: [],
    };
  },

  created() {
    this.paths = this.value && this.value.length ? this.value.map((p) => ({ ...p })) : [];
  },

  watch: {
    value(val) {
      if (JSON.stringify(val) !== JSON.stringify(this.paths)) {
        this.paths = (val || []).map((p) => ({ ...p }));
      }
    },
  },

  methods: {
    emitPaths() {
      this.$emit(
        'input',
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
