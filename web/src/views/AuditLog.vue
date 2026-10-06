<template>
  <div class="AuditLog">
    <v-toolbar flat>
      <v-toolbar-title>{{ $t('audit_log') }}</v-toolbar-title>
      <v-spacer />
      <v-menu offset-y>
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            color="primary"
            :disabled="!features.audit_log_filters"
            data-testid="audit-export"
            v-bind="attrs"
            v-on="on"
          >
            <v-icon left>mdi-download</v-icon>
            {{ $t('audit_export') }}
          </v-btn>
        </template>
        <v-list>
          <v-list-item :href="exportUrl('csv')">{{ $t('audit_export_csv') }}</v-list-item>
          <v-list-item :href="exportUrl('jsonl')">{{ $t('audit_export_jsonl') }}</v-list-item>
        </v-list>
      </v-menu>
    </v-toolbar>

    <v-divider />

    <v-alert
      v-if="!features.audit_log_filters"
      text
      color="hsl(348deg, 86%, 61%)"
      class="PageAlert"
      data-testid="audit-pro-notice"
    >
      <span v-html="$t('audit_filters_only_pro')"></span>
      <v-btn
        dark
        class="ml-2"
        color="hsl(348deg, 86%, 61%)"
        @click="upgradeToPro('audit_log_filters')"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>
    </v-alert>

    <AuditFilters
      :value="filters"
      :disabled="!features.audit_log_filters"
      :users="users"
      :projects="projects"
      :style="features.audit_log_filters ? '' : 'opacity: 0.5'"
      @input="setFilters"
    />

    <v-data-table
      :headers="headers"
      :items="events"
      :loading="loading"
      item-key="seq"
      hide-default-footer
      disable-pagination
      disable-sort
      class="mt-2 AuditLog__table"
      :no-data-text="$t('audit_no_events')"
      @click:row="selected = $event"
    >
      <template v-slot:item.timestamp="{ item }">
        <span style="white-space: nowrap">{{ localTime(item.timestamp) }}</span>
      </template>

      <template v-slot:item.actor="{ item }">
        <span style="white-space: nowrap">
          <v-icon small class="mr-1">{{ actorIcon(item.actor) }}</v-icon>
          {{ actorLabel(item.actor) }}
        </span>
      </template>

      <template v-slot:item.action="{ item }">{{ title(item) }}</template>

      <template v-slot:item.target="{ item }">
        <span v-if="item.target">
          <span class="text--secondary">{{ targetType(item.target.type) }}</span>
          {{ item.target.name || item.target.id }}
        </span>
      </template>

      <template v-slot:item.project="{ item }">{{ projectName(item) }}</template>

      <template v-slot:item.outcome="{ item }">
        <span style="white-space: nowrap">
          <v-icon v-if="item.outcome === 'failure'" small color="error">mdi-close-circle</v-icon>
          <v-icon v-else-if="item.reason" small color="warning">mdi-alert-circle</v-icon>
          <v-icon v-else small color="success">mdi-check-circle</v-icon>
          <span v-if="item.reason" class="ml-1">{{ reason(item.reason) }}</span>
        </span>
      </template>

      <template v-slot:item.ip="{ item }">{{ item.source ? item.source.ip : '' }}</template>
    </v-data-table>

    <div class="d-flex justify-end pa-4">
      <v-btn text :disabled="newer === null" data-testid="audit-latest" @click="load({})">
        {{ $t('audit_latest') }}
      </v-btn>
      <v-btn
        text
        :disabled="newer === null"
        data-testid="audit-newer"
        @click="load({ after: newer })"
      >
        <v-icon left>mdi-chevron-left</v-icon>
        {{ $t('audit_newer') }}
      </v-btn>
      <v-btn
        text
        :disabled="older === null"
        data-testid="audit-older"
        @click="load({ before: older })"
      >
        {{ $t('audit_older') }}
        <v-icon right>mdi-chevron-right</v-icon>
      </v-btn>
    </div>

    <v-navigation-drawer
      :value="selected !== null"
      right
      fixed
      temporary
      width="480"
      @input="onDrawer"
    >
      <AuditEventCard
        v-if="selected"
        :event="selected"
        :project-name="projectName(selected)"
        :pivots="features.audit_log_filters"
        @close="selected = null"
        @pivot="pivot"
      />
    </v-navigation-drawer>
  </div>
</template>
<script>
import axios from 'axios';
import dayjs from 'dayjs';
import PageMixin from '@/components/PageMixin';
import AuditFilters from '@/components/AuditFilters.vue';
import AuditEventCard from '@/components/AuditEventCard.vue';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import {
  eventTitle, reasonLabel, targetTypeLabel, actorLabel, actorIcon, filterParams,
} from '@/lib/auditEvent';

export default {
  mixins: [PageMixin],

  components: { AuditFilters, AuditEventCard },

  data() {
    return {
      events: [],
      older: null,
      newer: null,
      loading: false,
      filters: {},
      selected: null,
      users: [],
      projects: [],
    };
  },

  computed: {
    headers() {
      return [
        { text: this.$t('audit_col_time'), value: 'timestamp' },
        { text: this.$t('audit_col_actor'), value: 'actor' },
        { text: this.$t('audit_col_action'), value: 'action' },
        { text: this.$t('audit_col_object'), value: 'target' },
        { text: this.$t('audit_col_project'), value: 'project' },
        { text: this.$t('audit_col_result'), value: 'outcome' },
        { text: this.$t('audit_col_ip'), value: 'ip' },
      ];
    },
  },

  async created() {
    await Promise.all([this.load({}), this.loadNames()]);
  },

  methods: {
    // Periods come from datetime-local inputs in the browser's time zone.
    requestParams(page) {
      const params = filterParams(this.filters);
      ['from', 'to'].forEach((name) => {
        if (params[name]) {
          params[name] = dayjs(params[name]).toISOString();
        }
      });
      return { ...params, ...page };
    },

    async load(page) {
      this.loading = true;
      try {
        const { data } = await axios.get('/api/audit/events', {
          params: this.requestParams(page),
          paramsSerializer: { indexes: null },
        });
        this.events = data.events;
        this.older = data.older;
        this.newer = data.newer;
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      } finally {
        this.loading = false;
      }
    },

    async loadNames() {
      const [users, projects] = await Promise.all([
        axios.get('/api/users'),
        axios.get('/api/projects'),
      ]);
      this.users = users.data;
      this.projects = projects.data;
    },

    setFilters(filters) {
      this.filters = filters;
      this.load({});
    },

    pivot(filter) {
      this.selected = null;
      this.setFilters({ ...this.filters, ...filter });
    },

    exportUrl(format) {
      const query = new URLSearchParams();
      Object.entries(this.requestParams({ format })).forEach(([name, value]) => {
        [].concat(value).forEach((v) => query.append(name, v));
      });
      return `/api/audit/events/export?${query}`;
    },

    projectName(event) {
      const id = event.scope && Number(event.scope.project_id);
      if (!id) {
        return '';
      }
      const project = this.projects.find((p) => p.id === id);
      return project ? project.name : this.$t('audit_project_deleted', { id });
    },

    localTime(timestamp) {
      return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss');
    },

    title(event) {
      return eventTitle(event, this.$i18n);
    },

    reason(value) {
      return reasonLabel(value, this.$i18n);
    },

    targetType(value) {
      return targetTypeLabel(value, this.$i18n);
    },

    actorLabel(actor) {
      return actorLabel(actor, this.$i18n);
    },

    actorIcon,

    onDrawer(open) {
      if (!open) {
        this.selected = null;
      }
    },
  },
};
</script>
<style lang="scss">
.AuditLog__table tbody tr {
  cursor: pointer;
}
</style>
