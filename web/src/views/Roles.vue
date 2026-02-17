<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      save-button-text="Save"
      :title="$t('editRole')"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <RoleForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :is-admin="true"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      :title="$t('deleteRole')"
      :text="$t('askDeleteRole')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-btn icon class="mr-4" @click="returnToProjects()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ $t('Roles') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        :disabled="!premiumFeatures.custom_roles_management"
        color="primary"
        @click="editItem('new')"
        >{{ $t('newRole') }}</v-btn
      >
    </v-toolbar>

    <TeamMenu v-if="projectId" :project-id="projectId" :system-info="systemInfo" />

    <v-divider style="margin-top: -1px" />

    <v-alert
      v-if="!premiumFeatures.custom_roles_management"
      text
      color="amber darken-3"
      class="PageAlert"
    >
      <span class="mr-1" v-html="$t('roles_only_enterprise')"></span>

      <v-btn
        dark
        depressed
        v-if="isAdmin"
        color="amber darken-3"
        href="https://semaphoreui.com/enterprise"
        target="_blank"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>

      <span v-else style="font-weight: bold">
        {{ $t('contact_admin_to_upgrade_enterprise') }}
      </span>
    </v-alert>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.permissions="{ item }">
        <TemplatePermissionsChips class="py-1" :permissions="item.permissions" />
      </template>
      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn icon class="mr-1" @click="askDeleteItem(item.slug)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn icon class="mr-1" @click="editItem(item.slug)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import EventBus from '@/event-bus';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ItemListPageBase from '@/components/ItemListPageBase';
import EditDialog from '@/components/EditDialog.vue';
import RoleForm from '@/components/EditRoleForm.vue';
import TeamMenu from '@/components/TeamMenu.vue';
import TemplatePermissionsChips from '@/components/TemplatePermissionsChips.vue';

export default {
  mixins: [ItemListPageBase],

  props: {
    premiumFeatures: Object,
    projectId: Number,
    systemInfo: Object,
  },

  components: {
    TeamMenu,
    YesNoDialog,
    RoleForm,
    EditDialog,
    TemplatePermissionsChips,
  },

  data() {
    return {};
  },

  computed: {
    IDFieldName() {
      return 'slug';
    },
  },

  watch: {
    async projectId() {
      await this.loadItems();
    },
  },

  methods: {
    getHeaders() {
      return [
        {
          text: this.$i18n.t('name'),
          value: 'name',
          width: '50%',
        },
        {
          text: this.$i18n.t('permissions'),
          value: 'permissions',
        },
        {
          text: this.$i18n.t('actions'),
          value: 'actions',
          sortable: false,
        },
      ];
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    getItemsUrl() {
      return this.projectId ? `/api/project/${this.projectId}/roles` : '/api/roles';
    },

    getSingleItemUrl() {
      return this.projectId
        ? `/api/project/${this.projectId}/roles/${this.itemId}`
        : `/api/roles/${this.itemId}`;
    },

    getEventName() {
      return 'i-role';
    },
  },
};
</script>
