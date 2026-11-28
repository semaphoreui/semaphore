<template>
  <div
    v-if="items != null"
    data-testid="invites-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditTeamMemberDialog
      v-model="editDialog"
      :project-id="projectId"
      :item-id="itemId"
      :invites-enabled="systemInfo.teams.invites_enabled"
      :invite-type="systemInfo.teams.invite_type"
      :roles="USER_ROLES"
      @save="loadItems()"
    />

    <YesNoDialog
      test-id="invitesDeleteDialog"
      :title="$t('deleteTeamMember')"
      :text="$t('askDeleteTMem')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="invites-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title class="flex-0-1" data-testid="invites-title">
        {{ $t('Invites') }}
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn
        color="primary"
        variant="elevated"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectUsers)"
        data-testid="invites-new"
      >{{ $t('newTeamMember') }}
      </v-btn>
    </v-toolbar>

    <TeamMenu :project-id="projectId" :system-info="systemInfo" />

    <v-divider style="margin-top: -1px;"/>

    <v-data-table
      data-testid="invites-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4 InvitesTable"
      :items-per-page="-1"
    >
      <template v-slot:item.name="{ item }">
        {{ item.user ? item.user.name : item.email }}
      </template>

      <template v-slot:item.role="{ item }">
        {{ USER_ROLES.find(r => r.slug === item.role).name }}
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-group
          density="compact"
          data-testid="invites-rowActions"
        >
          <v-btn
            @click="askDeleteItem(item.id)"
            v-if="can(USER_PERMISSIONS.manageProjectUsers)"
            data-testid="invites-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </v-btn-group>
      </template>
    </v-data-table>
  </div>

</template>
<style lang="scss">
// Not an inline style: the utility `mt-4` of Vuetify 4 has no `!important` and would lose to it.
.InvitesTable {
  max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px);
  margin: auto;
}
</style>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import axios from 'axios';
import { USER_PERMISSIONS, USER_ROLES } from '@/lib/constants';
import EditTeamMemberDialog from '@/components/EditTeamMemberDialog.vue';
import TeamMenu from '@/components/TeamMenu.vue';

export default {
  components: { TeamMenu, EditTeamMemberDialog },
  mixins: [ItemListPageBase],

  props: {
    systemInfo: Object,
  },

  data() {
    return {
      USER_ROLES,
    };
  },

  methods: {
    async updateProjectInvite(invite) {
      await axios({
        method: 'put',
        url: `/api/project/${this.projectId}/invites/${invite.id}`,
        responseType: 'json',
        data: invite,
      });
      await this.loadItems();
    },

    allowActions() {
      return this.can(USER_PERMISSIONS.manageProjectUsers);
    },

    getHeaders() {
      return [
        {
          title: this.$t('name'),
          key: 'name',
          width: '40%',
        },
        {
          title: this.$t('username'),
          key: 'username',
          width: '30%',
        },
        {
          title: this.$t('role'),
          key: 'role',
          width: '30%',
        },
        {
          key: 'actions',
          sortable: false,
          width: '0%',
        }];
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/invites/${this.itemId}`;
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/invites`;
    },
    getEventName() {
      return 'i-repositories';
    },
  },
};
</script>
