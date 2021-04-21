<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="isLoaded"
  >
      <v-row>
        <v-col cols="12" md="6" class="pb-0">
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
        </v-col>

        <v-col cols="12" md="6" class="pb-0">
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

    <v-textarea
      outlined
      class="mt-4"
      v-model="item.arguments"
      label="Extra CLI Arguments"
      :disabled="formSaving"
      placeholder='*MUST* be a JSON array!
Example: ["-i", "@myinventory.sh", "--private-key=/there/id_rsa", "-vvvv"]'
      rows="4"
    ></v-textarea>

        </v-col>
      </v-row>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],

  props: {
    sourceItemId: String,
  },

  data() {
    return {
      item: null,
      keys: null,
      inventory: null,
      repositories: null,
      environment: null,
    };
  },

  watch: {
    needReset(val) {
      if (val) {
        this.item.template_id = this.templateId;
      }
    },

    sourceItemId(val) {
      this.item.template_id = val;
    },
  },

  async created() {
    if (this.sourceItemId) {
      this.item = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/templates/${this.sourceItemId}`,
        responseType: 'json',
      })).data;
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
  },

  computed: {
    isLoaded() {
      if (this.isNew && this.sourceItemId == null) {
        return true;
      }

      return this.keys != null
        && this.repositories != null
        && this.inventory != null
        && this.environment != null
        && this.item != null;
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
