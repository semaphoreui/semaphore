<template>
  <div data-testid="inventoryHosts-page" :data-state="state">
    <EditDialog
      v-model="editDialog"
      test-id="inventoryHostDialog"
      help-doc="inventory-host"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('inventory_hosts_add') : $t('inventory_hosts_edit')"
      :max-width="450"
      @save="load()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <InventoryHostForm
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
      test-id="inventoryHostDeleteDialog"
      :title="$t('inventory_hosts_delete')"
      :text="$t('inventory_hosts_delete_confirm', { name: deleteName })"
      @yes="deleteHost()"
    />

    <div class="d-flex flex-wrap align-center px-4 pt-4" style="gap: 16px">
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
      <v-spacer></v-spacer>
      <v-btn
        v-if="canManage"
        color="primary"
        variant="elevated"
        data-testid="inventoryHosts-add"
        @click="editHost('new')"
      >{{ $t('inventory_hosts_add') }}</v-btn>
    </div>

    <v-data-table
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

      <template v-slot:item.name="{ item }">
        <span data-testid="inventoryHosts-name">{{ item.name }}</span>
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

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import { USER_PERMISSIONS } from '@/lib/constants';
import PermissionsCheck from '@/components/PermissionsCheck';
import ProjectMixin from '@/components/ProjectMixin';
import EditDialog from '@/components/EditDialog.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import InventoryHostForm from '@/components/InventoryHostForm.vue';

export default {
  mixins: [PermissionsCheck, ProjectMixin],

  components: { EditDialog, YesNoDialog, InventoryHostForm },

  props: {
    inventoryId: Number,
  },

  data() {
    return {
      rows: null,
      total: 0,
      loading: false,
      offset: 0,
      count: 50,
      searchText: '',
      searchTimer: null,
      loadSeq: 0,
      itemId: null,
      editDialog: false,
      deleteDialog: false,
      deleteHostId: null,
      deleteName: '',
      busyIds: [],
    };
  },

  computed: {
    search() {
      return this.$route.query.search || '';
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
        { title: this.$t('description'), key: 'description', sortable: false },
        { title: this.$t('enabled'), key: 'enabled', sortable: false },
        ...(this.canManage ? [{
          title: '', key: 'actions', sortable: false, align: 'end',
        }] : []),
      ];
    },
  },

  watch: {
    search() {
      this.searchText = this.search;
      this.offset = 0;
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
    await this.load();
  },

  beforeUnmount() {
    clearTimeout(this.searchTimer);
  },

  methods: {
    hostsUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/hosts`;
    },

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
      this.load();
    },

    async deleteHost() {
      try {
        await axios.delete(`${this.hostsUrl()}/${this.deleteHostId}`);
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
      await this.load();
    },

    async setEnabled(host, enabled) {
      this.busyIds.push(host.id);
      try {
        await axios.put(`${this.hostsUrl()}/${host.id}/enabled`, { enabled });
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
      await this.load();
      this.busyIds = this.busyIds.filter((id) => id !== host.id);
    },

    async load() {
      this.loadSeq += 1;
      const seq = this.loadSeq;
      this.loading = true;
      try {
        const response = await axios.get(this.hostsUrl(), {
          params: {
            search: this.search || undefined,
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
        this.rows = response.data;
        this.total = Number(response.headers['x-total-count']) || 0;
      } catch (err) {
        if (seq !== this.loadSeq) return;
        this.rows = [];
        this.total = 0;
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      } finally {
        if (seq === this.loadSeq) this.loading = false;
      }
    },
  },
};
</script>
