<template>
  <div
    v-if="items != null"
    data-testid="roleMappings-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="roleMappingDialog"
      save-button-text="Save"
      :title="$t('newRoleMapping')"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <RoleMappingForm
          item-id="new"
          :providers="providers"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      :title="$t('deleteRoleMapping')"
      :text="$t('askDeleteRoleMapping')"
      v-model="deleteItemDialog"
      test-id="roleMappingsDeleteDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-btn
        icon
        variant="text"
        class="mr-4"
        @click="returnToProjects()"
        data-testid="roleMappings-back"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="roleMappings-title"
      >{{ $t('roleMappings') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="!features.custom_roles_management"
        color="primary"
        variant="elevated"
        @click="editItem('new')"
        data-testid="roleMappings-new"
        >{{ $t('newRoleMapping') }}</v-btn
      >
    </v-toolbar>

    <v-divider />

    <v-alert
      v-if="!features.custom_roles_management"
      variant="tonal"
      color="amber-darken-3"
      class="PageAlert"
      data-testid="roleMappings-licenseAlert"
    >
      <span
        class="mr-1"
        v-html="$t('roles_only_enterprise')"
        data-testid="roleMappings-licenseAlertText"
      ></span>

      <v-btn
        variant="flat"
        color="amber-darken-3"
        href="https://semaphoreui.com/enterprise?utm_source=app&utm_content=feature_roles"
        target="_blank"
        data-testid="roleMappings-upgrade"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>
    </v-alert>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4 FixedPageSizeTable"
      :items-per-page-options="[20]"
      :items-per-page="20"
      data-testid="roleMappings-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.provider="{ item }">
        {{ (providers.find((p) => p.id === item.provider) || { name: item.provider }).name }}
        <v-chip
          v-if="!item.provider_configured"
          size="small"
          label
          variant="flat"
          color="amber-darken-3"
          class="ml-1"
          data-testid="roleMappings-rowProviderNotConfigured"
        >{{ $t('providerNotConfigured') }}</v-chip>
      </template>
      <template v-slot:item.role="{ item }">
        <div data-testid="roleMappings-rowRoleName">{{ item.role_name }}</div>
        <div
          class="text-medium-emphasis text-body-small"
          data-testid="roleMappings-rowRoleSlug"
        >{{ item.role }}</div>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn
          icon
          variant="text"
          @click="askDeleteItem(item.id)"
          data-testid="roleMappings-rowDelete"
        >
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import ItemListPageBase from '@/components/ItemListPageBase';
import RoleMappingForm from '@/components/RoleMappingForm.vue';

export default {
  mixins: [ItemListPageBase],

  props: {
    features: Object,
  },

  components: {
    RoleMappingForm,
  },

  data() {
    return {
      providers: [],
    };
  },

  methods: {
    async beforeLoadItems() {
      this.providers = (await axios({
        method: 'get',
        url: '/api/auth/login',
        responseType: 'json',
      })).data.oidc_providers || [];
    },

    getHeaders() {
      return [
        {
          title: this.$t('roleMappingProvider'),
          key: 'provider',
        },
        {
          title: this.$t('roleMappingGroup'),
          key: 'group',
        },
        {
          title: this.$t('project'),
          key: 'project_name',
        },
        {
          title: this.$t('role'),
          key: 'role',
        },
        {
          title: this.$t('actions'),
          key: 'actions',
          sortable: false,
        },
      ];
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    getItemsUrl() {
      return '/api/role_mappings';
    },

    getSingleItemUrl() {
      return `/api/role_mappings/${this.itemId}`;
    },

    getEventName() {
      return 'i-role-mapping';
    },
  },
};
</script>
