<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('newUser') : $t('editUser')"
      :dont-close-on-save="itemId === 'new'"
      @save="loadItems()"
      :hide-buttons="hideEditDialogButtons"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <UserForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave($event); onUserSaved($event)"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :is-admin="isAdmin"
          :can-manage-global-roles="canAssignGlobalRoles"
          @hide-action-buttons="hideEditDialogButtons = true"
          @show-action-buttons="hideEditDialogButtons = false"
          :auth-methods="authMethods"
          :login-with-password="(systemInfo || {}).login_with_password"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      :title="$t('deleteUser')"
      :text="$t('askDeleteUser')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-btn icon class="mr-4" @click="returnToProjects()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ $t('users') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        v-if="tab === 'users' && canManageUsers"
        color="primary"
        @click="editItem('new')"
      >
        {{ $t('newUser') }}
      </v-btn>
      <v-btn
        v-if="tab === 'global-roles' && canManageGlobalRoles"
        :disabled="!(features || {}).custom_roles_management"
        color="primary"
        @click="$refs.globalRoles.editItem('new')"
      >
        {{ $t('newGlobalRole') }}
      </v-btn>
    </v-toolbar>

    <v-tabs v-if="canManageGlobalRoles" v-model="tab" class="px-4">
      <v-tab key="users" href="#users" data-testid="users-tab-users">{{ $t('users') }}</v-tab>
      <v-tab key="global-roles" href="#global-roles" data-testid="users-tab-global-roles">
        {{ $t('globalRoles') }}
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px" />

    <GlobalRolesList
      v-if="tab === 'global-roles'"
      ref="globalRoles"
      :user="user"
      :is-admin="isAdmin"
      :features="features"
    />

    <v-data-table
      v-else
      :headers="headers"
      :items="items"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.external="{ item }">
        <v-icon v-if="item.external">mdi-checkbox-marked</v-icon>
        <v-icon v-else>mdi-checkbox-blank-outline</v-icon>
      </template>

      <template v-slot:item.alert="{ item }">
        <v-icon v-if="item.alert">mdi-checkbox-marked</v-icon>
        <v-icon v-else>mdi-checkbox-blank-outline</v-icon>
      </template>

      <template v-slot:item.admin="{ item }">
        <v-icon v-if="item.admin">mdi-checkbox-marked</v-icon>
        <v-icon v-else>mdi-checkbox-blank-outline</v-icon>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn icon class="mr-1" @click="askDeleteItem(item.id)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn icon class="mr-1" @click="editItem(item.id)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import EventBus from '@/event-bus';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ItemListPageBase from '@/components/ItemListPageBase';
import EditDialog from '@/components/EditDialog.vue';
import UserForm from '@/components/UserForm.vue';
import GlobalRolesList from '@/components/GlobalRolesList.vue';
import { GLOBAL_PERMISSIONS } from '@/lib/constants';
import { hasGlobalPermission } from '@/lib/role-permissions';

export default {
  mixins: [ItemListPageBase],

  props: {
    authMethods: Object,
    systemInfo: Object,
    features: Object,
  },

  computed: {
    canManageUsers() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageUsers);
    },
    canManageGlobalRoles() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageRoles);
    },

    // Assigning global roles is a Pro feature, like managing the roles themselves.
    canAssignGlobalRoles() {
      return this.canManageGlobalRoles && !!(this.features || {}).custom_roles_management;
    },
  },

  components: {
    YesNoDialog,
    UserForm,
    EditDialog,
    GlobalRolesList,
  },

  data() {
    return {
      tab: 'users',
      hideEditDialogButtons: false,
    };
  },

  methods: {
    // Called from data() of the list mixin, before computed properties exist,
    // so the props are read directly.
    allowActions() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageUsers);
    },

    // A created user stays open for editing, so that roles can be assigned right away.
    async onUserSaved(e) {
      if (e.action !== 'new' || !e.item || !e.item.id) {
        return;
      }
      this.itemId = e.item.id;
      EventBus.$emit('i-snackbar', { color: 'success', text: this.$i18n.t('userCreated') });
      await this.loadItems();
    },

    getHeaders() {
      return [
        {
          text: this.$i18n.t('name'),
          value: 'name',
          width: '50%',
        },
        {
          text: this.$i18n.t('username'),
          value: 'username',
        },
        {
          text: this.$i18n.t('email'),
          value: 'email',
        },
        {
          text: this.$i18n.t('alert'),
          value: 'alert',
        },
        {
          text: this.$i18n.t('admin'),
          value: 'admin',
        },
        {
          text: this.$i18n.t('external'),
          value: 'external',
          width: '50%',
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
      return '/api/users';
    },

    getSingleItemUrl() {
      return `/api/users/${this.itemId}`;
    },

    getEventName() {
      return 'i-user';
    },
  },
};
</script>
