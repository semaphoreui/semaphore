<template>
  <div data-testid="inventorySources-page" :data-state="state">
    <EditDialog
      v-model="editDialog"
      test-id="inventorySourceDialog"
      help-doc="inventory-source"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('inventory_sources_add') : $t('inventory_sources_edit')"
      :max-width="450"
      @save="onSaved()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <InventorySourceForm
          :project-id="projectId"
          :inventory-id="inventoryId"
          :item-id="itemId"
          :need-save="needSave"
          :need-reset="needReset"
          @save="onSave"
          @error="onError"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      v-model="deleteDialog"
      test-id="inventorySourceDeleteDialog"
      :title="$t('inventory_sources_delete')"
      :text="deleteText"
      @yes="deleteSource()"
    />

    <div class="PagePanel">
      <div class="PagePanel__main">
        <v-spacer></v-spacer>
        <v-btn
          v-if="canManage"
          color="primary"
          variant="elevated"
          data-testid="inventorySources-add"
          @click="editSource('new')"
        >{{ $t('inventory_sources_add') }}</v-btn>
      </div>
    </div>

    <v-divider />

    <v-progress-linear
      v-if="loading"
      indeterminate
      color="primary-darken-2"
      bg-opacity=".3"
      data-testid="inventorySources-loading"
    ></v-progress-linear>

    <div v-if="state === 'empty'" class="pa-4" data-testid="inventorySources-empty">
      {{ $t('inventory_sources_empty') }}
    </div>

    <div v-if="state === 'error'" class="pa-4">
      <v-btn
        variant="outlined"
        data-testid="inventorySources-refresh"
        @click="load()"
      >{{ $t('refresh') }}</v-btn>
    </div>

    <v-card
      v-for="source in sources || []"
      :key="source.id"
      variant="outlined"
      class="mx-4 mt-3"
      :data-testid="`inventorySources-card-${source.id}`"
    >
      <v-card-text>
        <div class="d-flex align-start">
          <div class="flex-grow-1" style="min-width: 0">
            <span class="text-body-large font-weight-bold" style="overflow-wrap: anywhere">
              {{ source.name }}
            </span>
            <span class="text-medium-emphasis"> · {{
              source.type === 'git' ? $t('inventory_sources_type_git') : source.type
            }}</span>
            <div class="text-medium-emphasis" style="overflow-wrap: anywhere">
              {{ location(source) }}
            </div>
          </div>

          <v-btn
            v-if="canManage && !$vuetify.display.xs"
            variant="outlined"
            color="primary"
            class="ml-2"
            :disabled="source.status === 'syncing' || syncingIds.includes(source.id)"
            :data-testid="`inventorySources-sync-${source.id}`"
            @click="sync(source)"
          >{{ $t('inventory_sources_sync') }}</v-btn>

          <v-menu v-if="canManage" location="bottom end">
            <template v-slot:activator="{ props }">
              <v-btn
                v-bind="props"
                icon="mdi-dots-vertical"
                variant="text"
                density="comfortable"
                :aria-label="$t('inventory_sources_menu', { name: source.name })"
                :data-testid="`inventorySources-menu-${source.id}`"
              ></v-btn>
            </template>
            <v-list>
              <v-list-item
                prepend-icon="mdi-pencil"
                :title="$t('edit')"
                :data-testid="`inventorySources-edit-${source.id}`"
                @click="editSource(source.id)"
              ></v-list-item>
              <v-list-item
                prepend-icon="mdi-delete"
                :title="$t('inventory_sources_delete')"
                :data-testid="`inventorySources-delete-${source.id}`"
                @click="askDelete(source)"
              ></v-list-item>
            </v-list>
          </v-menu>
        </div>

        <div
          v-if="source.status"
          class="mt-2"
          :data-testid="`inventorySources-status-${source.id}`"
        >
          <v-chip
            size="small"
            class="font-weight-bold"
            :color="statusColor(source.status)"
          >{{ $t(`inventory_sources_status_${source.status}`) }}</v-chip>
          <span
            v-if="source.status !== 'syncing' && source.last_sync"
            class="text-medium-emphasis ml-1"
            :title="new Date(source.last_sync).toLocaleString()"
          >{{ $filters.formatDate(source.last_sync) }}</span>
          <span
            v-if="source.status === 'failed' && source.last_success"
            class="text-medium-emphasis"
            :title="new Date(source.last_success).toLocaleString()"
          > · {{ $t('inventory_sources_last_success', {
            time: $filters.formatDate(source.last_success),
          }) }}</span>
        </div>

        <div
          v-if="source.status === 'success' && source.last_result"
          class="mt-1"
          :data-testid="`inventorySources-result-${source.id}`"
        >
          {{ $t('inventory_sources_totals', counts(source)) }}
          <span class="text-medium-emphasis">
            · {{ $t('inventory_sources_changes', changes(source)) }}
            <template v-if="source.last_result.variables_changed">
              · {{ $t('inventory_sources_variables_changed') }}
            </template>
          </span>
        </div>

        <div
          v-if="!source.status && (source.host_count || source.group_count)"
          class="mt-1"
          :data-testid="`inventorySources-result-${source.id}`"
        >{{ $t('inventory_sources_totals', counts(source)) }}</div>

        <div
          v-if="source.status === 'success' && source.last_result?.skipped_count"
          class="text-medium-emphasis"
          :data-testid="`inventorySources-skipped-${source.id}`"
        >{{ skippedText(source.last_result) }}</div>

        <div
          v-if="source.status === 'failed' && source.last_error"
          class="mt-1 text-error"
          style="overflow-wrap: anywhere"
          :data-testid="`inventorySources-error-${source.id}`"
        >{{ source.last_error }}</div>

        <div
          v-if="source.status === 'failed' && source.last_success"
          class="text-medium-emphasis"
          :data-testid="`inventorySources-result-${source.id}`"
        >{{ $t('inventory_sources_kept', counts(source)) }}</div>

        <div v-if="canManage && $vuetify.display.xs" class="mt-3 InventorySources__phoneActions">
          <v-btn
            variant="outlined"
            color="primary"
            :disabled="source.status === 'syncing' || syncingIds.includes(source.id)"
            :data-testid="`inventorySources-sync-${source.id}`"
            @click="sync(source)"
          >{{ $t('inventory_sources_sync') }}</v-btn>
        </div>
      </v-card-text>
    </v-card>
  </div>
</template>

<script>
import { toRaw } from 'vue';
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import { USER_PERMISSIONS } from '@/lib/constants';
import PermissionsCheck from '@/components/PermissionsCheck';
import ProjectMixin from '@/components/ProjectMixin';
import EditDialog from '@/components/EditDialog.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import InventorySourceForm from '@/components/InventorySourceForm.vue';

const POLL_MS = 2000;

export default {
  mixins: [PermissionsCheck, ProjectMixin],

  components: { EditDialog, YesNoDialog, InventorySourceForm },

  props: {
    inventoryId: Number,
  },

  data() {
    return {
      sources: null,
      repositories: [],
      loading: false,
      failed: false,
      loadSeq: 0,
      timer: null,
      itemId: null,
      editDialog: false,
      deleteDialog: false,
      deleteSourceId: null,
      deleteText: '',
      syncingIds: [],
    };
  },

  computed: {
    canManage() {
      return this.can(USER_PERMISSIONS.manageProjectResources);
    },

    state() {
      if (this.sources == null) return this.failed ? 'error' : 'loading';
      return this.sources.length === 0 ? 'empty' : 'ready';
    },

    hasSyncing() {
      return (this.sources || []).some((source) => source.status === 'syncing');
    },
  },

  watch: {
    inventoryId() {
      this.reload();
    },

    projectId() {
      this.reload();
    },

    hasSyncing(syncing) {
      if (syncing && this.timer == null) {
        this.poll();
      } else if (!syncing && this.timer != null) {
        clearTimeout(this.timer);
        this.timer = null;
      }
    },
  },

  async created() {
    await this.loadRepositories();
    await this.load();
  },

  beforeUnmount() {
    clearTimeout(this.timer);
    this.timer = null;
  },

  methods: {
    sourcesUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/sources`;
    },

    statusColor(status) {
      return { success: 'success', failed: 'error', syncing: 'primary' }[status];
    },

    // The repository name, its branch and the path in it, a path left empty is the root
    location(source) {
      const repository = this.repositories.find((r) => r.id === source.repository_id);
      return [
        repository && repository.name,
        repository && repository.git_branch,
        (source.params || {}).path || '/',
      ].filter(Boolean).join(' · ');
    },

    counts(source) {
      return {
        hosts: this.$t('inventory_sources_hosts', { count: source.host_count }, source.host_count),
        groups: this.$t('inventory_sources_groups', { count: source.group_count }, source.group_count),
      };
    },

    changes(source) {
      const result = source.last_result;
      return {
        added: result.hosts_added + result.groups_added,
        changed: result.hosts_changed + result.groups_changed,
        removed: result.hosts_removed + result.groups_removed,
      };
    },

    skippedText(result) {
      const shown = result.skipped.slice(0, 3).map((s) => `${s.name} (${s.reason})`).join(', ');
      return this.$t('inventory_sources_skipped', {
        count: result.skipped_count,
        items: result.skipped_count > 3 ? `${shown}, …` : shown,
      });
    },

    editSource(id) {
      this.itemId = id;
      this.editDialog = true;
    },

    askDelete(source) {
      this.deleteSourceId = source.id;
      this.deleteText = this.$t('inventory_sources_delete_confirm', {
        name: source.name, ...this.counts(source),
      });
      this.deleteDialog = true;
    },

    async onSaved() {
      EventBus.$emit('i-inventory-counts');
      await this.load();
    },

    async deleteSource() {
      try {
        await axios.delete(`${this.sourcesUrl()}/${this.deleteSourceId}`);
        EventBus.$emit('i-inventory-counts');
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
      await this.load();
    },

    async sync(source) {
      this.syncingIds.push(source.id);
      try {
        await axios.post(`${this.sourcesUrl()}/${source.id}/sync`);
      } catch (err) {
        // Someone else started it, the list shows that
        if (!err.response || err.response.status !== 409) {
          EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
        }
      }
      try {
        await this.load();
      } finally {
        this.syncingIds = this.syncingIds.filter((id) => id !== source.id);
      }
    },

    // Nothing of the previous inventory stays on the screen or in an open dialog
    async reload() {
      // An answer for the previous inventory must not fill the cards
      this.loadSeq += 1;
      this.sources = null;
      this.editDialog = false;
      this.deleteDialog = false;
      await this.loadRepositories();
      await this.load();
    },

    async loadRepositories() {
      try {
        this.repositories = await this.loadProjectResources('repositories');
      } catch {
        // The cards then show the path only
        this.repositories = [];
      }
    },

    // The next read waits for the previous one, so a slow answer is not dropped for a newer request
    poll() {
      const timer = setTimeout(async () => {
        await this.load(true);
        if (toRaw(this.timer) !== timer) return;
        this.timer = null;
        if (this.hasSyncing) this.poll();
      }, POLL_MS);
      this.timer = timer;
    },

    async load(background = false) {
      this.loadSeq += 1;
      const seq = this.loadSeq;
      this.loading = this.sources == null;
      try {
        const response = await axios.get(this.sourcesUrl());
        if (seq !== this.loadSeq) return;
        // A fast sync can start and end between two reads, only its time shows that
        const finished = (this.sources || []).some((before) => response.data.some((now) => (
          now.id === before.id && now.status !== 'syncing'
          && (before.status === 'syncing' || now.last_sync !== before.last_sync))));
        this.sources = response.data;
        this.failed = false;
        if (finished) EventBus.$emit('i-inventory-counts');
      } catch (err) {
        if (seq !== this.loadSeq) return;
        // An empty list would say that the inventory has no sources
        this.failed = this.sources == null;
        if (!background) EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      } finally {
        if (seq === this.loadSeq) this.loading = false;
      }
    },
  },
};
</script>
