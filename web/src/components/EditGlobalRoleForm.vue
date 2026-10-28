<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
    data-testid="globalRole-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :value="formError"
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
      outlined
      dense
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.slug"
      data-testid="globalRole-slug"
      :label="$t('slug')"
      @input="slugTouched = true"
      :rules="slugRules"
      outlined
      dense
      required
      :disabled="formSaving || !isNew"
      :hint="$t('slugHint')"
    ></v-text-field>

    <v-subheader
      class="pl-0"
      data-testid="globalRole-permissionsTitle"
    >{{ $t('globalPermissions') }}</v-subheader>
    <div
      class="text--secondary caption mb-3"
      data-testid="globalRole-permissionsHint"
    >{{ $t('globalPermissionsHint') }}</div>

    <v-checkbox
      v-for="p in GLOBAL_ROLE_PERMISSIONS"
      :key="p.permission"
      class="mt-0"
      :input-value="hasPermission(p.permission)"
      :label="$t(p.label)"
      :disabled="formSaving"
      @change="setPermission(p.permission, $event)"
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
      this.$set(this.item, 'permissions', enabled ? current | permission : current & ~permission);
    },

    afterReset() {
      this.slugTouched = false;
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
