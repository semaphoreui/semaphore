<template>
  <div>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        <router-link :to="`/project/${projectId}/workflows`">
          {{ $t('workflows') }}
        </router-link>
        <span class="ml-2">/ {{ $t('workflowRun') }} #{{ runId }}</span>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn icon @click="loadData()" :title="$t('refresh')">
        <v-icon>mdi-refresh</v-icon>
      </v-btn>
    </v-toolbar>

    <v-divider />

    <div
      v-if="details != null"
      class="pa-4"
      style="max-width: calc(var(--breakpoint-lg) - var(--nav-drawer-width) - 200px); margin: auto"
    >
      <v-card class="mb-4 pa-4">
        <div class="d-flex align-center">
          <div>
            <div class="text-h6">{{ workflow ? workflow.name : '' }}</div>
            <div class="text-caption" v-if="workflow && workflow.description">
              {{ workflow.description }}
            </div>
          </div>
          <v-spacer />
          <v-chip :color="statusColor(details.run.status)" small>
            {{ details.run.status }}
          </v-chip>
        </div>
        <div class="text-caption mt-2">
          <span v-if="details.run.start">
            <strong>{{ $t('start') }}:</strong> {{ details.run.start }}
          </span>
          <span v-if="details.run.end" class="ml-3">
            <strong>{{ $t('end') }}:</strong> {{ details.run.end }}
          </span>
        </div>
      </v-card>

      <v-data-table
        :headers="headers"
        :items="details.nodes"
        item-key="node.id"
        hide-default-footer
        :items-per-page="Number.MAX_VALUE"
      >
        <template v-slot:item.id="{ item }">
          #{{ item.node.id }}
        </template>
        <template v-slot:item.template="{ item }">
          <router-link
            v-if="item.task"
            :to="`/project/${projectId}/templates/${item.task.template_id}`"
          >{{ item.task.tpl_alias || item.task.template_id }}</router-link>
          <span v-else>{{ templateName(item.node.template_id) }}</span>
        </template>
        <template v-slot:item.status="{ item }">
          <v-chip
            v-if="item.task"
            :color="statusColor(item.task.status)"
            small
          >{{ item.task.status }}</v-chip>
          <span v-else class="text-caption text--secondary">{{ $t('workflowNodePending') }}</span>
        </template>
        <template v-slot:item.task="{ item }">
          <router-link
            v-if="item.task"
            :to="`/project/${projectId}/templates/${item.task.template_id}/tasks`"
          >#{{ item.task.id }}</router-link>
          <span v-else>—</span>
        </template>
      </v-data-table>
    </div>

    <div v-else class="pa-4 text-center">
      <v-progress-circular indeterminate color="primary" />
    </div>
  </div>
</template>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    projectId: Number,
  },
  data() {
    return {
      details: null,
      workflow: null,
      templates: [],
      pollHandle: null,
    };
  },
  computed: {
    workflowId() {
      return parseInt(this.$route.params.workflowId, 10);
    },
    runId() {
      return parseInt(this.$route.params.runId, 10);
    },
    headers() {
      return [
        { text: this.$i18n.t('workflowNodeId'), value: 'id', sortable: false },
        { text: this.$i18n.t('taskTemplate'), value: 'template', sortable: false },
        { text: this.$i18n.t('status'), value: 'status', sortable: false },
        { text: this.$i18n.t('workflowTaskColumn'), value: 'task', sortable: false },
      ];
    },
  },
  async created() {
    await this.loadData();
    this.pollHandle = setInterval(() => {
      if (this.details && this.details.run.status === 'running') {
        this.loadData();
      } else if (this.pollHandle) {
        clearInterval(this.pollHandle);
        this.pollHandle = null;
      }
    }, 5000);
  },
  beforeDestroy() {
    if (this.pollHandle) clearInterval(this.pollHandle);
  },
  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },
    statusColor(status) {
      switch (status) {
        case 'success':
          return 'success';
        case 'failed':
        case 'error':
        case 'stopped':
          return 'error';
        case 'running':
          return 'primary';
        default:
          return 'grey';
      }
    },
    templateName(id) {
      const t = this.templates.find((x) => x.id === id);
      return t ? t.name : `#${id}`;
    },
    async loadData() {
      try {
        const [details, workflow, templates] = await Promise.all([
          axios.get(
            `/api/project/${this.projectId}/workflows/${this.workflowId}/runs/${this.runId}`,
          ),
          axios.get(`/api/project/${this.projectId}/workflows/${this.workflowId}`),
          axios.get(`/api/project/${this.projectId}/templates`),
        ]);
        this.details = details.data;
        this.workflow = workflow.data;
        this.templates = templates.data || [];
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },
  },
};
</script>
