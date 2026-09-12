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
    </v-tabs>

    <v-divider style="margin-top: -1px;" />

    <router-view
      :project-id="projectId"
      :workflow="item"
    ></router-view>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import YesNoDialog from '@/components/YesNoDialog.vue';
import PermissionsCheck from '@/components/PermissionsCheck';
import ProjectMixin from '@/components/ProjectMixin';
import { USER_PERMISSIONS } from '@/lib/constants';

export default {
  components: {
    YesNoDialog,
  },

  mixins: [PermissionsCheck, ProjectMixin],

  props: {
    projectId: Number,
  },

  data() {
    return {
      item: null,
      deleteDialog: null,
      USER_PERMISSIONS,
    };
  },

  computed: {
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

    async runWorkflow() {
      try {
        const run = (await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/workflows/${this.itemId}/run`,
          responseType: 'json',
        })).data;

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
