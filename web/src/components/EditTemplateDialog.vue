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
          :fields="fields"
      />
    </template>
  </EditDialog>
</template>

<style scoped lang="scss">

</style>

<script>

import TemplateForm from './TemplateForm.vue';
import EditDialog from './EditDialog.vue';
import AppsMixin from './AppsMixin';

const ANSIBLE_FIELDS = {
  playbook: {
    label: 'playbookFilename',
  },
  inventory: {
    label: 'inventory2',
  },
  repository: {
    label: 'repository',
  },
  environment: {
    label: 'environment3',
  },
  vault: {
    label: 'vaultPassword2',
  },
};

const TERRAFORM_FIELDS = {
  ...ANSIBLE_FIELDS,
  playbook: {
    label: 'Subdirectory path (Optional)',
    optional: true,
  },
  inventory: {
    label: 'Default Workspace',
  },
  vault: undefined,
};

const UNKNOWN_APP_FIELDS = {
  ...ANSIBLE_FIELDS,
  playbook: {
    label: 'Script Filename *',
  },
  inventory: undefined,
  vault: undefined,
};

const APP_FIELDS = {
  '': ANSIBLE_FIELDS,
  ansible: ANSIBLE_FIELDS,
  terraform: TERRAFORM_FIELDS,
  tofu: TERRAFORM_FIELDS,
};

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

  computed: {
    fields() {
      return APP_FIELDS[this.itemApp] || UNKNOWN_APP_FIELDS;
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
