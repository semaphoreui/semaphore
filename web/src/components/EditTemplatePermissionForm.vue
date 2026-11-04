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
      :items="groupedRoles"
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
      :disabled="formSaving || (p.permission === SEE && verbChecked)"
    ></v-checkbox>

  </v-form>
</template>

<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';
import { ROLE_PERMISSIONS } from '@/lib/constants';
import groupRolesForSelect from '@/lib/roles';

// Seeing the object: a verb on it includes it.
const SEE = 128;

function pack(checked) {
  const mask = Object.keys(checked).filter((k) => checked[k]).reduce((res, k) => res | k, 0);
  return mask === 0 ? 0 : mask | SEE;
}

export default {
  mixins: [ItemFormBase],

  props: {
    objectId: [Number, String],
    scope: {
      type: String,
      default: 'template',
    },
  },

  data() {
    return {
      ROLE_PERMISSIONS,
      SEE,
      availableRoles: [],
      permissions: {},
    };
  },

  computed: {
    groupedRoles() {
      return groupRolesForSelect(this.availableRoles, (key) => this.$i18n.t(key));
    },

    verbChecked() {
      return Object.keys(this.permissions).some((k) => Number(k) !== SEE && this.permissions[k]);
    },
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

        this.item.permissions = pack(newPermissions);
      },
      deep: true,
    },

    // Watch item.permissions and update checkboxes
    'item.permissions': {
      handler(newPermissions) {
        if (newPermissions === undefined || newPermissions === null) return;

        this.permissions = ROLE_PERMISSIONS[this.scope].reduce((res, p) => ({
          ...res,
          [p.permission]: !!(this.item.permissions & p.permission),
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
      return `/api/project/${this.projectId}/${this.scope}s/${this.objectId}/perms`;
    },

    getSingleItemUrl() {
      return `${this.getItemsUrl()}/${this.itemId}`;
    },

    getNewItem() {
      return {
        role_slug: null,
        permissions: 0,
      };
    },

    beforeSave() {
      // Ensure permissions are properly set before saving
      if (this.item) {
        this.item.permissions = pack(this.permissions);
      }
    },

    afterLoadData() {
      // Initialize permissions checkboxes after loading data
      if (this.item && this.item.permissions !== undefined) {
        this.permissions = ROLE_PERMISSIONS[this.scope].reduce((res, p) => ({
          ...res,
          [p.permission]: !!(this.item.permissions & p.permission),
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
