<template>
  <div
    v-if="items != null"
    data-testid="activeTasks-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <v-toolbar flat>
      <v-btn
        icon
        variant="text"
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
      class="mt-4 FixedPageSizeTable ActiveTasksTable"
      :items-per-page-options="[20]"
      :items-per-page="20"
      :show-first-last-page="false"
      :page-text="items.length === 0 ? '–' : undefined"
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
            variant="text"
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
<style lang="scss">
// The table of active tasks as Vuetify 2 drew it.
.ActiveTasksTable {
  // The line over the footer: the top border of `.v-data-footer` in Vuetify 2 — as wide as the
  // footer (16 px in from both sides, App.vue) and in the border colour (.12), not a divider over
  // the whole table at .12 of the text colour.
  > .v-divider {
    margin: 0 16px;
    border-color: rgba(var(--v-border-color), var(--v-border-opacity));
    opacity: 1;
  }

  // The footer of Vuetify 2 (`.v-data-footer`): 8 px padding, the range 24 px from the left and
  // 32 px before the arrows (no rows-per-page select pushes them right), 36 px round arrow buttons
  // 14 px apart, in the icon colour of the theme (.54) when enabled.
  > .v-data-table-footer {
    justify-content: flex-start;
    padding: 0 8px;

    > .v-data-table-footer__info {
      min-width: 0;
      margin: 0 32px 0 24px;
      padding: 0;
    }

    .v-pagination__list > li {
      margin: 0;

      &.v-pagination__prev {
        margin-inline-end: 7px;
      }

      &.v-pagination__next {
        margin-inline-start: 7px;
      }
    }

    .v-pagination .v-btn {
      width: 36px;
      height: 36px;
    }

    .v-pagination .v-btn.v-theme--light:not(.v-btn--disabled) {
      color: rgba(0, 0, 0, .54);
      opacity: 1;
    }
  }
}
</style>
