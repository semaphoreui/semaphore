<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
    data-testid="restoreProject-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="restoreProject-error"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="projectName"
      :label="$t('projectName')"
      :rules="[v => !!v || $t('project_name_required')]"
      required
      :disabled="formSaving"
      data-testid="restoreProject-name"
    ></v-text-field>

    <v-file-input
      show-size
      truncate-length="15"
      :placeholder="$t('Backup file')"
      @change="setFile"
      data-testid="restoreProject-file"
    ></v-file-input>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  data() {
    return {
      projectName: null,
    };
  },
  methods: {

    beforeSave() {
      this.item.meta.name = this.projectName;
    },

    getSaveAction() {
      return 'restore';
    },

    /**
     * @param file {File}
     */
    async setFile(file) {
      if (file == null) {
        this.item = {};
        return;
      }
      this.item = JSON.parse(await file.text());
    },

    getItemsUrl() {
      return '/api/projects/restore';
    },
  },
};
</script>
