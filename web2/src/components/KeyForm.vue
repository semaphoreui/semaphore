<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      label="Key Name"
      :rules="[v => !!v || 'Name is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

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

    <v-textarea
      v-model="item.key"
      label="Public Key"
      :disabled="formSaving"
      v-if="item.type === 'ssh'"
    ></v-textarea>

    <v-textarea
      v-model="item.secret"
      label="Private Key"
      :disabled="formSaving"
      v-if="item.type === 'ssh'"
    ></v-textarea>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  data() {
    return {
      inventoryTypes: [{
        id: 'ssh',
        name: 'SSH Key',
      }, {
        id: 'none',
        name: 'None',
      }],
    };
  },
  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/keys`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/keys/${this.itemId}`;
    },
  },
};
</script>
