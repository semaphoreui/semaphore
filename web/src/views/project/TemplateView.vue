<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="!isLoaded" data-testid="template-page" data-state="loading">
    <v-progress-linear
      indeterminate
      color="primary-darken-2"
      data-testid="template-loading"
    ></v-progress-linear>
  </div>
  <div v-else data-testid="template-page" data-state="ready">
    <NewTaskDialog
      v-model="newTaskDialog"
      :project-id="projectId"
      :template="item"
    />

    <EditTemplateDialog
      v-model="editDialog"
      :project-id="projectId"
      :item-app="item.app"
      :item-id="itemId"
      @save="loadData()"
      :features="features"
      :task-type="item.type"
      test-id="templateDialog"
    ></EditTemplateDialog>

    <EditTemplateDialog
      v-model="copyDialog"
      :project-id="projectId"
      :item-app="item.app"
      item-id="new"
      :source-item-id="itemId"
      @save="onTemplateCopied"
      :features="features"
      :task-type="item.type"
      test-id="templateCopyDialog"
    ></EditTemplateDialog>

    <ObjectRefsDialog
      object-title="template"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
      test-id="templateRefsDialog"
    />

    <YesNoDialog
      :title="$t('deleteTemplate')"
      :text="$t('askDeleteTemp')"
      v-model="deleteDialog"
      @yes="remove()"
      test-id="templateDeleteDialog"
    />

    <YesNoDialog
      :title="$t('stopAllTasks')"
      :text="$t('askStopAllTasks')"
      v-model="stopAllDialog"
      @yes="stopAllTasks(false)"
      test-id="templateStopAllDialog"
    />

    <YesNoDialog
      :title="$t('forceStopAllTasks')"
      :text="$t('askForceStopAllTasks')"
      v-model="forceStopAllDialog"
      @yes="stopAllTasks(true)"
      test-id="templateForceStopAllDialog"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="template-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title class="breadcrumbs" data-testid="template-title">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          data-testid="template-breadcrumbTemplates"
          :to="
            viewId
              ? `/project/${projectId}/views/${viewId}/templates/`
              : `/project/${projectId}/templates/`
          "
        >
          {{ $t('taskTemplates2') }}
        </router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item" data-testid="template-name">{{ item.name }}</span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-menu offset-y v-if="canStop">
        <template v-slot:activator="{ props }">
          <v-btn
            v-bind="props"
            color="grey"
            class="mr-3 pr-2"
            :disabled="stopAllTasksProgress"
            data-testid="template-stopAllMenu"
          >
            {{ $t('stopAll') }}
            <v-icon>mdi-chevron-down</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item
            @click="stopAllDialog = true"
            data-testid="template-stopAllMenu-stop"
          >
            <v-list-item-icon>
              <v-icon>mdi-stop</v-icon>
            </v-list-item-icon>
            <v-list-item-title>{{ $t('stop') }}</v-list-item-title>
          </v-list-item>
          <v-list-item
            @click="forceStopAllDialog = true"
            data-testid="template-stopAllMenu-forceStop"
          >
            <v-list-item-icon>
              <v-icon>mdi-alert-octagon</v-icon>
            </v-list-item-icon>
            <v-list-item-title>{{ $t('forceStop') }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>

      <v-btn
        v-if="canRun"
        color="primary"
        variant="flat"
        class="mr-3"
        @click="newTaskDialog = true"
        data-testid="template-run"
      >
        {{ $t(TEMPLATE_TYPE_ACTION_TITLES[item.type]) }}
      </v-btn>

      <v-btn
        icon
        color="error"
        @click="askDelete()"
        v-if="canUpdate"
        data-testid="template-delete"
      >
        <v-icon>mdi-delete</v-icon>
      </v-btn>

      <v-btn icon @click="copyDialog = true" v-if="canUpdate" data-testid="template-copy">
        <v-icon>mdi-content-copy</v-icon>
      </v-btn>

      <v-btn icon @click="editDialog = true" v-if="canUpdate" data-testid="template-edit">
        <v-icon>mdi-pencil</v-icon>
      </v-btn>
    </v-toolbar>

    <SingleLineEditable
      class="mx-4 TemplateView__description"
      v-model="item.description"
      @save="updateDescription()"
      v-if="item.description || can(USER_PERMISSIONS.manageProjectResources)"
      :can-edit="can(USER_PERMISSIONS.manageProjectResources)"
      data-testid="template-description"
    />

    <v-tabs data-testid="template-tabs">
      <v-tab
        data-testid="template-tab-tasks"
        :to="`/project/${item.project_id}${
          $route.params.viewId ? `/views/${$route.params.viewId}` : ''
        }/templates/${item.id}/tasks`">{{ $t('template_tasks') }}
      </v-tab>
      <v-tab
        data-testid="template-tab-details"
        :to="`/project/${item.project_id}${
          $route.params.viewId ? `/views/${$route.params.viewId}` : ''
        }/templates/${item.id}/details`">{{ $t('template_details') }}
      </v-tab>
      <v-tab
        v-if="isPro && can(USER_PERMISSIONS.manageProjectResources)"
        data-testid="template-tab-perms"
        :to="`/project/${item.project_id}${
          $route.params.viewId ? `/views/${$route.params.viewId}` : ''
        }/templates/${item.id}/perms`">{{ $t('Permissions') }}
      </v-tab>
      <v-tab
        v-if="['terraform', 'tofu'].includes(item.app)"
        data-testid="template-tab-state"
        :to="`/project/${item.project_id}${
          $route.params.viewId ? `/views/${$route.params.viewId}` : ''
        }/templates/${item.id}/state`"
      >
        {{ $t('template_tf_workspaces') }}
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;" />

    <router-view
      :project-id="projectId"
      :template="item"
      :inventory="inventory"
      :environment="environment"
      :repositories="repositories"
      :features="features"
      :is-admin="isAdmin"
      @update-template="loadData"
      :need-update="needLoadData"
    ></router-view>
  </div>
</template>

<style lang="scss">
@use 'sass:map';
@use '@/scss/variables' as settings;

.TemplateView__description {
  font-size: 14px;
  margin-bottom: 12px;
}

@media #{map.get(settings.$display-breakpoints, 'md-and-up')} {
  .TemplateView__description {
    transform: translateY(-12px);
    margin-bottom: 0;
  }
}

</style>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import YesNoDialog from '@/components/YesNoDialog.vue';
import {
  TEMPLATE_TYPE_ACTION_TITLES,
  TEMPLATE_TYPE_ICONS,
  TEMPLATE_TYPE_TITLES,
  USER_PERMISSIONS,
} from '@/lib/constants';
import ObjectRefsDialog from '@/components/ObjectRefsDialog.vue';
import NewTaskDialog from '@/components/NewTaskDialog.vue';
import EditTemplateDialog from '@/components/EditTemplateDialog.vue';
import PermissionsCheck from '@/components/PermissionsCheck';
import SingleLineEditable from '@/components/SingleLineEditable.vue';
import ProjectMixin from '@/components/ProjectMixin';

export default {
  components: {
    SingleLineEditable,
    YesNoDialog,
    ObjectRefsDialog,
    NewTaskDialog,
    EditTemplateDialog,
  },

  props: {
    projectId: Number,
    userPermissions: Number,
    features: Object,
  },

  mixins: [PermissionsCheck, ProjectMixin],

  data() {
    return {
      item: null,
      inventory: null,
      environment: null,
      repositories: null,
      deleteDialog: null,
      editDialog: null,
      copyDialog: null,
      TEMPLATE_TYPE_ICONS,
      TEMPLATE_TYPE_TITLES,
      TEMPLATE_TYPE_ACTION_TITLES,
      itemRefs: null,
      itemRefsDialog: null,
      newTaskDialog: null,
      stopAllDialog: null,
      forceStopAllDialog: null,
      USER_PERMISSIONS,

      needLoadData: false,
      stopAllTasksProgress: false,
    };
  },

  computed: {
    canRun() {
      return this.can(USER_PERMISSIONS.runProjectTasks);
    },

    canStop() {
      return this.can(USER_PERMISSIONS.runProjectTasks);
    },

    canUpdate() {
      return this.can(USER_PERMISSIONS.manageProjectResources);
    },

    viewId() {
      if (/^-?\d+$/.test(this.$route.params.viewId)) {
        return parseInt(this.$route.params.viewId, 10);
      }
      return this.$route.params.viewId;
    },

    itemId() {
      if (/^-?\d+$/.test(this.$route.params.templateId)) {
        return parseInt(this.$route.params.templateId, 10);
      }
      return this.$route.params.templateId;
    },
    isNew() {
      return this.itemId === 'new';
    },
    isLoaded() {
      return this.item && this.inventory && this.environment && this.repositories;
    },

    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },
  },

  watch: {
    async itemId() {
      await this.loadData();
    },
  },

  async created() {
    if (this.isNew) {
      await this.$router.replace({
        path: `/project/${this.projectId}/templates/new/edit`,
      });
    } else {
      await this.loadData();
    }
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async stopAllTasks(force) {
      try {
        this.stopAllTasksProgress = true;
        await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/templates/${this.itemId}/stop_all_tasks`,
          data: {
            force,
          },
          responseType: 'json',
        });

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'All running tasks have been requested to stop',
        });

        this.needLoadData = true;
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.stopAllDialog = false;
        this.forceStopAllDialog = false;
        this.stopAllTasksProgress = false;

        setTimeout(() => {
          this.needLoadData = false;
        }, 100);
      }
    },

    async askDelete() {
      this.itemRefs = (
        await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/templates/${this.itemId}/refs`,
          responseType: 'json',
        })
      ).data;

      if (this.itemRefs.integrations.length > 0) {
        this.itemRefsDialog = true;
        return;
      }

      this.deleteDialog = true;
    },

    async remove() {
      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/templates/${this.itemId}`,
          responseType: 'json',
        });

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: `Template "${this.item.name}" deleted`,
        });

        await this.$router.push({
          path: `/project/${this.projectId}/templates`,
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

    async onTemplateCopied(e) {
      await this.$router.push({
        path: `/project/${this.projectId}/templates/${e.item.id}`,
      });
    },

    async loadData() {
      [
        this.item,
        this.inventory,
        this.environment,
        this.repositories,
      ] = await Promise.all([
        this.loadProjectResource('templates', this.itemId),
        this.loadProjectResources('inventory'),
        this.loadProjectResources('environment'),
        this.loadProjectResources('repositories'),
      ]);
    },

    async updateDescription() {
      try {
        await axios({
          method: 'put',
          url: `/api/project/${this.projectId}/templates/${this.itemId}/description`,
          responseType: 'json',
          data: {
            description: this.item.description,
          },
        });
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
