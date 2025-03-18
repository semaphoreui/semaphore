<template>
  <EditDialog
      v-if="isAppsLoaded"
      :max-width="1200"
      :min-content-height="457"
      v-model="dialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :icon="getAppIcon(itemApp)"
      :icon-color="getAppColor(itemApp)"
      :title="(itemId === 'new' ? $t('newTemplate') : $t('editTemplate')) +
        ' \'' + getAppTitle(itemApp) + '\''"
      @save="onSave"
      :content-class="`EditTemplateDialog EditTemplateDialog--${id}`"
  >
    <template v-slot:form="{ onSave, onError, needSave, needReset }">
      <TemplateForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :source-item-id="sourceItemId"
          :app="itemApp"
          @resize="onFormResize"
      />
    </template>
  </EditDialog>
</template>

<style lang="scss">
.EditTemplateDialog {
  width: auto;
  .v-card__text {
    overflow-x: auto;
  }
}

@media #{map-get($display-breakpoints, 'sm-and-down')} {
  .EditTemplateDialog {
    width: auto !important;
  }
}
</style>

<script>

import TemplateForm from './TemplateForm.vue';
import EditDialog from './EditDialog.vue';
import AppsMixin from './AppsMixin';

export default {
  components: {
    TemplateForm,
    EditDialog,
  },

  mixins: [AppsMixin],

  props: {
    value: Boolean,
    itemApp: String,
    projectId: Number,
    itemId: [String, Number],
    sourceItemId: Number,
  },

  data() {
    return {
      id: Math.round(Math.random() * 1000000),
      dialog: false,
    };
  },

  watch: {
    async dialog(val) {
      this.$emit('input', val);
    },

    async value(val) {
      this.dialog = val;
    },
  },

  methods: {
    onFormResize(e) {
      const contentEl = document.querySelector(`.EditTemplateDialog--${this.id}`);
      contentEl.style.width = `${e.width + 50}px`;
    },

    onSave(e) {
      this.$emit('save', e);
    },
  },

};
</script>
