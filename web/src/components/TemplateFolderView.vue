<template>
  <div v-if="folders && folders.length > 0">
    <div class="mt-4">
      <v-card
        v-for="folder in folders"
        :key="folder.name"
        class="template-folder mb-4"
        elevation="2"
      >
        <v-card-title
          @click.stop="toggleFolder(folder.name)"
          class="folder-header clickable"
          :class="{ 'folder-expanded': expandedFolders.includes(folder.name) }"
        >
          <div class="d-flex align-center">
            <v-icon class="mr-3" :color="getFolderColor(folder.name)">
              {{ getFolderIcon(folder.name) }}
            </v-icon>
            <div class="flex-grow-1">
              <div class="text-h6">{{ folder.name }}</div>
              <div class="text-caption text--secondary">
                {{ folder.count }} {{ folder.count === 1 ? 'template' : 'templates' }}
              </div>
            </div>
            <v-icon class="folder-arrow">
              {{ expandedFolders.includes(folder.name) ? 'mdi-chevron-up' : 'mdi-chevron-down' }}
            </v-icon>
          </div>
        </v-card-title>

        <v-expand-transition>
          <div v-if="expandedFolders.includes(folder.name)">
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
          </div>
        </v-expand-transition>
      </v-card>
    </div>
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

  watch: {
    folders: {
      handler(newFolders) {
        console.log('Folders updated:', newFolders);
        if (newFolders && newFolders.length > 0) {
          console.log('First folder templates:', newFolders[0].templates);
        }
      },
      immediate: true,
    },
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

    toggleFolder(folderName) {
      console.log('Toggling folder:', folderName);
      console.log('Current expanded folders:', this.expandedFolders);
      const index = this.expandedFolders.indexOf(folderName);
      if (index > -1) {
        this.expandedFolders.splice(index, 1);
        console.log('Collapsed folder:', folderName);
      } else {
        this.expandedFolders.push(folderName);
        console.log('Expanded folder:', folderName);
      }
      console.log('New expanded folders:', this.expandedFolders);
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

.folder-header {
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.folder-header:hover {
  background-color: rgba(0, 0, 0, 0.04);
}

.folder-header.folder-expanded {
  background-color: rgba(0, 0, 0, 0.08);
}

.folder-arrow {
  transition: transform 0.2s ease;
}

.clickable {
  user-select: none;
}
</style>
