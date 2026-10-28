<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div
    v-if="items != null"
    data-testid="globalRoles-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="globalRoleDialog"
      save-button-text="Save"
      :title="itemId === 'new' ? $t('newGlobalRole') : $t('editGlobalRole')"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <GlobalRoleForm
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
      test-id="globalRolesDeleteDialog"
      :title="$t('deleteGlobalRole')"
      :text="$t('askDeleteGlobalRole')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <YesNoDialog
      test-id="globalRolesForceDeleteDialog"
      :title="$t('deleteGlobalRole')"
      :text="roleImpact ? $t('askDeleteAssignedGlobalRole', roleImpact) : ''"
      :yes-button-title="$t('deleteWithAssignments')"
      :max-width="520"
      v-model="forceDeleteDialog"
      @yes="deleteItem(itemId, true)"
    />

    <v-alert
      v-if="!features.global_roles_management"
      variant="tonal"
      color="hsl(348deg, 86%, 61%)"
      class="PageAlert"
      data-testid="globalRoles-licenseAlert"
    >
      <span
        class="mr-1"
        v-html="$t('global_roles_only_pro')"
        data-testid="globalRoles-licenseAlertText"
      ></span>

      <v-btn
        dark
        variant="flat"
        v-if="isAdmin"
        color="hsl(348deg, 86%, 61%)"
        @click="upgradeToPro('global_roles_management')"
        data-testid="globalRoles-upgrade"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>

      <span v-else style="font-weight: bold" data-testid="globalRoles-contactAdmin">
        {{ $t('contact_admin_to_upgrade') }}
      </span>
    </v-alert>

    <v-alert
      variant="tonal"
      density="compact"
      type="info"
      class="mx-4 mt-4"
      data-testid="globalRoles-hint"
    >
      {{ $t('globalRolesHint') }}
    </v-alert>

    <v-data-table
      :headers="headers"
      :items="items"
      item-value="id"
      class="mt-4"
      :items-per-page-options="[20]"
      data-testid="globalRoles-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.name="{ item }">
        <div data-testid="globalRoles-rowName">{{ item.name }}</div>
        <div
          class="text-medium-emphasis text-caption"
          data-testid="globalRoles-rowSlug"
        >{{ item.slug }}</div>
      </template>
      <template v-slot:item.permissions="{ item }">
        <TemplatePermissionsChips
          class="py-1"
          scope="global"
          :permissions="item.permissions || 0"
          data-testid="globalRoles-rowPermissions"
        />
      </template>
      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap" data-testid="globalRoles-rowActions">
          <v-btn
            icon
            class="mr-1"
            @click="askDeleteItem(item.id)"
            data-testid="globalRoles-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn icon class="mr-1" @click="editItem(item.id)" data-testid="globalRoles-rowEdit">
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
import GlobalRoleForm from '@/components/EditGlobalRoleForm.vue';
import TemplatePermissionsChips from '@/components/TemplatePermissionsChips.vue';
import { GLOBAL_PERMISSIONS } from '@/lib/constants';
import { hasGlobalPermission } from '@/lib/role-permissions';

/**
 * The global roles of the instance: roles that grant instance-wide permissions to the users
 * they are assigned to. Shown on the Users page; the parent opens the "new" dialog through
 * editItem('new').
 */
export default {
  mixins: [ItemListPageBase],

  props: {
    features: Object,
  },

  components: {
    YesNoDialog,
    GlobalRoleForm,
    EditDialog,
    TemplatePermissionsChips,
  },

  data() {
    return {
      roleImpact: null,
      forceDeleteDialog: false,
    };
  },

  watch: {
    // The plan may be read after the list: a change of the feature reloads the roles.
    'features.global_roles_management': async function onFeature() {
      await this.loadItems();
    },
  },

  methods: {
    // Without the Pro feature the API answers 403: the list stays empty and the page shows the
    // upgrade alert instead of nothing at all.
    async loadItems() {
      if (!(this.features || {}).global_roles_management) {
        this.items = [];
        return;
      }
      try {
        await ItemListPageBase.methods.loadItems.call(this);
      } catch (err) {
        if (err.response && err.response.status === 403) {
          this.items = [];
          return;
        }
        throw err;
      }
    },

    // Called from data() of the list mixin, before computed properties exist,
    // so the props are read directly.
    allowActions() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageRoles);
    },

    getHeaders() {
      return [
        {
          title: this.$t('name'),
          key: 'name',
          width: '50%',
        },
        {
          title: this.$t('globalPermissions'),
          key: 'permissions',
        },
        {
          title: this.$t('actions'),
          key: 'actions',
          sortable: false,
        },
      ];
    },

    async deleteItem(itemId, force = false) {
      this.itemId = itemId;
      try {
        const item = this.items.find((x) => x.id === itemId);
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
        if (impact) {
          this.roleImpact = impact;
          this.forceDeleteDialog = true;
          return;
        }
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
    },

    getItemsUrl() {
      return '/api/global-roles';
    },

    getSingleItemUrl() {
      return `/api/global-roles/${this.itemId}`;
    },

    getEventName() {
      return 'i-global-role';
    },
  },
};
</script>
