<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
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
              "
              class="pa-2"
              >{{ createdPublicKey }}</pre
            >

            <CopyClipboardButton
              style="position: absolute; right: 10px; top: 10px"
              :text="createdPublicKey"
            />
          </div>
        </div>
      </template>
    </EditDialog>

    <ObjectRefsDialog
      object-title="access key"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      :title="$t('deleteKey')"
      :text="$t('askDeleteKey')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('keyStore') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
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
        >{{ $t('empty') }}</v-chip>
        <v-chip
          v-if="item.synchronized"
          x-small
          class="ml-2"
        >{{ $t('synchronized') }}</v-chip>
      </template>
      <template v-slot:item.type="{ item }">
        <code>{{ item.type }}</code>
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

      const isGeneratedOnCreate = e && e.action === 'new';
      const isGeneratedOnUpdate = e && e.action === 'edit' && e.item && e.item.generate_ssh_key;
      if (!isGeneratedOnCreate && !isGeneratedOnUpdate) {
        this.createdPublicKey = '';
        return;
      }

      const itemId = e && e.item ? e.item.id : null;
      const reloadedItem = itemId ? this.items.find((x) => x.id === itemId) : null;
      const sourceItem = reloadedItem || (e || {}).item;
      const publicKey = this.extractPublicKey(sourceItem);

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
