<template>
  <EditDialog
    v-model="dialog"
    :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
    :title="title"
    @save="onSave"
    test-id="templatePermDialog"
  >
    <template v-slot:form="{ onSave, onError, needSave, needReset }">
      <EditTemplatePermissionForm
        :project-id="projectId"
        :object-id="objectId"
        :item-id="itemId"
        :scope="scope"
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
    scope: String,
    objectId: [String, Number],
    itemId: [String, Number],
  },

  emits: ['update:modelValue', 'save'],

  data() {
    return {
      dialog: false,
    };
  },

  computed: {
    title() {
      const key = this.scope === 'workflow' ? 'workflowPermission' : 'templatePermission';
      return this.$t(key, { expr: this.itemId === 'new' ? this.$t('nnew') : this.$t('edit') });
    },
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
