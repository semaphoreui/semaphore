<template>
  <EditDialog
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
      <TerraformTemplateForm
          v-if="['terraform', 'tofu'].includes(itemApp)"
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :source-item-id="sourceItemId"
          :app="itemApp"
      />
      <TemplateForm
          v-else-if="['', 'ansible'].includes(itemApp)"
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :source-item-id="sourceItemId"
      />
      <ShellTemplateForm
          v-else
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :source-item-id="sourceItemId"
          :app-id="itemApp"
      />
    </template>
  </EditDialog>
</template>

<style scoped lang="scss">

</style>

<script>

import TerraformTemplateForm from './TerraformTemplateForm.vue';
import ShellTemplateForm from './ShellTemplateForm.vue';
import TemplateForm from './TemplateForm.vue';
import EditDialog from './EditDialog.vue';
import AppsMixin from './AppsMixin';

export default {
  components: {
    ShellTemplateForm,
    TerraformTemplateForm,
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
