<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div class="pb-3">
    <v-row>
      <v-col cols="12" md="6">
        <v-card
          v-if="template"
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
        >
          <v-card-title>Template info</v-card-title>
          <v-card-text>
            <v-simple-table class="TaskDetails__table">
              <template v-slot:default>
                <tbody>
                <tr>
                  <td><b>App</b></td>
                  <td>{{ getAppTitle(template.app) }}</td>
                </tr>
                <tr>
                  <td><b>Template</b></td>
                  <td>
                    <RouterLink :to="`/project/${projectId}/templates/${template.id}`">
                      {{ template.name }}
                    </RouterLink>
                  </td>
                </tr>
                </tbody>
              </template>
            </v-simple-table>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="6">
        <v-card
          v-if="item.commit_hash"
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
        >
          <v-card-title>Commit info</v-card-title>

          <v-card-text>
            <v-simple-table class="TaskDetails__table">
              <template v-slot:default>
                <tbody>
                <tr>
                  <td><b>Message</b></td>
                  <td>{{ item.commit_message }}</td>
                </tr>
                <tr>
                  <td><b>Hash</b></td>
                  <td>{{ item.commit_hash }}</td>
                </tr>
                </tbody>
              </template>
            </v-simple-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12" md="6">
        <v-card
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
          class="mb-5"
        >
          <v-card-title>Running info</v-card-title>
          <v-card-text>
            <v-simple-table class="pa-0 TaskDetails__table">
              <template v-slot:default>
                <tbody>
                <tr>
                  <td><b>Message</b></td>
                  <td>{{ item.message || '—' }}</td>
                </tr>
                <tr v-if="item.user_id != null">
                  <td><b>{{ $t('author') }}</b></td>
                  <td>{{ user?.name || '—' }}</td>
                </tr>
                <tr v-else-if="item.integration_id != null">
                  <td><b>{{ $t('integration') }}</b></td>
                  <td>{{ item.integration_id }}</td>
                </tr>
                <tr v-else-if="item.schedule_id != null">
                  <td><b>{{ $t('schedule') }}</b></td>
                  <td>{{ item.schedule_id }}</td>
                </tr>
                <tr>
                  <td><b>{{ $t('created') }}</b></td>
                  <td>{{ item.created | formatDate }}</td>
                </tr>
                <tr>
                  <td><b>{{ $t('started') }}</b></td>
                  <td>{{ item.start | formatDate }}</td>
                </tr>
                <tr>
                  <td><b>{{ $t('end') }}</b></td>
                  <td>{{ item.end | formatDate }}</td>
                </tr>
                <tr>
                  <td><b>{{ $t('duration') }}</b></td>
                  <td>{{ [item.start, item.end] | formatMilliseconds }}</td>
                </tr>
                </tbody>
              </template>
            </v-simple-table>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card
          v-if="item?.params"
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
          class="mb-5"
        >
          <v-card-title>Task parameters</v-card-title>
          <v-card-text>
            <v-simple-table class="pa-0 TaskDetails__table">
              <template v-slot:default>
                <tbody>
                <tr>
                  <td><b>Branch</b></td>
                  <td>
                    {{ item.get_branch || '—' }}
                  </td>
                </tr>
                <tr>
                  <td><b>Limit</b></td>
                  <td>
                    {{ item.params.limit ? 'Yes' : 'No' }}
                  </td>
                </tr>
                <tr>
                  <td><b>Debug</b></td>
                  <td>
                    {{ item.params.debug ? 'Yes' : 'No' }}
                  </td>
                </tr>
                <tr>
                  <td><b>Debug level</b></td>
                  <td>{{ item.params.debug_level || '—' }}</td>
                </tr>
                <tr>
                  <td><b>Diff</b> <code>--diff</code></td>
                  <td>{{ item.params.diff ? 'Yes' : 'No' }}</td>
                </tr>
                <tr>
                  <td><b>Dry run</b> <code>--check</code></td>
                  <td>{{ item.params.dry_run ? 'Yes' : 'No' }}</td>
                </tr>
                <tr>
                  <td><b>Environment</b></td>
                  <td>
                    {{ !item.environment || item.environment === '{}' ? '—' : item.environment }}
                  </td>
                </tr>
                </tbody>
              </template>
            </v-simple-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row v-if="taskFiles.length > 0">
      <v-col cols="12">
        <v-card
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
          class="mb-5"
        >
          <v-card-title>
            <v-icon class="mr-2">mdi-file-download</v-icon>
            Task Files
          </v-card-title>
          <v-card-text>
            <v-data-table
              :headers="fileHeaders"
              :items="taskFiles"
              :items-per-page="10"
              class="elevation-0"
              hide-default-footer
            >
              <template v-slot:item.filename="{ item }">
                <div class="d-flex align-center">
                  <v-icon class="mr-2">{{ getFileIcon(item.mime_type) }}</v-icon>
                  {{ item.filename }}
                </div>
              </template>
              <template v-slot:item.file_size="{ item }">
                {{ formatFileSize(item.file_size) }}
              </template>
              <template v-slot:item.created="{ item }">
                {{ item.created | formatDate }}
              </template>
              <template v-slot:item.actions="{ item }">
                <v-btn
                  small
                  color="primary"
                  @click="downloadFile(item)"
                >
                  <v-icon left>mdi-download</v-icon>
                  Download
                </v-btn>
                <v-btn
                  small
                  color="error"
                  class="ml-2"
                  @click="deleteFile(item)"
                >
                  <v-icon left>mdi-delete</v-icon>
                  Delete
                </v-btn>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<style lang="scss">
.TaskDetails__table {
  background-color: transparent !important;
  .v-data-table__wrapper {
    padding-left: 0 !important;
    padding-right: 0 !important;
  }
}

</style>

<script>

import ProjectMixin from '@/components/ProjectMixin';
import AppsMixin from '@/components/AppsMixin';

export default {
  props: {
    item: Object,
    user: Object,
    projectId: Number,
  },

  mixins: [ProjectMixin, AppsMixin],

  data() {
    return {
      template: null,
      taskFiles: [],
      fileHeaders: [
        { text: 'Filename', value: 'filename' },
        { text: 'Size', value: 'file_size' },
        { text: 'Type', value: 'mime_type' },
        { text: 'Created', value: 'created' },
        { text: 'Actions', value: 'actions', sortable: false },
      ],
    };
  },

  watch: {
    async item() {
      if (this.item?.template_id !== this.template?.id) {
        await this.loadData();
      }
    },
  },

  computed: {},

  async created() {
    await this.loadData();
  },

  methods: {
    async loadData() {
      this.template = await this.loadProjectResource('templates', this.item.template_id);
      await this.loadTaskFiles();
    },

    async loadTaskFiles() {
      try {
        const response = await this.$http.get(
          `/api/project/${this.projectId}/tasks/${this.item.id}/files`,
        );
        this.taskFiles = response.data || [];
      } catch (error) {
        console.error('Failed to load task files:', error);
        this.taskFiles = [];
      }
    },

    getFileIcon(mimeType) {
      if (!mimeType) return 'mdi-file';

      if (mimeType.startsWith('text/')) return 'mdi-file-document';
      if (mimeType.startsWith('image/')) return 'mdi-file-image';
      if (mimeType.startsWith('video/')) return 'mdi-file-video';
      if (mimeType.startsWith('audio/')) return 'mdi-file-music';
      if (mimeType.includes('pdf')) return 'mdi-file-pdf-box';
      if (mimeType.includes('zip') || mimeType.includes('tar') || mimeType.includes('gzip')) {
        return 'mdi-file-archive';
      }
      if (mimeType.includes('json')) return 'mdi-code-json';
      if (mimeType.includes('xml')) return 'mdi-file-xml';
      if (mimeType.includes('csv')) return 'mdi-file-delimited';

      return 'mdi-file';
    },

    formatFileSize(bytes) {
      if (!bytes) return '0 B';

      const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(1024));
      return `${Math.round((bytes / (1024 ** i)) * 100) / 100} ${sizes[i]}`;
    },

    async downloadFile(file) {
      try {
        const response = await this.$http.get(
          `/api/project/${this.projectId}/tasks/${this.item.id}/files/${file.id}`,
          {
            responseType: 'blob',
          },
        );

        const url = window.URL.createObjectURL(new Blob([response.data]));
        const link = document.createElement('a');
        link.href = url;
        link.setAttribute('download', file.filename);
        document.body.appendChild(link);
        link.click();
        link.remove();
        window.URL.revokeObjectURL(url);
      } catch (error) {
        console.error('Failed to download file:', error);
        this.$toast?.error('Failed to download file');
      }
    },

    async deleteFile(file) {
      if (!window.confirm(`Are you sure you want to delete "${file.filename}"?`)) {
        return;
      }

      try {
        await this.$http.delete(
          `/api/project/${this.projectId}/tasks/${this.item.id}/files/${file.id}`,
        );
        await this.loadTaskFiles(); // Reload the list
        this.$toast?.success('File deleted successfully');
      } catch (error) {
        console.error('Failed to delete file:', error);
        this.$toast?.error('Failed to delete file');
      }
    },
  },
};
</script>
