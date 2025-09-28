<template>
  <div v-if="items != null">
    <EditTemplatePermissionDialog
      v-model="editDialog"
      :project-id="projectId"
      :template-id="templateId"
      :item-id="itemId"
      @save="loadItems()"
    />

    <YesNoDialog
      :title="$t('deleteTemplatePermission')"
      :text="$t('askDeleteTemplatePermission')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('templatePermissions') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >{{ $t('newTemplatePermission') }}
      </v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto;"
    >
      <template v-slot:item.role="{ item }">
        <v-chip
          :color="getRoleColor(item.role_id)"
          text-color="white"
          small
        >
          {{ getRoleName(item.role_id) }}
        </v-chip>
      </template>

      <template v-slot:item.permissions="{ item }">
        <div class="permissions-list">
          <v-chip
            v-if="item.permissions & 1"
            small
            color="blue"
            text-color="white"
            class="mr-1 mb-1"
          >
            {{ $t('canRunProjectTasks') }}
          </v-chip>
          <v-chip
            v-if="item.permissions & 2"
            small
            color="green"
            text-color="white"
            class="mr-1 mb-1"
          >
            {{ $t('canUpdateProject') }}
          </v-chip>
          <v-chip
            v-if="item.permissions & 4"
            small
            color="orange"
            text-color="white"
            class="mr-1 mb-1"
          >
            {{ $t('canManageProjectResources') }}
          </v-chip>
          <v-chip
            v-if="item.permissions & 8"
            small
            color="red"
            text-color="white"
            class="mr-1 mb-1"
          >
            {{ $t('canManageProjectUsers') }}
          </v-chip>
          <span v-if="item.permissions === 0" class="text--secondary">
            {{ $t('noPermissions') }}
          </span>
        </div>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn
            @click="editItem(item.id)"
            v-if="can(USER_PERMISSIONS.manageProjectResources)"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
          <v-btn
            @click="askDeleteItem(item.id)"
            v-if="can(USER_PERMISSIONS.manageProjectResources)"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>
  </div>
</template>

<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import EditTemplatePermissionDialog from '@/components/EditTemplatePermissionDialog.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import axios from 'axios';
import { USER_PERMISSIONS } from '@/lib/constants';

export default {
  components: {
    EditTemplatePermissionDialog,
    YesNoDialog,
  },
  mixins: [ItemListPageBase],

  props: {
    projectId: Number,
    template: Object,
    repositories: Array,
    inventory: Array,
    environment: Array,
  },

  data() {
    return {
      USER_PERMISSIONS,
      availableRoles: [],
    };
  },

  computed: {
    templateId() {
      return this.template.id;
    },
  },

  async created() {
    await this.loadRoles();
  },

  methods: {
    async loadRoles() {
      try {
        const response = await axios.get('/api/roles');
        this.availableRoles = response.data;
      } catch (error) {
        // eslint-disable-next-line no-console
        console.error('Failed to load roles:', error);
      }
    },

    getRoleName(roleId) {
      const role = this.availableRoles.find((r) => r.id === roleId);
      return role ? role.name : `Role ${roleId}`;
    },

    getRoleColor(roleId) {
      const role = this.availableRoles.find((r) => r.id === roleId);
      if (!role) return 'gray';

      // Color based on role slug or default colors
      const colorMap = {
        owner: 'red',
        manager: 'orange',
        task_runner: 'blue',
        guest: 'gray',
      };

      return colorMap[role.slug] || 'primary';
    },

    allowActions() {
      return this.can(USER_PERMISSIONS.manageProjectResources);
    },

    getHeaders() {
      return [
        {
          text: this.$i18n.t('role'),
          value: 'role',
          width: '25%',
        },
        {
          text: this.$i18n.t('permissions'),
          value: 'permissions',
          width: '65%',
        },
        {
          value: 'actions',
          sortable: false,
          width: '10%',
        }];
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/templates/${this.templateId}/perms/${this.itemId}`;
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/templates/${this.templateId}/perms`;
    },

    getEventName() {
      return 'i-template-perms';
    },
  },
};
</script>

<style scoped>
.permissions-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
</style>
