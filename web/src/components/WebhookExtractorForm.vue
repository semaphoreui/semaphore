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
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import WebhookExtractorFormBase from '@/components/WebhookExtractorFormBase';

export default {
  mixins: [ItemFormBase, WebhookExtractorFormBase],
  data() {
    return {
      projectId: this.$route.params.projectId,
      webhookId: this.$route.params.webhookId,
    };
  },
  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/webhook/${this.webhookId}/extractors`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/webhook/${this.webhookId}/extractor/${this.itemId}`;
    },
  },
};
</script>
