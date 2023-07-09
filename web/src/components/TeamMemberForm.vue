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

    <v-select
      v-model="item.user_id"
      :label="$t('user')"
      :items="users"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || $t('user_required')]"
      required
      :disabled="formSaving"
    ></v-select>

    <v-checkbox
      v-model="item.admin"
      :label="$t('administrator')"
    ></v-checkbox>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      users: null,
      userId: null,
      teamMembers: null,
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
