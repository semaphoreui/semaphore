<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null"
    class="restore-project-form"
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
      @update:model-value="setFile"
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

<style lang="scss">
// An underlined field of Vuetify 2 is a 4 px top margin, 12 px for the label and the 32 px line;
// the 48 px field of Vuetify 4 holds the margin. Where that margin collapsed in Vuetify 2 the
// fields stood 4 px higher: under the 16 px margin of the error alert, and at the top of the
// page form, where it merged into the margin of `.project-settings-form` (the form moves, the
// button under its 300 px box stays).
.restore-project-form > .v-alert + .v-input {
  margin-top: -4px;
}

.project-settings-form .restore-project-form:not(:has(> .v-alert)) {
  position: relative;
  top: -4px;
}

// The file field (`v-file-input` of Vuetify 2): the paperclip 9 px from the line
// (`.v-input__prepend-outer` margin-right), an icon of the theme (.54 light, white dark) at rest,
// not the medium emphasis of the colour around (.6 × .6 in a dialog); the placeholder in the
// secondary text colour (`.v-file-input__text--placeholder`), not the disabled one.
.restore-project-form .v-file-input {
  > .v-input__prepend {
    margin-inline-end: 9px;
  }

  &:not(.v-input--focused):not(.v-input--error) > .v-input__prepend > .v-icon {
    opacity: 1;
    color: rgba(0, 0, 0, .54);
  }

  .v-field__input::placeholder {
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
    opacity: 1;
  }
}

.v-theme--dark .restore-project-form .v-file-input:not(.v-input--focused):not(.v-input--error)
  > .v-input__prepend > .v-icon {
  color: #fff;
}
</style>
