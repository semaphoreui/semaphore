<template>
  <EditDialog
    v-model="dialog"
    :save-button-text="$t(TEMPLATE_TYPE_ACTION_TITLES[templateType])"
    :title="$t('newTask')"
    @save="closeDialog"
    @close="closeDialog"
  >
    <template v-slot:title={}>
      <v-icon small class="mr-4">{{ TEMPLATE_TYPE_ICONS[templateType] }}</v-icon>
      <span class="breadcrumbs__item">{{ templateAlias }}</span>
      <v-icon>mdi-chevron-right</v-icon>
      <span class="breadcrumbs__item">{{ $t('newTask') }}</span>
    </template>

    <template v-slot:form="{ onSave, onError, needSave, needReset }">
      <TaskForm
        v-if="['terraform', 'tofu'].includes(templateApp)"
        :project-id="projectId"
        item-id="new"
        :template-id="templateId"
        @save="onSave"
        @error="onError"
        :need-save="needSave"
        :need-reset="needReset"
      />
      <TaskForm
        v-else
        :project-id="projectId"
        item-id="new"
        :template-id="templateId"
        @save="onSave"
        @error="onError"
        :need-save="needSave"
        :need-reset="needReset"
      />
    </template>
  </EditDialog>
</template>
<script>
import TaskForm from './TaskForm.vue';
import EditDialog from './EditDialog.vue';

import { TEMPLATE_TYPE_ACTION_TITLES, TEMPLATE_TYPE_ICONS } from '../lib/constants';
import EventBus from '../event-bus';

export default {
  components: {
    TaskForm,
    EditDialog,
  },
  props: {
    value: Boolean,
    projectId: Number,
    templateId: [Number, String],
    templateType: String,
    templateAlias: String,
    templateApp: String,
  },
  data() {
    return {
      dialog: false,
      TEMPLATE_TYPE_ACTION_TITLES,
      TEMPLATE_TYPE_ICONS,
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
    closeDialog(e) {
      this.dialog = false;
      if (e) {
        EventBus.$emit('i-show-task', {
          taskId: e.item.id,
        });
        this.$emit('save', e);
      }
      this.$emit('close');
    },
  },
};
</script>
