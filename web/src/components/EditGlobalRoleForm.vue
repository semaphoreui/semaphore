<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null"
    data-testid="globalRole-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="globalRole-error"
    >{{ formError }}
    </v-alert>

    <v-text-field
      v-model="item.name"
      data-testid="globalRole-name"
      :label="$t('name')"
      :rules="[v => !!v || $t('name_required')]"
      variant="outlined"
      density="compact"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.slug"
      data-testid="globalRole-slug"
      :label="$t('slug')"
      @update:model-value="slugTouched = true"
      :rules="slugRules"
      variant="outlined"
      density="compact"
      required
      :disabled="formSaving || !isNew"
      :hint="$t('slugHint')"
    ></v-text-field>

    <v-list-subheader
      class="pl-0"
      data-testid="globalRole-permissionsTitle"
    >{{ $t('globalPermissions') }}</v-list-subheader>
    <div
      class="text-medium-emphasis text-caption mb-3"
      data-testid="globalRole-permissionsHint"
    >{{ $t('globalPermissionsHint') }}</div>

    <v-checkbox
      v-for="p in GLOBAL_ROLE_PERMISSIONS"
      :key="p.permission"
      class="mt-0"
      :model-value="hasPermission(p.permission)"
      :label="$t(p.label)"
      :disabled="formSaving"
      @update:model-value="setPermission(p.permission, $event)"
      :data-testid="`globalRole-${p.label}`"
    ></v-checkbox>
  </v-form>
</template>

<script>
import ItemFormBase from '@/components/ItemFormBase';
import { ROLE_PERMISSIONS } from '@/lib/constants';
import { isValidSlug, slugFromName } from '@/lib/roles';

/** The form of a global role: a name, a slug and the instance-wide permissions it grants. */
export default {
  mixins: [ItemFormBase],

  data() {
    return {
      GLOBAL_ROLE_PERMISSIONS: ROLE_PERMISSIONS.global,
      slugTouched: false,
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
  },

  methods: {
    validateSlug(value) {
      if (!value) return true; // Required validation is handled separately
      return isValidSlug(value) || this.$t('invalidSlugFormat');
    },

    getItemsUrl() {
      return '/api/global-roles';
    },

    getSingleItemUrl() {
      return `/api/global-roles/${this.itemId}`;
    },

    getNewItem() {
      return {
        name: '',
        slug: '',
        permissions: 0,
      };
    },

    hasPermission(permission) {
      // eslint-disable-next-line no-bitwise
      return ((this.item.permissions || 0) & permission) === permission;
    },

    setPermission(permission, enabled) {
      const current = this.item.permissions || 0;
      // eslint-disable-next-line no-bitwise
      this.item.permissions = enabled ? current | permission : current & ~permission;
    },

    afterReset() {
      this.slugTouched = false;
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
