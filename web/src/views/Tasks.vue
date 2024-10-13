<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <v-toolbar flat>
      <v-btn
        icon
        class="mr-4"
        @click="returnToProjects()"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ $t('tasks') }}</v-toolbar-title>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.task_id="{ item }">
        <router-link
          :to="'/project/' + item.project_id + '?t=' + item.task_id"
        >
          #{{ item.task_id }}
        </router-link>
      </template>

      <template v-slot:item.project_id="{ item }">
        <router-link
          :to="'/project/' + item.project_id"
        >
          #{{ item.project_id }}
        </router-link>
      </template>

      <template v-slot:item.status="{item}">
        <div class="pr-4">
          <TaskStatus :status="item.status"/>
        </div>
      </template>

      <template v-slot:item.location="{item}">
        <div v-if="item.location === 'queue'">Queue</div>
        <div v-else-if="item.runner_id">
          Runner #{{ item.runner_id }}
        </div>
        <div v-else>Local Running</div>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn
            icon
            class="mr-1"
            @click="deleteItem(item.task_id)"
          >
            <v-icon>mdi-stop</v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import EventBus from '@/event-bus';
import ItemListPageBase from '@/components/ItemListPageBase';
import TaskStatus from '@/components/TaskStatus.vue';

export default {
  mixins: [ItemListPageBase],

  components: {
    TaskStatus,
  },

  props: {
  },

  computed: {
  },

  data() {
    return {
      newRunnerTokenDialog: null,
      newRunner: null,
      updateTimer: null,
    };
  },

  created() {
    this.updateTimer = setInterval(() => {
      this.loadItems();
    }, 10000);
  },

  destroyed() {
    clearInterval(this.updateTimer);
  },

  methods: {

    stopTask(taskId) {
      console.log(taskId);
    },

    getHeaders() {
      return [{
        text: this.$i18n.t('task', {}),
        value: 'task_id',
      }, {
        text: this.$i18n.t('project'),
        value: 'project_id',
      }, {
        text: this.$i18n.t('username'),
        value: 'username',
      }, {
        text: this.$i18n.t('status'),
        value: 'status',
      }, {
        text: this.$i18n.t('location'),
        value: 'location',
      }, {
        text: this.$i18n.t('actions'),
        value: 'actions',
        sortable: false,
        width: 70,
      }];
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    getItemsUrl() {
      return '/api/tasks';
    },

    getSingleItemUrl() {
      return `/api/tasks/${this.itemId}`;
    },

    getEventName() {
      return 'i-task';
    },
  },
};
</script>
