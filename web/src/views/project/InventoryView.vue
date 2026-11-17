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
      </v-tab>
      <v-tab data-testid="inventoryView-tab-variables" :to="`${listPath}/${inventoryId}/variables`">
        {{ $t('inventory_tab_variables') }}
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;" />

    <router-view
      :project-id="projectId"
      :inventory-id="inventoryId"
      :user-permissions="userPermissions"
      :is-admin="isAdmin"
    ></router-view>
  </div>
</template>

<script>
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import ProjectMixin from '@/components/ProjectMixin';

export default {
  mixins: [ProjectMixin],

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
    await this.load();
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async load() {
      this.loadSeq += 1;
      const seq = this.loadSeq;
      this.state = 'loading';
      this.notFound = false;
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
    },
  },
};
</script>
