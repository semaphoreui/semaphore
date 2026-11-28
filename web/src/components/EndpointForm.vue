<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null"
    data-testid="endpointForm-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="endpointForm-error"
    >{{ formError }} </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      data-testid="endpointForm-name"
    ></v-text-field>

    <v-text-field
      v-model="item.description"
      :label="$t('description')"
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      data-testid="endpointForm-description"
    ></v-text-field>

    <IntegrationAuthForm v-model="item" :project-id="projectId" :disabled="formSaving" />
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import IntegrationAuthForm from '@/components/IntegrationAuthForm.vue';

export default {
  components: { IntegrationAuthForm },
  mixins: [ItemFormBase],

  methods: {
    getNewItem() {
      return {
        name: '',
        description: '',
        auth_method: '',
        auth_header: '',
        auth_secret_id: null,
      };
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/endpoints`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/endpoints/${this.itemId}`;
    },

    afterLoadData() {
      // The API omits empty fields; add them so the inputs can bind.
      ['auth_method', 'auth_header', 'description'].forEach((field) => {
        if (this.item[field] === undefined) {
          this.item[field] = '';
        }
      });
      if (this.item.auth_secret_id === undefined) {
        this.item.auth_secret_id = null;
      }
    },
  },
};
</script>
