<template>
  <div v-if="state === 'error'" data-testid="inventoryView-page" data-state="error">
    <v-alert type="info" variant="tonal" class="ma-4" data-testid="inventoryView-notFound">
      {{ notFound ? $t('inventory_not_found') : $t('inventory_not_managed') }}
      <router-link :to="listPath" data-testid="inventoryView-back">
        {{ $t('inventory') }}
      </router-link>
    </v-alert>
  </div>
  <div v-else-if="state === 'loading'" data-testid="inventoryView-page" data-state="loading">
    <v-progress-linear
      indeterminate
      color="primary-darken-2"
      bg-opacity=".3"
      data-testid="inventoryView-loading"
    ></v-progress-linear>
  </div>
  <div v-else data-testid="inventoryView-page" data-state="ready">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()" data-testid="inventoryView-drawerToggle"/>
      <v-toolbar-title class="breadcrumbs flex-0-1" data-testid="inventoryView-title">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="listPath"
          data-testid="inventoryView-breadcrumbInventory"
        >{{ $t('inventory') }}</router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item" data-testid="inventoryView-name">{{ inventory.name }}</span>
      </v-toolbar-title>
    </v-toolbar>

    <v-tabs data-testid="inventoryView-tabs">
      <v-tab data-testid="inventoryView-tab-hosts" :to="`${listPath}/${inventoryId}/hosts`">
        {{ $t('inventory_tab_hosts') }}
        <v-chip
          v-if="hostCount !== 0"
          size="x-small"
          class="ml-2 font-weight-bold"
          style="cursor: inherit"
          data-testid="inventoryView-hostCount"
        >{{ hostCount ?? '?' }}</v-chip>
      </v-tab>
      <v-tab data-testid="inventoryView-tab-groups" :to="`${listPath}/${inventoryId}/groups`">
        {{ $t('inventory_tab_groups') }}
        <v-chip
          v-if="groupCount !== 0"
          size="x-small"
          class="ml-2 font-weight-bold"
          style="cursor: inherit"
          data-testid="inventoryView-groupCount"
        >{{ groupCount ?? '?' }}</v-chip>
      </v-tab>
      <v-tab data-testid="inventoryView-tab-variables" :to="`${listPath}/${inventoryId}/variables`">
        {{ $t('inventory_tab_variables') }}
        <v-chip
          v-if="inventory.variable_count"
          size="x-small"
          class="ml-2 font-weight-bold"
          style="cursor: inherit"
          data-testid="inventoryView-variableCount"
        >{{ inventory.variable_count }}</v-chip>
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;" />

    <router-view
      :project-id="projectId"
      :inventory-id="inventoryId"
      :user-permissions="userPermissions"
      :is-admin="isAdmin"
      :system-info="systemInfo"
    ></router-view>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import ProjectMixin from '@/components/ProjectMixin';
import PageMixin from '@/components/PageMixin';

export default {
  mixins: [ProjectMixin, PageMixin],

  props: {
    userPermissions: Number,
    isAdmin: Boolean,
  },

  data() {
    return {
      inventory: null,
      state: 'loading',
      notFound: false,
      loadSeq: 0,
      hostCount: 0,
      groupCount: 0,
      countSeq: 0,
    };
  },

  computed: {
    inventoryId() {
      return Number(this.$route.params.inventoryId);
    },

    listPath() {
      return `/project/${this.projectId}/inventory`;
    },
  },

  watch: {
    async inventoryId() {
      await this.load();
    },
    async projectId() {
      await this.load();
    },
  },

  async created() {
    EventBus.$on('i-inventory-variables', this.onVariablesSaved);
    EventBus.$on('i-inventory-counts', this.loadCounts);
    await this.load();
  },

  beforeUnmount() {
    EventBus.$off('i-inventory-variables', this.onVariablesSaved);
    EventBus.$off('i-inventory-counts', this.loadCounts);
  },

  methods: {
    // The tab that saved may be gone by now, so the save comes over the bus
    async onVariablesSaved({ projectId, inventoryId }) {
      if (this.state !== 'ready' || projectId !== this.projectId || inventoryId !== this.inventoryId) {
        return;
      }
      this.loadSeq += 1;
      const seq = this.loadSeq;
      try {
        const inventory = await this.loadProjectResource('inventory', this.inventoryId);
        if (seq === this.loadSeq) this.inventory = inventory;
      } catch {
        // The number stays as it was until the page is opened again
      }
    },

    async loadCounts() {
      if (this.state !== 'ready') return;
      this.countSeq += 1;
      const seq = this.countSeq;
      const url = `/api/project/${this.projectId}/inventory/${this.inventoryId}`;
      const [hosts, groups] = await Promise.allSettled([
        axios.get(`${url}/hosts`, { params: { count: 1 } }),
        axios.get(`${url}/groups`),
      ]);
      if (seq !== this.countSeq) return;
      const hostCount = hosts.status === 'fulfilled'
        ? Number(hosts.value.headers['x-total-count'])
        : NaN;
      this.hostCount = Number.isFinite(hostCount) ? hostCount : null;
      this.groupCount = groups.status === 'fulfilled' ? groups.value.data.length : null;
    },

    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async load() {
      this.loadSeq += 1;
      const seq = this.loadSeq;
      this.state = 'loading';
      this.notFound = false;
      this.hostCount = 0;
      this.groupCount = 0;
      // Counts of the inventory shown before must not land on this one
      this.countSeq += 1;
      try {
        const inventory = await this.loadProjectResource('inventory', this.inventoryId);
        if (seq !== this.loadSeq) return;
        this.inventory = inventory;
        this.state = inventory.type === 'managed' ? 'ready' : 'error';
      } catch (err) {
        if (seq !== this.loadSeq) return;
        this.state = 'error';
        this.notFound = !!err.response && err.response.status === 404;
        if (!this.notFound) {
          EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
        }
      }
      await this.loadCounts();
    },
  },
};
</script>
