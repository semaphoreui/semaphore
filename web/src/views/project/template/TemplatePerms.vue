<template>
  <div
    v-if="items != null"
    data-testid="templatePerms-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
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
      test-id="templatePermsDeleteDialog"
    />

    <v-alert
      v-if="!features.custom_roles_management"
      variant="tonal"
      color="amber-darken-3"
      class="PageAlert"
      data-testid="templatePerms-licenseAlert"
    >
      <span
        class="mr-1"
        v-html="$t('roles_only_enterprise')"
        data-testid="templatePerms-licenseText"
      ></span>

      <v-btn
        variant="flat"
        v-if="isAdmin"
        color="amber-darken-3"
        href="https://semaphoreui.com/enterprise?utm_source=app&utm_content=feature_roles"
        target="_blank"
        data-testid="templatePerms-upgrade"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>

      <span v-else style="font-weight: bold" data-testid="templatePerms-contactAdmin">
        {{ $t('contact_admin_to_upgrade_enterprise') }}
      </span>
    </v-alert>

    <v-btn
      :disabled="!features.custom_roles_management"
      color="primary"
      @click="editItem('new')"
      style="position: absolute; right: 16px"
      data-testid="templatePerms-new"
      >{{ $t('Add Role') }}
    </v-btn>

    <v-data-table
      :headers="headers"
      :items="items"
      item-value="role_slug"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      :mobile-breakpoint="600"
      data-testid="templatePerms-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.role_slug="{ item }">
        <div class="d-flex align-center">
          <div>
            <div data-testid="templatePerms-rowName">{{ getRoleName(item.role_slug) }}</div>
            <div
              class="text-medium-emphasis text-caption"
              data-testid="templatePerms-rowSlug"
            >{{ item.role_slug }}</div>
          </div>
          <v-chip
            v-if="getRole(item.role_slug)"
            size="x-small"
            label
            class="ml-2"
            data-testid="templatePerms-rowScope"
          >
            {{ $t(roleScope(getRole(item.role_slug))) }}
          </v-chip>
        </div>
      </template>

      <template v-slot:item.permissions="{ item }">
        <TemplatePermissionsChips
          :permissions="item.permissions"
          scope="template"
          data-testid="templatePerms-rowPermissions"
        />
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-group
          density="compact"
          data-testid="templatePerms-rowActions"
        >
          <v-btn @click="editItem(item.id)" data-testid="templatePerms-rowEdit">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
          <v-btn @click="askDeleteItem(item.id)" data-testid="templatePerms-rowDelete">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </v-btn-group>
      </template>
    </v-data-table>
  </div>
</template>

<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import EditTemplatePermissionDialog from '@/components/EditTemplatePermissionDialog.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import TemplatePermissionsChips from '@/components/TemplatePermissionsChips.vue';
import axios from 'axios';
import { USER_PERMISSIONS } from '@/lib/constants';
import { roleScope } from '@/lib/roles';

export default {
  components: {
    EditTemplatePermissionDialog,
    YesNoDialog,
    TemplatePermissionsChips,
  },
  mixins: [ItemListPageBase],

  props: {
    projectId: Number,
    template: Object,
    repositories: Array,
    inventory: Array,
    environment: Array,
    features: Object,
    isAdmin: Boolean,
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
        const response = await axios.get(
          `/api/project/${this.template.project_id}/roles/all`,
        );
        this.availableRoles = response.data;
      } catch (error) {
        // eslint-disable-next-line no-console
        console.error('Failed to load roles:', error);
      }
    },

    getRole(slug) {
      return this.availableRoles.find((role) => role.slug === slug);
    },

    getRoleName(slug) {
      const role = this.getRole(slug);
      return role ? role.name : slug;
    },

    roleScope,

    allowActions() {
      return true;
    },

    getHeaders() {
      return [
        {
          title: this.$t('role'),
          key: 'role_slug',
          width: '25%',
        },
        {
          title: this.$t('permissions'),
          key: 'permissions',
          width: '65%',
        },
        {
          key: 'actions',
          sortable: false,
          width: '10%',
        },
      ];
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
