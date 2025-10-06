<template>
  <div v-if="folders && folders.length > 0">
    <!-- Debug info -->
    <div
      v-if="true"
      style="background: #f0f0f0; padding: 10px; margin: 10px 0; border-radius: 4px;"
    >
      <strong>Debug Info:</strong><br>
      Folders count: {{ folders ? folders.length : 'null' }}<br>
      Folders: {{ JSON.stringify(folders, null, 2) }}
    </div>
    <v-expansion-panels
      v-model="expandedFolders"
      multiple
      class="mt-4"
    >
      <v-expansion-panel
        v-for="folder in folders"
        :key="folder.name"
        class="template-folder"
      >
        <v-expansion-panel-header>
          <div class="d-flex align-center">
            <v-icon class="mr-3" :color="getFolderColor(folder.name)">
              {{ getFolderIcon(folder.name) }}
            </v-icon>
            <div>
              <div class="text-h6">{{ folder.name }}</div>
              <div class="text-caption text--secondary">
                {{ folder.count }} {{ folder.count === 1 ? 'template' : 'templates' }}
              </div>
            </div>
          </div>
        </v-expansion-panel-header>

        <v-expansion-panel-content>
          <v-data-table
            :headers="headers"
            :items="folder.templates"
            :items-per-page="Number.MAX_VALUE"
            hide-default-footer
            class="template-table"
          >
            <template v-slot:item.name="{ item }">
              <v-icon
                class="mr-3"
                small
              >
                {{ getAppIcon(item.app) }}
              </v-icon>

              <router-link
                :to="getTemplateUrl(item.id)"
                class="task-name-link"
              >
                {{ item.name }}
              </router-link>
            </template>

            <template v-slot:item.version="{ item }">
              <TaskLink
                v-if="item.last_task && item.last_task.tpl_type !== ''"
                :disabled="true"
                :status="item.last_task.status"
                :task-id="item.last_task.tpl_type === 'build'
                    ? item.last_task.id
                    : (item.last_task.build_task || {}).id"
                :label="item.last_task.tpl_type === 'build'
                    ? item.last_task.version
                    : (item.last_task.build_task || {}).version"
                :tooltip="item.last_task.tpl_type === 'build'
                    ? item.last_task.message
                    : (item.last_task.build_task || {}).message"
              />
              <div v-else>&mdash;</div>
            </template>

            <template v-slot:item.status="{ item }">
              <div class="mt-2 mb-2 d-flex" v-if="item.last_task != null">
                <TaskStatus :status="item.last_task.status"/>
              </div>
              <div v-else class="mt-3 mb-2 d-flex" style="color: gray;">
                {{ $t('notLaunched') }}
              </div>
            </template>

            <template v-slot:item.last_task="{ item }">
              <div class="mt-2 mb-2" v-if="item.last_task != null" style="line-height: 1">
                <TaskLink
                  :task-id="item.last_task.id"
                  :label="'#' + item.last_task.id"
                  :tooltip="item.last_task.message"
                />
              </div>
              <div v-else class="mt-3 mb-2 d-flex" style="color: gray;">
                {{ $t('notLaunched') }}
              </div>
            </template>

            <template v-slot:item.actions="{ item }">
              <div style="white-space: nowrap">
                <v-btn
                  icon
                  small
                  class="mr-1"
                  @click="runTask(item)"
                  :disabled="!canRunTask(item)"
                >
                  <v-icon>mdi-play</v-icon>
                </v-btn>

                <v-btn
                  icon
                  small
                  class="mr-1"
                  @click="editTemplate(item)"
                  v-if="can(USER_PERMISSIONS.manageProjectResources)"
                >
                  <v-icon>mdi-pencil</v-icon>
                </v-btn>

                <v-btn
                  icon
                  small
                  @click="deleteTemplate(item)"
                  v-if="can(USER_PERMISSIONS.manageProjectResources)"
                >
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
              </div>
            </template>
          </v-data-table>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>
  </div>

  <div v-else-if="!loading" class="text-center pa-8">
    <v-icon size="64" color="grey">mdi-folder-open</v-icon>
    <div class="text-h6 mt-4">No templates found</div>
    <div class="text--secondary">Create your first template to get started</div>
  </div>
</template>

<script>
import TaskLink from '@/components/TaskLink.vue';
import TaskStatus from '@/components/TaskStatus.vue';
import { TEMPLATE_TYPE_ICONS } from '@/lib/constants';

export default {
  name: 'TemplateFolderView',

  components: {
    TaskLink,
    TaskStatus,
  },

  props: {
    folders: {
      type: Array,
      default: () => [],
    },
    loading: {
      type: Boolean,
      default: false,
    },
    projectId: {
      type: Number,
      required: true,
    },
    viewId: {
      type: Number,
      default: null,
    },
  },

  data() {
    return {
      expandedFolders: [],
      TEMPLATE_TYPE_ICONS,
    };
  },

  computed: {
    headers() {
      return [
        {
          text: this.$i18n.t('name'),
          value: 'name',
          sortable: false,
        },
        {
          text: this.$i18n.t('version'),
          value: 'version',
          sortable: false,
          width: '120px',
        },
        {
          text: this.$i18n.t('status'),
          value: 'status',
          sortable: false,
          width: '120px',
        },
        {
          text: this.$i18n.t('lastTask'),
          value: 'last_task',
          sortable: false,
          width: '120px',
        },
        {
          text: this.$i18n.t('actions'),
          value: 'actions',
          sortable: false,
          width: '150px',
        },
      ];
    },
  },

  methods: {
    getTemplateUrl(templateId) {
      if (this.viewId) {
        return `/project/${this.projectId}/views/${this.viewId}/templates/${templateId}/details`;
      }
      return `/project/${this.projectId}/templates/${templateId}/details`;
    },

    getFolderIcon(folderName) {
      if (folderName === 'No Folder') {
        return 'mdi-folder-outline';
      }
      if (folderName.includes('STIG')) {
        return 'mdi-shield-check';
      }
      if (folderName.includes('CIS')) {
        return 'mdi-shield-star';
      }
      return 'mdi-folder';
    },

    getFolderColor(folderName) {
      if (folderName === 'No Folder') {
        return 'grey';
      }
      if (folderName.includes('STIG')) {
        return 'red';
      }
      if (folderName.includes('CIS')) {
        return 'green';
      }
      return 'primary';
    },

    getAppIcon(app) {
      const iconMap = {
        ansible: 'mdi-ansible',
        terraform: 'mdi-terraform',
        bash: 'mdi-bash',
        powershell: 'mdi-powershell',
        python: 'mdi-language-python',
        scc: 'mdi-shield-check',
      };
      return iconMap[app] || 'mdi-cog';
    },

    canRunTask(template) {
      // Add logic to determine if task can be run
      // For now, all tasks can be run
      return template && template.app;
    },

    runTask(template) {
      this.$emit('run-task', template);
    },

    editTemplate(template) {
      this.$emit('edit-template', template);
    },

    deleteTemplate(template) {
      this.$emit('delete-template', template);
    },

    can(permission) {
      // This should be passed from parent component
      return this.$parent.can ? this.$parent.can(permission) : true;
    },
  },
};
</script>

<style scoped>
.template-folder {
  margin-bottom: 8px;
}

.template-table {
  background: transparent;
}

.task-name-link {
  text-decoration: none;
  color: inherit;
}

.task-name-link:hover {
  text-decoration: underline;
}
</style>
