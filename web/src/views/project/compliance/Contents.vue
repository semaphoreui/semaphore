<template>
  <div class="contents-container">
    <!-- Header -->
    <div class="d-flex justify-space-between align-center mb-6">
      <div>
        <h2 class="text-h5 mb-2">{{ $t('compliance.contents') }}</h2>
        <p class="text-body-2 text--secondary">
          {{ $t('compliance.contentsDescription') }}
        </p>
      </div>
      <v-btn
        color="primary"
        @click="uploadDialog = true"
        :disabled="!canEdit"
      >
        <v-icon left>mdi-upload</v-icon>
        {{ $t('compliance.uploadContent') }}
      </v-btn>
    </div>

    <!-- Contents List -->
    <v-card>
      <v-data-table
        :headers="headers"
        :items="contents"
        :loading="loading"
        :items-per-page="25"
        class="elevation-1"
        @click:row="viewContent"
        style="cursor: pointer"
      >
        <!-- Name column -->
        <template v-slot:item.name="{ item }">
          <div class="d-flex align-center">
            <v-icon class="mr-2" color="primary">mdi-file-document</v-icon>
            <span class="font-weight-medium">{{ item.name }}</span>
          </div>
        </template>

        <!-- Source column -->
        <template v-slot:item.source="{ item }">
          <span v-if="item.source" class="text-body-2">{{ item.source }}</span>
          <span v-else class="text-body-2 text--secondary">
            {{ $t('compliance.unknownSource') }}
          </span>
        </template>

        <!-- Created column -->
        <template v-slot:item.created="{ item }">
          {{ formatDate(item.created) }}
        </template>

        <!-- Actions column -->
        <template v-slot:item.actions="{ item }">
          <v-btn
            icon
            small
            @click.stop="viewContent(item)"
            :title="$t('compliance.viewProfiles')"
          >
            <v-icon small>mdi-eye</v-icon>
          </v-btn>
          <v-btn
            v-if="canEdit"
            icon
            small
            color="error"
            @click.stop="deleteContent(item)"
            :title="$t('compliance.deleteContent')"
          >
            <v-icon small>mdi-delete</v-icon>
          </v-btn>
        </template>

        <!-- Empty state -->
        <template v-slot:no-data>
          <div class="text-center py-8">
            <v-icon size="64" color="grey lighten-1" class="mb-4">mdi-file-document-outline</v-icon>
            <h3 class="text-h6 mb-2">{{ $t('compliance.noContents') }}</h3>
            <p class="text-body-2 text--secondary mb-4">
              {{ $t('compliance.noContentsDescription') }}
            </p>
            <v-btn
              color="primary"
              @click="uploadDialog = true"
              :disabled="!canEdit"
            >
              <v-icon left>mdi-upload</v-icon>
              {{ $t('compliance.uploadFirstContent') }}
            </v-btn>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- Upload Dialog -->
    <v-dialog v-model="uploadDialog" max-width="600px" persistent>
      <v-card>
        <v-card-title>
          <span class="text-h6">{{ $t('compliance.uploadContent') }}</span>
        </v-card-title>

        <v-card-text>
          <v-form ref="uploadForm" v-model="uploadFormValid">
            <v-text-field
              v-model="uploadData.name"
              :label="$t('compliance.contentName')"
              :rules="[v => !!v || $t('validation.required')]"
              required
              class="mb-4"
            />

            <v-text-field
              v-model="uploadData.source"
              :label="$t('compliance.contentSource')"
              hint="e.g., SCAP Security Guide, Custom Policy"
              persistent-hint
              class="mb-4"
            />

            <v-file-input
              v-model="uploadData.file"
              :label="$t('compliance.selectFile')"
              :rules="[v => !!v || $t('validation.required')]"
              accept=".xml"
              show-size
              required
            />

            <v-alert
              type="info"
              class="mt-4"
              outlined
            >
              <div class="text-body-2">
                <strong>{{ $t('compliance.supportedFormats') }}:</strong><br>
                • SCAP DataStream (.xml)<br>
                • XCCDF Benchmark files
              </div>
            </v-alert>
          </v-form>
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn
            text
            @click="uploadDialog = false"
            :disabled="uploading"
          >
            {{ $t('cancel') }}
          </v-btn>
          <v-btn
            color="primary"
            @click="uploadContent"
            :loading="uploading"
            :disabled="!uploadFormValid"
          >
            {{ $t('compliance.upload') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Content Details Dialog -->
    <v-dialog v-model="contentDialog" max-width="800px">
      <v-card v-if="selectedContent">
        <v-card-title>
          <span class="text-h6">{{ selectedContent.name }}</span>
          <v-spacer />
          <v-btn icon @click="contentDialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>

        <v-card-text>
          <v-row>
            <v-col cols="12" md="6">
              <v-list dense>
                <v-list-item>
                  <v-list-item-content>
                    <v-list-item-title>{{ $t('compliance.name') }}</v-list-item-title>
                    <v-list-item-subtitle>{{ selectedContent.name }}</v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
                <v-list-item v-if="selectedContent.source">
                  <v-list-item-content>
                    <v-list-item-title>{{ $t('compliance.source') }}</v-list-item-title>
                    <v-list-item-subtitle>{{ selectedContent.source }}</v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
                <v-list-item>
                  <v-list-item-content>
                    <v-list-item-title>{{ $t('compliance.uploaded') }}</v-list-item-title>
                    <v-list-item-subtitle>
                      {{ formatDate(selectedContent.created) }}
                    </v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
              </v-list>
            </v-col>
            <v-col cols="12" md="6">
              <div class="text-h6 mb-3">{{ $t('compliance.availableProfiles') }}</div>
              <v-list dense>
                <v-list-item v-for="profile in selectedContent.profiles" :key="profile.id">
                  <v-list-item-icon>
                    <v-icon color="primary">mdi-shield-check</v-icon>
                  </v-list-item-icon>
                  <v-list-item-content>
                    <v-list-item-title>{{ profile.title }}</v-list-item-title>
                    <v-list-item-subtitle v-if="profile.severity">
                      {{ $t('compliance.severity') }}: {{ profile.severity }}
                    </v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
              </v-list>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  name: 'ComplianceContents',

  props: {
    projectId: {
      type: Number,
      required: true,
    },
    contents: {
      type: Array,
      default: () => [],
    },
  },

  data() {
    return {
      loading: false,
      uploadDialog: false,
      contentDialog: false,
      uploadFormValid: false,
      uploading: false,
      selectedContent: null,
      uploadData: {
        name: '',
        source: '',
        file: null,
      },
      headers: [
        { text: this.$t('compliance.name'), value: 'name', sortable: true },
        { text: this.$t('compliance.source'), value: 'source', sortable: true },
        { text: this.$t('compliance.uploaded'), value: 'created', sortable: true },
        {
          text: this.$t('actions'), value: 'actions', sortable: false, width: '120px',
        },
      ],
    };
  },

  computed: {
    canEdit() {
      return this.$props.userPermissions?.includes('manageProjectResources') || false;
    },
  },

  async mounted() {
    await this.loadProfiles();
  },

  methods: {
    async loadProfiles() {
      // Load profiles for each content
      const contentsWithProfiles = [...this.contents];
      const profilePromises = contentsWithProfiles.map(async (content, index) => {
        try {
          const response = await axios.get(
            `/api/project/${this.projectId}/compliance/contents/${content.id}/profiles`,
          );
          contentsWithProfiles[index].profiles = response.data;
        } catch (error) {
          console.error(`Failed to load profiles for content ${content.id}:`, error);
          contentsWithProfiles[index].profiles = [];
        }
      });
      await Promise.all(profilePromises);
      this.$emit('update-contents', contentsWithProfiles);
    },

    async uploadContent() {
      if (!this.$refs.uploadForm.validate()) {
        return;
      }

      this.uploading = true;
      try {
        const formData = new FormData();
        formData.append('file', this.uploadData.file);
        formData.append('name', this.uploadData.name);
        formData.append('source', this.uploadData.source);

        await axios.post(`/api/project/${this.projectId}/compliance/contents`, formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        });

        this.uploadDialog = false;
        this.resetUploadForm();
        this.$emit('refresh');
        this.$store.dispatch('showSuccess', this.$t('compliance.contentUploaded'));
      } catch (error) {
        this.$store.dispatch('showError', getErrorMessage(error));
      } finally {
        this.uploading = false;
      }
    },

    async deleteContent(content) {
      // eslint-disable-next-line no-alert
      if (!window.confirm(this.$t('compliance.confirmDeleteContent', { name: content.name }))) {
        return;
      }

      try {
        await axios.delete(`/api/project/${this.projectId}/compliance/contents/${content.id}`);
        this.$emit('refresh');
        this.$store.dispatch('showSuccess', this.$t('compliance.contentDeleted'));
      } catch (error) {
        this.$store.dispatch('showError', getErrorMessage(error));
      }
    },

    viewContent(content) {
      this.selectedContent = content;
      this.contentDialog = true;
    },

    resetUploadForm() {
      this.uploadData = {
        name: '',
        source: '',
        file: null,
      };
      this.$refs.uploadForm.reset();
    },

    formatDate(date) {
      return new Date(date).toLocaleDateString();
    },
  },
};
</script>

<style lang="scss" scoped>
.contents-container {
  // Add any custom styles here
}
</style>
