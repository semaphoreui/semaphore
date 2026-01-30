<template>
  <div v-if="items != null">
    <ObjectRefsDialog
        object-title="storage"
        :object-refs="itemRefs"
        :project-id="projectId"
        v-model="itemRefsDialog"
    />

    <YesNoDialog
        :title="$t('deleteStorage')"
        :text="$t('askDeleteStorage')"
        v-model="deleteItemDialog"
        @yes="deleteItem(itemId)"
    />

    <EditDialog
        v-model="editDialog"
        :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
        :title="`${itemId === 'new' ? $t('nnew') : $t('edit')} ${itemType} Storage`"
        :max-width="450"
        @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <SecretStorageForm
            :project-id="projectId"
            :item-id="itemId"
            :item-type="itemType"
            @save="onSave"
            @error="onError"
            :need-save="needSave"
            :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('keyStore') }}</v-toolbar-title>
      <v-spacer></v-spacer>

      <v-menu offset-y>
        <template v-slot:activator="{ on, attrs }">
          <v-btn
              class="pr-2"
              v-bind="attrs"
              v-on="on"
              color="primary"
              v-if="can(USER_PERMISSIONS.manageProjectResources)"
          >
            New Storage
            <v-icon>mdi-chevron-down</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item
              link
              @click="
              editItem('new');
              itemType = 'vault';
            "
              :disabled="!premiumFeatures.secret_storage_management"
          >
            <v-list-item-icon>
              <v-icon>$vuetify.icons.hashicorp_vault</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Hashicorp Vault</v-list-item-title>
          </v-list-item>

          <v-list-item
              link
              @click="
              editItem('new');
              itemType = 'dvls';
            "
              :disabled="!premiumFeatures.secret_storage_management"
          >
            <v-list-item-icon>
              <v-icon>$vuetify.icons.dvls</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Devolutions Server</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-toolbar>

    <v-tabs class="pl-4">
      <v-tab key="keys" :to="`/project/${projectId}/keys`" data-testid="keystore-keys">
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

    <v-divider style="margin-top: -1px"/>

    <v-alert
        v-if="!premiumFeatures.secret_storage_management"
        text
        color="hsl(348deg, 86%, 61%)"
        class="PageAlert"
    >
      <span class="mr-1" v-html="$t('secret_storage_only_pro')"></span>

      <v-btn dark v-if="isAdmin" color="hsl(348deg, 86%, 61%)" @click="upgradeToPro()">
        {{ $t('upgrade_to_pro') }}
      </v-btn>

      <span v-else style="font-weight: bold">
        {{ $t('contact_admin_to_upgrade') }}
      </span>
    </v-alert>

    <v-data-table
        :headers="headers"
        :items="items"
        hide-default-footer
        class="mt-4"
        :items-per-page="Number.MAX_VALUE"
        style="
          max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px);
          margin: auto
        "
    >
      <template v-slot:item.name="{ item }">
        <v-icon class="mr-3" small>
          {{ getIcon(item) }}
        </v-icon>

        <span class="mr-2">{{ item.name }}</span>

        <v-chip v-if="item.readonly" style="transform: translateY(-1px)" color="info" small>
          Read only
        </v-chip>
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

<style scoped lang="scss"></style>

<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import SecretStorageForm from '@/components/SecretStorageForm.vue';
import EventBus from '@/event-bus';

export default {
  components: { SecretStorageForm },
  mixins: [ItemListPageBase],
  data() {
    return {
      itemType: 'vault',
    };
  },

  props: {
    systemInfo: Object,
  },

  computed: {
    premiumFeatures() {
      return this.systemInfo?.premium_features || {};
    },
  },

  methods: {
    getIcon(item) {
      switch (item.type) {
        case 'vault':
          return '$vuetify.icons.hashicorp_vault';
        case 'dvls':
          return '$vuetify.icons.dvls';
        default:
          return '';
      }
    },

    upgradeToPro() {
      EventBus.$emit('i-subscription', {});
    },

    getHeaders() {
      return [
        {
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
      return `/api/project/${this.projectId}/secret_storages`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/secret_storages/${this.itemId}`;
    },
    getEventName() {
      return 'i-secret-storage';
    },
  },
};
</script>
