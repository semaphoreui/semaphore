<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="!isLoaded">
    <v-progress-linear
      indeterminate
      color="primary darken-2"
    ></v-progress-linear>
  </div>
  <div v-else>

    <NewTaskDialog
      v-model="newTaskDialog"
      :project-id="projectId"
      :template-id="itemId"
      :template-alias="item.name"
      :template-type="item.type"
      :template-app="item.app"
    />

    <EditTemplateDialogue
        v-model="editDialog"
        :project-id="projectId"
        :item-app="item.app"
        :item-id="itemId"
        @save="loadData()"
    ></EditTemplateDialogue>

    <EditTemplateDialogue
        v-model="copyDialog"
        :project-id="projectId"
        :item-app="item.app"
        item-id="new"
        :source-item-id="itemId"
        @save="onTemplateCopied"
    ></EditTemplateDialogue>

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

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title class="breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="viewId
              ? `/project/${projectId}/views/${viewId}/templates/`
              : `/project/${projectId}/templates/`"
        >
          {{ $t('taskTemplates2') }}
        </router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ item.name }}</span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn color="primary" depressed class="mr-3" @click="newTaskDialog = true">
        {{ $t(TEMPLATE_TYPE_ACTION_TITLES[item.type]) }}
      </v-btn>

      <v-btn
        icon
        color="error"
        @click="askDelete()"
        v-if="canUpdate"
      >
        <v-icon>mdi-delete</v-icon>
      </v-btn>

      <v-btn
        icon
        @click="copyDialog = true"
        v-if="canUpdate"
      >
        <v-icon>mdi-content-copy</v-icon>
      </v-btn>

      <v-btn
        icon
        @click="editDialog = true"
        v-if="canUpdate"
      >
        <v-icon>mdi-pencil</v-icon>
      </v-btn>
    </v-toolbar>

    <v-container>
      <v-alert
        text
        type="info"
        class="mb-0 ml-4 mr-4 mb-2"
        v-if="item.description"
      >{{ item.description }}
      </v-alert>

      <v-row>
        <v-col>
          <v-list subheader dense>
            <v-list-item>
              <v-list-item-icon>
                <v-icon>mdi-book-play</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>{{ $t('playbook') }}</v-list-item-title>
                <v-list-item-subtitle>{{ item.playbook }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list subheader dense>
            <v-list-item>
              <v-list-item-icon>
                <v-icon>{{ TEMPLATE_TYPE_ICONS[item.type] }}</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>{{ $t('type') }}</v-list-item-title>
                <v-list-item-subtitle>{{ $t(TEMPLATE_TYPE_TITLES[item.type]) }}
                </v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list subheader dense>
            <v-list-item>
              <v-list-item-icon>
                <v-icon>mdi-monitor</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>{{ $t('inventory') }}</v-list-item-title>
                <v-list-item-subtitle>
                  {{ (inventory.find((x) => x.id === item.inventory_id) || {name: '—'}).name }}
                </v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list subheader dense>
            <v-list-item>
              <v-list-item-icon>
                <v-icon>mdi-code-braces</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>{{ $t('environment') }}</v-list-item-title>
                <v-list-item-subtitle>
                  {{ environment.find((x) => x.id === item.environment_id).name }}
                </v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list subheader dense>
            <v-list-item>
              <v-list-item-icon>
                <v-icon>mdi-git</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>{{ $t('repository2') }}</v-list-item-title>
                <v-list-item-subtitle>
                  {{ repositories.find((x) => x.id === item.repository_id).name }}
                </v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
      </v-row>
    </v-container>

    <TaskList :template="item"/>
  </div>
</template>
<style lang="scss">

</style>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import YesNoDialog from '@/components/YesNoDialog.vue';
import TaskList from '@/components/TaskList.vue';
import {
  TEMPLATE_TYPE_ACTION_TITLES,
  TEMPLATE_TYPE_ICONS,
  TEMPLATE_TYPE_TITLES,
  USER_PERMISSIONS,
} from '@/lib/constants';
import ObjectRefsDialog from '@/components/ObjectRefsDialog.vue';
import NewTaskDialog from '@/components/NewTaskDialog.vue';
import EditTemplateDialogue from '@/components/EditTemplateDialog.vue';
import PermissionsCheck from '@/components/PermissionsCheck';

export default {
  components: {
    YesNoDialog,
    TaskList,
    ObjectRefsDialog,
    NewTaskDialog,
    EditTemplateDialogue,
  },

  props: {
    projectId: Number,
    userPermissions: Number,
  },

  mixins: [PermissionsCheck],

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
      USER_PERMISSIONS,
    };
  },

  computed: {
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
      return this.item
        && this.inventory
        && this.environment
        && this.repositories;
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

    async askDelete() {
      this.itemRefs = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates/${this.itemId}/refs`,
        responseType: 'json',
      })).data;

      if (this.itemRefs.templates.length > 0) {
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
      this.item = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates/${this.itemId}`,
        responseType: 'json',
      })).data;

      this.inventory = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/inventory`,
        responseType: 'json',
      })).data;

      this.environment = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/environment`,
        responseType: 'json',
      })).data;

      this.repositories = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/repositories`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
