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
      label="Project Name"
      :rules="[v => !!v || 'Project name is required']"
      required
      :disabled="formSaving || !item.admin"
    ></v-text-field>

    <v-checkbox
      v-model="item.alert"
      label="Allow alerts for this project"
      :disabled="formSaving || !item.admin"
    ></v-checkbox>

    <v-text-field
      v-model="item.alert_chat"
      label="Chat ID"
      :disabled="formSaving || !item.admin"
    ></v-text-field>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  methods: {
    getItemsUrl() {
      return '/api/projects';
    },
    getSingleItemUrl() {
      return `/api/project/${this.itemId}`;
    },
  },
};
</script>
