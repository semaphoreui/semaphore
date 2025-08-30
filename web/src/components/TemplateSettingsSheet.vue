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

    <v-divider class="my-4"></v-divider>

    <h4>{{ $t('tabs') }}</h4>
    <div class="d-flex flex-row flex-wrap">
      <v-checkbox
        class="mr-6"
        :label="$t('placeAllTabAtEnd')"
        v-model="settings.tabs.allTabAtEnd"
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
          tabs: {},
        };
      }

      // Initialize column settings
      this.headers.forEach((header) => {
        if (!this.settings.columns[header.value]) {
          this.settings.columns[header.value] = {
            visible: true,
          };
        }
      });

      // Initialize tab settings
      if (!this.settings.tabs) {
        this.settings.tabs = {};
      }
      if (this.settings.tabs.allTabAtEnd === undefined) {
        this.settings.tabs.allTabAtEnd = false;
      }

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
