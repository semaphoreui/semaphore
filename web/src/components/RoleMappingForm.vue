<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null && projects != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-select
      v-model="item.provider"
      :label="$t('roleMappingProvider')"
      :items="providers"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || $t('provider_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-select>

    <v-text-field
      v-model="item.group"
      :label="$t('roleMappingGroup')"
      :rules="[v => !!v || $t('group_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-select
      v-model="item.project_id"
      :label="$t('project')"
      :items="projects"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || $t('project_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-select>

    <v-select
      v-model="item.role"
      :label="$t('role')"
      :items="groupedRoles"
      item-value="slug"
      item-text="name"
      :rules="[v => !!v || $t('role_required')]"
      required
      :disabled="formSaving || item.project_id == null"
      outlined
      dense
    >
      <template v-slot:item="{ item: role }">
        <v-list-item-content>
          <v-list-item-title>{{ role.name }}</v-list-item-title>
          <v-list-item-subtitle>{{ role.slug }}</v-list-item-subtitle>
        </v-list-item-content>
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
      return groupRolesForSelect(this.roles, (key) => this.$i18n.t(key));
    },
  },

  watch: {
    'item.project_id': async function onProjectChange(projectId) {
      if (projectId == null) {
        this.roles = [];
        return;
      }
      this.roles = (await axios({
        method: 'get',
        url: `/api/project/${projectId}/roles/all`,
        responseType: 'json',
      })).data;
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
