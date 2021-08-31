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
      v-model="item.name"
      label="Name"
      :rules="[v => !!v || 'Name is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-select
      v-model="item.ssh_key_id"
      label="Access Key"
      :items="keys"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'Access Key is required']"
      required
      :disabled="formSaving"
    ></v-select>

    <v-select
      v-model="item.type"
      label="Type"
      :rules="[v => !!v || 'Type is required']"
      :items="inventoryTypes"
      item-value="id"
      item-text="name"
      required
      :disabled="formSaving"
    ></v-select>

    <v-text-field
      v-model="item.inventory"
      label="Path to Inventory file"
      :rules="[v => !!v || 'Path to Inventory file is required']"
      required
      :disabled="formSaving"
      v-if="item.type === 'file'"
    ></v-text-field>

    <v-textarea
      v-model="item.inventory"
      label="Inventory"
      :rules="[v => !!v || 'Inventory is required']"
      required
      :disabled="formSaving"
      v-if="item.type === 'static'"
      solo
    ></v-textarea>
    <div v-if="item.type === 'static'">
      Static inventory example:
      <pre>
    [website]
    172.18.8.40
    172.18.8.41
      </pre>
    </div>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],
  data() {
    return {
      keys: null,
      inventoryTypes: [{
        id: 'static',
        name: 'Static',
      }, {
        id: 'file',
        name: 'File',
      }],
    };
  },
  async created() {
    this.keys = (await axios({
      keys: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },
  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.itemId}`;
    },
  },
};
</script>
