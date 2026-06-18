<template>
  <div>
    <v-data-table
      class="mt-4"
      hide-default-footer
      :headers="headers"
      :items="runs || []"
      :loading="runs == null"
      :items-per-page="Number.MAX_VALUE"
      :no-data-text="$t('notLaunched')"
    >
      <template v-slot:item.id="{ item: run }">
        <router-link
          :to="`/project/${projectId}/workflows/${workflow.id}/runs/${run.id}`"
        >#{{ run.id }}</router-link>
      </template>
      <template v-slot:item.status="{ item: run }">
        <v-chip :color="statusColor(run.status)" small>{{ run.status }}</v-chip>
      </template>
      <template v-slot:item.version="{ item: run }">
        {{ run.version || '—' }}
      </template>
      <template v-slot:item.start="{ item: run }">
        {{ run.start | formatDate }}
      </template>
      <template v-slot:item.end="{ item: run }">
        {{ run.end | formatDate }}
      </template>
    </v-data-table>
  </div>
</template>

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
        { text: this.$i18n.t('workflowRun'), value: 'id', sortable: false },
        { text: this.$i18n.t('status'), value: 'status', sortable: false },
        { text: this.$i18n.t('version'), value: 'version', sortable: false },
        { text: this.$i18n.t('start'), value: 'start', sortable: false },
        { text: this.$i18n.t('end'), value: 'end', sortable: false },
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
