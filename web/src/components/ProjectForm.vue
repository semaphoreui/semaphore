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
      :label="$t(projectNameTitle)"
      :rules="[v => !!v || $t('project_name_required')]"
      required
      :disabled="formSaving"
      data-testid="newProject-name"
      outlined
      dense
    ></v-text-field>

    <v-textarea
      v-model="item.description"
      label="Project Description"
      :disabled="formSaving"
      data-testid="newProject-description"
      outlined
      rows="3"
      auto-grow
      counter="500"
      hint="Optional description for your project"
      persistent-hint
    ></v-textarea>

    <v-switch
      class="mt-0"
      v-model="item.alert"
      :label="$t('allowAlertsForThisProject')"
      data-testid="newProject-alert"
      inset
    ></v-switch>

    <v-switch
      v-if="itemId === 'new'"
      v-model="item.demo"
      label="Demo"
      style="position: absolute; left: 24px; bottom: 15px;"
      hide-details
    />

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  props: {
    projectNameTitle: {
      type: String,
      default: 'projectName',
    },
  },
  created() {
    // Set default values for new projects
    if (this.itemId === 'new' && this.item) {
      this.item.alert = true;
    }
  },
  methods: {
    getItemsUrl() {
      return '/api/projects';
    },
    getSingleItemUrl() {
      return `/api/project/${this.itemId}`;
    },
    beforeSave() {
      // Set alerts enabled by default for new projects
      if (this.itemId === 'new') {
        this.item.alert = true;
      }
    },
  },
};
</script>
