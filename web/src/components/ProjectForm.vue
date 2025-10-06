<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}</v-alert>

    <v-text-field
      ref="projectNameField"
      v-model="item.name"
      :label="$t(projectNameTitle)"
      :rules="[v => !!v || $t('project_name_required')]"
      required
      :disabled="formSaving"
      data-testid="newProject-name"
      outlined
      dense
    ></v-text-field>

    <v-textarea
      v-model="item.description"
      label="Project Description"
      :disabled="formSaving"
      data-testid="newProject-description"
      outlined
      rows="3"
      auto-grow
      counter="500"
      hint="Optional description for your project"
      persistent-hint
    ></v-textarea>

    <v-switch
      v-if="itemId === 'new'"
      v-model="item.import"
      label="Import"
      class="mt-4"
      data-testid="newProject-import"
      hide-details
    />

    <v-text-field
      v-if="itemId === 'new' && item.import"
      v-model="item.path"
      label="Path"
      :disabled="formSaving"
      data-testid="newProject-path"
      outlined
      dense
      class="mt-4"
    ></v-text-field>

    <v-switch
      v-if="itemId === 'new'"
      v-model="item.demo"
      label="Demo"
      class="mt-4"
      hide-details
    />

    <!-- Alert Settings Section -->
    <v-divider class="my-6"></v-divider>
    <v-subheader class="px-0">
      <v-icon class="mr-2">mdi-bell</v-icon>
      {{ $t('alertSettings') }}
    </v-subheader>

    <v-switch
      v-model="item.alert"
      :label="$t('enableAlerts')"
      class="mt-2"
      hide-details
      hint="Enable alert notifications for this project"
      persistent-hint
    />

    <!-- Compliance Framework Section -->
    <v-divider class="my-6"></v-divider>
    <v-subheader class="px-0">
      <v-icon class="mr-2">mdi-shield-check</v-icon>
      {{ $t('complianceFrameworkOptional') }}
    </v-subheader>

    <v-switch
      v-model="complianceEnabled"
      :label="$t('enableComplianceFramework')"
      class="mt-2"
      hide-details
      @change="onComplianceToggle"
    />

    <v-expand-transition>
      <div v-if="complianceEnabled">
        <v-row>
          <v-col cols="12" md="6">
            <v-select
              v-model="complianceFramework"
              :items="supportedFrameworks"
              label="Compliance Framework"
              :disabled="formSaving"
              outlined
              dense
              hint="Select compliance framework (CIS or STIG)"
              persistent-hint
              @change="onFrameworkChange"
            />
          </v-col>
          <v-col cols="12" md="6">
            <v-select
              v-model="complianceOS"
              :items="supportedOS"
              label="Operating System"
              :disabled="formSaving"
              outlined
              dense
              hint="Select target operating system"
              persistent-hint
              @change="onOSChange"
            />
          </v-col>
        </v-row>

        <v-switch
          v-model="enableSTIG"
          label="Import STIG/CIS Ansible Tasks"
          class="mt-2"
          hide-details
          hint="Automatically import compliance tasks from Ansible Lockdown repositories"
          persistent-hint
        />

        <!-- Import Button for Existing Projects -->
        <div v-if="itemId !== 'new' && complianceFramework && complianceOS" class="mt-4">
          <v-btn
            color="primary"
            :loading="importingTasks"
            :disabled="formSaving || importingTasks"
            @click="importComplianceTasks"
            outlined
          >
            <v-icon left>mdi-download</v-icon>
            Import {{ complianceFramework }} {{ complianceOS }} Tasks
          </v-btn>

          <v-btn
            v-if="hasImportedTasks"
            color="success"
            @click="viewImportedTasks"
            class="ml-2"
            outlined
          >
            <v-icon left>mdi-eye</v-icon>
            View Imported Tasks
          </v-btn>
        </div>

        <v-alert
          v-if="complianceEnabled && complianceFramework && complianceOS"
          type="info"
          outlined
          class="mt-4"
        >
          <div class="text-body-2">
            <strong>Selected:</strong> {{ complianceFramework }} {{ complianceOS }}<br>
            <span v-if="enableSTIG">
              <strong>Tasks:</strong>
              <span v-if="hasImportedTasks">
                ✅ Ansible compliance tasks have been imported from
              </span>
              <span v-else>
                🔄 Will automatically import Ansible compliance tasks from
              </span>
              <a href="https://github.com/ansible-lockdown" target="_blank" rel="noopener">
                Ansible Lockdown repositories
              </a>
            </span>
          </div>
        </v-alert>
      </div>
    </v-expand-transition>

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],
  props: {
    projectNameTitle: {
      type: String,
      default: 'projectName',
    },
  },
  data() {
    return {
      complianceEnabled: false,
      complianceFramework: '',
      complianceOS: '',
      enableSTIG: false,
      supportedFrameworks: ['CIS', 'STIG'],
      supportedOS: [
        'RHEL7', 'RHEL8', 'RHEL9',
        'UBUNTU18', 'UBUNTU20', 'UBUNTU22', 'UBUNTU24',
        'DEBIAN11', 'DEBIAN12',
        'AMAZON2', 'AMAZON2023',
        'SUSE15',
        'WINDOWS10', 'WINDOWS11', 'WINDOWS2016', 'WINDOWS2019', 'WINDOWS2022',
      ],
      importingTasks: false,
      hasImportedTasks: false,
    };
  },
  created() {
    // Set default values for new projects
    if (this.itemId === 'new' && this.item) {
      this.item.alert = true; // Default to enabled for new projects
      this.item.path = '/local/path';
    }
  },
  async mounted() {
    // Load existing compliance settings for existing projects
    if (this.itemId !== 'new' && this.item) {
      this.loadExistingComplianceSettings();
      await this.checkForImportedTasks();
    }

    // Focus the project name field when component is mounted
    this.$nextTick(() => {
      if (this.$refs.projectNameField) {
        this.$refs.projectNameField.focus();
      }
    });
  },
  methods: {
    getItemsUrl() {
      return '/api/projects';
    },
    getSingleItemUrl() {
      return `/api/project/${this.itemId}`;
    },
    beforeSave() {
      // Set compliance fields if enabled
      if (this.complianceEnabled) {
        this.item.compliance_framework = this.complianceFramework;
        this.item.compliance_os = this.complianceOS;
        this.item.enable_stig = this.enableSTIG;
      } else {
        // Clear compliance fields if not enabled
        this.item.compliance_framework = null;
        this.item.compliance_os = null;
        this.item.enable_stig = false;
      }
    },
    async onComplianceToggle() {
      if (!this.complianceEnabled) {
        // Reset compliance fields when disabled
        this.complianceFramework = '';
        this.complianceOS = '';
        this.enableSTIG = false;
      } else {
        // When enabling compliance, set default values if not already set
        if (!this.complianceFramework) {
          this.complianceFramework = 'STIG'; // Default to STIG
        }
        if (!this.complianceOS) {
          this.complianceOS = 'RHEL9'; // Default to RHEL9
        }
        this.enableSTIG = true;

        // Auto-import compliance tasks if framework and OS are selected
        if (this.complianceFramework && this.complianceOS && this.itemId !== 'new') {
          await this.importComplianceTasks();
        }
      }
    },
    async onFrameworkChange() {
      // Auto-enable STIG import when framework is selected
      if (this.complianceFramework && this.complianceOS) {
        this.enableSTIG = true;

        // Auto-import compliance tasks if both framework and OS are selected
        if (this.itemId !== 'new') {
          await this.importComplianceTasks();
        }
      }
    },
    async onOSChange() {
      // Auto-enable STIG import when OS is selected
      if (this.complianceFramework && this.complianceOS) {
        this.enableSTIG = true;

        // Auto-import compliance tasks if both framework and OS are selected
        if (this.itemId !== 'new') {
          await this.importComplianceTasks();
        }
      }
    },
    loadExistingComplianceSettings() {
      // Load existing compliance settings from the project
      if (this.item.compliance_framework) {
        this.complianceEnabled = true;
        this.complianceFramework = this.item.compliance_framework;
      }
      if (this.item.compliance_os) {
        this.complianceOS = this.item.compliance_os;
      }
      if (this.item.enable_stig) {
        this.enableSTIG = this.item.enable_stig;
      }
    },
    async checkForImportedTasks() {
      // Check if there are already imported compliance tasks for this project
      try {
        const response = await axios.get(`/api/project/${this.itemId}/folders/templates`);
        const folders = response.data.folders || [];
        this.hasImportedTasks = folders.some((folder) =>
          folder.name.includes('STIG') || folder.name.includes('CIS')
        );
      } catch (error) {
        console.error('Failed to check for imported tasks:', error);
        this.hasImportedTasks = false;
      }
    },
    async importComplianceTasks() {
      if (!this.complianceFramework || !this.complianceOS) {
        this.$emit('error', 'Please select both compliance framework and operating system');
        return;
      }

      this.importingTasks = true;
      try {
        const response = await axios.post(`/api/project/${this.itemId}/lockdown/import`, {
          framework: this.complianceFramework,
          os: this.complianceOS,
        });

        if (response.data.success) {
          this.$emit(
            'success',
            `Successfully imported ${this.complianceFramework} `
            + `${this.complianceOS} compliance tasks`,
          );
          this.hasImportedTasks = true;

          // Update the project with compliance settings
          this.item.compliance_framework = this.complianceFramework;
          this.item.compliance_os = this.complianceOS;
          this.item.enable_stig = true;
        } else {
          this.$emit('error', response.data.message || 'Failed to import compliance tasks');
        }
      } catch (error) {
        console.error('Failed to import compliance tasks:', error);
        this.$emit('error', 'Failed to import compliance tasks. Please try again.');
      } finally {
        this.importingTasks = false;
      }
    },
    viewImportedTasks() {
      // Navigate to the templates page to view imported tasks
      this.$router.push(`/project/${this.itemId}/templates`);
    },
  },
};
</script>
