<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item"
    data-testid="integrationValueForm-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <!-- mb-3: in Vuetify 2 the 16 px under the alert took in the 4 px top margin of the Name
         field; in Vuetify 4 those 4 px sit inside the field and add up. -->
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2 mb-3"
      data-testid="integrationValueForm-error"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      data-testid="integrationValueForm-name"
      label="Name"
      :rules="[(v) => !!v || 'Name is required']"
      required
      :disabled="formSaving"
    ></v-text-field>
    <!-- mb-n3: the bottom −12 px of a Vuetify 2 row under a `pb-0` column (the grid of Vuetify 4
         has no negative margins). -->
    <v-row class="mb-n3">
      <v-col cols="12" md="12" class="pb-0">
        <div class="ml-4 mr-4 mt-6">
          <v-select
            v-model="item.value_source"
            data-testid="integrationValueForm-valueSource"
            label="Source of the Value"
            :items="valueSources"
            item-value="id"
            item-title="text"
            :rules="[(v) => !!v || 'Value Source is required']"
            variant="outlined"
            density="compact"
            required
            :disabled="formSaving"
          >
          </v-select>
          <v-select
            v-model="item.body_data_type"
            data-testid="integrationValueForm-bodyDataType"
            label="Data Type of Body"
            v-if="item.value_source == 'body'"
            :items="bodyDataTypes"
            item-value="id"
            item-title="text"
            :rules="[(v) => !!v || 'Body Data Type is required']"
            variant="outlined"
            density="compact"
            required
            :disabled="formSaving"
          >
          </v-select>
          <v-text-field
            v-model="item.key"
            data-testid="integrationValueForm-key"
            label="Key *"
            :rules="[(v) => !!v || 'Key is required']"
            variant="outlined"
            density="compact"
            required
            :disabled="formSaving"
          >
          </v-text-field>
          <v-select
            v-model="item.variable_type"
            data-testid="integrationValueForm-variableType"
            label="Variable Usage *"
            :items="variableTypes"
            item-value="id"
            item-title="text"
            :rules="[(v) => !!v || 'Variable Type is required']"
            variant="outlined"
            density="compact"
            required
            :disabled="formSaving"
          >
          </v-select>
          <v-text-field
            v-model="item.variable"
            data-testid="integrationValueForm-variable"
            label="Variable *"
            :rules="[(v) => !!v || 'Variable is required']"
            variant="outlined"
            density="compact"
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
  props: {
    integrationId: Number,
  },
  data() {
    return {
      EXTRACT_VALUE_TYPE_ICONS,
      EXTRACT_VALUE_TYPE_TITLES,
      valueSources: [
        {
          id: 'body',
          text: 'Body',
        },
        {
          id: 'header',
          text: 'Header',
        },
      ],
      bodyDataTypes: [
        {
          id: 'json',
          text: 'JSON',
        },
        {
          id: 'string',
          text: 'String',
        },
      ],
      variableTypes: [
        {
          id: 'environment',
          text: 'Variables',
        },
        {
          id: 'task',
          text: 'Task Params',
        },
      ],
    };
  },
  computed: {
    // integrationId() {
    //   if (/^-?\d+$/.test(this.$route.params.integrationId)) {
    //     return parseInt(this.$route.params.integrationId, 10);
    //   }
    //   return this.$route.params.integrationId;
    // },
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
