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
      v-for="p in ROLE_PERMISSIONS[scope]"
      :key="p.permission"
      class="mt-0"
      v-model="permissions[p.permission]"
      :label="$t(p.label)"
      :disabled="formSaving"
    ></v-checkbox>

  </v-form>
</template>

<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';
import { ROLE_PERMISSIONS } from '@/lib/constants';

export default {
  mixins: [ItemFormBase],

  props: {
    templateId: [Number, String],
    scope: {
      type: String,
      default: 'default',
    },
  },

  data() {
    return {
      ROLE_PERMISSIONS,
      availableRoles: [],
      permissions: {},
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

        this.item.permissions = Object.keys(newPermissions)
          .filter((k) => newPermissions[k])
          .reduce((res, k) => res | k, 0);
      },
      deep: true,
    },

    // Watch item.permissions and update checkboxes
    'item.permissions': {
      handler(newPermissions) {
        if (newPermissions === undefined || newPermissions === null) return;

        this.permissions = [1, 2, 4, 8].reduce((res, k) => ({
          ...res,
          [k]: !!(this.item.permissions & k),
        }), {});
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
        this.item.permissions = Object.keys(this.permissions)
          .filter((k) => this.permissions[k])
          .reduce((res, k) => res | k, 0);

        this.item.template_id = parseInt(this.templateId, 10);
        this.item.project_id = this.projectId;
      }
    },

    afterLoadData() {
      // Initialize permissions checkboxes after loading data
      if (this.item && this.item.permissions !== undefined) {
        this.permissions = [1, 2, 4, 8].reduce((res, k) => ({
          ...res,
          [k]: !!(this.item.permissions & k),
        }), {});
      }
    },

    afterReset() {
      // Reset permissions checkboxes
      this.permissions = {};
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
