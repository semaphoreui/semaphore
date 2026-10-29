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
      class="pb-2 mb-4"
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
    >
      <template v-slot:item="{ props, item: { raw: role } }">
        <v-list-item
          v-bind="props"
          :subtitle="role.slug"
          :data-testid="`roleMapping-roleOption-${role.slug}`"
        />
      </template>
    </v-select>
  </v-form>
</template>
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
