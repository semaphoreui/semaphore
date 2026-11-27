<template>
  <div
    v-if="items != null && keys != null"
    data-testid="proxies-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="proxiesDialog"
      help-doc="proxy"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="`${itemId === 'new' ? $t('nnew') : $t('edit')} Proxy`"
      @save="loadItems()"
      :max-width="450"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <ProxyForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      object-title="proxy"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
      test-id="proxiesRefsDialog"
    />

    <YesNoDialog
      test-id="proxiesDeleteDialog"
      :title="$t('deleteProxy')"
      :text="$t('askDeleteProxy')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="proxies-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title data-testid="proxies-title">{{ $t('proxies') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        variant="elevated"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        data-testid="proxies-new"
      >{{ $t('newProxy') }}</v-btn>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4 ProxiesTable"
      :items-per-page="-1"
      data-testid="proxies-table"
    >
      <template v-slot:item.host="{ item }">
        {{ item.user ? `${item.user}@${item.host}` : item.host }}<span
          v-if="item.port"
        >:<code>{{ item.port }}</code></span>
      </template>

      <template v-slot:item.ssh_key_id="{ item }">
        {{ keyName(item.ssh_key_id) }}
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-group density="compact" data-testid="proxies-rowActions">
          <v-btn @click="askDeleteItem(item.id)" data-testid="proxies-rowDelete">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn @click="editItem(item.id)" data-testid="proxies-rowEdit">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-group>
      </template>
    </v-data-table>
  </div>
</template>
<style lang="scss">
// Not an inline style: the utility `mt-4` has no `!important` and would lose to it.
.ProxiesTable {
  max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 100px);
  margin: auto;
}
</style>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import ProxyForm from '@/components/ProxyForm.vue';
import axios from 'axios';

export default {
  mixins: [ItemListPageBase],
  components: { ProxyForm },

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
    keyName(keyID) {
      return this.keys.find((k) => k.id === keyID)?.name || '—';
    },

    getHeaders() {
      return [{
        title: this.$t('name'),
        key: 'name',
        width: '25%',
      },
      {
        title: this.$t('type'),
        key: 'type',
        width: '15%',
      },
      {
        title: this.$t('proxyHost'),
        key: 'host',
        width: '35%',
      },
      {
        title: this.$t('accessKey'),
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
      return `/api/project/${this.projectId}/proxies`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/proxies/${this.itemId}`;
    },
    getEventName() {
      return 'i-proxies';
    },
  },
};
</script>
