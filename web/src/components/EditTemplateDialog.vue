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
          v-if="['terraform', 'tofu'].includes(itemApp.id)"
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :source-item-id="sourceItemId"
          :app="itemApp.id"
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
    itemApp: Object,
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
    getAppColor(item) {
      if (APP_ICONS[item.id]) {
        return this.$vuetify.theme.dark ? APP_ICONS[item.id].darkColor : APP_ICONS[item.id].color;
      }

      return item.color || 'grey';
    },

    getAppTitle(item) {
      return APP_TITLE[item.id] || item.title;
    },

    getAppIcon(item) {
      return APP_ICONS[item.id] ? APP_ICONS[item.id].icon : `mdi-${item.icon}`;
    },

    onSave(e) {
      this.$emit('save', e);
    },
  },

};
</script>
