<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <UserDialog
      :user-id="itemId"
      v-model="editDialog"
      @saved="onItemSaved"
    />

    <YesNoDialog
      title="Delete user"
      text="Are you really want to delete this user?"
      yes-button-title="Yes"
      no-button-title="Cancel"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat color="white">
      <v-btn
        icon
        class="mr-4"
        @click="returnToProjects()"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>Users</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem()"
      >New User</v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
    >
      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-tooltip bottom>
            <template v-slot:activator="{ on, attrs }">
              <v-btn
                icon
                class="mr-1"
                v-bind="attrs"
                v-on="on"
                @click="askDeleteItem(item.id)"
              >
                <v-icon>mdi-delete</v-icon>
              </v-btn>
            </template>
            <span>Delete user</span>
          </v-tooltip>

          <v-tooltip bottom>
            <template v-slot:activator="{ on, attrs }">
              <v-btn
                icon
                class="mr-1"
                v-bind="attrs"
                v-on="on"
                @click="editItem(item.id)"
              >
                <v-icon>mdi-pencil</v-icon>
              </v-btn>
            </template>
            <span>Edit user</span>
          </v-tooltip>
        </div>
      </template>
    </v-data-table>
  </div>

</template>
<style lang="scss">

</style>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import UserDialog from '@/components/UserDialog.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import { getErrorMessage } from '@/lib/error';

export default {
  components: {
    YesNoDialog,
    UserDialog,
  },
  data() {
    return {
      headers: [
        {
          text: 'Name',
          value: 'name',
        },
        {
          text: 'Username',
          value: 'username',
        },
        {
          text: 'Email',
          value: 'email',
        },
        {
          text: 'Alert',
          value: 'alert',
        },
        {
          text: 'Admin',
          value: 'admin',
        },
        {
          text: 'External',
          value: 'external',
        },
        {
          text: 'Actions',
          value: 'actions',
          sortable: false,
        },
      ],
      items: null,
      itemId: null,
      editDialog: null,
      deleteItemDialog: null,
    };
  },

  async created() {
    await this.loadItems();
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    async onItemSaved() {
      await this.loadItems();
    },

    askDeleteItem(itemId) {
      this.itemId = itemId;
      this.deleteItemDialog = true;
    },

    async deleteItem(itemId) {
      try {
        const item = this.items.find((x) => x.id === itemId);

        await axios({
          method: 'delete',
          url: `/api/users/${itemId}`,
          responseType: 'json',
        });

        EventBus.$emit('i-user', {
          action: 'delete',
          item,
        });

        await this.loadItems();
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },

    editItem(itemId = 'new') {
      this.itemId = itemId;
      this.editDialog = true;
    },

    async loadItems() {
      this.items = (await axios({
        method: 'get',
        url: '/api/users',
        responseType: 'json',
      })).data;
    },
  },
};
</script>
