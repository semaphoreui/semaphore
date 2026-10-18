<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
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
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-btn icon class="mr-4" @click="returnToProjects()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ $t('roleMappings') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="!features.custom_roles_management"
        color="primary"
        @click="editItem('new')"
        >{{ $t('newRoleMapping') }}</v-btn
      >
    </v-toolbar>

    <v-divider />

    <v-alert v-if="!features.custom_roles_management" text color="amber darken-3" class="PageAlert">
      <span class="mr-1" v-html="$t('roles_only_enterprise')"></span>

      <v-btn
        dark
        depressed
        color="amber darken-3"
        href="https://semaphoreui.com/enterprise?utm_source=app&utm_content=feature_roles"
        target="_blank"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>
    </v-alert>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.provider="{ item }">
        {{ (providers.find((p) => p.id === item.provider) || { name: item.provider }).name }}
        <v-chip
          v-if="!item.provider_configured"
          small
          label
          dark
          color="amber darken-3"
          class="ml-1"
        >{{ $t('providerNotConfigured') }}</v-chip>
      </template>
      <template v-slot:item.role="{ item }">
        <div>{{ item.role_name }}</div>
        <div class="text--secondary caption">{{ item.role }}</div>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn icon @click="askDeleteItem(item.id)">
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
          text: this.$i18n.t('roleMappingProvider'),
          value: 'provider',
        },
        {
          text: this.$i18n.t('roleMappingGroup'),
          value: 'group',
        },
        {
          text: this.$i18n.t('project'),
          value: 'project_name',
        },
        {
          text: this.$i18n.t('role'),
          value: 'role',
        },
        {
          text: this.$i18n.t('actions'),
          value: 'actions',
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
