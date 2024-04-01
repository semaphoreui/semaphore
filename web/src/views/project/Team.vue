<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="(this.itemId === 'new' ? 'Link' : $t('save'))"
      :title="$t('teamMember', {expr: this.itemId === 'new' ? $t('nnew') : $t('edit')})"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TeamMemberForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

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

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
    >
      <template v-slot:item.role="{ item }">
        <v-select
          v-model="item.role"
          :items="USER_ROLES"
          item-value="slug"
          item-text="title"
          :style="{width: '200px'}"
          @change="updateProjectUser(item)"
          v-if="can(USER_PERMISSIONS.manageProjectUsers)"
        />
        <div v-else>{{ USER_ROLES.find(r => r.slug === item.role).title }}</div>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn
          icon
          @click="askDeleteItem(item.id)"
          v-if="can(USER_PERMISSIONS.manageProjectUsers)"
        >
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </template>
    </v-data-table>
  </div>

</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import TeamMemberForm from '@/components/TeamMemberForm.vue';
import axios from 'axios';
import { USER_PERMISSIONS, USER_ROLES } from '@/lib/constants';

export default {
  components: { TeamMemberForm },
  mixins: [ItemListPageBase],
  data() {
    return {
      USER_ROLES,
    };
  },

  methods: {
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
          width: '50%',
        },
        {
          text: this.$i18n.t('username'),
          value: 'username',
        },
        {
          text: this.$i18n.t('role'),
          value: 'role',
        },
        {
          text: this.$i18n.t('actions'),
          value: 'actions',
          sortable: false,
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
