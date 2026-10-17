<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
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
      :title="$t('deleteGlobalRole')"
      :text="$t('askDeleteGlobalRole')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <YesNoDialog
      :title="$t('deleteGlobalRole')"
      :text="roleImpact ? $t('askDeleteAssignedGlobalRole', roleImpact) : ''"
      :yes-button-title="$t('deleteWithAssignments')"
      :max-width="520"
      v-model="forceDeleteDialog"
      @yes="deleteItem(itemId, true)"
    />

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

    <v-alert text dense type="info" class="mx-4 mt-4">
      {{ $t('globalRolesHint') }}
    </v-alert>

    <v-data-table
      :headers="headers"
      :items="items"
      item-key="id"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.name="{ item }">
        <div>{{ item.name }}</div>
        <div class="text--secondary caption">{{ item.slug }}</div>
      </template>
      <template v-slot:item.permissions="{ item }">
        <TemplatePermissionsChips
          class="py-1"
          scope="global"
          :permissions="item.permissions || 0"
        />
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

  methods: {
    // Called from data() of the list mixin, before computed properties exist,
    // so the props are read directly.
    allowActions() {
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
          text: this.$i18n.t('globalPermissions'),
          value: 'permissions',
        },
        {
          text: this.$i18n.t('actions'),
          value: 'actions',
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
