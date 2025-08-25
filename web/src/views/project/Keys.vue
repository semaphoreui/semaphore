<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="`${itemId === 'new' ? $t('nnew') : $t('edit')} Key`"
      :max-width="450"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <KeyForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :support-storages="premiumFeatures.secret_storages"
        />
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

    <v-tabs class="pl-4">
      <v-tab
        key="keys"
        :to="`/project/${projectId}/keys`"
        data-testid="keystore-keys"
      >
        Keys
      </v-tab>

      <v-tab
        key="storages"
        :to="`/project/${projectId}/secret_storages`"
        data-testid="keystore-storages"
      >
        Storages
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;" />

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

export default {
  components: { KeyForm },

  mixins: [ItemListPageBase, PageMixin],

  props: {
    systemInfo: Object,
  },

  methods: {
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
