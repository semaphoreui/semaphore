<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null"
    data-testid="templatePerm-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="templatePerm-error"
    >{{ formError }}
    </v-alert>

    <v-select
      v-model="item.role_slug"
      :items="groupedRoles"
      item-value="slug"
      item-title="name"
      :label="$t('role')"
      :rules="[v => !!v || $t('role_required')]"
      required
      variant="outlined"
      density="compact"
      :disabled="formSaving"
      :menu-props="{ location: 'top', origin: 'top', maxHeight: 304 }"
      :list-props="{ class: 'template-perm-form__roles' }"
      data-testid="templatePerm-roleSlug"
    >
      <template v-slot:item="{ item: role, props }">
        <v-list-item v-bind="props" :title="undefined">
          <div :data-testid="`templatePerm-roleSlug-${role.slug}`">
            <v-list-item-title>{{ role.name }}</v-list-item-title>
            <v-list-item-subtitle>{{ role.slug }}</v-list-item-subtitle>
          </div>
        </v-list-item>
      </template>
    </v-select>

    <v-list-subheader
      class="pl-0"
      data-testid="templatePerm-permissionsTitle"
    >{{ $t('permissions') }}</v-list-subheader>

    <v-checkbox
      v-for="p in ROLE_PERMISSIONS[scope]"
      :key="p.permission"
      class="mt-0"
      v-model="permissions[p.permission]"
      :label="$t(p.label)"
      :disabled="formSaving || (p.permission === SEE && verbChecked)"
      :data-testid="`templatePerm-permission-${p.permission}`"
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
      return groupRolesForSelect(this.availableRoles, (key) => this.$t(key));
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
.v-list-subheader {
  font-weight: 500;
  font-size: 14px;
}
</style>

<style lang="scss">
// The role list of the dense select is a dense list of Vuetify 2 (`.v-list--dense`): headers
// 40 px high, 12 px, 8 px from the edge; two-line items 50 px (content padding 8 px, a title of
// 16 px on a 16 px line — App.vue of Vuetify 2 kept the dense title at 1rem and normal weight —,
// 2 px, then a subtitle of 13 px, weight 500, in the secondary text colour). The menu is at most
// 304 px high (`menu-props` of the Vuetify 2 `v-select`; Vuetify 4: 310), opening over the field
// as the VSelect defaults in plugins/vuetify.js say (a `menu-props` prop replaces them).
.template-perm-form__roles {
  > .v-list-subheader {
    min-height: 40px;
    padding-inline: 8px;
    font-size: .75rem;
  }

  > .v-list-item .v-list-item__content {
    padding-block: 8px;
  }

  > .v-list-item .v-list-item-title {
    line-height: 1rem;

    &:not(:last-child) {
      margin-bottom: 2px;
    }
  }

  > .v-list-item .v-list-item-subtitle {
    font-size: .8125rem;
    font-weight: 500;
    line-height: 1rem;
    opacity: 1;
  }

  > .v-list-item.v-theme--light .v-list-item-subtitle {
    color: rgba(0, 0, 0, .6);
  }

  > .v-list-item.v-theme--dark .v-list-item-subtitle {
    color: rgba(255, 255, 255, .7);
  }
}
</style>
