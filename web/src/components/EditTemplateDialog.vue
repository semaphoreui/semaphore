<template>
  <EditDialog
      v-if="isAppsLoaded"
      :max-width="700"
      :min-content-height="457"
      v-model="dialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :icon="getAppIcon(itemApp)"
      :icon-color="getAppColor(itemApp)"
      :title="(itemId === 'new' ? $t('newTemplate') : $t('editTemplate')) +
        ' \'' + getAppTitle(itemApp) + '\''"
      @save="onSave"
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
      />
    </template>
  </EditDialog>
</template>
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
    onSave(e) {
      this.$emit('save', e);
    },
  },

};
</script>
