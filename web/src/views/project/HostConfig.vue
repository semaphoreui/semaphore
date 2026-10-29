<template>
  <div
    v-if="items != null && keys != null"
    data-testid="hostConfig-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="hostConfigDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('addMapping') : $t('editMapping')"
      @save="loadItems()"
      :max-width="450"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <HostConfigForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      test-id="hostConfigDeleteDialog"
      :title="$t('deleteMapping')"
      :text="$t('askDeleteMapping')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="hostConfig-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="hostConfig-title"
      >{{ $t('hostConfig') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        variant="elevated"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        data-testid="hostConfig-new"
      >{{ $t('addMapping') }}</v-btn>
    </v-toolbar>

    <v-divider />

    <v-alert
      v-if="items.length === 0"
      type="info"
      variant="tonal"
      class="mt-4"
      style="max-width: 800px; margin: auto;"
      data-testid="hostConfig-empty"
    >{{ $t('hostConfigEmpty') }}</v-alert>

    <v-data-table
      v-else
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="-1"
      data-testid="hostConfig-table"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 100px); margin: auto;"
    >
      <template v-slot:item.type="{ item }">
        <v-chip
          size="small"
          :data-testid="`hostConfig-rowType-${item.type === 'url' ? 'url' : 'host'}`"
        >{{
          item.type === 'url' ? $t('hostConfigTypeUrl') : $t('hostConfigTypeHost')
        }}</v-chip>
      </template>

      <template v-slot:item.ssh_key_id="{ item }">
        {{ (keys.find((k) => k.id === item.ssh_key_id) || {}).name }}
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-group
          density="compact"
          data-testid="hostConfig-rowActions"
        >
          <v-btn
            @click="askDeleteItem(item.id)"
            data-testid="hostConfig-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn
            @click="editItem(item.id)"
            data-testid="hostConfig-rowEdit"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-group>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import axios from 'axios';
import ItemListPageBase from '@/components/ItemListPageBase';
import HostConfigForm from '@/components/HostConfigForm.vue';

export default {
  mixins: [ItemListPageBase],
  components: { HostConfigForm },

  data() {
    return {
      keys: null,
    };
  },

  async created() {
    this.keys = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },

  methods: {
    getHeaders() {
      return [{
        title: this.$t('hostConfigHostOrUrl'),
        key: 'host',
        width: '50%',
      },
      {
        title: this.$t('type'),
        key: 'type',
        width: '15%',
      },
      {
        title: this.$t('credential'),
        key: 'ssh_key_id',
        width: '25%',
      },
      {
        title: this.$t('actions'),
        key: 'actions',
        sortable: false,
      }];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/host_configs`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/host_configs/${this.itemId}`;
    },

    getEventName() {
      return 'i-host-configs';
    },
  },
};
</script>
