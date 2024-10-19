<template>
  <PageBottomSheet v-model="sheet" v-if="settings">
    <h4>{{ $t('columns') }}</h4>
    <div class="d-flex flex-row flex-wrap">
      <v-checkbox
        class="mr-6"
        v-for="(header) in headers.filter((header) => header.value !== 'actions')"
        :key="header.value"
        :label="header.text"
        v-model="settings.columns[header.value].visible"
        @change="saveSettings()"
      />
    </div>
  </PageBottomSheet>
</template>

<script>
import PageBottomSheet from '@/components/PageBottomSheet.vue';

export default {
  props: {
    value: Boolean,
    tableName: String,
    headers: Array,
  },

  components: { PageBottomSheet },

  data() {
    return {
      sheet: false,
      settings: null,
    };
  },

  watch: {
    async sheet(val) {
      this.$emit('input', val);
    },

    async value(val) {
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
        if (!this.settings.columns[header.value]) {
          this.settings.columns[header.value] = {
            visible: true,
          };
        }
      });

      this.$emit('change', {
        settings: this.settings,
        headers: this.headers.filter((header) => {
          const column = this.settings.columns[header.value];
          return !column || column.visible;
        }),
      });
    },
  },
};
</script>
