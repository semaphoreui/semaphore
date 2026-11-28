<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null && endpoints != null"
    data-testid="integrationGroupForm-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="integrationGroupForm-error"
    >{{ formError }} </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      data-testid="integrationGroupForm-name"
    ></v-text-field>

    <!-- The list 1 px under the field and at most 304 px high, as in the integration form. -->
    <v-autocomplete
      v-model="item.endpoint_id"
      :label="$t('endpoint')"
      :items="endpoints"
      item-value="id"
      item-title="name"
      :rules="[(v) => !!v || $t('endpointRequired')]"
      :disabled="formSaving"
      :no-data-text="$t('noEndpoints')"
      variant="outlined"
      density="compact"
      class="IntegrationGroupForm__endpoint"
      :menu-props="{ maxHeight: 304, offset: 1 }"
      data-testid="integrationGroupForm-endpointId"
    ></v-autocomplete>

    <v-checkbox
      v-model="item.auth_override"
      :label="$t('overrideAuth')"
      :hint="$t('overrideAuthHint')"
      persistent-hint
      :disabled="formSaving"
      class="mt-0 mb-6 IntegrationGroupForm__override"
      data-testid="integrationGroupForm-authOverride"
    ></v-checkbox>

    <IntegrationAuthForm
      v-if="item.auth_override"
      v-model="item"
      :project-id="projectId"
      :disabled="formSaving"
    />
  </v-form>
</template>
<script>
import axios from 'axios';
import ItemFormBase from '@/components/ItemFormBase';
import IntegrationAuthForm from '@/components/IntegrationAuthForm.vue';

export default {
  components: { IntegrationAuthForm },
  mixins: [ItemFormBase],

  data() {
    return {
      endpoints: null,
    };
  },

  async created() {
    this.endpoints = (
      await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/endpoints`,
        responseType: 'json',
      })
    ).data || [];
  },

  methods: {
    getNewItem() {
      return {
        name: '',
        endpoint_id: null,
        auth_override: false,
        auth_method: '',
        auth_header: '',
        auth_secret_id: null,
      };
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/integration_groups`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integration_groups/${this.itemId}`;
    },

    afterLoadData() {
      // The API omits empty fields; add them so the inputs can bind.
      ['auth_method', 'auth_header'].forEach((field) => {
        if (this.item[field] === undefined) {
          this.item[field] = '';
        }
      });
      if (this.item.auth_secret_id === undefined) {
        this.item.auth_secret_id = null;
      }
      if (this.item.auth_override === undefined) {
        this.item.auth_override = false;
      }
    },

    beforeSave() {
      // Inherited auth: the server ignores the group's own fields, send them empty.
      if (!this.item.auth_override) {
        this.item.auth_method = '';
        this.item.auth_header = '';
        this.item.auth_secret_id = null;
      }
    },
  },
};
</script>
