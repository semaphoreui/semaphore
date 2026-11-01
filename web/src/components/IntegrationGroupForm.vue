<template>
  <v-form ref="form" lazy-validation v-model="formValid" v-if="item != null && endpoints != null">
    <v-alert :value="formError" color="error" class="pb-2">{{ formError }} </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-autocomplete
      v-model="item.endpoint_id"
      :label="$t('endpoint')"
      :items="endpoints"
      item-value="id"
      item-text="name"
      :rules="[(v) => !!v || $t('endpointRequired')]"
      :disabled="formSaving"
      :no-data-text="$t('noEndpoints')"
      outlined
      dense
      class="IntegrationGroupForm__endpoint"
    ></v-autocomplete>

    <v-checkbox
      v-model="item.auth_override"
      :label="$t('overrideAuth')"
      :hint="$t('overrideAuthHint')"
      persistent-hint
      :disabled="formSaving"
      class="mt-0 mb-4 IntegrationGroupForm__override"
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
      // The API omits empty fields; add them reactively so the inputs can bind.
      ['auth_method', 'auth_header'].forEach((field) => {
        if (this.item[field] === undefined) {
          this.$set(this.item, field, '');
        }
      });
      if (this.item.auth_secret_id === undefined) {
        this.$set(this.item, 'auth_secret_id', null);
      }
      if (this.item.auth_override === undefined) {
        this.$set(this.item, 'auth_override', false);
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
