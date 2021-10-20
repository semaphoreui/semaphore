<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="!isLoaded">
    <v-progress-linear
        indeterminate
        color="primary darken-2"
    ></v-progress-linear>
  </div>
  <div v-else>

    <EditDialog
        :max-width="700"
        v-model="editDialog"
        save-button-text="Save"
        title="Edit Template"
        @save="loadData()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TemplateForm
            :project-id="projectId"
            :item-id="itemId"
            @save="onSave"
            @error="onError"
            :need-save="needSave"
            :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <EditDialog
        v-model="copyDialog"
        save-button-text="Create"
        title="New Template"
        @save="onTemplateCopied"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TemplateForm
            :project-id="projectId"
            item-id="new"
            :source-item-id="itemId"
            @save="onSave"
            @error="onError"
            :need-save="needSave"
            :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <YesNoDialog
        title="Delete template"
        text="Are you really want to delete this template?"
        v-model="deleteDialog"
        @yes="remove()"
    />

    <EditDialog
        v-model="newTaskDialog"
        :save-button-text="'Re' + getActionButtonTitle()"
        @save="onTaskCreated"
    >
      <template v-slot:title={}>
        <v-icon class="mr-4">{{ TEMPLATE_TYPE_ICONS[item.type] }}</v-icon>
        <span class="breadcrumbs__item">{{ item.alias }}</span>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">New Task</span>
      </template>

      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TaskForm
            :project-id="projectId"
            item-id="new"
            :template-id="itemId"
            @save="onSave"
            @error="onError"
            :need-save="needSave"
            :need-reset="needReset"
            :commit-hash="sourceTask == null ? null : sourceTask.commit_hash"
            :commit-message="sourceTask == null ? null : sourceTask.commit_message"
            :build_task="sourceTask == null ? null : sourceTask.build_task"
        />
      </template>
    </EditDialog>

    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title class="breadcrumbs">
        <router-link
            class="breadcrumbs__item breadcrumbs__item--link"
            :to="`/project/${projectId}/templates/`"
        >Task Templates
        </router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ item.alias }}</span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn
          icon
          color="error"
          @click="deleteDialog = true"
      >
        <v-icon>mdi-delete</v-icon>
      </v-btn>

      <v-btn
          icon
          color="black"
          @click="copyDialog = true"
      >
        <v-icon>mdi-content-copy</v-icon>
      </v-btn>

      <v-btn
          icon
          color="black"
          @click="editDialog = true"
      >
        <v-icon>mdi-pencil</v-icon>
      </v-btn>
    </v-toolbar>

    <v-container class="pa-0">

      <v-alert
          text
          type="info"
          class="mb-0 ml-4 mr-4 mb-2"
          v-if="item.description"
      >{{ item.description }}
      </v-alert>

      <v-row>
        <v-col>
          <v-list two-line subheader>
            <v-list-item>
              <v-list-item-icon>
                <v-icon>mdi-book-play</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>Playbook</v-list-item-title>
                <v-list-item-subtitle>{{ item.playbook }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list two-line subheader>
            <v-list-item>
              <v-list-item-icon>
                <v-icon>{{ TEMPLATE_TYPE_ICONS[item.type] }}</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>Type</v-list-item-title>
                <v-list-item-subtitle>{{ TEMPLATE_TYPE_TITLES[item.type] }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list two-line subheader>
            <v-list-item>
              <v-list-item-icon>
                <v-icon>mdi-monitor</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>Inventory</v-list-item-title>
                <v-list-item-subtitle>
                  {{ inventory.find((x) => x.id === item.inventory_id).name }}
                </v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list two-line subheader>
            <v-list-item>
              <v-list-item-icon>
                <v-icon>mdi-code-braces</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>Environment</v-list-item-title>
                <v-list-item-subtitle>
                  {{ environment.find((x) => x.id === item.environment_id).name }}
                </v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list two-line subheader>
            <v-list-item>
              <v-list-item-icon>
                <v-icon>mdi-git</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>Repository</v-list-item-title>
                <v-list-item-subtitle>
                  {{ repositories.find((x) => x.id === item.repository_id).name }}
                </v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
      </v-row>
    </v-container>

    <v-data-table
        :headers="headers"
        :items="tasks"
        :footer-props="{ itemsPerPageOptions: [20] }"
        class="mt-0"
    >
      <template v-slot:item.id="{ item }">
        <div style="display: flex; justify-content: left; align-items: center;">
          <a @click="showTaskLog(item.id)">#{{ item.id }}</a>
          <v-tooltip color="black" right max-width="350" transition="fade-transition">
            <template v-slot:activator="{ on, attrs }">
              <v-icon
                  v-bind="attrs"
                  v-on="on"
                  v-if="item.message"
                  class="ml-1"
                  color="gray"
                  small
              >mdi-information
              </v-icon>
            </template>
            <span>{{ item.message }}</span>
          </v-tooltip>
        </div>
      </template>

      <template v-slot:item.version="{ item }">
        <div v-if="item.version != null || item.build_task != null">
          <v-icon
              small
              class="mr-2"
              :color="item.status === 'success' ? 'success' : 'red'"
          >mdi-{{ item.status === 'success' ? 'check' : 'close' }}
          </v-icon>

          <span v-if="item.version">{{ item.version }}</span>

          <v-tooltip
              v-else
              color="black"
              right
              max-width="350"
              transition="fade-transition"
          >
            <template v-slot:activator="{ on, attrs }">
              <a
                  @click="showTaskLog(item.build_task_id)"
                  v-bind="attrs"
                  v-on="on"
                  style="border-bottom: 1px dashed gray; text-decoration: none !important;"
              >{{ item.build_task.version }}</a>
            </template>
            <span>{{ item.build_task.message }}</span>
          </v-tooltip>

        </div>
        <div v-else>&mdash;</div>
      </template>

      <template v-slot:item.status="{ item }">
        <TaskStatus :status="item.status"/>
      </template>

      <template v-slot:item.start="{ item }">
        {{ item.start | formatDate }}
      </template>

      <template v-slot:item.end="{ item }">
        {{ [item.start, item.end] | formatMilliseconds }}
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn text color="black" class="pl-1 pr-2" @click="createTask(item)">
          <v-icon class="pr-1">mdi-replay</v-icon>
          Re{{ getActionButtonTitle() }}
        </v-btn>
      </template>
    </v-data-table>
  </div>
</template>
<style lang="scss">

</style>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import YesNoDialog from '@/components/YesNoDialog.vue';
import EditDialog from '@/components/EditDialog.vue';
import TemplateForm from '@/components/TemplateForm.vue';
import TaskForm from '@/components/TaskForm.vue';
import TaskStatus from '@/components/TaskStatus.vue';
import { TEMPLATE_TYPE_ACTION_TITLES, TEMPLATE_TYPE_ICONS, TEMPLATE_TYPE_TITLES } from '../../lib/constants';

export default {
  components: {
    YesNoDialog, EditDialog, TemplateForm, TaskStatus, TaskForm,
  },

  props: {
    projectId: Number,
  },

  data() {
    return {
      TEMPLATE_TYPE_ICONS,
      TEMPLATE_TYPE_TITLES,
      TEMPLATE_TYPE_ACTION_TITLES,
      headers: [
        {
          text: 'Task ID',
          value: 'id',
          sortable: false,
        },
        {
          text: 'Version',
          value: 'version',
          sortable: false,
        },
        {
          text: 'Status',
          value: 'status',
          sortable: false,
        },
        {
          text: 'User',
          value: 'user_name',
          sortable: false,
        },
        {
          text: 'Start',
          value: 'start',
          sortable: false,
        },
        {
          text: 'Duration',
          value: 'end',
          sortable: false,
        },
        {
          text: 'Actions',
          value: 'actions',
          sortable: false,
          width: '0%',
        },
      ],
      tasks: null,
      item: null,
      inventory: null,
      environment: null,
      repositories: null,
      deleteDialog: null,
      editDialog: null,
      copyDialog: null,
      taskLogDialog: null,
      taskId: null,
      newTaskDialog: null,
      sourceTask: null,
    };
  },

  computed: {
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
          && this.tasks
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
    getActionButtonTitle() {
      return TEMPLATE_TYPE_ACTION_TITLES[this.item.type];
    },

    onTaskCreated(e) {
      EventBus.$emit('i-show-task', {
        taskId: e.item.id,
      });
    },

    createTask(task) {
      this.sourceTask = task;
      this.newTaskDialog = true;
    },

    showTaskLog(taskId) {
      EventBus.$emit('i-show-task', {
        taskId,
      });
    },

    showDrawer() {
      EventBus.$emit('i-show-drawer');
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
          text: `Template "${this.item.alias}" deleted`,
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

      this.tasks = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates/${this.itemId}/tasks/last`,
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
