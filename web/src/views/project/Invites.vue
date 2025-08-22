<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditTeamMemberDialog
      v-model="editDialog"
      :project-id="projectId"
      :item-id="itemId"
      :invites-enabled="systemInfo.teams.invites_enabled"
      :invite-type="systemInfo.teams.invite_type"
      @save="loadItems()"
    />

    <YesNoDialog
      :title="$t('deleteTeamMember')"
      :text="$t('askDeleteTMem')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('Invites') }}</v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectUsers)"
      >{{ $t('newTeamMember') }}
      </v-btn>
    </v-toolbar>

    <v-tabs class="pl-4" v-if="systemInfo.teams.invites_enabled">
      <v-tab
        key="team"
        :to="`/project/${projectId}/team`"
        data-testid="team-members"
      >
        Members
      </v-tab>

      <v-tab
        key="invites"
        :to="`/project/${projectId}/invites`"
        data-testid="team-invites"
      >
        Invites
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;"/>

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto;"
    >
      <template v-slot:item.name="{ item }">
        {{ item.user ? item.user.name : item.email }}
      </template>

      <template v-slot:item.role="{ item }">
        {{ USER_ROLES.find(r => r.slug === item.role).title }}
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
import axios from 'axios';
import { USER_PERMISSIONS, USER_ROLES } from '@/lib/constants';
import EditTeamMemberDialog from '@/components/EditTeamMemberDialog.vue';

export default {
  components: { EditTeamMemberDialog },
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
