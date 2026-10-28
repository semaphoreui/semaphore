<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div
    v-if="items != null"
    data-testid="keystore-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="keystoreDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="`${itemId === 'new' ? $t('nnew') : $t('edit')} Key`"
      :max-width="450"
      @save="loadItemsAndShowPublicKey($event)"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <KeyForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :support-storages="features.secret_storages"
        />
      </template>
    </EditDialog>

    <EditDialog
      :max-width="700"
      v-model="createdPublicKeyDialog"
      test-id="keystorePublicKeyDialog"
      :save-button-text="null"
      title="Generated SSH Public Key"
      hide-buttons
    >
      <template v-slot:form="{}">
        <div class="mb-4">
          <div style="position: relative">
            <pre
              style="
                overflow: auto;
                background: gray;
                color: white;
                border-radius: 10px;
                margin-top: 5px;
                white-space: normal;
              "
              class="pa-2"
              data-testid="keystorePublicKey-value"
              >{{ createdPublicKey }}</pre
            >

            <CopyClipboardButton
              style="position: absolute; right: 0; top: 0"
              :text="createdPublicKey"
              test-id="keystorePublicKey-copy"
            />
          </div>
        </div>
      </template>
    </EditDialog>

    <ObjectRefsDialog
      test-id="keystoreRefsDialog"
      object-title="access key"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      test-id="keystoreDeleteDialog"
      :title="$t('deleteKey')"
      :text="$t('askDeleteKey')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="keystore-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title data-testid="keystore-title">{{ $t('keyStore') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        data-testid="keystore-new"
      >{{ $t('newKey') }}</v-btn>
    </v-toolbar>

    <KeyStoreMenu v-if="isPro" :project-id="projectId" />
    <v-divider v-else />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      data-testid="keystore-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto;"
    >
      <template v-slot:item.name="{ item }">
        {{ item.name }}
        <v-chip
          color="error"
          v-if="item.empty && item.type !== 'none'"
          small
          style="font-weight: bold;"
          class="ml-2"
          data-testid="keystore-rowEmpty"
        >{{ $t('empty') }}</v-chip>
        <v-chip
          v-if="item.synchronized"
          x-small
          class="ml-2"
          data-testid="keystore-rowSynchronized"
        >{{ $t('synchronized') }}</v-chip>
      </template>
      <template v-slot:item.type="{ item }">
        <code data-testid="keystore-rowType">{{ item.type }}</code>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn-toggle
          dense
          :value-comparator="() => false"
          data-testid="keystore-rowActions"
        >
          <v-btn
            @click="askDeleteItem(item.id)"
            data-testid="keystore-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn
            @click="editItem(item.id)"
            data-testid="keystore-rowEdit"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>

  </div>

</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import KeyForm from '@/components/KeyForm.vue';
import PageMixin from '@/components/PageMixin';
import KeyStoreMenu from '@/components/KeyStoreMenu.vue';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';

export default {
  components: {
    CopyClipboardButton,
    KeyStoreMenu,
    KeyForm,
  },

  mixins: [ItemListPageBase, PageMixin],

  props: {
    systemInfo: Object,
  },

  computed: {
    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },
  },

  data() {
    return {
      createdPublicKeyDialog: false,
      createdPublicKey: '',
    };
  },

  methods: {
    async loadItemsAndShowPublicKey(e) {
      await this.loadItems();

      const generatedCreatePublicKey = e && e.action === 'new'
        ? this.extractPublicKey((e || {}).item)
        : '';
      const isGeneratedOnCreate = generatedCreatePublicKey !== '';
      const isGeneratedOnUpdate = e && e.action === 'edit' && e.item && e.item.generate_ssh_key;
      if (!isGeneratedOnCreate && !isGeneratedOnUpdate) {
        this.createdPublicKey = '';
        return;
      }

      const itemId = e && e.item ? e.item.id : null;
      const reloadedItem = itemId ? this.items.find((x) => x.id === itemId) : null;
      const publicKey = generatedCreatePublicKey
        || this.extractPublicKey(reloadedItem)
        || this.extractPublicKey((e || {}).item);

      if (!publicKey) {
        this.createdPublicKey = '';
        return;
      }

      this.createdPublicKey = publicKey;
      this.createdPublicKeyDialog = true;
    },

    extractPublicKey(item) {
      if (!item || !item.plain) {
        return '';
      }

      try {
        const plain = JSON.parse(item.plain);
        return plain.public_key || '';
      } catch (e) {
        return '';
      }
    },

    getHeaders() {
      return [{
        text: this.$i18n.t('name'),
        value: 'name',
        width: '60%',
      },
      {
        text: this.$i18n.t('type'),
        value: 'type',
        width: '40%',
      },
      {
        value: 'actions',
        sortable: false,
        width: '0%',
      },
      ];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/keys`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/keys/${this.itemId}`;
    },
    getEventName() {
      return 'i-keys';
    },
  },
};
</script>
