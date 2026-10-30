<template>
  <div
    data-testid="workflowRuns-page"
    :data-state="runs == null ? 'loading' : (runs.length === 0 ? 'empty' : 'ready')"
  >
    <v-data-table
      class="mt-4 workflow-runs-table"
      hide-default-footer
      :headers="headers"
      :items="runs || []"
      :loading="runs == null"
      :items-per-page="-1"
      :no-data-text="$t('notLaunched')"
      data-testid="workflowRuns-table"
      :data-state="runs == null ? 'loading' : (runs.length === 0 ? 'empty' : 'ready')"
    >
      <template v-slot:item.id="{ item: run }">
        <router-link
          :to="`/project/${projectId}/workflows/${workflow.id}/runs/${run.id}`"
          data-testid="workflowRuns-rowId"
        >#{{ run.id }}</router-link>
      </template>
      <template v-slot:item.status="{ item: run }">
        <v-chip
          :color="statusColor(run.status)"
          size="small"
          :data-testid="`workflowRuns-rowStatus-${run.status}`"
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
  </div>
</template>

<style lang="scss">
@use 'sass:map';
@use '@/scss/variables' as settings;

// Status chips: a small chip of Vuetify 2 is 24 px high (`$chip-sizes`), Vuetify 4 makes it 26 px.
.workflow-runs-table .v-chip.v-chip--size-small {
  --v-chip-height: 24px;
}

// The loading line in the disabled text colour, as the no-data line (both are
// `.v-data-table__empty-wrapper` of Vuetify 2).
.workflow-runs-table .v-data-table-rows-loading {
  color: rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
}

// On a phone (mobile rows below 600 px, `mobile-breakpoint`) both lines start at the left edge, as
// the rows of the cards do; the table has no sortable column, so no mobile header marks it.
@media #{map.get(settings.$display-breakpoints, 'xs')} {
  .workflow-runs-table :is(.v-data-table-rows-loading, .v-data-table-rows-no-data) {
    text-align: start;
  }
}
</style>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    projectId: Number,
    workflow: Object,
  },

  data() {
    return {
      runs: null,
    };
  },

  computed: {
    headers() {
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
    'workflow.id': function reloadOnWorkflowChange() {
      this.runs = null;
      this.loadData();
    },
  },

  async created() {
    await this.loadData();
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

    async loadData() {
      try {
        this.runs = (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/workflows/${this.workflow.id}/runs`,
          responseType: 'json',
        })).data;
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
    },
  },
};
</script>
