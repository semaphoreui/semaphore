<template>
  <div v-if="item == null">
    <v-progress-linear indeterminate color="primary darken-2"></v-progress-linear>
  </div>
  <div v-else>
    <YesNoDialog
      :title="$t('deleteWorkflow')"
      :text="$t('askDeleteWorkflow')"
      v-model="deleteDialog"
      @yes="remove()"
    />

    <WorkflowRunDialog
      v-model="runDialog"
      :workflow="item"
      :running="runStarting"
      @run="startRun($event)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title class="breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/workflows/`"
        >
          {{ $t('workflows') }}
        </router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ item.name }}</span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn
        v-if="canRun"
        color="primary"
        depressed
        class="mr-3"
        @click="runWorkflow()"
        data-testid="workflow-run"
      >
        {{ $t('run') }}
      </v-btn>

      <v-btn icon color="error" @click="deleteDialog = true" v-if="canUpdate">
        <v-icon>mdi-delete</v-icon>
      </v-btn>

      <v-btn
        icon
        :to="`/project/${projectId}/workflows/${itemId}/edit`"
        v-if="canUpdate"
      >
        <v-icon>mdi-pencil</v-icon>
      </v-btn>
    </v-toolbar>

    <v-tabs>
      <v-tab :to="`/project/${projectId}/workflows/${itemId}/runs`">
        {{ $t('workflowRuns') }}
      </v-tab>
      <v-tab :to="`/project/${projectId}/workflows/${itemId}/stats`">
        {{ $t('workflowStats') }}
      </v-tab>
      <v-tab
        v-if="isPro && can(USER_PERMISSIONS.editWorkflows)"
        :to="`/project/${projectId}/workflows/${itemId}/perms`"
      >
        {{ $t('Permissions') }}
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;" />

    <router-view
      :project-id="projectId"
      :workflow="item"
      :features="features"
      :is-admin="isAdmin"
    ></router-view>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import YesNoDialog from '@/components/YesNoDialog.vue';
import WorkflowRunDialog from '@/components/WorkflowRunDialog.vue';
import PermissionsCheck from '@/components/PermissionsCheck';
import ProjectMixin from '@/components/ProjectMixin';
import { USER_PERMISSIONS } from '@/lib/constants';

export default {
  components: {
    YesNoDialog,
    WorkflowRunDialog,
  },

  mixins: [PermissionsCheck, ProjectMixin],

  props: {
    projectId: Number,
    features: Object,
  },

  data() {
    return {
      item: null,
      deleteDialog: null,
      runDialog: false,
      runStarting: false,
      USER_PERMISSIONS,
    };
  },

  computed: {
    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },

    canRun() {
      return this.can(USER_PERMISSIONS.runProjectTasks);
    },

    canUpdate() {
      return this.can(USER_PERMISSIONS.manageProjectResources);
    },

    itemId() {
      return parseInt(this.$route.params.workflowId, 10);
    },
  },

  watch: {
    async itemId() {
      await this.loadData();
    },
  },

  async created() {
    await this.loadData();
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    // A workflow with a survey asks for its answers first; one without
    // starts at once, with no request body, as before.
    runWorkflow() {
      if ((this.item.survey_vars || []).length === 0) {
        this.startRun(null);
        return;
      }
      this.runDialog = true;
    },

    async startRun(environment) {
      this.runStarting = true;
      try {
        const run = (await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/workflows/${this.itemId}/runs`,
          responseType: 'json',
          ...(environment ? { data: { environment } } : {}),
        })).data;
        this.runDialog = false;

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.$t('workflowRunStarted'),
        });

        await this.$router.push(
          `/project/${this.projectId}/workflows/${this.itemId}/runs/${run.id}`,
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

    async remove() {
      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/workflows/${this.itemId}`,
          responseType: 'json',
        });

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: `Workflow "${this.item.name}" deleted`,
        });

        await this.$router.push({
          path: `/project/${this.projectId}/workflows`,
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.deleteDialog = false;
      }
    },

    async loadData() {
      try {
        this.item = await this.loadProjectResource('workflows', this.itemId);
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
