<template>
  <PageBottomSheet v-model="sheet" v-if="settings">
    <h4 data-testid="tableSettings-title">{{ $t('columns') }}</h4>
    <div class="d-flex flex-row flex-wrap">
      <v-checkbox
        class="mr-6"
        v-for="(header) in columns"
        :key="header.value"
        :label="header.text"
        :data-testid="`tableSettings-column-${header.value}`"
        v-model="settings.columns[header.value].visible"
        @update:model-value="saveSettings()"
      />
    </div>
  </PageBottomSheet>
</template>

<script>
import PageBottomSheet from '@/components/PageBottomSheet.vue';

export default {
  props: {
    modelValue: Boolean,
    tableName: String,
    headers: Array,
  },

  components: { PageBottomSheet },

  emits: ['update:modelValue', 'change'],

  data() {
    return {
      sheet: false,
      settings: null,
    };
  },

  computed: {
    // Checkbox per column: Vuetify 3 headers carry `key` / `title`.
    columns() {
      return this.headers
        .filter((header) => header.key !== 'actions')
        .map((header) => ({ value: header.key, text: header.title }));
    },
  },

  watch: {
    async sheet(val) {
      this.$emit('update:modelValue', val);
    },

    async modelValue(val) {
      this.sheet = val;
    },
    headers() {
      this.loadSettings();
    },
  },

  created() {
    this.loadSettings();
  },

  methods: {
    saveSettings() {
      localStorage.setItem(`${this.tableName}__settings`, JSON.stringify(this.settings));
      this.loadSettings();
    },
    loadSettings() {
      if (localStorage.getItem(`${this.tableName}__settings`)) {
        this.settings = JSON.parse(
          localStorage.getItem(`${this.tableName}__settings`),
        );
      } else {
        this.settings = {
          columns: {},
        };
      }

      this.headers.forEach((header) => {
        if (!this.settings.columns[header.key]) {
          this.settings.columns[header.key] = {
            visible: true,
          };
        }
      });

      this.$emit('change', {
        settings: this.settings,
        headers: this.headers.filter((header) => {
          const column = this.settings.columns[header.key];
          return !column || column.visible;
        }),
      });
    },
  },
};
</script>

<style lang="scss">
// Column labels in the secondary text colour, as every `.v-label` of Vuetify 2 (.6 light, .7 dark);
// Vuetify 4 writes a checkbox label in the colour around it (.87 on the sheet).
.page-bottom-sheet .v-checkbox .v-label {
  color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
  opacity: 1;
}
</style>
