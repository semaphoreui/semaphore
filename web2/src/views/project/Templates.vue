<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="users != null">
    <v-dialog
      v-model="deleteUserDialog"
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
            @click="deleteUserDialog = false"
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
      v-model="userDialog"
      max-width="290"
      persistent
    >
      <v-card>
        <v-card-title class="headline">{{ isNewUser ? 'New user' : 'Edit user' }}</v-card-title>

        <v-alert
          :value="userFormError"
          color="error"
        >
          {{ userFormError }}
        </v-alert>

        <v-card-text>
          <v-form
            ref="userForm"
            lazy-validation
            v-model="userFormValid"
          >
            <div style="display: none">
              <input type="password" tabindex="-1"/>
            </div>

            <v-text-field
              v-model="user.username"
              label="Username"
              :rules="[v => !!v || 'Username is required']"
              required
              :disabled="!isNewUser || userFormSaving"
            ></v-text-field>

            <v-text-field
              v-model="user.email"
              label="Email"
              :rules="[v => !!v || 'Email is required']"
              required
              :disabled="userFormSaving"
            ></v-text-field>

            <v-text-field
              v-model="user.fullName"
              label="Full Name"
              :rules="[v => !!v || 'Full Name is required']"
              required
              :disabled="userFormSaving"
            ></v-text-field>

            <v-select
              :items="userTypes"
              v-model="user.type"
              label="Type"
              :disabled="userFormSaving"
            ></v-select>

            <v-text-field
              v-model="user.password"
              label="Password"
              type="password"
              :rules="[v => !v || v.length === 0 || v.length >= 6 || 'Min 6 characters']"
              :disabled="userFormSaving"
            ></v-text-field>
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="blue darken-1"
            flat="flat"
            :disabled="userFormSaving"
            @click="userDialog = false"
          >
            Cancel
          </v-btn>
          <v-btn
            color="blue darken-1"
            flat="flat"
            :disabled="userFormSaving"
            @click="saveUser"
          >
            Save
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-toolbar flat color="white">
      <v-toolbar-title>Users</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn color="primary" @click="addUser">New User</v-btn>
    </v-toolbar>

    <v-divider></v-divider>

    <v-container>
      <v-data-table
        :headers="headers"
        :items="users"
        hide-actions
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
  data() {
    return {
      userTypes: ['admin', 'user'],
      headers: [
        {
          text: 'Username',
          value: 'username',
        },
        {
          text: 'Full Name',
          value: 'fullName',
          sortable: false,
        },
        {
          text: 'Email',
          value: 'email',
          sortable: false,
        },
        {
          text: 'Type',
          value: 'type',
          sortable: false,
        },
        {
          text: '',
          sortable: false,
        },
      ],
      users: null,

      user: {},
      isNewUser: false,

      userDialog: false,
      userFormValid: false,
      userFormError: null,
      userFormSaving: false,
      username: '',
      fullName: '',
      email: '',
      type: '',
      password: '',

      deleteUserDialog: false,
      deleteUsername: null,
    };
  },

  async created() {
    await this.loadUsers();
  },

  methods: {
    askDeleteUser(username) {
      this.deleteUsername = username;
      this.deleteUserDialog = true;
    },

    async deleteUser() {
      try {
        await axios({
          method: 'delete',
          url: `/api/users/${this.deleteUsername}`,
          responseType: 'json',
        });

        const userIndex = this.users.findIndex((user) => user.username === this.deleteUsername);
        if (userIndex !== -1) {
          this.users.splice(userIndex, 1);
        }

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: `User "${this.deleteUsername}" deleted`,
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.deleteUserDialog = false;
      }
    },

    async addUser() {
      await this.editUser();
    },

    async editUser(username) {
      this.isNewUser = !username;
      this.userFormError = null;

      this.user = username ? (await axios({
        method: 'get',
        url: `/api/templates/${username}`,
        responseType: 'json',
      })).data : { type: 'user' };

      this.$refs.userForm.resetValidation();
      this.userDialog = true;
    },

    async saveUser() {
      this.userFormError = null;

      if (!this.$refs.userForm.validate()) {
        return;
      }

      this.userFormSaving = true;
      try {
        await axios({
          method: this.isNewUser ? 'post' : 'put',
          url: this.isNewUser ? '/api/templates' : `/api/templates/${this.user.id}`,
          responseType: 'json',
          data: this.user,
        });

        if (this.isNewUser) {
          this.users.push(this.user);
        } else {
          const userIndex = this.users.findIndex((user) => this.user.id === user.id);
          if (userIndex !== -1) {
            this.users.splice(userIndex, 1, this.user);
          }
        }

        this.userDialog = false;

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.isNewUser ? `User "${this.user.username}" created` : `User "${this.user.username}" changed`,
        });
      } catch (err) {
        this.userFormError = getErrorMessage(err);
      } finally {
        this.userFormSaving = false;
      }
    },

    async loadUsers() {
      this.users = (await axios({
        method: 'get',
        url: '/api/templates',
        responseType: 'json',
      })).data;
    },
  },
};
</script>
