<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <v-dialog
      v-model="deleteItemDialog"
      max-width="290">
      <v-card>
        <v-card-title class="headline">Delete template</v-card-title>

        <v-card-text>
          Are you really want to delete this template?
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>

          <v-btn
            color="blue darken-1"
            flat="flat"
            @click="deleteItemDialog = false"
          >
            Cancel
          </v-btn>

          <v-btn
            color="blue darken-1"
            flat="flat"
            @click="deleteUser()"
          >
            Yes
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog
      v-model="itemDialog"
      max-width="290"
      persistent
    >
      <v-card>
        <v-card-title class="headline">
          {{ isNewItem ? 'New Template' : 'Edit Template' }}
        </v-card-title>

        <v-alert
          :value="itemFormError"
          color="error"
        >
          {{ itemFormError }}
        </v-alert>

        <v-card-text>
          <v-form
            ref="itemForm"
            lazy-validation
            v-model="itemFormValid"
          >
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="blue darken-1"
            text
            :disabled="itemFormSaving"
            @click="itemDialog = false"
          >
            Cancel
          </v-btn>
          <v-btn
            color="blue darken-1"
            text
            :disabled="itemFormSaving"
            @click="saveUser"
          >
            Save
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-toolbar flat color="white">
      <v-toolbar-title>Task Templates</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn color="primary" @click="addUser">New template</v-btn>
    </v-toolbar>

    <v-divider></v-divider>

    <v-container>
      <v-data-table
        :headers="headers"
        :items="items"
        hide-default-footer
      >
        <template v-slot:items="props">
          <td>{{ props.item.username }}</td>
          <td>{{ props.item.fullName }}</td>
          <td>{{ props.item.email }}</td>
          <td>{{ props.item.type }}</td>
          <td style="width: 80px;" class="pa-2 text-xs-right">
            <v-icon
              small
              class="mr-2 pa-1"
              @click="editUser(props.item.username)"
            >
              edit
            </v-icon>
            <v-icon
              small
              class="pa-1"
              @click="askDeleteUser(props.item.username)"
            >
              delete
            </v-icon>
          </td>
        </template>
      </v-data-table>
    </v-container>
  </div>

</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    projectId: Number,
  },
  data() {
    return {
      headers: [
        {
          text: 'Alias',
          value: 'alias',
        },
        {
          text: 'Playbook',
          value: 'playbook',
          sortable: false,
        },
        {
          text: 'SSH key',
          value: 'email',
          sortable: false,
        },
        {
          text: 'Inventory',
          value: 'inventory',
          sortable: false,
        },
        {
          text: 'Environment',
          value: 'environment',
          sortable: false,
        },
        {
          text: 'Repository',
          value: 'repository',
          sortable: false,
        },
      ],
      items: null,

      item: {},
      isNewItem: false,

      itemDialog: false,
      itemFormValid: false,
      itemFormError: null,
      itemFormSaving: false,
      username: '',
      fullName: '',
      email: '',
      type: '',
      password: '',

      deleteItemDialog: false,
      deleteItemId: null,
    };
  },

  async created() {
    await this.loadUsers();
  },

  methods: {
    askDeleteUser(username) {
      this.deleteItemId = username;
      this.deleteItemDialog = true;
    },

    async deleteUser() {
      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/${this.deleteItemId}`,
          responseType: 'json',
        });

        const userIndex = this.items.findIndex((item) => item.username === this.deleteItemId);
        if (userIndex !== -1) {
          this.items.splice(userIndex, 1);
        }

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: `User "${this.deleteItemId}" deleted`,
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.deleteItemDialog = false;
      }
    },

    async addUser() {
      await this.editUser();
    },

    async editUser(username) {
      this.isNewItem = !username;
      this.itemFormError = null;

      this.item = username ? (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates/${username}`,
        responseType: 'json',
      })).data : { type: 'item' };
      this.itemDialog = true;
    },

    async saveUser() {
      this.itemFormError = null;

      if (!this.$refs.itemForm.validate()) {
        return;
      }

      this.itemFormSaving = true;
      try {
        await axios({
          method: this.isNewItem ? 'post' : 'put',
          url: this.isNewItem
            ? `/api/project/${this.projectId}/templates`
            : `/api/project/${this.projectId}/templates/${this.item.id}`,
          responseType: 'json',
          data: this.item,
        });

        if (this.isNewItem) {
          this.items.push(this.item);
        } else {
          const userIndex = this.items.findIndex((item) => this.item.id === item.id);
          if (userIndex !== -1) {
            this.items.splice(userIndex, 1, this.item);
          }
        }

        this.itemDialog = false;

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.isNewItem ? `User "${this.item.username}" created` : `User "${this.item.username}" changed`,
        });
      } catch (err) {
        this.itemFormError = getErrorMessage(err);
      } finally {
        this.itemFormSaving = false;
      }
    },

    async loadUsers() {
      this.items = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
