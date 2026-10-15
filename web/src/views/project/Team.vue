<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null && roles != null">
    <EditTeamMemberDialog
      v-model="editDialog"
      :project-id="projectId"
      :item-id="itemId"
      :invites-enabled="systemInfo.teams.invites_enabled"
      :invite-type="systemInfo.teams.invite_type"
      :roles="roles"
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
        v-if="systemInfo.teams.members_can_leave"
        color="error"
        @click="leftProject()"
        class="mr-2"
        >{{ $t('LeaveProject') }}
      </v-btn>
      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectUsers)"
        >{{ $t('newTeamMember') }}
      </v-btn>
    </v-toolbar>

    <TeamMenu v-if="isPro" :project-id="projectId" :system-info="systemInfo" />

    <v-divider style="margin-top: -1px" />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto"
    >
      <template v-slot:item.roles="{ item }">
        <v-autocomplete
          v-if="can(USER_PERMISSIONS.manageProjectUsers)"
          v-model="item.roles"
          :items="roleItems(item)"
          item-value="slug"
          item-text="name"
          multiple
          small-chips
          deletable-chips
          hide-details
          dense
          class="pt-0 mt-0"
          @change="updateProjectUser(item)"
        >
          <template v-slot:item="{ item: role }">
            <v-list-item-content>
              <v-list-item-title>{{ role.name }}</v-list-item-title>
              <v-list-item-subtitle>{{ role.slug }}</v-list-item-subtitle>
            </v-list-item-content>
          </template>
        </v-autocomplete>
        <div v-else>
          <v-chip v-for="role in roleItems(item).filter((r) => item.roles.includes(r.slug))"
                  :key="role.slug" small class="mr-1">{{ role.name }}</v-chip>
        </div>
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
import { USER_PERMISSIONS } from '@/lib/constants';
import groupRolesForSelect from '@/lib/roles';
import TeamMenu from '@/components/TeamMenu.vue';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  components: { TeamMenu, EditTeamMemberDialog },
  mixins: [ItemListPageBase],

  props: {
    systemInfo: Object,
  },

  data() {
    return {
      roles: null,
    };
  },

  computed: {
    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },
  },

  async created() {
    this.roles = (
      await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/roles/all`,
        responseType: 'json',
      })
    ).data;
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
      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/me`,
          responseType: 'json',
        });
        window.location.reload();
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
    },

    roleItems(member) {
      // A role missing from the list (a custom role in a build without them) still shows its slug.
      const missing = member.roles
        .filter((slug) => !this.roles.some((role) => role.slug === slug))
        .map((slug) => ({ slug, name: slug }));
      return [...groupRolesForSelect(this.roles, (key) => this.$i18n.t(key)), ...missing];
    },

    async updateProjectUser(user) {
      // Removing a member is the delete button, the server refuses an empty role list.
      if (user.roles.length > 0) {
        try {
          await axios({
            method: 'put',
            url: `/api/project/${this.projectId}/users/${user.id}`,
            responseType: 'json',
            data: { roles: user.roles },
          });
        } catch (err) {
          EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
        }
      } else {
        EventBus.$emit('i-snackbar', { color: 'error', text: this.$i18n.t('memberNeedsRole') });
      }
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
          value: 'roles',
          sortable: false,
          width: '30%',
        },
        {
          value: 'actions',
          sortable: false,
          width: '0%',
        },
      ];
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
