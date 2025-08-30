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
import axios from 'axios';

export default {
  props: {
    value: Boolean,
    tableName: String,
    headers: Array,
    projectId: Number,
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
        await this.loadAllTabSettings();
      }
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

    async loadAllTabSettings() {
      try {
        const response = await axios.get(`/api/project/${this.projectId}/views/all-tab-settings`);
        this.allTabAtEnd = response.data.allTabAtEnd;
      } catch (error) {
        console.error('Failed to load All tab settings:', error);
        this.allTabAtEnd = false;
      }
    },

    async saveAllTabSettings() {
      try {
        await axios.post(`/api/project/${this.projectId}/views/all-tab-settings`, {
          allTabAtEnd: this.allTabAtEnd,
        });

        // Emit the change to parent component
        this.$emit('change', {
          settings: { ...this.settings, tabs: { allTabAtEnd: this.allTabAtEnd } },
          headers: this.headers.filter((header) => {
            const column = this.settings.columns[header.value];
            return !column || column.visible;
          }),
        });
      } catch (error) {
        console.error('Failed to save All tab settings:', error);
        // Revert the change if save failed
        this.allTabAtEnd = !this.allTabAtEnd;
      }
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
