<template>
  <div
    style="overflow: hidden"
    class="pb-5"
    data-testid="taskSummary-panel"
    :data-state="hosts == null ? 'loading' : (hosts.length === 0 ? 'empty' : 'ready')"
  >
    <v-alert
      variant="tonal"
      color="hsl(348deg, 86%, 61%)"
      class="AnsibleStageView__demo PageAlert"
      v-if="!features.task_summary"
      data-testid="taskSummary-demoAlert"
    >
      <span class="mr-2">
        This is <b>DEMO</b> data. Task summary available only in <b>PRO</b> version.
      </span>

      <v-btn
        class="ml-2"
        color="hsl(348deg, 86%, 61%)"
        @click="upgradeToPro('task_summary')"
        data-testid="taskSummary-upgrade"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>
    </v-alert>

    <div class="pl-5 pt-5 d-flex" style="column-gap: 10px">
      <div class="AnsibleServerStatus AnsibleServerStatus--ok">
        <div
          class="AnsibleServerStatus__count"
          data-testid="taskSummary-okServers"
        >{{ okServers }}</div>
        <div class="AnsibleServerStatus__title">OK SERVERS</div>
      </div>

      <div class="AnsibleServerStatus AnsibleServerStatus--bad">
        <div
          class="AnsibleServerStatus__count"
          data-testid="taskSummary-notOkServers"
        >{{ notOkServers }}</div>
        <div class="AnsibleServerStatus__title">NOT OK SERVERS</div>
      </div>
    </div>

    <v-btn-toggle
      class="AnsibleStageView__filter pl-5 mt-8 mb-3"
      density="compact"
      v-model="tab"
      mandatory
      data-testid="taskSummary-filter"
    >
      <v-btn
        value="notOkServers"
        data-testid="taskSummary-filter-notOkServers"
      > Not ok servers </v-btn>
      <v-btn value="allServers" data-testid="taskSummary-filter-allServers"> All servers </v-btn>
    </v-btn-toggle>

    <v-data-table
      v-if="tab === 'notOkServers'"
      hide-default-footer
      show-expand
      v-model:expanded="expanded"
      expand-strategy="single"
      :headers="notOkServersHeaders"
      :items="failedTasks"
      :items-per-page="-1"
      class="AnsibleStageView__errors w-100"
      data-testid="taskSummary-errorsTable"
      :data-state="hosts == null ? 'loading' : (failedTasks.length === 0 ? 'empty' : 'ready')"
    >
      <template v-slot:item.error="{ item }">
        <div
          style="overflow: hidden; color: #ff5252; max-width: 400px; text-overflow: ellipsis"
          data-testid="taskSummary-errorText"
        >
          {{ item.error }}
        </div>
      </template>
      <!-- the slot draws the whole row: Vuetify 3 does not wrap it in a <tr> -->
      <template v-slot:expanded-row="{ columns, item }">
        <tr class="AnsibleStageView__expandedRow">
          <td :colspan="columns.length">
            <pre
              style="
                overflow: auto;
                background: gray;
                font-size: 14px;
                color: white;
                border-radius: 10px;
                white-space: pre-wrap;
                margin-top: 5px;
                margin-bottom: 5px;
              "
              class="pa-2"
              data-testid="taskSummary-errorDetails"
              >{{ item.error.trim() }}</pre
            >
          </td>
        </tr>
      </template>
    </v-data-table>

    <v-table
      v-else-if="tab === 'allServers'"
      data-testid="taskSummary-hostsTable"
      :data-state="hosts == null ? 'loading' : (hosts.length === 0 ? 'empty' : 'ready')"
    >
      <template v-slot:default>
        <thead>
          <tr>
            <th>Host</th>
            <th>Changed</th>
            <th>Failed</th>
            <th>Ignored</th>
            <th>Ok</th>
            <th>Rescued</th>
            <th>Skipped</th>
            <th>Unreachable</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(host, index) in hosts" :key="index">
            <td data-testid="taskSummary-hostName">{{ host.host }}</td>

            <td
              :style="{
                color: host.changed > 0 ? 'rgb(170,85,0)' : undefined,
                'font-weight': host.changed > 0 ? 'bold' : undefined,
              }"
              data-testid="taskSummary-hostChanged"
            >
              {{ host.changed }}
            </td>

            <td
              :style="{
                color: host.failed > 0 ? 'red' : undefined,
                'font-weight': host.failed > 0 ? 'bold' : undefined,
              }"
              data-testid="taskSummary-hostFailed"
            >
              {{ host.failed }}
            </td>

            <td
              :style="{
                color: host.ignored > 0 ? 'red' : undefined,
                'font-weight': host.ignored > 0 ? 'bold' : undefined,
              }"
              data-testid="taskSummary-hostIgnored"
            >
              {{ host.ignored }}
            </td>

            <td
              :style="{
                color: host.ok > 0 ? 'green' : undefined,
                'font-weight': host.ok > 0 ? 'bold' : undefined,
              }"
              data-testid="taskSummary-hostOk"
            >
              {{ host.ok }}
            </td>

            <td
              :style="{
                'font-weight': host.rescued > 0 ? 'bold' : undefined,
              }"
              data-testid="taskSummary-hostRescued"
            >
              {{ host.rescued }}
            </td>

            <td
              :style="{
                color: host.skipped > 0 ? 'rgb(0,170,170)' : undefined,
                'font-weight': host.skipped > 0 ? 'bold' : undefined,
              }"
              data-testid="taskSummary-hostSkipped"
            >
              {{ host.skipped }}
            </td>

            <td
              :style="{
                color: host.unreachable > 0 ? 'red' : undefined,
                'font-weight': host.unreachable > 0 ? 'bold' : undefined,
              }"
              data-testid="taskSummary-hostUnreachable"
            >
              {{ host.unreachable }}
            </td>
          </tr>
        </tbody>
      </template>
    </v-table>
  </div>
</template>
<style lang="scss">
// The active filter button under the .18 overlay of Vuetify 2 (`$btn-states` activated), not the
// .12 of Vuetify 4.
.AnsibleStageView__filter > .v-btn--active > .v-btn__overlay {
  opacity: calc(.18 * var(--v-theme-overlay-multiplier));
}

// The shadow of the Upgrade button is not cut by the alert content (Vuetify 4 clips
// `.v-alert__content`, Vuetify 2 did not).
.AnsibleStageView__demo > .v-alert__content {
  overflow: visible;
}

// One thin line between the buttons: every button but the first has no left border
// (`.v-btn-toggle > .v-btn.v-btn:not(:first-child)`).
.AnsibleStageView__filter.v-btn-group > .v-btn:not(:first-child):not(.v-btn--variant-text) {
  border-inline-start-width: 0;
}

// The expand column of Vuetify 2: a 24 px icon (`.v-data-table__expand-icon`), 16 px to the next
// column; Vuetify 4 has a 28 px button and 8 px. A focused icon (the one just clicked) has the
// circle of `.v-icon--link:focus::after`: the icon colour at .12, 1.3 times the icon.
.AnsibleStageView__errors thead th:first-child {
  padding-right: 16px;
}

.AnsibleStageView__errors .v-data-table__td--expanded-row {
  padding-right: 16px;

  > .v-btn {
    width: 24px;
    height: 24px;
    // the circle is wider than the icon
    overflow: visible;

    &:focus > .v-btn__overlay {
      border-radius: 50%;
      opacity: .12;
      transform: scale(1.3);
    }
  }
}

// Rows as cards on a phone: a flex row of Vuetify 2 (`.v-data-table__mobile-row`: space-between,
// the header 16 px from the value), not the two equal grid columns of Vuetify 4; the expand icon at
// the right end. The value takes the width of its content, so a long error widens the table as it
// did.
.AnsibleStageView__errors .v-data-table__tr--mobile > td {
  display: flex;
  justify-content: space-between;
  column-gap: 0;

  > .v-data-table__td-title {
    padding-right: 16px;
  }

  &.v-data-table__td--expanded-row {
    justify-content: flex-end;
  }
}

// "No data available" on a phone at the left edge, as the rows (see History.vue).
@media (max-width: 599.98px) {
  .AnsibleStageView__errors .v-data-table-rows-no-data {
    text-align: start;
  }
}

// The opened error on a phone is as wide as its text (Vuetify 2: the rows are blocks of a flex
// column, `.v-data-table--mobile tbody`, so the lone cell of this row shrinks to its content).
.AnsibleStageView__errors .v-data-table__tr--mobile + .AnsibleStageView__expandedRow {
  display: block;
}

.AnsibleServerStatus {
  text-align: center;
  width: 250px;
  font-weight: bold;
  color: white;
  font-size: 24px;
  line-height: 1.2;
  border-radius: 8px;
}

.AnsibleServerStatus__count {
  padding-top: 10px;
  font-size: 80px;
  line-height: 1;
}

.AnsibleServerStatus--ok {
  background-color: #4caf50;
}

.AnsibleServerStatus--bad {
  background-color: #ff5252;
}

.AnsibleServerStatus__title {
  padding-bottom: 10px;
}
</style>

<script>
import ProjectMixin from '@/components/ProjectMixin';

export default {
  props: {
    projectId: Number,
    taskId: Number,
    features: Object,
  },

  mixins: [ProjectMixin],

  data() {
    return {
      stages: null,
      okServers: 0,
      notOkServers: 0,
      tab: 'notOkServers',
      failedTasks: [],
      hosts: null,
      // opened rows of the errors table (expand-strategy="single": one at a time)
      expanded: [],
      notOkServersHeaders: [
        {
          key: 'data-table-expand',
          // as narrow as its content: the 24 px icon with 2 px before it (App.vue) and 16 px
          // after (Vuetify 2: `width: 1px`; Vuetify 4: 48 px)
          width: 1,
        },
        {
          title: 'Server',
          key: 'host',
          sortable: false,
        },
        {
          title: 'Task',
          key: 'task',
          sortable: false,
        },
        {
          title: 'Error',
          key: 'error',
          sortable: false,
        },
      ],
    };
  },

  watch: {
    // the errors table is drawn anew on every switch of the filter: no row comes back opened
    tab() {
      this.expanded = [];
    },

    async taskId() {
      await this.loadData();
      this.calcStats();
    },
  },

  async created() {
    await this.loadData();
    this.calcStats();
  },

  methods: {
    async loadData() {
      if (this.features.task_summary) {
        [this.failedTasks, this.hosts, this.stages] = await Promise.all([
          this.loadProjectEndpoint(`/tasks/${this.taskId}/ansible/errors`),
          this.loadProjectEndpoint(`/tasks/${this.taskId}/ansible/hosts`),
          this.loadProjectEndpoint(`/tasks/${this.taskId}/stages`),
        ]);
      } else {
        [this.failedTasks, this.hosts, this.stages] = this.getDemoData();
      }

      this.hosts.forEach((host) => {
        if (host.unreachable) {
          this.failedTasks.push({
            host: host.host,
            task: '—',
            error: 'Host is unreachable',
          });
        }
      });
    },

    calcStats() {
      this.hosts.forEach((host) => {
        if (host.failed > 0 || host.unreachable > 0) {
          this.notOkServers += 1;
        } else {
          this.okServers += 1;
        }
      });
    },

    getDemoData() {
      const failedTasks = [
        {
          id: 1,
          host: 'web-01.prod.example.com',
          task: 'Install nginx package',
          error: 'fatal: [web-01.prod.example.com]: FAILED! => {"changed": false, "msg": "No package matching \'nginx\' is available"}',
        },
        {
          id: 2,
          host: 'web-02.prod.example.com',
          task: 'Start nginx service',
          error: 'fatal: [web-02.prod.example.com]: FAILED! => {"changed": false, "msg": "Could not find the requested service nginx: host"}',
        },
        {
          id: 3,
          host: 'db-01.prod.example.com',
          task: 'Apply database migrations',
          error: 'fatal: [db-01.prod.example.com]: FAILED! => {"changed": false, "msg": "Migration failed: relation \\"users\\" already exists", "rc": 1}',
        },
        {
          id: 4,
          host: 'cache-01.prod.example.com',
          task: 'Configure redis maxmemory',
          error: 'fatal: [cache-01.prod.example.com]: FAILED! => {"changed": false, "msg": "Destination /etc/redis/redis.conf does not exist!"}',
        },
      ];

      const hosts = [
        {
          host: 'web-01.prod.example.com', changed: 3, failed: 1, ignored: 0, ok: 12, rescued: 0, skipped: 2, unreachable: 0,
        },
        {
          host: 'web-02.prod.example.com', changed: 2, failed: 1, ignored: 0, ok: 11, rescued: 0, skipped: 2, unreachable: 0,
        },
        {
          host: 'web-03.prod.example.com', changed: 4, failed: 0, ignored: 0, ok: 15, rescued: 0, skipped: 1, unreachable: 0,
        },
        {
          host: 'db-01.prod.example.com', changed: 1, failed: 1, ignored: 0, ok: 8, rescued: 0, skipped: 3, unreachable: 0,
        },
        {
          host: 'db-02.prod.example.com', changed: 0, failed: 0, ignored: 0, ok: 9, rescued: 0, skipped: 3, unreachable: 0,
        },
        {
          host: 'cache-01.prod.example.com', changed: 0, failed: 1, ignored: 0, ok: 5, rescued: 0, skipped: 0, unreachable: 0,
        },
        {
          host: 'worker-01.prod.example.com', changed: 2, failed: 0, ignored: 1, ok: 10, rescued: 0, skipped: 1, unreachable: 0,
        },
        {
          host: 'worker-02.prod.example.com', changed: 0, failed: 0, ignored: 0, ok: 0, rescued: 0, skipped: 0, unreachable: 1,
        },
      ];

      const stages = [
        {
          name: 'Gathering Facts', ok: 7, failed: 0, changed: 0,
        },
        {
          name: 'Install packages', ok: 5, failed: 2, changed: 3,
        },
        {
          name: 'Configure services', ok: 6, failed: 1, changed: 4,
        },
        {
          name: 'Start services', ok: 6, failed: 1, changed: 2,
        },
        {
          name: 'Run migrations', ok: 5, failed: 1, changed: 1,
        },
      ];

      return [failedTasks, hosts, stages];
    },
  },
};
</script>
