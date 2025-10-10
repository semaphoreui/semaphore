<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-select
      v-model="item.role_slug"
      :items="availableRoles"
      item-value="slug"
      item-text="name"
      :label="$t('role')"
      :rules="[v => !!v || $t('role_required')]"
      required
      outlined
      dense
      :disabled="formSaving"
    >
      <template v-slot:item="{ item: role }">
        <v-list-item-content>
          <v-list-item-title>{{ role.name }}</v-list-item-title>
          <v-list-item-subtitle>{{ role.slug }}</v-list-item-subtitle>
        </v-list-item-content>
      </template>
    </v-select>

    <v-subheader class="pl-0">{{ $t('permissions') }}</v-subheader>

    <v-checkbox
      class="mt-0"
      v-model="permissions.canRunProjectTasks"
      :label="$t('canRunProjectTasks')"
      :disabled="formSaving"
    ></v-checkbox>

    <v-checkbox
      v-if="templateId == null"
      class="mt-0"
      v-model="permissions.canUpdateProject"
      :label="$t('canUpdateProject')"
      :disabled="formSaving"
    ></v-checkbox>

    <v-checkbox
      class="mt-0"
      v-model="permissions.canManageProjectResources"
      :label="$t('canManageProjectResources')"
      :disabled="formSaving"
    ></v-checkbox>

    <v-checkbox
      v-if="templateId == null"
      class="mt-0"
      v-model="permissions.canManageProjectUsers"
      :label="$t('canManageProjectUsers')"
      :disabled="formSaving"
    ></v-checkbox>

  </v-form>
</template>

<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  mixins: [ItemFormBase],

  props: {
    templateId: [Number, String],
  },

  data() {
    return {
      availableRoles: [],
      permissions: {
        canRunProjectTasks: false,
        canUpdateProject: false,
        canManageProjectResources: false,
        canManageProjectUsers: false,
      },
    };
  },

  async created() {
    await this.loadRoles();
    await this.loadData();
  },

  watch: {
    // Watch permissions and update the item.permissions value
    permissions: {
      handler(newPermissions) {
        if (!this.item) return;

        let permissionValue = 0;
        if (newPermissions.canRunProjectTasks) permissionValue |= 1;
        if (newPermissions.canUpdateProject) permissionValue |= 2;
        if (newPermissions.canManageProjectResources) permissionValue |= 4;
        if (newPermissions.canManageProjectUsers) permissionValue |= 8;

        this.item.permissions = permissionValue;
      },
      deep: true,
    },

    // Watch item.permissions and update checkboxes
    'item.permissions': {
      handler(newPermissions) {
        if (newPermissions === undefined || newPermissions === null) return;

        this.permissions.canRunProjectTasks = !!(newPermissions & 1);
        this.permissions.canUpdateProject = !!(newPermissions & 2);
        this.permissions.canManageProjectResources = !!(newPermissions & 4);
        this.permissions.canManageProjectUsers = !!(newPermissions & 8);
      },
      immediate: true,
    },
  },

  methods: {
    async loadRoles() {
      try {
        const response = await axios.get(`/api/project/${this.projectId}/roles/all`);
        this.availableRoles = response.data;
      } catch (error) {
        this.formError = getErrorMessage(error);
      }
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/templates/${this.templateId}/perms`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/templates/${this.templateId}/perms/${this.itemId}`;
    },

    getNewItem() {
      return {
        role_slug: null,
        template_id: parseInt(this.templateId, 10),
        project_id: this.projectId,
        permissions: 0,
      };
    },

    beforeSave() {
      // Ensure permissions are properly set before saving
      if (this.item) {
        let permissionValue = 0;
        if (this.permissions.canRunProjectTasks) permissionValue |= 1;
        if (this.permissions.canUpdateProject) permissionValue |= 2;
        if (this.permissions.canManageProjectResources) permissionValue |= 4;
        if (this.permissions.canManageProjectUsers) permissionValue |= 8;

        this.item.permissions = permissionValue;
        this.item.template_id = parseInt(this.templateId, 10);
        this.item.project_id = this.projectId;
      }
    },

    afterLoadData() {
      // Initialize permissions checkboxes after loading data
      if (this.item && this.item.permissions !== undefined) {
        this.permissions.canRunProjectTasks = !!(this.item.permissions & 1);
        this.permissions.canUpdateProject = !!(this.item.permissions & 2);
        this.permissions.canManageProjectResources = !!(this.item.permissions & 4);
        this.permissions.canManageProjectUsers = !!(this.item.permissions & 8);
      }
    },

    afterReset() {
      // Reset permissions checkboxes
      this.permissions = {
        canRunProjectTasks: false,
        canUpdateProject: false,
        canManageProjectResources: false,
        canManageProjectUsers: false,
      };
    },
  },
};
</script>

<style scoped>
.v-subheader {
  font-weight: 500;
  font-size: 14px;
}
</style>
