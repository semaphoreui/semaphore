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
        <v-select
          v-model="item.match_type"
          label="Match on *"
          :items="matchTypes"
          item-value="id"
          item-text="text"
          :rules="[v => !!v || 'Match source is required']"
          outlined
          dense
          required
          :disabled="formSaving"
        ></v-select>
        <v-select
          v-model="item.body_data_type"
          v-if="item.match_type == 'body'"
          label="Body Data Format *"
          :items="bodyDataFormats"
          item-value="id"
          item-text="text"
          :rules="[v => !!v || 'Body Data Format is required']"
          outlined
          dense
          required
          :disabled="formSaving"
        ></v-select>
        <v-text-field
          v-model="item.key"
          label="Key *"
          :rules="[v => !!v || 'Key is required']"
          outlined
          dense
          required
          :disabled="formSaving"
        ></v-text-field>
        <v-select
          v-model="item.method"
          label="Comparison Method *"
          :items="methods"
          item-value="id"
          item-text="text"
          :rules="[v => !!v || 'Comparison Method is required']"
          outlined
          dense
          required
          :disabled="formSaving"
        ></v-select>
        <v-text-field
          v-model="item.value"
          label="Value *"
          :rules="[v => !!v || 'Value is required']"
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
import { MATCHER_TYPE_ICONS, MATCHER_TYPE_TITLES } from '../lib/constants';

export default {
  mixins: [ItemFormBase, IntegrationExtractorChildValueFormBase],
  data() {
    return {
      MATCHER_TYPE_ICONS,
      MATCHER_TYPE_TITLES,
      matchTypes: [{
        id: 'body',
        text: 'Body',
      }, {
        id: 'header',
        text: 'Header',
      }],
      bodyDataFormats: [{
        id: 'json',
        text: 'JSON',
      }, {
        id: 'string',
        text: 'String',
      }],
      methods: [{
        id: 'equals',
        text: '==',
      }, {
        id: 'unequals',
        text: '!=',
      }, {
        id: 'contains',
        text: 'Contains',
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
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/matchers`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/matchers/${this.itemId}`;
    },
  },
};
</script>
