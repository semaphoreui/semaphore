<template>
  <EditDialog
    v-model="dialog"
    :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
    :title="$t('templatePermission', { expr: itemId === 'new' ? $t('nnew') : $t('edit') })"
    @save="onSave"
  >
    <template v-slot:form="{ onSave, onError, needSave, needReset }">
      <EditTemplatePermissionForm
        :project-id="projectId"
        :template-id="templateId"
        :item-id="itemId"
        @save="onSave"
        @error="onError"
        :need-save="needSave"
        :need-reset="needReset"
      />
    </template>
  </EditDialog>
</template>

<script>
import EditDialog from './EditDialog.vue';
import EditTemplatePermissionForm from './EditTemplatePermissionForm.vue';

export default {
  components: {
    EditDialog,
    EditTemplatePermissionForm,
  },

  props: {
    value: Boolean,
    projectId: Number,
    templateId: [String, Number],
    itemId: [String, Number],
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
