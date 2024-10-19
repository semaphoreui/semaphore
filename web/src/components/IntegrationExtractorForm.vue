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
import IntegrationExtractorFormBase from '@/components/IntegrationExtractorFormBase';

export default {
  mixins: [ItemFormBase, IntegrationExtractorFormBase],
  data() {
    return {
      projectId: this.$route.params.projectId,
      integrationId: this.$route.params.integrationId,
    };
  },
  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/integrations`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.integrationId}`;
    },
  },
};
</script>
