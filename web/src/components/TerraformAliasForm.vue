<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
    data-testid="tfAlias-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
      data-testid="tfAlias-error"
    >{{ formError }}</v-alert>

    <v-select
      v-model="item.auth_key_id"
      :label="$t('Auth key')"
      :items="keys"
      item-value="id"
      item-text="name"
      outlined
      dense
      required
      :disabled="formSaving"
      data-testid="tfAlias-authKeyId"
    ></v-select>

  </v-form>
</template>
<style>
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],

  components: {
  },

  props: {
    inventoryId: Number,
  },

  data() {
    return {
      keys: null,
    };
  },

  async created() {
    this.keys = (await axios({
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data.filter((key) => key.type === 'login_password');
  },

  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/terraform/aliases`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/terraform/aliases/${this.itemId}`;
    },
  },
};
</script>
