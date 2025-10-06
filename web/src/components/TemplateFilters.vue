<template>
  <v-card class="mb-4" elevation="1">
    <v-card-title class="py-2">
      <v-icon class="mr-2">mdi-filter</v-icon>
      Filters
      <v-spacer></v-spacer>
      <v-btn
        v-if="hasActiveFilters"
        text
        small
        @click="clearFilters"
        color="primary"
      >
        Clear All
      </v-btn>
    </v-card-title>

    <v-card-text class="pt-0">
      <v-row>
        <!-- Search -->
        <v-col cols="12" md="4">
          <v-text-field
            v-model="filters.search"
            label="Search"
            placeholder="Search templates..."
            outlined
            dense
            clearable
            prepend-inner-icon="mdi-magnify"
            @input="onFilterChange"
          ></v-text-field>
        </v-col>

        <!-- App Filter -->
        <v-col cols="12" md="2">
          <v-select
            v-model="filters.app"
            :items="appOptions"
            label="App"
            outlined
            dense
            clearable
            @change="onFilterChange"
          ></v-select>
        </v-col>

        <!-- Tags Filter -->
        <v-col cols="12" md="3">
          <v-combobox
            v-model="filters.tags"
            :items="availableTags"
            label="Tags"
            outlined
            dense
            multiple
            chips
            closable-chips
            @change="onFilterChange"
          ></v-combobox>
        </v-col>

        <!-- Labels Filter -->
        <v-col cols="12" md="3">
          <v-combobox
            v-model="filters.labels"
            :items="availableLabels"
            label="Labels"
            outlined
            dense
            multiple
            chips
            closable-chips
            @change="onFilterChange"
          >
            <template v-slot:selection="{ item, index }">
              <v-chip
                :key="index"
                :color="getLabelColor(item)"
                text-color="white"
                small
                close
                @click:close="removeLabel(index)"
              >
                <v-icon left small>{{ getLabelIcon(item) }}</v-icon>
                {{ item }}
              </v-chip>
            </template>
            <template v-slot:item="{ item }">
              <v-list-item>
                <v-list-item-avatar>
                  <v-icon :color="getLabelColor(item)">{{ getLabelIcon(item) }}</v-icon>
                </v-list-item-avatar>
                <v-list-item-content>
                  <v-list-item-title>{{ item }}</v-list-item-title>
                  <v-list-item-subtitle>{{ getLabelDescription(item) }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </template>
          </v-combobox>
        </v-col>
      </v-row>

      <!-- Active Filters Display -->
      <div v-if="hasActiveFilters" class="mt-2">
        <v-chip
          v-for="(filter, key) in activeFilters"
          :key="key"
          class="mr-2 mb-2"
          close
          @click:close="removeFilter(key)"
          color="primary"
          small
        >
          {{ filter }}
        </v-chip>
      </div>
    </v-card-text>
  </v-card>
</template>

<script>
import { getLabelConfig } from '@/lib/labelConfig';

export default {
  name: 'TemplateFilters',
  props: {
    availableTags: {
      type: Array,
      default: () => [],
    },
    availableLabels: {
      type: Array,
      default: () => [],
    },
    appOptions: {
      type: Array,
      default: () => [
        { text: 'Ansible', value: 'ansible' },
        { text: 'Terraform', value: 'terraform' },
        { text: 'OpenTofu', value: 'tofu' },
        { text: 'Terragrunt', value: 'terragrunt' },
        { text: 'Bash', value: 'bash' },
        { text: 'PowerShell', value: 'powershell' },
        { text: 'Python', value: 'python' },
        { text: 'SCC', value: 'scc' },
        { text: 'Pulumi', value: 'pulumi' },
      ],
    },
  },
  data() {
    return {
      filters: {
        search: '',
        app: null,
        tags: [],
        labels: [],
      },
    };
  },
  computed: {
    hasActiveFilters() {
      return this.filters.search
        || this.filters.app
        || this.filters.tags.length > 0
        || this.filters.labels.length > 0;
    },
    activeFilters() {
      const active = {};
      if (this.filters.search) {
        active.Search = `"${this.filters.search}"`;
      }
      if (this.filters.app) {
        const appName = this.appOptions.find((app) => app.value === this.filters.app)?.text
          || this.filters.app;
        active.App = appName;
      }
      if (this.filters.tags.length > 0) {
        active.Tags = this.filters.tags.join(', ');
      }
      if (this.filters.labels.length > 0) {
        active.Labels = this.filters.labels.join(', ');
      }
      return active;
    },
  },
  methods: {
    onFilterChange() {
      this.$emit('filter-change', { ...this.filters });
    },
    clearFilters() {
      this.filters = {
        search: '',
        app: null,
        tags: [],
        labels: [],
      };
      this.onFilterChange();
    },
    removeFilter(filterKey) {
      switch (filterKey) {
        case 'Search':
          this.filters.search = '';
          break;
        case 'App':
          this.filters.app = null;
          break;
        case 'Tags':
          this.filters.tags = [];
          break;
        case 'Labels':
          this.filters.labels = [];
          break;
        default:
          break;
      }
      this.onFilterChange();
    },
    getLabelColor(labelName) {
      const config = getLabelConfig(labelName);
      return config.color;
    },
    getLabelIcon(labelName) {
      const config = getLabelConfig(labelName);
      return config.categoryIcon;
    },
    getLabelDescription(labelName) {
      const config = getLabelConfig(labelName);
      return config.description;
    },
    removeLabel(index) {
      this.filters.labels.splice(index, 1);
      this.onFilterChange();
    },
  },
};
</script>
