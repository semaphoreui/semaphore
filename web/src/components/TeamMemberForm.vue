<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="teamMembers != null && users != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}
    </v-alert>

    <div class="d-flex justify-end mb-1" v-if="inviteType === 'both'">
      <v-btn-toggle
        v-model="selectedInviteType"
        tile
        group
      >
        <v-btn value="email" small class="mr-0" style="border-radius: 4px;">
          Email
        </v-btn>
        <v-btn value="username" small class="mr-0" style="border-radius: 4px;">
          Username
        </v-btn>
      </v-btn-toggle>
    </div>

    <v-text-field
      v-if="selectedInviteType === 'email'"
      type="email"
      :label="$t('email')"
      v-model="item.email"
      outlined
      dense
    />

    <v-autocomplete
      v-else
      v-model="item.user_id"
      :label="$t('user')"
      :items="users"
      item-value="id"
      :item-text="(itm) => `${itm.username} (${itm.name})`"
      :rules="[v => !!v || $t('user_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-autocomplete>

    <v-select
      v-model="item.role"
      :label="$t('role')"
      :items="userRoles"
      item-value="slug"
      item-text="name"
      :rules="[v => !!v || $t('user_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-select>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import { USER_ROLES } from '@/lib/constants';

export default {
  mixins: [ItemFormBase],

  props: {
    invitesEnabled: Boolean,
    inviteType: String,
    roles: Array,
  },

  computed: {
    userRoles() {
      return [...USER_ROLES, ...(this.roles || [])];
    },
  },

  data() {
    return {
      users: null,
      userId: null,
      teamMembers: null,
      USER_ROLES,
      selectedInviteType: this.inviteType === 'both' ? 'username' : this.inviteType,
    };
  },

  async created() {
    this.teamMembers = (await axios({
      method: 'get',
      url: this.getItemsUrl(),
      responseType: 'json',
    })).data;

    this.users = (await axios({
      method: 'get',
      url: '/api/users',
      responseType: 'json',
    })).data.filter((user) => !this.teamMembers.some((teamMember) => user.id === teamMember.id));
  },

  methods: {
    getItemsUrl() {
      if (this.invitesEnabled) {
        return `/api/project/${this.projectId}/invites`;
      }
      return `/api/project/${this.projectId}/users`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/users/${this.itemId}`;
    },
  },
};
</script>
