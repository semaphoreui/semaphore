<template>
  <div data-testid="inventoryGroups-page" :data-state="state">
    <EditDialog
      v-model="editDialog"
      test-id="inventoryGroupDialog"
      help-doc="inventory-group"
      :save-button-text="saveText"
      :title="itemId === 'new' ? $t('inventory_groups_add') : $t('inventory_groups_edit')"
      :max-width="450"
      @save="onSaved()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <InventoryGroupForm
          :project-id="projectId"
          :inventory-id="inventoryId"
          :item-id="itemId"
          :groups="groups || []"
          :sources="sourceNames"
          :need-save="needSave"
          :need-reset="needReset"
          @save="onSave"
          @error="onError"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      v-model="deleteDialog"
      test-id="inventoryGroupDeleteDialog"
      :title="$t('inventory_groups_delete')"
      :text="$t('inventory_groups_delete_confirm', { name: deleteName })"
      @yes="deleteGroup()"
    />

    <div class="PagePanel">
      <div class="PagePanel__main">
        <v-text-field
            v-model="searchText"
            :label="$t('inventory_groups_search')"
            prepend-inner-icon="mdi-magnify"
            density="compact"
            variant="outlined"
            hide-details
            clearable
            style="max-width: 320px"
            data-testid="inventoryGroups-search"
        />
        <v-select
          v-if="sources.length > 0 || sourceFilter !== ''"
          v-model="sourceFilter"
          :items="sourceFilterItems"
          :label="$t('inventory_source_filter')"
          density="compact"
          variant="outlined"
          hide-details
          style="max-width: 240px"
          data-testid="inventoryGroups-sourceFilter"
        />
        <v-spacer></v-spacer>
        <v-btn
            v-if="canManage"
            color="primary"
            variant="elevated"
            data-testid="inventoryGroups-add"
            @click="editGroup('new')"
        >{{ $t('inventory_groups_add') }}</v-btn>
      </div>
    </div>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="rows"
      :items-per-page="-1"
      :loading="loading"
      hide-default-footer
      class="mt-2"
      data-testid="inventoryGroups-table"
      :data-state="state"
    >
      <template v-slot:no-data>
        <span data-testid="inventoryGroups-empty">{{ $t('inventory_groups_empty') }}</span>
      </template>

      <template v-slot:item.name="{ item }">
        <div>
          <span
            class="text-wrap"
            style="overflow-wrap: anywhere"
            data-testid="inventoryGroups-name"
          >
            {{ item.name }}
          </span>
          <div
            v-if="item.source_id != null"
            class="text-body-small text-medium-emphasis"
            style="overflow-wrap: anywhere"
            :data-testid="`inventoryGroups-source-${item.id}`"
          >
            <v-icon size="x-small">mdi-sync</v-icon>
            {{ sourcesLoaded
              ? (sourceNames[item.source_id] || $t('inventory_source_unknown'))
              : '' }}
          </div>
        </div>
      </template>

      <template v-slot:item.description="{ item }">
        <span
          class="text-wrap"
          style="overflow-wrap: anywhere"
          data-testid="inventoryGroups-description"
        >
          {{ item.description }}
        </span>
      </template>

      <template v-slot:item.host_count="{ item }">
        <v-btn
          variant="text"
          density="compact"
          class="px-1"
          style="min-width: 0"
          :to="`/project/${projectId}/inventory/${inventoryId}/hosts?group=${item.id}`"
          :aria-label="$t('inventory_groups_hosts_label', {
            name: item.name,
            count: item.host_count,
          })"
          data-testid="inventoryGroups-hosts"
        >{{ item.host_count }}</v-btn>
      </template>

      <template v-slot:item.children="{ item }">
        <div class="text-wrap">
          <v-chip
            v-for="id in item.children"
            :key="id"
            size="small"
            class="InventoryGroups__chip mr-1 my-1"
            data-testid="inventoryGroups-child"
          >{{ groupNames[id] }}</v-chip>
        </div>
      </template>

      <template v-slot:item.variables="{ item }">
        <span v-if="item.variable_count" data-testid="inventoryGroups-variables">
          {{ item.variable_count }}
        </span>
        <span v-else data-testid="inventoryGroups-noVariables">&mdash;</span>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-group density="compact" data-testid="inventoryGroups-rowActions">
          <v-btn
            v-if="item.source_id == null"
            :disabled="busyIds.includes(item.id)"
            :aria-label="$t('inventory_groups_delete')"
            data-testid="inventoryGroups-rowDelete"
            @click="askDelete(item)"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn
            :disabled="busyIds.includes(item.id)"
            :aria-label="$t('inventory_groups_edit')"
            data-testid="inventoryGroups-rowEdit"
            @click="editGroup(item.id)"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-group>
      </template>
    </v-data-table>
  </div>
</template>

<style>
/* A long group name wraps inside its chip instead of widening the table */
.v-chip.v-chip--size-small.InventoryGroups__chip {
  height: auto;
  min-height: var(--v-chip-height);
  padding-top: 4px;
  padding-bottom: 4px;
  white-space: normal;
  overflow-wrap: anywhere;
}
</style>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import { USER_PERMISSIONS } from '@/lib/constants';
import PermissionsCheck from '@/components/PermissionsCheck';
import ProjectMixin from '@/components/ProjectMixin';
import EditDialog from '@/components/EditDialog.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import InventoryGroupForm from '@/components/InventoryGroupForm.vue';

export default {
  mixins: [PermissionsCheck, ProjectMixin],

  components: { EditDialog, YesNoDialog, InventoryGroupForm },

  props: {
    inventoryId: Number,
  },

  data() {
    return {
      groups: null,
      sources: [],
      sourcesLoaded: false,
      sourceFilter: '',
      loading: false,
      searchText: '',
      loadSeq: 0,
      sourcesSeq: 0,
      itemId: null,
      editDialog: false,
      deleteDialog: false,
      deleteGroupId: null,
      deleteName: '',
      busyIds: [],
    };
  },

  computed: {
    canManage() {
      return this.can(USER_PERMISSIONS.manageProjectResources);
    },

    rows() {
      const search = (this.searchText || '').toLowerCase();
      return (this.groups || []).filter((group) => group.name.toLowerCase().includes(search)
        && (this.sourceFilter === '' || (this.sourceFilter === 'manual'
          ? group.source_id == null
          : group.source_id === this.sourceFilter)));
    },

    sourceFilterItems() {
      return [
        { title: this.$t('inventory_source_all'), value: '' },
        { title: this.$t('inventory_source_manual'), value: 'manual' },
        ...this.sources.map((source) => ({ title: source.name, value: source.id })),
      ];
    },

    sourceNames() {
      return Object.fromEntries(this.sources.map((source) => [source.id, source.name]));
    },

    // An imported group has nothing to save
    saveText() {
      const group = (this.groups || []).find((g) => g.id === this.itemId);
      if (group != null && group.source_id != null) return undefined;
      return this.itemId === 'new' ? this.$t('create') : this.$t('save');
    },

    state() {
      if (this.loading || this.groups == null) return 'loading';
      return this.groups.length === 0 ? 'empty' : 'ready';
    },

    groupNames() {
      return Object.fromEntries((this.groups || []).map((group) => [group.id, group.name]));
    },

    headers() {
      return [
        { title: this.$t('name'), key: 'name', sortable: false },
        { title: this.$t('description'), key: 'description', sortable: false },
        { title: this.$t('inventory_tab_hosts'), key: 'host_count', sortable: false },
        { title: this.$t('inventory_groups_children'), key: 'children', sortable: false },
        { title: this.$t('inventory_tab_variables'), key: 'variables', sortable: false },
        ...(this.canManage ? [{
          title: '', key: 'actions', sortable: false, align: 'end',
        }] : []),
      ];
    },
  },

  watch: {
    inventoryId() {
      this.reload();
    },

    projectId() {
      this.reload();
    },
  },

  async created() {
    await Promise.all([this.load(), this.loadSources()]);
  },

  methods: {
    groupsUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/groups`;
    },

    editGroup(id) {
      this.itemId = id;
      this.editDialog = true;
    },

    askDelete(group) {
      this.deleteGroupId = group.id;
      this.deleteName = group.name;
      this.deleteDialog = true;
    },

    async onSaved() {
      EventBus.$emit('i-inventory-counts');
      await this.load();
    },

    async deleteGroup() {
      const id = this.deleteGroupId;
      this.busyIds.push(id);
      try {
        await axios.delete(`${this.groupsUrl()}/${id}`);
        EventBus.$emit('i-inventory-counts');
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
      await this.load();
      this.busyIds = this.busyIds.filter((busyId) => busyId !== id);
    },

    // Nothing of the previous inventory stays on the screen or in an open dialog
    reload() {
      this.groups = null;
      this.sources = [];
      this.sourcesLoaded = false;
      this.sourceFilter = '';
      this.editDialog = false;
      this.deleteDialog = false;
      this.load();
      this.loadSources();
    },

    // Without the names the rows still show that a group is imported
    async loadSources() {
      this.sourcesSeq += 1;
      const seq = this.sourcesSeq;
      try {
        const response = await axios.get(
          `/api/project/${this.projectId}/inventory/${this.inventoryId}/sources`,
        );
        if (seq !== this.sourcesSeq) return;
        this.sources = Array.isArray(response.data) ? response.data : [];
      } catch (err) {
        if (seq !== this.sourcesSeq) return;
        this.sources = [];
      }
      this.sourcesLoaded = true;
    },

    async load() {
      this.loadSeq += 1;
      const seq = this.loadSeq;
      this.loading = true;
      try {
        const response = await axios.get(this.groupsUrl());
        if (seq !== this.loadSeq) return;
        this.groups = response.data;
      } catch (err) {
        if (seq !== this.loadSeq) return;
        this.groups = [];
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      } finally {
        if (seq === this.loadSeq) this.loading = false;
      }
    },
  },
};
</script>
