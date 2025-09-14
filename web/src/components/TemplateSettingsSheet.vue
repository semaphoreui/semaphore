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
        v-model="allTabAtEnd"
        @change="saveAllTabSettings()"
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
    projectId: Number,
    views: Array,
  },

  components: { PageBottomSheet },

  data() {
    return {
      sheet: false,
      settings: null,
      allTabAtEnd: false,
    };
  },

  watch: {
    async sheet(val) {
      this.$emit('input', val);
    },

    async value(val) {
      this.sheet = val;
      if (val) {
        this.updateAllTabFromViews();
      }
    },

    headers() {
      this.loadSettings();
    },

    views() {
      this.updateAllTabFromViews();
    },
  },

  created() {
    this.loadSettings();
    this.updateAllTabFromViews();
  },

  methods: {
    saveSettings() {
      localStorage.setItem(`${this.tableName}__settings`, JSON.stringify(this.settings));
      this.loadSettings();
    },

    updateAllTabFromViews() {
      if (!this.views || this.views.length === 0) {
        this.allTabAtEnd = false;
        return;
      }

      // Calculate allTabAtEnd from the views array using the same logic as backend
      let allView = null;
      let maxCustomPosition = -1;

      for (let i = 0; i < this.views.length; i += 1) {
        const view = this.views[i];
        if (view.type === 'all' && !view.hidden) {
          allView = view;
        } else if (!view.hidden && view.position > maxCustomPosition) {
          maxCustomPosition = view.position;
        }
      }

      // If no All view exists or it's hidden, default to beginning
      if (!allView) {
        this.allTabAtEnd = false;
        return;
      }

      // If All view position is greater than all other visible views, it should be at end
      this.allTabAtEnd = allView.position > maxCustomPosition;
    },

    async saveAllTabSettings() {
      // Emit the change to parent component which will handle the API call
      this.$emit('change', {
        settings: { ...this.settings, tabs: { allTabAtEnd: this.allTabAtEnd } },
        headers: this.headers.filter((header) => {
          const column = this.settings.columns[header.value];
          return !column || column.visible;
        }),
        allTabAtEnd: this.allTabAtEnd,
      });
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

      // Initialize column settings
      this.headers.forEach((header) => {
        if (!this.settings.columns[header.value]) {
          this.settings.columns[header.value] = {
            visible: true,
          };
        }
      });

      this.$emit('change', {
        settings: { ...this.settings, tabs: { allTabAtEnd: this.allTabAtEnd } },
        headers: this.headers.filter((header) => {
          const column = this.settings.columns[header.value];
          return !column || column.visible;
        }),
      });
    },
  },
};
</script>
