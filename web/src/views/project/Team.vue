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

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('team2') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >{{ $t('newTeamMember') }}</v-btn>
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
          :disabled="!isUserAdmin()"
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
          text: this.$i18n.t('name'),
          value: 'name',
          width: '50%',
        },
        {
          text: this.$i18n.t('username'),
          value: 'username',
        },
        {
          text: this.$i18n.t('email'),
          value: 'email',
          width: '50%',
        },
        {
          text: this.$i18n.t('admin'),
          value: 'admin',
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
    isUserAdmin() {
      return (this.items.find((x) => x.id === this.userId) || {}).admin;
    },
  },
};
</script>
