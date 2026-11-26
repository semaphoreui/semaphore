<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null && projects != null"
    data-testid="roleMapping-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="roleMapping-error"
    >{{ formError }}
    </v-alert>

    <v-select
      v-model="item.provider"
      :label="$t('roleMappingProvider')"
      :items="providers"
      item-value="id"
      item-title="name"
      :rules="[v => !!v || $t('provider_required')]"
      required
      data-testid="roleMapping-provider"
      :disabled="formSaving"
      variant="outlined"
      density="compact"
    ></v-select>

    <v-text-field
      v-model="item.group"
      :label="$t('roleMappingGroup')"
      :rules="[v => !!v || $t('group_required')]"
      required
      data-testid="roleMapping-group"
      :disabled="formSaving"
      variant="outlined"
      density="compact"
    ></v-text-field>

    <v-select
      v-model="item.project_id"
      :label="$t('project')"
      :items="projects"
      item-value="id"
      item-title="name"
      :rules="[v => !!v || $t('project_required')]"
      required
      data-testid="roleMapping-projectId"
      :disabled="formSaving"
      variant="outlined"
      density="compact"
    ></v-select>

    <v-select
      v-model="item.role"
      :label="$t('role')"
      :items="groupedRoles"
      item-value="slug"
      item-title="name"
      :rules="[v => !!v || $t('role_required')]"
      required
      data-testid="roleMapping-role"
      :disabled="formSaving || item.project_id == null"
      variant="outlined"
      density="compact"
      :menu-props="{
        location: 'top', origin: 'top', contentClass: 'RoleMappingRoleMenu', maxHeight: 304,
      }"
    >
      <template v-slot:item="{ props, item: role }">
        <v-list-item
          v-bind="props"
          :subtitle="role.slug"
          :data-testid="`roleMapping-roleOption-${role.slug}`"
        />
      </template>
    </v-select>
  </v-form>
</template>
<style lang="scss">
// The role list: the dense list of a dense select of Vuetify 2 (`.v-list--dense`), at most 304 px
// high (`menuProps.maxHeight`) — group headers 40 px with the text 8 px in (`.v-subheader`,
// 12 px), two-line items 50 px: the name on a 16 px line, the slug 2 px under it at 13 px medium
// in the secondary text colour (`.v-list-item__subtitle`).
.RoleMappingRoleMenu {
  .v-list-subheader {
    min-height: 40px;
    padding-inline: 8px;
    font-size: 12px;
    line-height: 18px;
  }

  .v-list-item {
    min-height: 50px;
    padding-block: 8px;
  }

  .v-list-item-title {
    line-height: 16px;
  }

  .v-list-item-subtitle {
    margin-top: 2px;
    font-size: 13px;
    font-weight: 500;
    line-height: 16px;
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
    opacity: 1;
  }
}
</style>
<script>
import axios from 'axios';
import ItemFormBase from '@/components/ItemFormBase';
import groupRolesForSelect from '@/lib/roles';

export default {
  mixins: [ItemFormBase],

  props: {
    providers: Array,
  },

  data() {
    return {
      projects: null,
      roles: [],
    };
  },

  computed: {
    groupedRoles() {
      return groupRolesForSelect(this.roles, (key) => this.$t(key));
    },
  },

  watch: {
    'item.project_id': async function onProjectChange(projectId) {
      if (projectId == null) {
        this.roles = [];
        return;
      }
      const roles = (await axios({
        method: 'get',
        url: `/api/project/${projectId}/roles/all`,
        responseType: 'json',
      })).data;
      // A late answer for a project chosen before must not replace the roles of the current one.
      if (projectId !== this.item.project_id) {
        return;
      }
      this.roles = roles;
      // A project role of the previous project would stay selected but hidden.
      if (!this.roles.some((role) => role.slug === this.item.role)) {
        this.item.role = null;
      }
    },
  },

  async created() {
    this.projects = (await axios({
      method: 'get',
      url: '/api/projects',
      responseType: 'json',
    })).data;
  },

  methods: {
    getItemsUrl() {
      return '/api/role_mappings';
    },

    getSingleItemUrl() {
      return `/api/role_mappings/${this.itemId}`;
    },

    getNewItem() {
      return {
        provider: null,
        group: '',
        project_id: null,
        role: null,
      };
    },

    getRequestOptions() {
      // The base form would replace project_id with the page project, which is none here.
      return { data: this.item };
    },
  },
};
</script>
