<template>
  <EditDialog
      :max-width="700"
      :min-content-height="457"
      v-model="dialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :icon="APP_ICONS[itemApp].icon"
      :icon-color="$vuetify.theme.dark ? APP_ICONS[itemApp].darkColor : APP_ICONS[itemApp].color"
      :title="(itemId === 'new' ? $t('newTemplate') : $t('editTemplate')) +
        ' \'' + APP_TITLE[itemApp] + '\''"
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
      />
      <BashTemplateForm
          v-else-if="itemApp === 'bash'"
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :source-item-id="sourceItemId"
      />
      <TemplateForm
          v-else
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :source-item-id="sourceItemId"
      />
    </template>
  </EditDialog>
</template>

<style scoped lang="scss">

</style>

<script>

import { APP_ICONS, APP_TITLE } from '../lib/constants';
import TerraformTemplateForm from './TerraformTemplateForm.vue';
import BashTemplateForm from './BashTemplateForm.vue';
import TemplateForm from './TemplateForm.vue';
import EditDialog from './EditDialog.vue';

export default {
  components: {
    BashTemplateForm,
    TerraformTemplateForm,
    TemplateForm,
    EditDialog,
  },

  props: {
    value: Boolean,
    itemApp: String,
    projectId: Number,
    itemId: [String, Number],
    sourceItemId: Number,
  },

  data() {
    return {
      APP_TITLE,
      APP_ICONS,
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
