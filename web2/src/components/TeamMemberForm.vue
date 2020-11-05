<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="users != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}</v-alert>

    <v-select
      v-model="item.user_id"
      label="User"
      :items="users"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'User is required']"
      required
      :disabled="formSaving"
    ></v-select>

    <v-checkbox
      v-model="item.admin"
      label="Administrator"
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
    };
  },

  async created() {
    this.users = (await axios({
      method: 'get',
      url: '/api/users',
      responseType: 'json',
    })).data;
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
