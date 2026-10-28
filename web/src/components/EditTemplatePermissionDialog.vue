<template>
  <EditDialog
    v-model="dialog"
    :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
    :title="$t('templatePermission', { expr: itemId === 'new' ? $t('nnew') : $t('edit') })"
    @save="onSave"
    test-id="templatePermDialog"
  >
    <template v-slot:form="{ onSave, onError, needSave, needReset }">
      <EditTemplatePermissionForm
        :project-id="projectId"
        :template-id="templateId"
        :item-id="itemId"
        scope="template"
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
    modelValue: Boolean,
    projectId: Number,
    templateId: [String, Number],
    itemId: [String, Number],
  },

  emits: ['update:modelValue', 'save'],

  data() {
    return {
      dialog: false,
    };
  },

  watch: {
    async dialog(val) {
      this.$emit('update:modelValue', val);
    },

    async modelValue(val) {
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
