<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <ItemDialog
      v-model="editDialog"
      :save-button-text="(this.itemId === 'new' ? 'Link' : 'Save')"
      :title="(this.itemId === 'new' ? 'New' : 'Edit') + ' Team Member'"
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
    </ItemDialog>

    <YesNoDialog
      title="Delete team member"
      text="Are you really want to delete the team member?"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Team</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >New Team Member</v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
    >
      <template v-slot:item.admin="{ item }">

        <v-switch
          v-model="item.admin"
          inset
          @change="item.admin ? grantAdmin(item.id) : refuseAdmin(item.id)"
        ></v-switch>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn
          icon
          :disabled="!isUserAdmin()"
          @click="askDeleteItem(item.id)"
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

export default {
  components: { TeamMemberForm },
  mixins: [ItemListPageBase],
  methods: {
    async grantAdmin(userId) {
      await axios({
        method: 'post',
        url: `/api/project/${this.projectId}/users/${userId}/admin`,
        responseType: 'json',
      });
      await this.loadItems();
    },
    async refuseAdmin(userId) {
      await axios({
        method: 'delete',
        url: `/api/project/${this.projectId}/users/${userId}/admin`,
        responseType: 'json',
      });
      await this.loadItems();
    },
    getHeaders() {
      return [
        {
          text: 'Name',
          value: 'name',
          width: '50%',
        },
        {
          text: 'Username',
          value: 'username',
        },
        {
          text: 'Email',
          value: 'email',
          width: '50%',
        },
        {
          text: 'Admin',
          value: 'admin',
        },
        {
          text: 'Actions',
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
    isUserAdmin() {
      return (this.items.find((x) => x.id === this.userId) || {}).admin;
    },
  },
};
</script>
