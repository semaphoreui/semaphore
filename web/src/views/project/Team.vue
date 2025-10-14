<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditTeamMemberDialog
      v-model="editDialog"
      :project-id="projectId"
      :item-id="itemId"
      :invites-enabled="systemInfo.teams.invites_enabled"
      :invite-type="systemInfo.teams.invite_type"
      :roles="systemInfo.roles"
      @save="openInvites()"
    />

    <YesNoDialog
      :title="$t('deleteTeamMember')"
      :text="$t('askDeleteTMem')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('team2') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        v-if="systemInfo.teams.memebers_can_leave"
        color="error"
        @click="leftProject()"
        class="mr-2"
        :disabled="userRole === 'owner'"
      >{{ $t('LeaveProject') }}
      </v-btn>
      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectUsers)"
      >{{ $t('newTeamMember') }}
      </v-btn>
    </v-toolbar>

    <TeamMenu
      v-if="isPro"
      :project-id="projectId"
      :system-info="systemInfo"
    />

    <v-divider style="margin-top: -1px;"/>

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto;"
    >
      <template v-slot:item.role="{ item }">
        <v-select
          hide-details
          v-model="item.role"
          :items="userRoles"
          item-value="slug"
          item-text="name"
          :style="{width: '200px'}"
          @change="updateProjectUser(item)"
          v-if="can(USER_PERMISSIONS.manageProjectUsers)"
          class="pt-0 mt-0"
        />
        <div v-else>{{ userRoles.find(r => r.slug === item.role).name }}</div>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn @click="askDeleteItem(item.id)" v-if="can(USER_PERMISSIONS.manageProjectUsers)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>
  </div>

</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import EditTeamMemberDialog from '@/components/EditTeamMemberDialog.vue';
import axios from 'axios';
import { USER_PERMISSIONS, USER_ROLES } from '@/lib/constants';
import TeamMenu from '@/components/TeamMenu.vue';

export default {
  components: { TeamMenu, EditTeamMemberDialog },
  mixins: [ItemListPageBase],

  props: {
    systemInfo: Object,
  },

  computed: {
    userRoles() {
      return [...USER_ROLES, ...this.systemInfo.roles];
    },

    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },
  },

  methods: {
    openInvites() {
      if (this.systemInfo.teams.invites_enabled) {
        this.$router.push(`/project/${this.projectId}/invites`);
        return;
      }
      this.loadItems();
    },

    async leftProject() {
      await axios({
        method: 'delete',
        url: `/api/project/${this.projectId}/me`,
        responseType: 'json',
      });
      window.location.reload();
    },

    async updateProjectUser(user) {
      await axios({
        method: 'put',
        url: `/api/project/${this.projectId}/users/${user.id}`,
        responseType: 'json',
        data: user,
      });
      await this.loadItems();
    },

    allowActions() {
      return this.can(USER_PERMISSIONS.manageProjectUsers);
    },

    getHeaders() {
      return [
        {
          text: this.$i18n.t('name'),
          value: 'name',
          width: '40%',
        },
        {
          text: this.$i18n.t('username'),
          value: 'username',
          width: '30%',
        },
        {
          text: this.$i18n.t('role'),
          value: 'role',
          width: '30%',
        },
        {
          value: 'actions',
          sortable: false,
          width: '0%',
        }];
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/users/${this.itemId}`;
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/users?sort=name&order=asc`;
    },
    getEventName() {
      return 'i-repositories';
    },
  },
};
</script>
