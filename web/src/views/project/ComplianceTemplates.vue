<template>
  <div>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Compliance Templates</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="importTasks"
        :loading="importing"
        :disabled="!canImport"
      >
        <v-icon left>mdi-download</v-icon>
        Import Tasks
      </v-btn>
    </v-toolbar>

    <v-container>
      <!-- Project Compliance Info -->
      <v-card class="mb-6" v-if="project.compliance_framework">
        <v-card-title>
          <v-icon class="mr-2">mdi-shield-check</v-icon>
          Compliance Framework
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12" md="4">
              <div class="text-subtitle-2">Framework</div>
              <div class="text-body-1">{{ project.compliance_framework }}</div>
            </v-col>
            <v-col cols="12" md="4">
              <div class="text-subtitle-2">Operating System</div>
              <div class="text-body-1">{{ project.compliance_os }}</div>
            </v-col>
            <v-col cols="12" md="4">
              <div class="text-subtitle-2">STIG Enabled</div>
              <div class="text-body-1">
                <v-chip :color="project.enable_stig ? 'success' : 'default'" small>
                  {{ project.enable_stig ? 'Yes' : 'No' }}
                </v-chip>
              </div>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- Import Tasks Section -->
      <v-card class="mb-6" v-if="!project.compliance_framework">
        <v-card-title>
          <v-icon class="mr-2">mdi-plus</v-icon>
          Import Compliance Tasks
        </v-card-title>
        <v-card-text>
          <v-form ref="importForm" v-model="importFormValid">
            <v-row>
              <v-col cols="12" md="6">
                <v-select
                  v-model="importData.framework"
                  :items="supportedFrameworks"
                  label="Compliance Framework"
                  :rules="[v => !!v || 'Framework is required']"
                  required
                  outlined
                  dense
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-select
                  v-model="importData.os"
                  :items="supportedOS"
                  label="Operating System"
                  :rules="[v => !!v || 'OS is required']"
                  required
                  outlined
                  dense
                />
              </v-col>
            </v-row>
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="primary"
            @click="importTasks"
            :loading="importing"
            :disabled="!importFormValid"
          >
            Import Tasks
          </v-btn>
        </v-card-actions>
      </v-card>

      <!-- Templates List -->
      <v-card>
        <v-card-title>
          <v-icon class="mr-2">mdi-format-list-bulleted</v-icon>
          Compliance Templates ({{ templates.length }})
        </v-card-title>
        <v-card-text>
          <v-data-table
            :headers="headers"
            :items="templates"
            :loading="loading"
            :items-per-page="25"
            class="elevation-1"
          >
            <template v-slot:item.name="{ item }">
              <div class="d-flex align-center">
                <v-icon class="mr-2" color="primary">mdi-shield-check</v-icon>
                <span class="font-weight-medium">{{ item.name }}</span>
              </div>
            </template>

            <template v-slot:item.description="{ item }">
              <span class="text-body-2">{{ item.description || 'No description' }}</span>
            </template>

            <template v-slot:item.app="{ item }">
              <v-chip small :color="getAppColor(item.app)">
                {{ item.app }}
              </v-chip>
            </template>

            <template v-slot:item.created="{ item }">
              {{ formatDate(item.created) }}
            </template>

            <template v-slot:item.actions="{ item }">
              <v-btn
                icon
                small
                @click="viewTemplate(item)"
                color="primary"
              >
                <v-icon>mdi-eye</v-icon>
              </v-btn>
              <v-btn
                icon
                small
                @click="createTask(item)"
                color="success"
              >
                <v-icon>mdi-play</v-icon>
              </v-btn>
            </template>
          </v-data-table>
        </v-card-text>
      </v-card>
    </v-container>

    <!-- Template Details Dialog -->
    <v-dialog v-model="templateDialog" max-width="800px">
      <v-card v-if="selectedTemplate">
        <v-card-title>
          <v-icon class="mr-2">mdi-shield-check</v-icon>
          {{ selectedTemplate.name }}
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12" md="6">
              <div class="text-subtitle-2">Description</div>
              <div class="text-body-1">{{ selectedTemplate.description || 'No description' }}</div>
            </v-col>
            <v-col cols="12" md="6">
              <div class="text-subtitle-2">Application</div>
              <v-chip :color="getAppColor(selectedTemplate.app)">
                {{ selectedTemplate.app }}
              </v-chip>
            </v-col>
          </v-row>
          <v-divider class="my-4"></v-divider>
          <div class="text-subtitle-2 mb-2">Playbook Content</div>
          <v-textarea
            :value="selectedTemplate.playbook"
            readonly
            outlined
            rows="15"
            class="font-family-monospace"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="createTask(selectedTemplate)">
            <v-icon left>mdi-play</v-icon>
            Create Task
          </v-btn>
          <v-btn @click="templateDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import axios from 'axios';

// Simple date formatter
function formatDate(dateString) {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
}

export default {
  name: 'ComplianceTemplates',

  data() {
    return {
      loading: false,
      importing: false,
      importFormValid: false,
      templateDialog: false,
      selectedTemplate: null,
      templates: [],
      project: {},
      importData: {
        framework: '',
        os: '',
      },
      supportedFrameworks: ['CIS', 'STIG'],
      supportedOS: [
        'RHEL7', 'RHEL8', 'RHEL9',
        'UBUNTU18', 'UBUNTU20', 'UBUNTU22', 'UBUNTU24',
        'DEBIAN11', 'DEBIAN12',
        'AMAZON2', 'AMAZON2023',
        'SUSE15',
        'WINDOWS10', 'WINDOWS11', 'WINDOWS2016', 'WINDOWS2019', 'WINDOWS2022',
      ],
      headers: [
        { text: 'Name', value: 'name', sortable: true },
        { text: 'Description', value: 'description', sortable: false },
        { text: 'Application', value: 'app', sortable: true },
        { text: 'Created', value: 'created', sortable: true },
        {
          text: 'Actions', value: 'actions', sortable: false, width: '120px',
        },
      ],
    };
  },

  computed: {
    canImport() {
      return this.project.compliance_framework && this.project.compliance_os;
    },
  },

  async mounted() {
    await this.loadProject();
    await this.loadTemplates();
  },

  methods: {
    async loadProject() {
      try {
        const response = await axios.get(`/api/project/${this.projectId}`);
        this.project = response.data;
      } catch (error) {
        console.error('Failed to load project:', error);
        this.$store.dispatch('showError', 'Failed to load project details');
      }
    },

    async loadTemplates() {
      this.loading = true;
      try {
        const response = await axios.get(`/api/project/${this.projectId}/lockdown/templates`);
        this.templates = response.data.templates || [];
      } catch (error) {
        console.error('Failed to load compliance templates:', error);
        this.$store.dispatch('showError', 'Failed to load compliance templates');
      } finally {
        this.loading = false;
      }
    },

    async importTasks() {
      this.importing = true;
      try {
        const data = this.project.compliance_framework
          ? { framework: this.project.compliance_framework, os: this.project.compliance_os }
          : this.importData;

        await axios.post(`/api/project/${this.projectId}/lockdown/import`, data);
        this.$store.dispatch('showSuccess', 'Compliance tasks imported successfully');
        await this.loadTemplates();
        await this.loadProject(); // Refresh project info
      } catch (error) {
        console.error('Failed to import compliance tasks:', error);
        this.$store.dispatch('showError', 'Failed to import compliance tasks');
      } finally {
        this.importing = false;
      }
    },

    viewTemplate(template) {
      this.selectedTemplate = template;
      this.templateDialog = true;
    },

    createTask(template) {
      // Navigate to create task page with template
      this.$router.push({
        name: 'NewTask',
        params: { projectId: this.projectId },
        query: { templateId: template.id },
      });
    },

    getAppColor(app) {
      const colors = {
        ansible: 'primary',
        terraform: 'orange',
        bash: 'success',
        powershell: 'blue',
        python: 'green',
      };
      return colors[app] || 'grey';
    },

    formatDate,
  },
};
</script>
