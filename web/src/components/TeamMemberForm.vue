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
    >{{ formError }}</v-alert>

    <v-autocomplete
      v-model="item.user_id"
      :label="$t('user')"
      :items="users"
      item-value="id"
      :item-text="(itm) => `${itm.username} (${itm.name})`"
      :rules="[v => !!v || $t('user_required')]"
      required
      :disabled="formSaving"
    ></v-autocomplete>

    <v-select
      v-model="item.role"
      :label="$t('role')"
      :items="USER_ROLES"
      item-value="slug"
      item-text="title"
      :rules="[v => !!v || $t('user_required')]"
      required
      :disabled="formSaving"
    ></v-select>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import { USER_ROLES } from '@/lib/constants';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      users: null,
      userId: null,
      teamMembers: null,
      USER_ROLES,
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
      return `/api/project/${this.projectId}/users`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/users/${this.itemId}`;
    },
  },
};
</script>
