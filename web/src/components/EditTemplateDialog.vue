<template>
  <EditDialog
      v-if="isAppsLoaded"
      :max-width="dialogWidth"
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
          :premium-features="premiumFeatures"
      />
    </template>
  </EditDialog>
</template>

<style lang="scss">
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
    premiumFeatures: Object,
  },

  data() {
    return {
      id: Math.round(Math.random() * 1000000),
      dialog: false,
    };
  },

  computed: {
    dialogWidth() {
      if (['ansible', 'terraform', 'tofu'].includes(this.itemApp)) {
        return 1200;
      }

      return 800;
    },
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
