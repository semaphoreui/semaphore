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
  <v-row>
    <v-col cols="12" md="12" class="pb-0">
      <div class="ml-4 mr-4 mt-6">
        <v-select v-model="item.value_source"
                  label="Source of the Value"
                  :items="valueSources"
                  item-value="id"
                  item-text="text"
                  :rules="[v => !!v || 'Value Source is required']"
                  outlined
                  dense
                  required
                  :disabled="formSaving">
        </v-select>
        <v-select v-model="item.body_data_type"
                  label="Data Type of Body"
                  v-if="item.value_source == 'body'"
                  :items="bodyDataTypes"
                  item-value="id"
                  item-text="text"
                  :rules="[v => !!v || 'Body Data Type is required']"
                  outlined
                  dense
                  required
                  :disabled="formSaving">
        </v-select>
        <v-text-field
          v-model="item.key"
          label="Key *"
          :rules="[v => !!v || 'Key is required']"
          outlined
          dense
          required
          :disabled="formSaving"
          ></v-text-field>
        <v-text-field
          v-model="item.variable"
          label="Variable *"
          :rules="[v => !!v || 'Variable is required']"
          outlined
          dense
          required
          :disabled="formSaving"
          ></v-text-field>
      </div>

    </v-col>
  </v-row>
</v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import IntegrationExtractorChildValueFormBase from './IntegrationExtractorChildValueFormBase';
import { EXTRACT_VALUE_TYPE_ICONS, EXTRACT_VALUE_TYPE_TITLES } from '../lib/constants';

export default {
  mixins: [ItemFormBase, IntegrationExtractorChildValueFormBase],

  data() {
    return {
      EXTRACT_VALUE_TYPE_ICONS,
      EXTRACT_VALUE_TYPE_TITLES,
      valueSources: [{
        id: 'body',
        text: 'Body',
      }, {
        id: 'header',
        text: 'Header',
      }],
      bodyDataTypes: [{
        id: 'json',
        text: 'JSON',
      }, {
        id: 'string',
        text: 'String',
      }],
    };
  },
  computed: {
    projectId() {
      if (/^-?\d+$/.test(this.$route.params.projectId)) {
        return parseInt(this.$route.params.projectId, 10);
      }
      return this.$route.params.projectId;
    },
    integrationId() {
      if (/^-?\d+$/.test(this.$route.params.integrationId)) {
        return parseInt(this.$route.params.integrationId, 10);
      }
      return this.$route.params.integrationId;
    },
  },
  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/values`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/values/${this.itemId}`;
    },
  },
};
</script>
