<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null && keys != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.inventory"
      label="Workspace name"
      :rules="[v => !!v || $t('path_required')]"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-select
      v-model="item.ssh_key_id"
      label="SSH key for private modules *"
      :rules="[v => !!v || 'Key is required']"
      dense
      required
      :items="sshKeys"
      item-value="id"
      item-text="name"
      :disabled="formSaving"
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

  props: {
    namePrefix: String,
    type: String,
    templateId: Number,
  },

  components: {
  },

  data() {
    return {
      keys: null,
    };
  },

  computed: {
    // A string key is a bare value with no login or private key, so it can not
    // authenticate a user; the picker offers only the types the executor accepts.
    sshKeys() {
      if (this.keys == null) {
        return null;
      }
      return this.keys.filter((key) => ['ssh', 'login_password', 'none'].includes(key.type));
    },
  },

  async created() {
    this.keys = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },

  methods: {
    async beforeSave() {
      this.item.name = this.namePrefix + this.item.inventory;
      this.item.type = this.type;
      this.item.template_id = this.templateId;
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.itemId}`;
    },
  },
};
</script>
