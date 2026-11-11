<template>
  <div data-testid="hosts-page" :data-state="state">
    <v-toolbar flat>
      <v-btn
        v-if="!projectId"
        icon
        variant="text"
        class="mr-4"
        data-testid="hosts-back"
        @click="returnToProjects()"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-app-bar-nav-icon v-else @click="showDrawer()" data-testid="hosts-drawerToggle"/>
      <v-toolbar-title data-testid="hosts-title">
        {{ $t('hosts') }}
        <span v-if="features.hosts" class="text-medium-emphasis" data-testid="hosts-total">
          {{ total }}
        </span>
      </v-toolbar-title>
    </v-toolbar>

    <v-divider v-if="!projectId" />

    <v-alert
      v-if="!features.hosts"
      variant="tonal"
      color="hsl(348deg, 86%, 61%)"
      class="PageAlert"
      data-testid="hosts-licenseAlert"
    >
      <span v-html="$t('hosts_only_pro')" data-testid="hosts-licenseAlertText"></span>
      <v-btn
        v-if="canManageSystem"
        class="ml-2 text-white"
        color="hsl(348deg, 86%, 61%)"
        @click="upgradeToPro('hosts')"
        data-testid="hosts-upgrade"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>
      <span v-else style="font-weight: bold" data-testid="hosts-contactAdmin">
        {{ $t('contact_admin_to_upgrade') }}
      </span>
    </v-alert>

    <template v-else>
      <v-alert
        v-if="limitExceeded"
        type="warning"
        variant="tonal"
        class="mx-4 mt-4"
        data-testid="hosts-limitAlert"
      >
        {{ $t('hosts_limit_exceeded', { used: nodesUsed, limit: nodes }) }}
      </v-alert>

      <div class="d-flex flex-wrap align-center px-4 pt-4" style="gap: 16px">
        <v-text-field
          v-model="searchText"
          :label="$t('hosts_search')"
          prepend-inner-icon="mdi-magnify"
          density="compact"
          variant="outlined"
          hide-details
          clearable
          style="max-width: 320px"
          data-testid="hosts-search"
        />
        <v-select
          v-if="projectId"
          :model-value="templateId"
          :items="templates"
          item-title="name"
          item-value="id"
          :label="$t('template')"
          density="compact"
          variant="outlined"
          hide-details
          clearable
          style="max-width: 260px"
          data-testid="hosts-template"
          @update:model-value="setList({ templateId: $event })"
        />
      </div>

      <v-chip-group
        :model-value="filter"
        mandatory
        selected-class="text-primary"
        class="px-4"
        data-testid="hosts-filters"
        @update:model-value="setList({ filter: $event })"
      >
        <v-chip
          v-for="f in filters"
          :key="f.key"
          :value="f.key"
          variant="outlined"
          filter
          :data-testid="`hosts-filter-${f.key}`"
        >
          {{ $t(f.label) }}
        </v-chip>
      </v-chip-group>

      <v-data-table
        :headers="headers"
        :items="rows"
        :items-per-page="-1"
        :loading="loading"
        hide-default-footer
        class="mt-2"
        data-testid="hosts-table"
        :data-state="state"
      >
        <template v-slot:no-data>
          <span data-testid="hosts-empty">{{ $t('hosts_empty') }}</span>
        </template>

        <template v-slot:item.name="{ item }">
          <router-link
            :to="projectId ? `/project/${projectId}/hosts/${item.id}` : `/hosts/${item.id}`"
            data-testid="hosts-name"
          >{{ item.display_name }}</router-link>
        </template>

        <template v-slot:item.last_status="{ item }">
          <v-chip
            v-if="!projectId && item.decommissioned"
            size="small"
            data-testid="hosts-rowDecommissioned"
          >{{ $t('host_decommissioned') }}</v-chip>
          <HostStatus v-else :status="item.last_status" data-testid="hosts-rowStatus"/>
        </template>

        <template v-slot:item.last_task="{ item }">
          <span v-if="item.last_task_id" data-testid="hosts-rowTask">
            <TaskLink
              :task-id="item.last_task_id"
              :label="item.last_template
                ? `#${item.last_task_id} ${item.last_template.name}`
                : `#${item.last_task_id}`"
            />
          </span>
          <span v-else>&mdash;</span>
        </template>

        <template v-slot:item.first_seen="{ item }">
          {{ $filters.formatDate(item.first_seen) }}
        </template>

        <template v-slot:item.last_seen="{ item }">
          {{ $filters.formatDate(item.last_seen) }}
        </template>

        <template v-slot:item.templates="{ item }">
          <span data-testid="hosts-rowTemplates">
            <template v-for="(t, i) in item.templates" :key="t.id">
              <span v-if="i > 0">, </span>
              <router-link :to="`/project/${projectId}/templates/${t.id}`">
                {{ t.name }}
              </router-link>
            </template>
          </span>
        </template>

        <template v-slot:item.projects="{ item }">
          <span data-testid="hosts-rowProjects">
            <template v-for="(p, i) in item.projects" :key="p.id">
              <span v-if="i > 0">, </span>
              <router-link :to="`/project/${p.id}/hosts?search=${encodeURIComponent(item.name)}`">
                {{ p.name }}
              </router-link>
            </template>
          </span>
        </template>

        <template v-slot:item.actions="{ item }">
          <v-tooltip v-if="!item.decommissioned" location="bottom">
            <template v-slot:activator="{ props }">
              <v-btn
                v-bind="props"
                icon
                variant="text"
                data-testid="hosts-decommission"
                @click="askDecommission(item)"
              >
                <v-icon>mdi-archive-arrow-down</v-icon>
              </v-btn>
            </template>
            <span>{{ $t('host_decommission') }}</span>
          </v-tooltip>
        </template>
      </v-data-table>

      <div class="d-flex align-center justify-end mt-2 mr-2 mb-4" data-testid="hosts-pagination">
        <span class="text-medium-emphasis mr-2">{{ $t('rowsPerPage') }}</span>

        <v-select
          :model-value="count"
          :items="[25, 50, 100]"
          :disabled="loading"
          density="compact"
          variant="underlined"
          hide-details
          style="max-width: 72px"
          class="mr-4"
          data-testid="hosts-perPage"
          @update:model-value="setList({ count: $event })"
        />

        <v-btn
          :disabled="loading || offset === 0"
          icon
          variant="text"
          data-testid="hosts-prevPage"
          @click="setList({ offset: Math.max(0, offset - count) })"
        >
          <v-icon>mdi-chevron-left</v-icon>
        </v-btn>

        <span class="mx-2 text-medium-emphasis" data-testid="hosts-range">
          {{ rangeStart }}&ndash;{{ rangeEnd }} {{ $t('hosts_range_of') }} {{ total }}
        </span>

        <v-btn
          :disabled="loading || offset + count >= total"
          icon
          variant="text"
          data-testid="hosts-nextPage"
          @click="setList({ offset: offset + count })"
        >
          <v-icon>mdi-chevron-right</v-icon>
        </v-btn>
      </div>

      <YesNoDialog
        v-model="decommissionDialog"
        :title="$t('host_decommission')"
        :text="$t('host_decommission_confirm', {
          name: decommissionHost ? decommissionHost.display_name : '',
        })"
        test-id="hosts-decommissionDialog"
        @yes="decommission()"
      />
    </template>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import ProjectMixin from '@/components/ProjectMixin';
import PageMixin from '@/components/PageMixin';
import HostStatus from '@/components/HostStatus.vue';
import TaskLink from '@/components/TaskLink.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import { GLOBAL_PERMISSIONS } from '@/lib/constants';
import { hasGlobalPermission } from '@/lib/role-permissions';

const FILTERS = [
  { key: 'all', label: 'all' },
  { key: 'failed', label: 'hosts_filter_failing' },
  { key: 'unreachable', label: 'hosts_filter_unreachable' },
  { key: 'not_seen', label: 'hosts_filter_not_seen' },
];

const NOT_SEEN_DAYS = 30;

export default {
  mixins: [ProjectMixin, PageMixin],

  components: { HostStatus, TaskLink, YesNoDialog },

  props: {
    user: Object,
  },

  data() {
    return {
      filters: FILTERS,
      rows: null,
      total: 0,
      loading: false,
      templates: [],
      filter: 'all',
      templateId: null,
      offset: 0,
      count: 25,
      searchText: '',
      searchTimer: null,
      loadSeq: 0,
      nodes: 0,
      nodesUsed: 0,
      decommissionDialog: false,
      decommissionHost: null,
    };
  },

  computed: {
    search() {
      return this.$route.query.search || '';
    },

    rangeStart() {
      return this.total === 0 ? 0 : this.offset + 1;
    },

    rangeEnd() {
      return this.offset + (this.rows ? this.rows.length : 0);
    },

    listUrl() {
      return this.projectId ? `/api/project/${this.projectId}/hosts` : '/api/hosts';
    },

    canManageSystem() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageSystem);
    },

    limitExceeded() {
      return !this.projectId && this.nodes > 0 && this.nodesUsed > this.nodes;
    },

    state() {
      if (this.loading || this.rows == null) return 'loading';
      return this.rows.length === 0 ? 'empty' : 'ready';
    },

    headers() {
      return [
        { title: this.$t('hosts_col_host'), key: 'name', sortable: false },
        { title: this.$t('hosts_col_last_status'), key: 'last_status', sortable: false },
        ...(this.projectId
          ? [{ title: this.$t('hosts_col_last_task'), key: 'last_task', sortable: false }]
          : [{ title: this.$t('hosts_col_first_seen'), key: 'first_seen', sortable: false }]),
        { title: this.$t('hosts_col_last_run'), key: 'last_seen', sortable: false },
        { title: this.$t('hosts_col_runs'), key: 'runs', sortable: false },
        ...(this.projectId
          ? [{ title: this.$t('hosts_col_templates'), key: 'templates', sortable: false }]
          : [{ title: this.$t('hosts_col_projects'), key: 'projects', sortable: false }]),
        ...(this.projectId ? [] : [{
          title: '', key: 'actions', sortable: false, align: 'end',
        }]),
      ];
    },
  },

  watch: {
    search() {
      this.searchText = this.search;
      this.offset = 0;
      this.load();
    },

    async projectId() {
      this.filter = 'all';
      this.templateId = null;
      this.offset = 0;
      await this.open();
    },

    searchText(value) {
      clearTimeout(this.searchTimer);
      if ((value || '') === this.search) return;
      this.searchTimer = setTimeout(() => {
        this.$router.replace({ query: { ...this.$route.query, search: value || undefined } });
      }, 300);
    },
  },

  async created() {
    await this.open();
  },

  beforeUnmount() {
    clearTimeout(this.searchTimer);
  },

  methods: {
    async open() {
      if (!this.projectId && !this.canManageSystem) {
        EventBus.$emit('i-open-last-project');
        return;
      }
      this.searchText = this.search;
      if (!this.features.hosts) return;
      await Promise.all([this.load(), this.projectId ? this.loadTemplates() : null]);
      if (!this.projectId) await this.loadLimit();
    },

    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    setList(changes) {
      Object.assign(this, { offset: 0, ...changes });
      this.load();
    },

    async loadTemplates() {
      try {
        this.templates = await this.loadProjectResources('templates');
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      }
    },

    askDecommission(host) {
      this.decommissionHost = host;
      this.decommissionDialog = true;
    },

    async decommission() {
      try {
        await axios.post(`/api/hosts/${this.decommissionHost.id}/decommission`);
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
        return;
      }
      await this.open();
    },

    async loadLimit() {
      try {
        const { data } = await axios.get('/api/subscription');
        this.nodes = data.nodes || 0;
        this.nodesUsed = data.nodes_used || 0;
      } catch {
        this.nodes = 0;
        this.nodesUsed = 0;
      }
    },

    async load() {
      if (!this.features.hosts) return;
      this.loadSeq += 1;
      const seq = this.loadSeq;
      this.loading = true;
      try {
        const response = await axios.get(this.listUrl, {
          params: {
            search: this.search || undefined,
            status: this.filter === 'failed' || this.filter === 'unreachable' ? this.filter : undefined,
            not_seen_days: this.filter === 'not_seen' ? NOT_SEEN_DAYS : undefined,
            template_id: this.templateId || undefined,
            offset: this.offset,
            count: this.count,
          },
        });
        if (seq !== this.loadSeq) return;
        this.rows = response.data;
        this.total = Number(response.headers['x-total-count']) || 0;
      } catch (err) {
        if (seq !== this.loadSeq) return;
        this.rows = [];
        this.total = 0;
        EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
      } finally {
        if (seq === this.loadSeq) this.loading = false;
      }
    },
  },
};
</script>
