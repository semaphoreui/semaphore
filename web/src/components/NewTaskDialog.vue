<template>
  <EditDialog
    v-model="dialog"
    :save-button-text="$t(TEMPLATE_TYPE_ACTION_TITLES[template?.type || ''])"
    :title="$t('newTask')"
    @save="closeDialog"
    @close="closeDialog"
    :test-id="testId"
  >
    <template v-slot:title={}>
      <v-icon
        size="small"
        class="mr-4"
        data-testid="newTask-templateTypeIcon"
      >{{ TEMPLATE_TYPE_ICONS[template?.type || ''] }}</v-icon>
      <span class="breadcrumbs__item" data-testid="newTask-templateName">{{ templateTitle }}</span>
      <v-icon>mdi-chevron-right</v-icon>
      <span class="breadcrumbs__item" data-testid="newTask-title">{{ $t('newTask') }}</span>
    </template>

    <template v-slot:form="{ onSave, onError, needSave, needReset }">
      <TaskForm
        :project-id="projectId"
        item-id="new"
        :template="template"
        @save="onSave"
        @error="onError"
        :need-save="needSave"
        :need-reset="needReset"
        :source-task="sourceTask"
      />
    </template>
  </EditDialog>
</template>
<script>
import { TEMPLATE_TYPE_ACTION_TITLES, TEMPLATE_TYPE_ICONS } from '@/lib/constants';
import TaskForm from './TaskForm.vue';
import EditDialog from './EditDialog.vue';

import EventBus from '../event-bus';

export default {
  components: {
    TaskForm,
    EditDialog,
  },
  props: {
    modelValue: Boolean,
    projectId: Number,
    template: Object,
    sourceTask: Object,
    testId: {
      type: String,
      default: 'newTaskDialog',
    },
  },
  emits: ['update:modelValue', 'save', 'close'],
  data() {
    return {
      dialog: false,
      TEMPLATE_TYPE_ACTION_TITLES,
      TEMPLATE_TYPE_ICONS,
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

  computed: {
    templateTitle() {
      let res = this.template?.name || '';
      if (res.length > 16) {
        res = `${res.substring(0, 14)}...`;
      }

      return res;
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
