<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="tasks != null">
    <EditDialog
        v-model="newTaskDialog"
        :save-button-text="$t('re', {getActionButtonTitle: getActionButtonTitle()})"
        @save="onTaskCreated"
    >
      <template v-slot:title={}>
        <v-icon class="mr-4">{{ TEMPLATE_TYPE_ICONS[template.type] }}</v-icon>
        <span class="breadcrumbs__item">{{ template.name }}</span>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ $t('newTask') }}</span>
      </template>

      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TerraformTaskForm
            v-if="['terraform', 'tofu'].includes(template.app)"
            :project-id="template.project_id"
            item-id="new"
            :template-id="template.id"
            @save="onSave"
            @error="onError"
            :need-save="needSave"
            :need-reset="needReset"
            :source-task="sourceTask"
        />
        <TaskForm
          v-else
          :project-id="template.project_id"
          item-id="new"
          :template-id="template.id"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :source-task="sourceTask"
        />
      </template>
    </EditDialog>

    <v-data-table
        :headers="headers"
        :items="tasks"
        :hide-default-footer="hideFooter"
        :footer-props="{ itemsPerPageOptions: [20] }"
        class="mt-0"
    >
      <template v-slot:item.id="{ item }">
        <TaskLink
            :task-id="item.id"
            :tooltip="item.message"
            :label="'#' + item.id"
        />
      </template>

      <template v-slot:item.version="{ item }">
        <div v-if="item.tpl_type !== ''">
          <TaskLink
              :disabled="item.tpl_type === 'build'"
              :task-id="item.build_task_id"
              :tooltip="item.tpl_type === 'build' ? item.message : (item.build_task || {}).message"
              :label="item.tpl_type === 'build' ? item.version : (item.build_task || {}).version"
              :status="item.status"
          />
        </div>
        <div v-else>&mdash;</div>
      </template>

      <template v-slot:item.status="{ item }">
        <TaskStatus :status="item.status"/>
      </template>

      <template v-slot:item.start="{ item }">
        {{ item.start | formatDate }}
      </template>

      <template v-slot:item.end="{ item }">
        {{ [item.start, item.end] | formatMilliseconds }}
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn text class="pl-1 pr-2" @click="createTask(item)">
          <v-icon class="pr-1">mdi-replay</v-icon>
          Re{{ getActionButtonTitle() }}
        </v-btn>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import TaskForm from '@/components/TaskForm.vue';
import TaskStatus from '@/components/TaskStatus.vue';
import TaskLink from '@/components/TaskLink.vue';
import EditDialog from '@/components/EditDialog.vue';
import { TEMPLATE_TYPE_ACTION_TITLES, TEMPLATE_TYPE_ICONS } from '@/lib/constants';
import TerraformTaskForm from '@/components/TerraformTaskForm.vue';

export default {
  components: {
    TerraformTaskForm,
    EditDialog,
    TaskStatus,
    TaskForm,
    TaskLink,
  },
  props: {
    template: Object,
    limit: Number,
    hideFooter: Boolean,
  },
  data() {
    return {
      headers: [
        {
          text: this.$i18n.t('taskId'),
          value: 'id',
          sortable: false,
        },
        {
          text: this.$i18n.t('version'),
          value: 'version',
          sortable: false,
        },
        {
          text: this.$i18n.t('status'),
          value: 'status',
          sortable: false,
        },
        {
          text: this.$i18n.t('user'),
          value: 'user_name',
          sortable: false,
        },
        {
          text: this.$i18n.t('start'),
          value: 'start',
          sortable: false,
        },
        {
          text: this.$i18n.t('duration'),
          value: 'end',
          sortable: false,
        },
        {
          text: this.$i18n.t('actions'),
          value: 'actions',
          sortable: false,
          width: '0%',
        },
      ],
      tasks: null,
      taskId: null,
      newTaskDialog: null,
      sourceTask: null,
      TEMPLATE_TYPE_ICONS,
    };
  },
  watch: {
    async template() {
      await this.loadData();
    },
  },
  async created() {
    await this.loadData();
  },
  methods: {
    async loadData() {
      this.tasks = null;
      this.tasks = (await axios({
        method: 'get',
        url: `/api/project/${this.template.project_id}/templates/${this.template.id}/tasks/last?limit=${this.limit || 200}`,
        responseType: 'json',
      })).data;
    },
    getActionButtonTitle() {
      return this.$i18n.t(TEMPLATE_TYPE_ACTION_TITLES[this.template.type]);
    },

    onTaskCreated(e) {
      EventBus.$emit('i-show-task', {
        taskId: e.item.id,
      });
    },

    createTask(task) {
      this.sourceTask = task;
      this.newTaskDialog = true;
    },
  },
};
</script>
