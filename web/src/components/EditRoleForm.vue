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

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[v => !!v || $t('name_required')]"
      outlined
      dense
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.slug"
      :label="$t('slug')"
      :rules="[v => !!v || $t('slug_required'), v => this.validateSlug(v)]"
      outlined
      dense
      required
      :disabled="formSaving"
      :hint="$t('slugHint')"
    ></v-text-field>

<!--    <v-divider class="my-4"></v-divider>-->

    <v-subheader class="pl-0">{{ $t('permissions') }}</v-subheader>

    <v-checkbox
      class="mt-0"
      v-model="permissions.canRunProjectTasks"
      :label="$t('canRunProjectTasks')"
      :disabled="formSaving"
    ></v-checkbox>

    <v-checkbox
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
      class="mt-0"
      v-model="permissions.canManageProjectUsers"
      :label="$t('canManageProjectUsers')"
      :disabled="formSaving"
    ></v-checkbox>

  </v-form>
</template>

<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      permissions: {
        canRunProjectTasks: false,
        canUpdateProject: false,
        canManageProjectResources: false,
        canManageProjectUsers: false,
      },
    };
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
    validateSlug(value) {
      if (!value) return true; // Required validation is handled separately

      // Slug should be lowercase, alphanumeric with underscores/hyphens
      const slugPattern = /^[a-z0-9_-]+$/;
      if (!slugPattern.test(value)) {
        return this.$t('invalidSlugFormat');
      }

      return true;
    },

    getItemsUrl() {
      if (this.projectId) {
        return `/api/project/${this.projectId}/roles`;
      }
      return '/api/roles';
    },

    getSingleItemUrl() {
      if (this.projectId) {
        return `/api/project/${this.projectId}/roles/${this.itemId}`;
      }
      return `/api/roles/${this.itemId}`;
    },

    getNewItem() {
      return {
        name: '',
        slug: '',
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
