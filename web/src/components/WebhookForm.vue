<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
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
        v-model="item.template_id"
        label="Task Template to run"
        clearable
        :items="templates"
        item-value="id"
        item-text="name"
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
      templates: [],
    };
  },
  async created() {
    this.templates = (await axios({
      templates: 'get',
      url: `/api/project/${this.projectId}/templates`,
      responseType: 'json',
    })).data;
  },
  methods: {
    getNewItem() {
      return {
        template_id: {},
      };
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/webhooks`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/webhooks/${this.itemId}`;
    },
  },
};
</script>
