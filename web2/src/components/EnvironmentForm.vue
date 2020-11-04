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
      label="Environment Name"
      :rules="[v => !!v || 'Name is required']"
      required
      :disabled="formSaving"
      class="mb-4"
    ></v-text-field>

    <v-textarea
      v-model="item.json"
      label="Environment (This has to be a JSON object)"
      :disabled="formSaving"
      solo
    ></v-textarea>

    <div>
      Must be valid JSON. You may use the key ENV to pass a json object which sets environmental
      variables for the ansible command execution environment
    </div>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/environment`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/environment/${this.itemId}`;
    },
  },
};
</script>
