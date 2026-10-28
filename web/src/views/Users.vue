<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div
    v-if="items != null"
    data-testid="users-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="userDialog"
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
      test-id="usersDeleteDialog"
      :title="$t('deleteUser')"
      :text="$t('askDeleteUser')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-btn icon class="mr-4" @click="returnToProjects()" data-testid="users-back">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title data-testid="users-title">{{ $t('users') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        v-if="tab === 'users' && canManageUsers"
        color="primary"
        @click="editItem('new')"
        data-testid="users-new"
      >
        {{ $t('newUser') }}
      </v-btn>

      <v-btn
        v-if="tab === 'global-roles' && canManageGlobalRoles"
        :disabled="!(features || {}).global_roles_management"
        color="primary"
        @click="$refs.globalRoles.editItem('new')"
        data-testid="globalRoles-new"
      >
        {{ $t('newGlobalRole') }}
      </v-btn>
    </v-toolbar>

    <v-tabs v-if="canManageGlobalRoles" v-model="tab" class="px-4" data-testid="users-tabs">
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
      :items-per-page-options="[20]"
      data-testid="users-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.external="{ item }">
        <v-icon
          v-if="item.external"
          data-testid="users-rowExternal-true"
        >mdi-checkbox-marked</v-icon>
        <v-icon v-else data-testid="users-rowExternal-false">mdi-checkbox-blank-outline</v-icon>
      </template>

      <template v-slot:item.alert="{ item }">
        <v-icon v-if="item.alert" data-testid="users-rowAlert-true">mdi-checkbox-marked</v-icon>
        <v-icon v-else data-testid="users-rowAlert-false">mdi-checkbox-blank-outline</v-icon>
      </template>

      <template v-slot:item.admin="{ item }">
        <v-icon v-if="item.admin" data-testid="users-rowAdmin-true">mdi-checkbox-marked</v-icon>
        <v-icon v-else data-testid="users-rowAdmin-false">mdi-checkbox-blank-outline</v-icon>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap" data-testid="users-rowActions">
          <v-btn icon class="mr-1" @click="askDeleteItem(item.id)" data-testid="users-rowDelete">
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn icon class="mr-1" @click="editItem(item.id)" data-testid="users-rowEdit">
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

const USERS_PATH = '/users';
const GLOBAL_ROLES_PATH = '/users/roles';

export default {
  mixins: [ItemListPageBase],

  props: {
    authMethods: Object,
    systemInfo: Object,
    features: Object,
  },

  computed: {
    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },

    canManageUsers() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageUsers);
    },

    // Global roles exist in the Pro edition only: the community build has no tab for them.
    canManageGlobalRoles() {
      return this.isPro && hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageRoles);
    },

    // Assigning global roles needs the Pro feature, like managing the roles themselves.
    canAssignGlobalRoles() {
      return this.canManageGlobalRoles && !!(this.features || {}).global_roles_management;
    },

    // The tab is the address: /users/roles shows the global roles, /users the users.
    // Someone who may not manage global roles sees the users whatever the address says.
    tab: {
      get() {
        if (this.canManageGlobalRoles && this.$route.path === GLOBAL_ROLES_PATH) {
          return 'global-roles';
        }
        return 'users';
      },
      set(value) {
        const path = value === 'global-roles' ? GLOBAL_ROLES_PATH : USERS_PATH;
        if (this.$route.path !== path) {
          this.$router.push(path);
        }
      },
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
      EventBus.$emit('i-snackbar', { color: 'success', text: this.$t('userCreated') });
      await this.loadItems();
    },

    getHeaders() {
      return [
        {
          title: this.$t('name'),
          key: 'name',
          width: '50%',
        },
        {
          title: this.$t('username'),
          key: 'username',
        },
        {
          title: this.$t('email'),
          key: 'email',
        },
        {
          title: this.$t('alert'),
          key: 'alert',
        },
        {
          title: this.$t('admin'),
          key: 'admin',
        },
        {
          title: this.$t('external'),
          key: 'external',
          width: '50%',
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
