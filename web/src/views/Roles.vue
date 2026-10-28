<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div
    v-if="items != null"
    data-testid="roles-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      test-id="roleDialog"
      v-model="editDialog"
      save-button-text="Save"
      :title="itemId === 'new' ? $t('newRole') : $t('editRole')"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <RoleForm
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
      test-id="rolesDeleteDialog"
      :title="$t('deleteRole')"
      :text="$t('askDeleteRole')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <YesNoDialog
      test-id="rolesForceDeleteDialog"
      :title="$t('deleteRole')"
      :text="roleImpact ? $t('askDeleteAssignedRole', roleImpact) : ''"
      :yes-button-title="$t('deleteWithAssignments')"
      :max-width="520"
      v-model="forceDeleteDialog"
      @yes="deleteItem(itemId, true)"
    />

    <v-toolbar flat>
      <v-btn icon class="mr-4" @click="returnToProjects()" data-testid="roles-back">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title data-testid="roles-title">{{ $t('Roles') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        v-if="canManageRoles"
        :disabled="!features.custom_roles_management"
        color="primary"
        @click="editItem('new')"
        data-testid="roles-new"
        >{{ $t('newRole') }}</v-btn
      >
    </v-toolbar>

    <TeamMenu v-if="projectId" :project-id="projectId" :system-info="systemInfo" />

    <v-divider style="margin-top: -1px" />

    <v-alert
      v-if="!features.custom_roles_management"
      text
      color="amber darken-3"
      class="PageAlert"
      data-testid="roles-licenseAlert"
    >
      <span
        class="mr-1"
        v-html="$t('roles_only_enterprise')"
        data-testid="roles-licenseAlertText"
      ></span>

      <v-btn
        dark
        depressed
        v-if="isAdmin"
        color="amber darken-3"
        href="https://semaphoreui.com/enterprise?utm_source=app&utm_content=feature_roles"
        target="_blank"
        data-testid="roles-upgrade"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>

      <span v-else style="font-weight: bold" data-testid="roles-contactAdmin">
        {{ $t('contact_admin_to_upgrade_enterprise') }}
      </span>
    </v-alert>

    <v-data-table
      data-testid="roles-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
      :headers="headers"
      :items="items"
      item-key="slug"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.name="{ item }">
        <div data-testid="roles-rowName">{{ item.name }}</div>
        <div class="text--secondary caption" data-testid="roles-rowSlug">{{ item.slug }}</div>
      </template>
      <template v-slot:item.scope="{ item }">
        <v-chip
          small
          label
          :data-testid="`roles-rowScope-${roleScope(item)}`"
        >{{ $t(roleScope(item)) }}</v-chip>
      </template>
      <template v-slot:item.permissions="{ item }">
        <TemplatePermissionsChips
          class="py-1"
          :permissions="item.permissions"
          data-testid="roles-rowPermissions"
        />
      </template>
      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap" data-testid="roles-rowActions">
          <v-btn
            icon
            class="mr-1"
            @click="askDeleteItem(item.slug)"
            data-testid="roles-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn icon class="mr-1" @click="editItem(item.slug)" data-testid="roles-rowEdit">
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
import TeamMenu from '@/components/TeamMenu.vue';
import TemplatePermissionsChips from '@/components/TemplatePermissionsChips.vue';
import { roleScope } from '@/lib/roles';
import { GLOBAL_PERMISSIONS, USER_PERMISSIONS } from '@/lib/constants';
import { hasGlobalPermission } from '@/lib/role-permissions';

/**
 * Project roles. Inside a project it lists the project's own roles; on the instance it lists the
 * shared roles every project can use. Global roles live on the Users page.
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
    EditDialog,
    TemplatePermissionsChips,
  },

  data() {
    return {
      roleImpact: null,
      forceDeleteDialog: false,
    };
  },

  computed: {
    IDFieldName() {
      return 'slug';
    },

    /** Project roles follow the project permission, shared roles the global one. */
    canManageRoles() {
      return this.allowActions();
    },
  },

  watch: {
    async projectId() {
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
      return [
        {
          text: this.$i18n.t('name'),
          value: 'name',
          width: '50%',
        },
        {
          text: this.$i18n.t('roleScope'),
          value: 'scope',
          sortable: false,
        },
        {
          text: this.$i18n.t('permissions'),
          value: 'permissions',
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
        if (impact && impact.mappings > 0 && !this.isAdmin) {
          EventBus.$emit('i-snackbar', { color: 'error', text: this.$i18n.t('roleUsedByMappings') });
          return;
        }
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
      return this.projectId ? `/api/project/${this.projectId}/roles` : '/api/roles';
    },

    getSingleItemUrl() {
      return this.projectId
        ? `/api/project/${this.projectId}/roles/${this.itemId}`
        : `/api/roles/${this.itemId}`;
    },

    getEventName() {
      return 'i-role';
    },
  },
};
</script>
