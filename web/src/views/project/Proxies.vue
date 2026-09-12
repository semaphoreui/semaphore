<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null && keys != null">
    <EditDialog
      v-model="editDialog"
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
    />

    <YesNoDialog
      :title="$t('deleteProxy')"
      :text="$t('askDeleteProxy')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('proxies') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >{{ $t('newProxy') }}</v-btn>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 100px); margin: auto;"
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
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn @click="askDeleteItem(item.id)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn @click="editItem(item.id)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>
  </div>
</template>
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
        text: this.$i18n.t('name'),
        value: 'name',
        width: '25%',
      },
      {
        text: this.$i18n.t('type'),
        value: 'type',
        width: '15%',
      },
      {
        text: this.$i18n.t('proxyHost'),
        value: 'host',
        width: '35%',
      },
      {
        text: this.$i18n.t('accessKey'),
        value: 'ssh_key_id',
        width: '25%',
      },
      {
        text: this.$i18n.t('actions'),
        value: 'actions',
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
