<template>
  <div v-if="state === 'error'" data-testid="host-page" data-state="error">
    <v-alert
      type="info"
      variant="tonal"
      class="ma-4"
      data-testid="host-notFound"
    >
      {{ $t('hosts_not_found') }}
      <router-link :to="listPath" data-testid="host-back">
        {{ $t('hosts') }}
      </router-link>
    </v-alert>
  </div>
  <div v-else-if="state === 'loading'" data-testid="host-page" data-state="loading">
    <v-progress-linear
      indeterminate
      color="primary-darken-2"
      bg-opacity=".3"
      data-testid="host-loading"
    ></v-progress-linear>
  </div>
  <div v-else data-testid="host-page" data-state="ready">
    <v-toolbar flat>
      <v-btn
        v-if="!projectId"
        icon
        variant="text"
        class="mr-4"
        data-testid="host-backToProjects"
        @click="returnToProjects()"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-app-bar-nav-icon v-else @click="showDrawer()" data-testid="host-drawerToggle"/>
      <v-toolbar-title class="breadcrumbs flex-0-1" data-testid="host-title">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="listPath"
          data-testid="host-breadcrumbHosts"
        >{{ $t('hosts') }}</router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item" data-testid="host-name">{{ host.display_name }}</span>
      </v-toolbar-title>
      <HostStatus :status="host.last_status" class="ml-2" data-testid="host-status"/>
      <span class="ml-2 text-medium-emphasis" data-testid="host-lastSeen">
        {{ $filters.formatDate(host.last_seen) }}
      </span>
    </v-toolbar>

    <v-list density="compact" class="HostView__fields" data-testid="host-fields">
      <v-list-item data-testid="host-firstSeen">
        <v-list-item-title>{{ $t('hosts_col_first_seen') }}</v-list-item-title>
        <template v-slot:append>{{ $filters.formatDate2(host.first_seen) }}</template>
      </v-list-item>
      <v-list-item data-testid="host-lastRun">
        <v-list-item-title>{{ $t('hosts_col_last_task') }}</v-list-item-title>
        <template v-slot:append>
          <span v-if="host.last_task_id == null">&mdash;</span>
          <router-link
            v-else-if="!projectId"
            :to="`/project/${host.last_project_id}?t=${host.last_task_id}`"
          >#{{ host.last_task_id }}</router-link>
          <TaskLink
            v-else
            :task-id="host.last_task_id"
            :label="'#' + host.last_task_id"
          />
        </template>
      </v-list-item>
      <v-list-item data-testid="host-runs">
        <v-list-item-title>{{ $t('hosts_col_runs') }}</v-list-item-title>
        <template v-slot:append>{{ host.runs }}</template>
      </v-list-item>
      <v-list-item data-testid="host-failed">
        <v-list-item-title>{{ $t('host_failed_last_10') }}</v-list-item-title>
        <template v-slot:append>{{ host.failed_in_last_10 }}</template>
      </v-list-item>
      <v-list-item data-testid="host-projects">
        <v-list-item-title>{{ $t('hosts_col_projects') }}</v-list-item-title>
        <template v-slot:append>
          <template v-for="(p, i) in host.projects" :key="p.id">
            <span v-if="i > 0">, </span>
            <router-link :to="`/project/${p.id}/hosts?search=${encodeURIComponent(host.name)}`">
              {{ p.name }}
            </router-link>
          </template>
        </template>
      </v-list-item>
    </v-list>

    <v-tabs data-testid="host-tabs">
      <v-tab data-testid="host-tab-tasks" :to="`${listPath}/${hostId}/tasks`">
        {{ $t('host_tab_tasks') }}
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;" />

    <router-view
      :project-id="projectId"
      :host-id="hostId"
    ></router-view>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import ProjectMixin from '@/components/ProjectMixin';
import HostStatus from '@/components/HostStatus.vue';
import TaskLink from '@/components/TaskLink.vue';

export default {
  mixins: [ProjectMixin],

  components: { HostStatus, TaskLink },

  data() {
    return {
      host: null,
      state: 'loading',
    };
  },

  computed: {
    hostId() {
      return Number(this.$route.params.hostId);
    },

    listPath() {
      return this.projectId ? `/project/${this.projectId}/hosts` : '/hosts';
    },
  },

  watch: {
    async hostId() {
      await this.load();
    },
    async projectId() {
      await this.load();
    },
  },

  async created() {
    await this.load();
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    async load() {
      this.state = 'loading';
      try {
        const response = await axios.get(
          this.projectId
            ? `/api/project/${this.projectId}/hosts/${this.hostId}`
            : `/api/hosts/${this.hostId}`,
        );
        this.host = response.data;
        this.state = 'ready';
      } catch (err) {
        this.state = 'error';
        if (!err.response || err.response.status !== 404) {
          EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
        }
      }
    },
  },
};
</script>

<style lang="scss">
.HostView__fields {
  max-width: 560px;
}
</style>
