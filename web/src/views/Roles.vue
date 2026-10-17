<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      save-button-text="Save"
      :title="editDialogTitle"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <GlobalRoleForm
          v-if="isGlobalTab"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :is-admin="true"
        />
        <RoleForm
          v-else
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :is-admin="true"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      :title="$t(isGlobalTab ? 'deleteGlobalRole' : 'deleteRole')"
      :text="$t(isGlobalTab ? 'askDeleteGlobalRole' : 'askDeleteRole')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <YesNoDialog
      :title="$t(isGlobalTab ? 'deleteGlobalRole' : 'deleteRole')"
      :text="roleImpact ? $t(forceDeleteQuestionKey, roleImpact) : ''"
      :yes-button-title="$t('deleteWithAssignments')"
      :max-width="520"
      v-model="forceDeleteDialog"
      @yes="deleteItem(itemId, true)"
    />

    <v-toolbar flat>
      <v-btn icon class="mr-4" @click="returnToProjects()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ $t('Roles') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        v-if="canManageRoles"
        :disabled="!features.custom_roles_management"
        color="primary"
        @click="editItem('new')"
        >{{ $t(isGlobalTab ? 'newGlobalRole' : 'newRole') }}</v-btn
      >
    </v-toolbar>

    <TeamMenu v-if="projectId" :project-id="projectId" :system-info="systemInfo" />

    <v-tabs v-else v-model="tab" class="px-4">
      <v-tab key="project" href="#project" data-testid="roles-tab-project">
        {{ $t('projectLevelRoles') }}
      </v-tab>
      <v-tab key="global" href="#global" data-testid="roles-tab-global">
        {{ $t('globalRoles') }}
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px" />

    <v-alert v-if="!features.custom_roles_management" text color="amber darken-3" class="PageAlert">
      <span class="mr-1" v-html="$t('roles_only_enterprise')"></span>

      <v-btn
        dark
        depressed
        v-if="isAdmin"
        color="amber darken-3"
        href="https://semaphoreui.com/enterprise?utm_source=app&utm_content=feature_roles"
        target="_blank"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>

      <span v-else style="font-weight: bold">
        {{ $t('contact_admin_to_upgrade_enterprise') }}
      </span>
    </v-alert>

    <v-alert v-if="isGlobalTab" text dense type="info" class="mx-4 mt-4">
      {{ $t('globalRolesHint') }}
    </v-alert>

    <v-data-table
      :headers="tableHeaders"
      :items="items"
      :item-key="IDFieldName"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.name="{ item }">
        <div>{{ item.name }}</div>
        <div class="text--secondary caption">{{ item.slug }}</div>
      </template>
      <template v-slot:item.scope="{ item }">
        <v-chip small label>{{ $t(roleScope(item)) }}</v-chip>
      </template>
      <template v-slot:item.permissions="{ item }">
        <TemplatePermissionsChips
          class="py-1"
          :scope="isGlobalTab ? 'global' : 'default'"
          :permissions="item.permissions || 0"
        />
      </template>
      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn icon class="mr-1" @click="askDeleteItem(item[IDFieldName])">
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn icon class="mr-1" @click="editItem(item[IDFieldName])">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ItemListPageBase from '@/components/ItemListPageBase';
import EditDialog from '@/components/EditDialog.vue';
import RoleForm from '@/components/EditRoleForm.vue';
import GlobalRoleForm from '@/components/EditGlobalRoleForm.vue';
import TeamMenu from '@/components/TeamMenu.vue';
import TemplatePermissionsChips from '@/components/TemplatePermissionsChips.vue';
import { roleScope } from '@/lib/roles';
import { GLOBAL_PERMISSIONS, USER_PERMISSIONS } from '@/lib/constants';
import { hasGlobalPermission } from '@/lib/role-permissions';

/**
 * The roles page. Inside a project it lists the project's own roles. On the instance it has
 * two tabs: project-level roles shared by every project, and global roles that grant
 * instance-wide permissions.
 */
export default {
  mixins: [ItemListPageBase],

  props: {
    features: Object,
    projectId: Number,
    systemInfo: Object,
  },

  components: {
    TeamMenu,
    YesNoDialog,
    RoleForm,
    GlobalRoleForm,
    EditDialog,
    TemplatePermissionsChips,
  },

  data() {
    return {
      tab: 'project',
      roleImpact: null,
      forceDeleteDialog: false,
    };
  },

  computed: {
    isGlobalTab() {
      return !this.projectId && this.tab === 'global';
    },

    /** Project roles are addressed by slug, global roles by id. */
    IDFieldName() {
      return this.isGlobalTab ? 'id' : 'slug';
    },

    forceDeleteQuestionKey() {
      return this.isGlobalTab ? 'askDeleteAssignedGlobalRole' : 'askDeleteAssignedRole';
    },

    editDialogTitle() {
      if (this.isGlobalTab) {
        return this.itemId === 'new' ? this.$t('newGlobalRole') : this.$t('editGlobalRole');
      }
      return this.itemId === 'new' ? this.$t('newRole') : this.$t('editRole');
    },

    tableHeaders() {
      const headers = [
        {
          text: this.$i18n.t('name'),
          value: 'name',
          width: '50%',
        },
      ];
      if (!this.isGlobalTab) {
        headers.push({
          text: this.$i18n.t('roleScope'),
          value: 'scope',
          sortable: false,
        });
      }
      headers.push({
        text: this.$i18n.t(this.isGlobalTab ? 'globalPermissions' : 'permissions'),
        value: 'permissions',
      });
      if (this.allowActions()) {
        headers.push({
          text: this.$i18n.t('actions'),
          value: 'actions',
          sortable: false,
        });
      }
      return headers;
    },

    /** Project roles follow the project permission, instance roles the global one. */
    canManageRoles() {
      return this.allowActions();
    },
  },

  watch: {
    async projectId() {
      await this.loadItems();
    },

    async tab() {
      this.items = null;
      await this.loadItems();
    },
  },

  methods: {
    // Called from data() of the list mixin, before computed properties exist,
    // so the props are read directly.
    allowActions() {
      if (this.projectId) {
        return this.can(USER_PERMISSIONS.manageProjectResources);
      }
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageRoles);
    },

    getHeaders() {
      return [];
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    roleScope,

    async deleteItem(itemId, force = false) {
      this.itemId = itemId;
      try {
        const item = this.items.find((x) => x[this.IDFieldName] === itemId);
        await axios({
          method: 'delete',
          url: force ? `${this.getSingleItemUrl()}?force=true` : this.getSingleItemUrl(),
          responseType: 'json',
        });
        EventBus.$emit(this.getEventName(), { action: 'delete', item });
        await this.loadItems();
      } catch (err) {
        const impact = err.response && err.response.status === 409
          && err.response.data && err.response.data.impact;
        if (impact && this.projectId && !this.can(USER_PERMISSIONS.manageProjectUsers)) {
          EventBus.$emit('i-snackbar', { color: 'error', text: this.$i18n.t('roleInUseNeedsManageUsers') });
          return;
        }
        if (impact) {
          this.roleImpact = impact;
          this.forceDeleteDialog = true;
          return;
        }
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
    },

    getItemsUrl() {
      if (this.projectId) {
        return `/api/project/${this.projectId}/roles`;
      }
      return this.isGlobalTab ? '/api/global-roles' : '/api/roles';
    },

    getSingleItemUrl() {
      if (this.projectId) {
        return `/api/project/${this.projectId}/roles/${this.itemId}`;
      }
      return this.isGlobalTab ? `/api/global-roles/${this.itemId}` : `/api/roles/${this.itemId}`;
    },

    getEventName() {
      return this.isGlobalTab ? 'i-global-role' : 'i-role';
    },
  },
};
</script>
