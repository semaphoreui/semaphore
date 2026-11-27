<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null"
    data-testid="role-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="role-error"
    >{{ formError }}
    </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[v => !!v || $t('name_required')]"
      variant="outlined"
      density="compact"
      required
      :disabled="formSaving"
      data-testid="role-name"
    ></v-text-field>

    <v-text-field
      v-model="item.slug"
      :label="$t('slug')"
      @update:model-value="slugTouched = true"
      :rules="slugRules"
      variant="outlined"
      density="compact"
      required
      :disabled="formSaving || !isNew"
      :hint="$t('slugHint')"
      data-testid="role-slug"
    ></v-text-field>

<!--    <v-divider class="my-4"></v-divider>-->

    <v-list-subheader
      class="pl-0"
      data-testid="role-permissionsTitle"
    >{{ $t('permissions') }}</v-list-subheader>

    <v-checkbox
      v-for="p in ROLE_PERMISSIONS.default"
      :key="p.permission"
      class="mt-0"
      v-model="permissions[p.permission]"
      :label="$t(p.label)"
      :disabled="formSaving"
      :data-testid="`role-${p.label}`"
    ></v-checkbox>

  </v-form>
</template>

<script>
import ItemFormBase from '@/components/ItemFormBase';
import { isValidSlug, slugFromName } from '@/lib/roles';
import { ROLE_PERMISSIONS } from '@/lib/constants';

const KNOWN = ROLE_PERMISSIONS.default.reduce((mask, p) => mask | p.permission, 0);

function checkboxes(mask) {
  return ROLE_PERMISSIONS.default.reduce(
    (res, p) => ({ ...res, [p.permission]: !!(mask & p.permission) }),
    {},
  );
}

// Bits the form does not show stay as they are.
function pack(mask, checked) {
  const kept = (mask || 0) & ~KNOWN;
  return Object.keys(checked).filter((k) => checked[k]).reduce((res, k) => res | k, kept);
}

/** The form of a project role: a name, a slug and the project permissions it grants. */
export default {
  mixins: [ItemFormBase],

  data() {
    return {
      slugTouched: false,
      permissions: checkboxes(0),
      ROLE_PERMISSIONS,
    };
  },

  computed: {
    slugRules() {
      if (!this.isNew) return [];
      return [(v) => !!v || this.$t('slug_required'), (v) => this.validateSlug(v)];
    },
  },

  watch: {
    'item.name': function onNameChange(name) {
      if (this.isNew && !this.slugTouched && this.item) {
        this.item.slug = slugFromName(name);
      }
    },

    permissions: {
      handler(checked) {
        if (!this.item) return;
        this.item.permissions = pack(this.item.permissions, checked);
      },
      deep: true,
    },

    'item.permissions': {
      handler(mask) {
        if (mask === undefined || mask === null) return;
        this.permissions = checkboxes(mask);
      },
      immediate: true,
    },
  },

  methods: {
    validateSlug(value) {
      if (!value) return true; // Required validation is handled separately
      return isValidSlug(value) || this.$t('invalidSlugFormat');
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
      if (this.item) {
        this.item.permissions = pack(this.item.permissions, this.permissions);
      }
    },

    afterLoadData() {
      if (this.item && this.item.permissions !== undefined) {
        this.permissions = checkboxes(this.item.permissions);
      }
    },

    afterReset() {
      this.slugTouched = false;
      this.permissions = checkboxes(0);
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
