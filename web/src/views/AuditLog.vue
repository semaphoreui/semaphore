<template>
  <div class="AuditLog" :class="{ 'AuditLog--fixed': !$vuetify.breakpoint.xs }">
    <v-toolbar flat class="flex-grow-0">
      <v-btn icon class="mr-4" data-testid="audit-back" @click="returnToProjects">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
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
          <v-list-item :href="exportUrl('csv')" download>
            {{ $t('audit_export_csv') }}
          </v-list-item>
          <v-list-item :href="exportUrl('jsonl')" download>
            {{ $t('audit_export_jsonl') }}
          </v-list-item>
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

    <div v-if="$vuetify.breakpoint.xs" class="d-flex justify-end px-4 pt-2">
      <v-btn
        text
        :disabled="loading || newer === null"
        data-testid="audit-newer-top"
        @click="load({ after: newer })"
      >
        <v-icon left>mdi-chevron-left</v-icon>
        {{ $t('audit_newer') }}
      </v-btn>
      <v-btn
        text
        :disabled="loading || older === null"
        data-testid="audit-older-top"
        @click="load({ before: older })"
      >
        {{ $t('audit_older') }}
        <v-icon right>mdi-chevron-right</v-icon>
      </v-btn>
    </div>

    <div v-if="failed && events.length" class="px-4 pt-2" data-testid="audit-load-failed">
      {{ $t('audit_load_failed') }}
      <v-btn
        text
        small
        color="primary"
        class="ml-2"
        data-testid="audit-retry"
        @click="load(lastPage)"
      >
        {{ $t('audit_retry') }}
      </v-btn>
    </div>

    <v-data-table
      fixed-header
      :headers="headers"
      :items="events"
      :loading="loading"
      item-key="seq"
      hide-default-footer
      disable-pagination
      disable-sort
      class="mt-2 AuditLog__table"
      @click:row="selected = $event"
    >
      <template v-slot:no-data>
        <span v-if="!failed">{{ $t('audit_no_events') }}</span>
        <span v-else data-testid="audit-load-failed">
          {{ $t('audit_load_failed') }}
          <v-btn
            text
            small
            color="primary"
            class="ml-2"
            data-testid="audit-retry"
            @click="load(lastPage)"
          >
            {{ $t('audit_retry') }}
          </v-btn>
        </span>
      </template>

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

    <div v-if="searchedTo" class="d-flex align-center px-4 pt-2" data-testid="audit-searched-to">
      <span class="text--secondary">
        {{ $t('audit_searched_to', { date: localTime(searchedTo.timestamp) }) }}
      </span>
      <v-btn
        text
        color="primary"
        class="ml-2"
        :disabled="loading"
        data-testid="audit-search-older"
        @click="load({ before: older })"
      >
        {{ $t('audit_search_older') }}
      </v-btn>
    </div>

    <div class="d-flex justify-end pa-4">
      <v-btn
        text
        :disabled="loading"
        data-testid="audit-latest"
        @click="load({})"
      >
        {{ $t('audit_latest') }}
      </v-btn>
      <v-btn
        text
        :disabled="loading || newer === null"
        data-testid="audit-newer"
        @click="load({ after: newer })"
      >
        <v-icon left>mdi-chevron-left</v-icon>
        {{ $t('audit_newer') }}
      </v-btn>
      <v-btn
        text
        :disabled="loading || older === null"
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

  props: {
    user: Object,
  },

  components: { AuditFilters, AuditEventCard },

  data() {
    return {
      events: [],
      older: null,
      newer: null,
      searchedTo: null,
      loading: false,
      failed: false,
      lastPage: {},
      requestId: 0,
      filters: {},
      selected: null,
      users: [],
      projects: [],
      namesLoaded: false,
      namesFailed: false,
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
    if (!this.user || !this.user.can_read_audit_log) {
      EventBus.$emit('i-open-last-project');
      return;
    }
    await Promise.all([this.load({}), this.loadNames()]);
  },

  methods: {
    returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    // Periods come from datetime-local inputs in the browser's time zone.
    requestParams(page) {
      const params = filterParams(this.filters);
      if (params.from) {
        params.from = dayjs(params.from).toISOString();
      }
      // The server excludes the end, so the whole chosen end minute is sent.
      if (params.to) {
        params.to = dayjs(params.to).add(1, 'minute').toISOString();
      }
      return { ...params, ...page };
    },

    async load(page) {
      this.requestId += 1;
      const id = this.requestId;
      this.lastPage = page;
      this.loading = true;
      try {
        const { data } = await axios.get('/api/audit/events', {
          params: this.requestParams(page),
          paramsSerializer: { indexes: null },
        });
        if (id !== this.requestId) {
          return;
        }
        this.events = data.events;
        this.older = data.older;
        this.newer = data.newer;
        this.searchedTo = data.searched_to;
        this.failed = false;
      } catch (err) {
        if (id === this.requestId) {
          this.failed = true;
          EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
        }
      } finally {
        if (id === this.requestId) {
          this.loading = false;
        }
      }
    },

    async loadNames() {
      const [users, projects] = await Promise.allSettled([
        axios.get('/api/users'),
        axios.get('/api/projects'),
      ]);
      if (users.status === 'fulfilled') {
        this.users = users.value.data;
      }
      if (projects.status === 'fulfilled') {
        this.projects = projects.value.data;
      }
      const failure = [users, projects].find((r) => r.status === 'rejected');
      if (failure) {
        this.namesFailed = true;
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(failure.reason) });
      } else {
        this.namesLoaded = true;
      }
    },

    setFilters(filters) {
      this.filters = filters;
      // A failed request must not leave events of the previous filters under the new ones.
      this.events = [];
      this.older = null;
      this.newer = null;
      this.searchedTo = null;
      this.load({});
    },

    pivot(filter) {
      this.selected = null;
      this.setFilters({ ...this.filters, ...filter });
    },

    exportUrl(format) {
      return axios.getUri({
        url: '/api/audit/events/export',
        params: this.requestParams({ format }),
        paramsSerializer: { indexes: null },
      });
    },

    projectName(event) {
      const id = event.scope && Number(event.scope.project_id);
      if (!id) {
        return '';
      }
      if (!this.namesLoaded && !this.namesFailed) {
        return '';
      }
      // The project list is capped, so a project missing from it may still exist.
      const project = this.projects.find((p) => p.id === id);
      return project ? project.name : this.$t('audit_project_id', { id });
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
.AuditLog--fixed {
  display: flex;
  flex-direction: column;
  height: 100vh;

  .AuditLog__table {
    flex: 1 1 auto;
    min-height: 0;
    overflow: hidden;

    .v-data-table__wrapper {
      height: 100%;
      overflow: auto;
    }
  }
}

.AuditLog__table tbody tr {
  cursor: pointer;
}
</style>
