<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="isLoaded"
  >
    <v-text-field
      v-model="item.alias"
      label="Playbook Alias"
      :rules="[v => !!v || 'Playbook Alias is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.playbook"
      label="Playbook Name"
      :rules="[v => !!v || 'Playbook Name is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-select
      v-model="item.ssh_key_id"
      label="SSH Key"
      :items="keys"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'SSH Key is required']"
      required
      :disabled="formSaving"
    ></v-select>

    <v-select
      v-model="item.inventory_id"
      label="Inventory"
      :items="inventory"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'Inventory is required']"
      required
      :disabled="formSaving"
    ></v-select>

    <v-select
      v-model="item.repository_id"
      label="Playbook Repository"
      :items="repositories"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'Playbook Repository is required']"
      required
      :disabled="formSaving"
    ></v-select>

    <v-select
      v-model="item.environment_id"
      label="Environment"
      :items="environment"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'Environment is required']"
      required
      :disabled="formSaving"
    ></v-select>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      item: null,
      keys: null,
      inventory: null,
      repositories: null,
      environment: null,
    };
  },

  async created() {
    this.keys = (await axios({
      keys: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
    this.repositories = (await axios({
      keys: 'get',
      url: `/api/project/${this.projectId}/repositories`,
      responseType: 'json',
    })).data;
    this.inventory = (await axios({
      keys: 'get',
      url: `/api/project/${this.projectId}/inventory`,
      responseType: 'json',
    })).data;
    this.environment = (await axios({
      keys: 'get',
      url: `/api/project/${this.projectId}/environment`,
      responseType: 'json',
    })).data;
  },

  computed: {
    isLoaded() {
      if (this.isNew) {
        return true;
      }
      return this.keys && this.repositories && this.inventory && this.environment && this.item;
    },
  },

  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/templates`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/templates/${this.itemId}`;
    },
  },
};
</script>
