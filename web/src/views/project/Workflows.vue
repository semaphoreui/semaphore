<template>
  <div v-if="items == null" data-testid="workflows-page" data-state="loading">
    <v-progress-linear
      indeterminate
      color="primary-darken-2"
      data-testid="workflows-loading"
    ></v-progress-linear>
  </div>
  <div
    v-else
    data-testid="workflows-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <YesNoDialog
      :title="$t('deleteWorkflow')"
      :text="$t('askDeleteWorkflow')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
      test-id="workflowsDeleteDialog"
    />

    <WorkflowRunDialog
      v-model="runDialog"
      :workflow="runDialogWorkflow"
      :running="runStarting"
      @run="startRun(runDialogWorkflow, $event)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="workflows-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title class="flex-0-1" data-testid="workflows-title">
        {{ $t('workflows') }}
      </v-toolbar-title>
      <v-spacer></v-spacer>

      <v-btn
        variant="elevated"
        color="primary"
        class="mr-1"
        @click="openEditor('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        data-testid="workflows-new"
      >
        {{ $t('newWorkflow') }}
      </v-btn>

      <v-btn
        icon
        variant="text"
        @click="settingsSheet = true"
        data-testid="workflows-tableSettings"
      >
        <v-icon>mdi-cog</v-icon>
      </v-btn>
    </v-toolbar>

    <v-divider style="margin-top: -1px;"/>

    <v-data-table
      hide-default-footer
      class="mt-4 workflows-table"
      show-expand
      :headers="tableHeaders"
      :items="items"
      :items-per-page="-1"
      :expanded="openedItems"
      @update:expanded="openedItems = $event.slice(-1)"
      data-testid="workflows-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.name="{ item }">
        <router-link
          :to="`/project/${projectId}/workflows/${item.id}`"
          data-testid="workflows-rowName"
        >
          {{ item.name }}
        </router-link>
      </template>

      <template v-slot:item.version="{ item }">
        <span
          v-if="item.last_run && item.last_run.version"
          data-testid="workflows-rowVersion"
        >{{ item.last_run.version }}</span>
        <div v-else data-testid="workflows-rowVersionEmpty">&mdash;</div>
      </template>

      <template v-slot:item.status="{ item }">
        <div class="mt-2 mb-2 d-flex" v-if="item.last_run != null">
          <v-chip
            :color="statusColor(item.last_run.status)"
            size="small"
            :data-testid="`workflows-rowStatus-${item.last_run.status}`"
          >
            {{ item.last_run.status }}
          </v-chip>
        </div>
        <div
          v-else
          class="mt-3 mb-2 d-flex"
          style="color: gray;"
          data-testid="workflows-rowNotLaunched"
        >{{ $t('notLaunched') }}</div>
      </template>

      <template v-slot:item.last_run="{ item }">
        <div class="mt-2 mb-2" v-if="item.last_run != null" style="line-height: 1">
          <router-link
            :to="`/project/${projectId}/workflows/${item.id}/runs/${item.last_run.id}`"
            data-testid="workflows-rowLastRun"
          >#{{ item.last_run.id }}</router-link>
          <div style="color: gray; font-size: 14px;" data-testid="workflows-rowLastRunStart">
            {{ $filters.formatDate(item.last_run.start) }}
          </div>
        </div>
        <div v-else data-testid="workflows-rowLastRunEmpty">&mdash;</div>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-group
          density="compact"
          data-testid="workflows-rowActions"
        >
          <v-btn
            v-if="can(USER_PERMISSIONS.runProjectTasks)"
            @click="runWorkflow(item)"
            :title="$t('workflowRunNow')"
            data-testid="workflows-rowRun"
          >
            <v-icon>mdi-play</v-icon>
          </v-btn>
        </v-btn-group>
      </template>

      <!-- the slot draws the whole row: Vuetify 3 does not wrap it in a <tr> as Vuetify 2 did -->
      <template v-slot:expanded-row="{ columns, item }">
        <tr v-if="openedItems.includes(item.id)">
          <td :colspan="columns.length">
            <v-data-table
              style="border: 1px solid lightgray; border-radius: 6px; margin: 10px 0;"
              hide-default-footer
              :headers="runHeaders"
              :items="runs[item.id] || []"
              :loading="runsLoading[item.id]"
              :items-per-page="5"
              :no-data-text="$t('notLaunched')"
              data-testid="workflows-runsTable"
              :data-state="runsLoading[item.id]
                ? 'loading'
                : ((runs[item.id] || []).length === 0 ? 'empty' : 'ready')"
            >
              <template v-slot:item.id="{ item: run }">
                <router-link
                  :to="`/project/${projectId}/workflows/${item.id}/runs/${run.id}`"
                  data-testid="workflows-runsRowId"
                >
                  #{{ run.id }}
                </router-link>
              </template>
              <template v-slot:item.status="{ item: run }">
                <v-chip
                  :color="statusColor(run.status)"
                  size="small"
                  :data-testid="`workflows-runsRowStatus-${run.status}`"
                >{{ run.status }}</v-chip>
              </template>
              <template v-slot:item.version="{ item: run }">
                {{ run.version || '—' }}
              </template>
              <template v-slot:item.start="{ item: run }">
                {{ $filters.formatDate(run.start) }}
              </template>
              <template v-slot:item.end="{ item: run }">
                {{ $filters.formatDate(run.end) }}
              </template>
            </v-data-table>
          </td>
        </tr>
      </template>
    </v-data-table>

    <TableSettingsSheet
      v-model="settingsSheet"
      table-name="project__workflow"
      :headers="headers"
      @change="onTableSettingsChange"
      data-testid="workflows-tableSettingsSheet"
    />
  </div>
</template>
<style lang="scss">
@use 'sass:map';
@use '@/scss/variables' as settings;

// The expand column, first (Vuetify 3 adds it last unless the headers place it).
.workflows-table thead th:first-child,
.workflows-table .v-data-table__td--expanded-row {
  padding-right: 0 !important;
}

// The expand toggle: the 24 px chevron icon of Vuetify 2 (`.v-data-table__expand-icon`), not a
// small icon button of 28 px, so the column is 26 px wide. Focused (after a click) it keeps the
// circle of a focused link icon (`.v-icon--link:focus::after`: currentColor at .12, scale 1.3).
.workflows-table .v-data-table__td--expanded-row > .v-btn {
  width: 24px;
  height: 24px;

  &:focus {
    overflow: visible;

    > .v-btn__overlay {
      border-radius: 50%;
      opacity: .12;
      transform: scale(1.3);
    }
  }
}

// Status chips: a small chip of Vuetify 2 is 24 px high (`$chip-sizes`), Vuetify 4 makes it 26 px.
.workflows-table .v-chip.v-chip--size-small {
  --v-chip-height: 24px;
}

// The loading line of the runs table in the disabled text colour, as the no-data line (both are
// `.v-data-table__empty-wrapper` of Vuetify 2).
.workflows-table .v-data-table-rows-loading {
  color: rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
}

// Mobile rows: the value sits at the right end of the row (`.v-data-table__mobile-row__cell`, a
// shrink-to-fit flex item); Vuetify 4 stretches it over the second grid column, so a block value
// (the status chip, the last run) started at the middle of the row.
.workflows-table .v-data-table__tr--mobile > td > .v-data-table__td-value {
  justify-self: end;
}

// Mobile rows: no expand cell (Vuetify 2 hid the first mobile row, the same cell).
@media #{map.get(settings.$display-breakpoints, 'sm-and-down')} {
  .workflows-table .v-data-table__tr--mobile > .v-data-table__td--expanded-row {
    display: none !important;
  }
}
</style>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import TableSettingsSheet from '@/components/TableSettingsSheet.vue';
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import WorkflowRunDialog from '@/components/WorkflowRunDialog.vue';

export default {
  components: {
    TableSettingsSheet,
    WorkflowRunDialog,
  },
  mixins: [ItemListPageBase],

  data() {
    return {
      settingsSheet: null,
      filteredHeaders: [],
      openedItems: [],
      runs: {},
      runsLoading: {},
      runDialog: false,
      runDialogWorkflow: null,
      runStarting: false,
    };
  },

  computed: {
    // The expand column first and as narrow as its toggle (`width: '1px'` of the Vuetify 2
    // default header), as Vuetify 2 drew it.
    tableHeaders() {
      return [{ key: 'data-table-expand', width: '1px' }, ...this.filteredHeaders];
    },

    runHeaders() {
      return [
        { title: this.$t('workflowRun'), key: 'id', sortable: false },
        { title: this.$t('status'), key: 'status', sortable: false },
        { title: this.$t('version'), key: 'version', sortable: false },
        { title: this.$t('start'), key: 'start', sortable: false },
        { title: this.$t('end'), key: 'end', sortable: false },
      ];
    },
  },

  watch: {
    // Ids of the expanded rows (Vuetify 3 keeps item values, not items): one at most,
    // as single-expand of Vuetify 2 did — the update handler keeps the last one.
    openedItems(ids) {
      ids.forEach((id) => this.loadRuns(id));
    },
  },

  methods: {
    statusColor(status) {
      switch (status) {
        case 'success':
        case 'approved':
          return 'success';
        case 'failed':
        case 'error':
        case 'stopped':
        case 'rejected':
          return 'error';
        case 'running':
        case 'pending':
          return 'primary';
        case 'approval':
          return 'warning';
        default:
          return 'grey';
      }
    },

    openEditor(id) {
      this.$router.push(
        id === 'new'
          ? `/project/${this.projectId}/workflows/new`
          : `/project/${this.projectId}/workflows/${id}/edit`,
      );
    },

    async loadRuns(workflowId) {
      if (this.runs[workflowId] != null) {
        return;
      }
      this.runsLoading[workflowId] = true;
      try {
        const runs = (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/workflows/${workflowId}/runs`,
          responseType: 'json',
        })).data;
        this.runs[workflowId] = runs;
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      } finally {
        this.runsLoading[workflowId] = false;
      }
    },

    onTableSettingsChange({ headers }) {
      this.filteredHeaders = headers;
    },

    getHeaders() {
      return [
        {
          title: this.$t('name'),
          key: 'name',
        },
        {
          key: 'actions',
          sortable: false,
          width: '0%',
        },
        {
          title: this.$t('version'),
          key: 'version',
          sortable: false,
        },
        {
          title: this.$t('status'),
          key: 'status',
          sortable: false,
        },
        {
          title: this.$t('workflowLastRun'),
          key: 'last_run',
          sortable: false,
        },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/workflows`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/workflows/${this.itemId}`;
    },

    getEventName() {
      return 'i-workflow';
    },

    // A workflow with a survey asks for its answers first; one without
    // starts at once, with no request body, as before.
    runWorkflow(workflow) {
      if ((workflow.survey_vars || []).length === 0) {
        this.startRun(workflow, null);
        return;
      }
      this.runDialogWorkflow = workflow;
      this.runDialog = true;
    },
    async startRun(workflow, environment) {
      this.runStarting = true;
      try {
        const run = (await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/workflows/${workflow.id}/runs`,
          responseType: 'json',
          ...(environment ? { data: { environment } } : {}),
        })).data;
        this.runDialog = false;
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.$t('workflowRunStarted'),
        });
        this.$router.push(
          `/project/${this.projectId}/workflows/${workflow.id}/runs/${run.id}`,
        );
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.runStarting = false;
      }
    },
  },
};
</script>
