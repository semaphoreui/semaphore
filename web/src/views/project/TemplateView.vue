<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="!isLoaded">
    <v-progress-linear indeterminate color="primary darken-2"></v-progress-linear>
  </div>
  <div v-else>
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
      :premium-features="premiumFeatures"
      :task-type="item.type"
    ></EditTemplateDialog>

    <EditTemplateDialog
      v-model="copyDialog"
      :project-id="projectId"
      :item-app="item.app"
      item-id="new"
      :source-item-id="itemId"
      @save="onTemplateCopied"
      :premium-features="premiumFeatures"
      :task-type="item.type"
    ></EditTemplateDialog>

    <ObjectRefsDialog
      object-title="template"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      :title="$t('deleteTemplate')"
      :text="$t('askDeleteTemp')"
      v-model="deleteDialog"
      @yes="remove()"
    />

    <YesNoDialog
      :title="$t('stopAllTasks')"
      :text="$t('askStopAllTasks')"
      v-model="stopAllDialog"
      @yes="stopAllTasks()"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title class="breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="
            viewId
              ? `/project/${projectId}/views/${viewId}/templates/`
              : `/project/${projectId}/templates/`
          "
        >
          {{ $t('taskTemplates2') }}
        </router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ item.name }}</span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn
        @click="stopAllDialog = true"
        color="grey"
        class="mr-3"
        v-if="canStop"
      >
        Stop all
      </v-btn>

      <v-btn
        v-if="canRun"
        color="primary"
        depressed
        class="mr-3"
        @click="newTaskDialog = true"
        data-testid="template-run"
      >
        {{ $t(TEMPLATE_TYPE_ACTION_TITLES[item.type]) }}
      </v-btn>

      <v-btn icon color="error" @click="askDelete()" v-if="canUpdate">
        <v-icon>mdi-delete</v-icon>
      </v-btn>

      <v-btn icon @click="copyDialog = true" v-if="canUpdate">
        <v-icon>mdi-content-copy</v-icon>
      </v-btn>

      <v-btn icon @click="editDialog = true" v-if="canUpdate">
        <v-icon>mdi-pencil</v-icon>
      </v-btn>
    </v-toolbar>

    <SingleLineEditable
      class="mx-4 TemplateView__description"
      v-model="item.description"
      @save="updateDescription()"
      v-if="item.description || can(USER_PERMISSIONS.manageProjectResources)"
      :can-edit="can(USER_PERMISSIONS.manageProjectResources)"
    />

    <v-tabs>
      <v-tab
        :to="`/project/${item.project_id}${
          $route.params.viewId ? `/views/${$route.params.viewId}` : ''
        }/templates/${item.id}/tasks`">{{ $t('template_tasks') }}
      </v-tab>
      <v-tab
        :to="`/project/${item.project_id}${
          $route.params.viewId ? `/views/${$route.params.viewId}` : ''
        }/templates/${item.id}/details`">{{ $t('template_details') }}
      </v-tab>
      <v-tab
        v-if="['terraform', 'tofu'].includes(item.app)"
        :to="`/project/${item.project_id}${
          $route.params.viewId ? `/views/${$route.params.viewId}` : ''
        }/templates/${item.id}/state`"
      >
        {{ $t('template_tf_workspaces') }}
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;" />

    <router-view
      class="mt-8"
      :template="item"
      :inventory="inventory"
      :environment="environment"
      :repositories="repositories"
      :premium-features="premiumFeatures"
      @update-template="loadData"
    ></router-view>
  </div>
</template>

<style lang="scss">

@import '~vuetify/src/styles/settings/_variables';

.TemplateView__description {
  font-size: 14px;
  margin-bottom: 12px;
}

@media #{map-get($display-breakpoints, 'md-and-up')} {
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
    premiumFeatures: Object,
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
      USER_PERMISSIONS,
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

    async stopAllTasks() {
      try {
        await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/templates/${this.itemId}/stop_all_tasks`,
          data: {
            force: true,
          },
          responseType: 'json',
        });

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'All running tasks have been requested to stop',
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.stopAllDialog = false;
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
