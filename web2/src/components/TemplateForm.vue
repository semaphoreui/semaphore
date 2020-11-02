<template>
  <v-form
    ref="itemForm"
    lazy-validation
    v-model="itemFormValid"
    v-if="isNewItem || isLoaded"
  >
    <v-text-field
      v-model="item.alias"
      label="Playbook Alias"
      :rules="[v => !!v || 'Playbook Alias is required']"
      required
      :disabled="itemFormSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.name"
      label="Playbook Name"
      :rules="[v => !!v || 'Playbook Name is required']"
      required
      :disabled="itemFormSaving"
    ></v-text-field>

    <v-select
      v-model="item.ssh_key_id"
      label="SSH Key"
      :items="keys"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'SSH Key is required']"
      required
      :disabled="itemFormSaving"
    ></v-select>

    <v-select
      v-model="item.inventory_id"
      label="Inventory"
      :items="inventory"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'Inventory is required']"
      required
      :disabled="itemFormSaving"
    ></v-select>

    <v-select
      v-model="item.repository_id"
      label="Playbook Repository"
      :items="repositories"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'Playbook Repository is required']"
      required
      :disabled="itemFormSaving"
    ></v-select>

    <v-select
      v-model="item.environment_id"
      label="Environment"
      :items="environment"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'Environment is required']"
      required
      :disabled="itemFormSaving"
    ></v-select>
  </v-form>
</template>
<script>
import axios from 'axios';

export default {
  props: {
    templateId: [Number, String],
    projectId: Number,
  },

  data() {
    return {
      item: null,
      keys: null,
      inventory: null,
      repositories: null,
      environment: null,
      itemFormValid: false,
      itemFormError: null,
      itemFormSaving: false,
    };
  },

  async created() {
    if (this.isNewItem) {
      this.item = {};
    }
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
    if (!this.isNewItem) {
      this.item = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates/${this.itemId}`,
        responseType: 'json',
      })).data;
    }
  },

  computed: {
    isLoaded() {
      if (this.isNewItem) {
        return true;
      }
      return this.keys && this.repositories && this.inventory && this.environment && this.item;
    },
    isNewItem() {
      return this.templateId === 'new';
    },
    itemId() {
      return this.templateId;
    },
  },

  methods: {
    async saveItem() {
      if (!this.$refs.itemForm.validate()) {
        return null;
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
      } finally {
        this.itemFormSaving = true;
      }

      return this.item;
    },
  },
};
</script>
