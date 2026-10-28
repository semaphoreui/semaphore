<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div
    v-if="items != null"
    data-testid="activeTasks-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <v-toolbar flat>
      <v-btn
        icon
        class="mr-4"
        @click="returnToProjects()"
        data-testid="activeTasks-back"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title data-testid="activeTasks-title">{{ $t('activeTasks') }}</v-toolbar-title>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :items-per-page-options="[20]"
      data-testid="activeTasks-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.task_id="{ item }">
        <router-link
          :to="'/project/' + item.project_id + '?t=' + item.task_id"
          data-testid="activeTasks-rowTask"
        >
          #{{ item.task_id }}
        </router-link>
      </template>

      <template v-slot:item.project_id="{ item }">
        <router-link
          :to="'/project/' + item.project_id"
          data-testid="activeTasks-rowProject"
        >
          #{{ item.project_id }}
        </router-link>
      </template>

      <template v-slot:item.status="{item}">
        <div class="pr-4">
          <TaskStatus :status="item.status" data-testid="activeTasks-rowStatus"/>
        </div>
      </template>

      <template v-slot:item.location="{item}">
        <div
          v-if="item.location === 'queue'"
          data-testid="activeTasks-rowLocation-queue"
        >Queue</div>
        <div
          v-else-if="item.runner_id"
          data-testid="activeTasks-rowLocation-runner"
        >
          Runner #{{ item.runner_id }}
        </div>
        <div v-else data-testid="activeTasks-rowLocation-local">Local Running</div>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn
            icon
            class="mr-1"
            @click="deleteItem(item.task_id)"
            data-testid="activeTasks-rowStop"
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

  unmounted() {
    clearInterval(this.updateTimer);
  },

  methods: {

    getHeaders() {
      return [{
        title: this.$t('task', {}),
        key: 'task_id',
      }, {
        title: this.$t('project'),
        key: 'project_id',
      }, {
        title: this.$t('username'),
        key: 'username',
      }, {
        title: this.$t('status'),
        key: 'status',
      }, {
        title: this.$t('location'),
        key: 'location',
      }, {
        title: this.$t('actions'),
        key: 'actions',
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
