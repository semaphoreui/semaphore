<template>
  <div data-testid="inventoryHosts-page" :data-state="state">
    <EditDialog
      v-model="editDialog"
      test-id="inventoryHostDialog"
      help-doc="inventory-host"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('inventory_hosts_add') : $t('inventory_hosts_edit')"
      :max-width="450"
      @save="onSaved()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <InventoryHostForm
          :project-id="projectId"
          :inventory-id="inventoryId"
          :item-id="itemId"
          :groups="groups || []"
          :need-save="needSave"
          :need-reset="needReset"
          @save="onSave"
          @error="onError"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      v-model="deleteDialog"
      test-id="inventoryHostDeleteDialog"
      :title="$t('inventory_hosts_delete')"
      :text="$t('inventory_hosts_delete_confirm', { name: deleteName })"
      @yes="deleteHost()"
    />

    <v-dialog v-model="groupsDialog" :persistent="applying" max-width="400">
      <v-card data-testid="inventoryHosts-groupsDialog">
        <v-card-title style="white-space: normal">
          {{ $t(
            groupsDialogAdd
              ? 'inventory_hosts_add_to_groups_title'
              : 'inventory_hosts_remove_from_groups_title',
            { count: selected.length },
            selected.length,
          ) }}
        </v-card-title>
        <v-card-text>
          <v-alert
            :model-value="!!groupsError"
            color="error"
            class="mb-4"
            data-testid="inventoryHosts-groupsDialogError"
          >{{ groupsError }}</v-alert>
          <v-select
            v-model="pickedGroups"
            :items="dialogGroups"
            item-title="name"
            item-value="id"
            :label="$t('inventory_tab_groups')"
            multiple
            chips
            :disabled="applying"
            variant="outlined"
            density="compact"
            hide-details
            data-testid="inventoryHosts-groupsDialogGroups"
          />
          <div class="mt-3 text-medium-emphasis" data-testid="inventoryHosts-groupsDialogHosts">
            {{ selectedHosts.map((host) => host.name).join(', ') }}
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            variant="text"
            :disabled="applying"
            data-testid="inventoryHosts-groupsDialogCancel"
            @click="groupsDialog = false"
          >{{ $t('cancel') }}</v-btn>
          <v-btn
            color="primary"
            variant="text"
            :disabled="pickedGroups.length === 0 || selected.length === 0 || applying"
            data-testid="inventoryHosts-groupsDialogApply"
            @click="applyGroups()"
          >{{ groupsDialogAdd ? $t('add') : $t('inventory_hosts_remove') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <div
      class="PagePanel"
      :style="hasSelection ? { position: 'sticky', top: 0 } : {}"
    >
      <div
        class="PagePanel__main"
        style="grid-area: 1 / 1"
        :style="hasSelection ? { visibility: 'hidden' } : {}"
        :inert="hasSelection || undefined"
        data-testid="inventoryHosts-toolbar"
      >
        <v-text-field
          v-model="searchText"
          :label="$t('inventory_hosts_search')"
          prepend-inner-icon="mdi-magnify"
          density="compact"
          variant="outlined"
          hide-details
          clearable
          style="max-width: 320px"
          data-testid="inventoryHosts-search"
        />
        <v-select
          :model-value="groupFilter"
          :items="groupFilterItems"
          :label="$t('inventory_hosts_group_filter')"
          density="compact"
          variant="outlined"
          hide-details
          style="max-width: 240px"
          data-testid="inventoryHosts-groupFilter"
          @update:model-value="setGroupFilter"
        />
        <v-spacer></v-spacer>
        <v-btn
          v-if="canManage"
          color="primary"
          variant="elevated"
          data-testid="inventoryHosts-add"
          @click="editHost('new')"
        >{{ $t('inventory_hosts_add') }}</v-btn>
      </div>

      <div
        v-if="hasSelection"
        class="PagePanel__multiselect"
        style="grid-area: 1 / 1"
        data-testid="inventoryHosts-selectionBar"
      >

        <span class="font-weight-medium flex-grow-1" role="status">
          {{ $t('inventory_hosts_selected', { count: selected.length }, selected.length) }}
        </span>

        <v-btn
          color="primary"
          variant="elevated"
          data-testid="inventoryHosts-addToGroups"
          @click="openGroupsDialog(true)"
        >{{ $t('inventory_hosts_add_to_groups') }}</v-btn>

        <v-btn
          variant="text"
          data-testid="inventoryHosts-removeFromGroups"
          @click="openGroupsDialog(false)"
        >{{ $t('inventory_hosts_remove_from_groups') }}</v-btn>

        <v-btn
          color="white"
          icon
          variant="text"
          density="compact"
          :aria-label="$t('inventory_hosts_clear_selection')"
          data-testid="inventoryHosts-clearSelection"
          @click="selected = []"
        >
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </div>
    </div>

    <v-divider />

    <v-data-table
      v-model="selected"
      :show-select="canManage"
      :select-all-label="$t('inventory_hosts_select_all')"
      item-value="id"
      :headers="headers"
      :items="rows || []"
      :items-per-page="-1"
      :loading="loading"
      hide-default-footer
      class="mt-2"
      data-testid="inventoryHosts-table"
      :data-state="state"
    >
      <template v-slot:no-data>
        <span data-testid="inventoryHosts-empty">{{ $t('inventory_hosts_empty') }}</span>
      </template>

      <template v-slot:item.data-table-select="{ item, props }">
        <v-checkbox-btn
          v-bind="props"
          :aria-label="$t('inventory_hosts_select', { name: item.name })"
          data-testid="inventoryHosts-select"
        />
      </template>

      <template v-slot:item.name="{ item }">
        <span class="text-wrap" style="overflow-wrap: anywhere" data-testid="inventoryHosts-name">
          <router-link
            v-if="features.hosts && item.host_id"
            :to="`/project/${projectId}/hosts/${item.host_id}`"
          >{{ item.name }}</router-link>
          <template v-else>{{ item.name }}</template>
        </span>
      </template>
      <template v-slot:item.last_run="{ item }">
        <span v-if="item.last_status" data-testid="inventoryHosts-lastRun">
          <HostStatus :status="item.last_status" />
          <TaskLink
            :task-id="item.last_task_id"
            :label="$filters.formatDate(item.last_run)"
            class="ml-2"
          />
        </span>
        <span v-else class="text-medium-emphasis" data-testid="inventoryHosts-neverRun">
          {{ $t('inventory_hosts_never_run') }}
        </span>
      </template>

      <template v-slot:item.description="{ item }">
        <span
          class="text-wrap"
          style="overflow-wrap: anywhere"
          data-testid="inventoryHosts-description"
        >
          {{ item.description }}
        </span>
      </template>

      <template v-slot:item.groups="{ item }">
        <div class="text-wrap">
          <v-chip
            v-for="id in knownGroups[item.id]"
            :key="id"
            size="small"
            class="InventoryHosts__chip mr-1 my-1"
            data-testid="inventoryHosts-groups"
          >{{ groupNames[id] }}</v-chip>
        </div>
      </template>

      <template v-slot:item.variables="{ item }">
        <span v-if="item.variable_count" data-testid="inventoryHosts-variables">
          {{ item.variable_count }}
        </span>
        <span v-else data-testid="inventoryHosts-noVariables">&mdash;</span>
      </template>

      <template v-slot:item.enabled="{ item }">
        <v-switch
          :model-value="item.enabled"
          :disabled="!canManage || loading || busyIds.includes(item.id)"
          color="primary"
          hide-details
          density="compact"
          :aria-label="$t('inventory_hosts_enabled_label', { name: item.name })"
          data-testid="inventoryHosts-enabled"
          @update:model-value="setEnabled(item, $event)"
        />
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-group density="compact" data-testid="inventoryHosts-rowActions">
          <v-btn
            :aria-label="$t('inventory_hosts_delete')"
            data-testid="inventoryHosts-rowDelete"
            @click="askDelete(item)"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn
            :aria-label="$t('inventory_hosts_edit')"
            data-testid="inventoryHosts-rowEdit"
            @click="editHost(item.id)"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-group>
      </template>
    </v-data-table>

    <div
      class="d-flex align-center justify-end mt-2 mr-2 mb-4"
      data-testid="inventoryHosts-pagination"
    >
      <v-btn
        :disabled="loading || offset === 0"
        icon
        variant="text"
        :aria-label="$t('inventory_hosts_prev_page')"
        data-testid="inventoryHosts-prevPage"
        @click="setOffset(Math.max(0, offset - count))"
      >
        <v-icon>mdi-chevron-left</v-icon>
      </v-btn>

      <span class="mx-2 text-medium-emphasis" data-testid="inventoryHosts-range">
        {{ rangeStart }}&ndash;{{ rangeEnd }} {{ $t('inventory_hosts_range_of') }} {{ total }}
      </span>

      <v-btn
        :disabled="loading || offset + count >= total"
        icon
        variant="text"
        :aria-label="$t('inventory_hosts_next_page')"
        data-testid="inventoryHosts-nextPage"
        @click="setOffset(offset + count)"
      >
        <v-icon>mdi-chevron-right</v-icon>
      </v-btn>
    </div>
  </div>
</template>

<style>
/* A long group name wraps inside its chip instead of widening the table */
.v-chip.v-chip--size-small.InventoryHosts__chip {
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
import PageMixin from '@/components/PageMixin';
import HostStatus from '@/components/HostStatus.vue';
import TaskLink from '@/components/TaskLink.vue';
import EditDialog from '@/components/EditDialog.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import InventoryHostForm from '@/components/InventoryHostForm.vue';

export default {
  mixins: [PermissionsCheck, ProjectMixin, PageMixin],

  components: {
    EditDialog, YesNoDialog, InventoryHostForm, HostStatus, TaskLink,
  },

  props: {
    inventoryId: Number,
  },

  data() {
    return {
      rows: null,
      groups: null,
      total: 0,
      loading: false,
      offset: 0,
      count: 50,
      searchText: '',
      searchTimer: null,
      loadSeq: 0,
      groupsSeq: 0,
      itemId: null,
      editDialog: false,
      deleteDialog: false,
      deleteHostId: null,
      deleteName: '',
      busyIds: [],
      selected: [],
      groupsDialog: false,
      groupsDialogAdd: true,
      pickedGroups: [],
      groupsError: null,
      applying: false,
    };
  },

  computed: {
    inventoryUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}`;
    },

    search() {
      return this.$route.query.search || '';
    },

    groupFilter() {
      if (this.$route.query.ungrouped === '1') return 'ungrouped';
      const id = Number(this.$route.query.group);
      return Number.isInteger(id) && id > 0 ? id : 'any';
    },

    groupFilterItems() {
      return [
        { title: this.$t('inventory_hosts_any_group'), value: 'any' },
        { title: this.$t('inventory_hosts_no_groups'), value: 'ungrouped' },
        ...(this.groups || []).map((group) => ({ title: group.name, value: group.id })),
      ];
    },

    groupNames() {
      return Object.fromEntries((this.groups || []).map((group) => [group.id, group.name]));
    },

    // A group created after the groups were loaded has no name to show yet
    knownGroups() {
      return Object.fromEntries((this.rows || []).map((host) => [
        host.id,
        host.groups.filter((id) => this.groupNames[id] !== undefined),
      ]));
    },

    hasSelection() {
      return this.selected.length > 0;
    },

    selectedHosts() {
      return (this.rows || []).filter((host) => this.selected.includes(host.id));
    },

    dialogGroups() {
      const groups = this.groups || [];
      if (this.groupsDialogAdd) return groups;
      return groups.filter((group) => this.selectedHosts
        .some((host) => host.groups.includes(group.id)));
    },

    canManage() {
      return this.can(USER_PERMISSIONS.manageProjectResources);
    },

    rangeStart() {
      return this.total === 0 ? 0 : this.offset + 1;
    },

    rangeEnd() {
      return this.offset + (this.rows ? this.rows.length : 0);
    },

    state() {
      if (this.loading || this.rows == null) return 'loading';
      return this.rows.length === 0 ? 'empty' : 'ready';
    },

    headers() {
      return [
        { title: this.$t('name'), key: 'name', sortable: false },
        ...(this.features.hosts
          ? [{ title: this.$t('inventory_hosts_last_run'), key: 'last_run', sortable: false }]
          : []),
        { title: this.$t('description'), key: 'description', sortable: false },
        { title: this.$t('inventory_tab_groups'), key: 'groups', sortable: false },
        { title: this.$t('inventory_tab_variables'), key: 'variables', sortable: false },
        { title: this.$t('enabled'), key: 'enabled', sortable: false },
        ...(this.canManage ? [{
          title: '', key: 'actions', sortable: false, align: 'end',
        }] : []),
      ];
    },
  },

  watch: {
    inventoryUrl() {
      this.offset = 0;
      this.selected = [];
      this.groups = null;
      this.editDialog = false;
      this.deleteDialog = false;
      this.groupsDialog = false;
      this.load();
      this.loadGroups();
    },

    search() {
      this.searchText = this.search;
      this.offset = 0;
      this.selected = [];
      this.load();
    },

    groupFilter() {
      this.offset = 0;
      this.selected = [];
      this.load();
    },

    searchText(value) {
      clearTimeout(this.searchTimer);
      if ((value || '') === this.search) return;
      this.searchTimer = setTimeout(() => {
        this.$router.replace({ query: { ...this.$route.query, search: value || undefined } });
      }, 300);
    },
  },

  async created() {
    this.searchText = this.search;
    await Promise.all([this.load(), this.loadGroups()]);
  },

  beforeUnmount() {
    clearTimeout(this.searchTimer);
  },

  methods: {
    editHost(id) {
      this.itemId = id;
      this.editDialog = true;
    },

    askDelete(host) {
      this.deleteHostId = host.id;
      this.deleteName = host.name;
      this.deleteDialog = true;
    },

    setOffset(offset) {
      this.offset = offset;
      this.selected = [];
      this.load();
    },

    setGroupFilter(value) {
      this.$router.replace({
        query: {
          ...this.$route.query,
          group: typeof value === 'number' ? value : undefined,
          ungrouped: value === 'ungrouped' ? '1' : undefined,
        },
      });
    },

    openGroupsDialog(add) {
      this.groupsDialogAdd = add;
      this.pickedGroups = [];
      this.groupsError = null;
      this.groupsDialog = true;
    },

    async applyGroups() {
      this.applying = true;
      this.groupsError = null;
      try {
        await axios.put(`${this.inventoryUrl}/host_groups`, {
          hosts: this.selected,
          add: this.groupsDialogAdd ? this.pickedGroups : [],
          remove: this.groupsDialogAdd ? [] : this.pickedGroups,
        });
        this.groupsDialog = false;
        this.selected = [];
        EventBus.$emit('i-inventory-counts');
        await this.load();
      } catch (err) {
        this.groupsError = getErrorMessage(err);
        // Someone else may have deleted a host or a group, a retry must not fail the same way
        await Promise.all([this.load(), this.loadGroups()]);
        this.pickedGroups = this.pickedGroups.filter((id) => this.groupNames[id] !== undefined);
      } finally {
        this.applying = false;
      }
    },

    async onSaved() {
      EventBus.$emit('i-inventory-counts');
      await this.load();
    },

    async deleteHost() {
      try {
        await axios.delete(`${this.inventoryUrl}/hosts/${this.deleteHostId}`);
        EventBus.$emit('i-inventory-counts');
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
      await this.load();
    },

    async setEnabled(host, enabled) {
      this.busyIds.push(host.id);
      try {
        await axios.put(`${this.inventoryUrl}/hosts/${host.id}/enabled`, { enabled });
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
      await this.load();
      this.busyIds = this.busyIds.filter((id) => id !== host.id);
    },

    async load() {
      // A group in the address that is not an id is dropped like a deleted one
      if (this.$route.query.group && this.groupFilter === 'any') {
        this.$router.replace({ query: { ...this.$route.query, group: undefined } });
      }
      this.loadSeq += 1;
      const seq = this.loadSeq;
      this.loading = true;
      try {
        const response = await axios.get(`${this.inventoryUrl}/hosts`, {
          params: {
            search: this.search || undefined,
            group: typeof this.groupFilter === 'number' ? this.groupFilter : undefined,
            ungrouped: this.groupFilter === 'ungrouped' ? 1 : undefined,
            offset: this.offset,
            count: this.count,
          },
        });
        if (seq !== this.loadSeq) return;
        if (response.data.length === 0 && this.offset > 0) {
          this.offset = Math.max(0, this.offset - this.count);
          await this.load();
          return;
        }
        // A server of the previous version answers without groups
        this.rows = response.data.map((host) => ({ ...host, groups: host.groups || [] }));
        this.selected = this.rows.map((host) => host.id).filter((id) => this.selected.includes(id));
        this.total = Number(response.headers['x-total-count']) || 0;
      } catch (err) {
        if (seq !== this.loadSeq) return;
        // The group of the filter was deleted, the watcher of the filter loads again
        if (err.response && err.response.status === 404 && typeof this.groupFilter === 'number') {
          this.$router.replace({ query: { ...this.$route.query, group: undefined } });
          this.loadGroups();
          return;
        }
        this.rows = [];
        this.selected = [];
        this.total = 0;
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      } finally {
        if (seq === this.loadSeq) this.loading = false;
      }
    },

    async loadGroups() {
      this.groupsSeq += 1;
      const seq = this.groupsSeq;
      try {
        const response = await axios.get(`${this.inventoryUrl}/groups`);
        if (seq !== this.groupsSeq) return;
        this.groups = response.data;
      } catch (err) {
        if (seq !== this.groupsSeq) return;
        this.groups = [];
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
    },
  },
};
</script>
