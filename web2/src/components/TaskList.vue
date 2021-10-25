<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <v-data-table
      :headers="headers"
      :items="tasks"
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
      <div v-if="item.version != null || item.build_task != null">
        <TaskLink
            :disabled="item.tpl_type === 'build'"
            :task-id="item.build_task_id"
            :tooltip="item.build_task ? item.build_task.message : null"
            :label="item.tpl_type === 'build' ? item.version : item.build_task.version"
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
      <v-btn text color="black" class="pl-1 pr-2" @click="createTask(item)">
        <v-icon class="pr-1">mdi-replay</v-icon>
        Re{{ getActionButtonTitle() }}
      </v-btn>
    </template>
  </v-data-table>
</template>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';

export default {
  data() {
    return {
      headers: [
        {
          text: 'Task ID',
          value: 'id',
          sortable: false,
        },
        {
          text: 'Version',
          value: 'version',
          sortable: false,
        },
        {
          text: 'Status',
          value: 'status',
          sortable: false,
        },
        {
          text: 'User',
          value: 'user_name',
          sortable: false,
        },
        {
          text: 'Start',
          value: 'start',
          sortable: false,
        },
        {
          text: 'Duration',
          value: 'end',
          sortable: false,
        },
        {
          text: 'Actions',
          value: 'actions',
          sortable: false,
          width: '0%',
        },
      ],
      tasks: null,
      taskId: null,
      newTaskDialog: null,
      sourceTask: null,
    };
  },
  async created() {
    this.tasks = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/templates/${this.itemId}/tasks/last`,
      responseType: 'json',
    })).data;

  },
  methods: {

    onTaskCreated(e) {
      EventBus.$emit('i-show-task', {
        taskId: e.item.id,
      });
    },

    createTask(task) {
      this.sourceTask = task;
      this.newTaskDialog = true;
    },

  }
};
</script>
