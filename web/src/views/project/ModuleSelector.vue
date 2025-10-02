<template>
  <div>
    <v-toolbar flat>
      <v-btn
        icon
        class="mr-4"
        @click="returnToProjects()"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Module Selector</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <DashboardMenu
      v-if="projectId"
      :project-id="projectId"
      :project-type="projectType"
      :can-update-project="can(USER_PERMISSIONS.updateProject)"
    />

    <v-container>
      <v-form ref="form" v-model="isValid" lazy-validation>
        <!-- Tabbed Interface -->
        <v-tabs v-model="activeTab" class="mb-6">
          <v-tab>Basic Configuration</v-tab>
          <v-tab>Kubernetes Options</v-tab>
        </v-tabs>

        <v-tabs-items v-model="activeTab">
          <!-- Basic Configuration Tab -->
          <v-tab-item>
            <!-- Line 1: Cloud Provider takes entire first line -->
        <v-row>
          <v-col cols="12">
            <v-select
              :items="cloudProviders"
              v-model="form.cloudProvider"
              label="Cloud Provider"
              :rules="[rules.required]"
              filled
              dense
              clearable
            />
          </v-col>
        </v-row>

        <!-- Line 2: Golden Image and STIG Compliant -->
        <v-row>
          <v-col cols="12" md="8">
            <v-select
              :items="goldenImages"
              v-model="form.goldenImage"
              label="Golden Image"
              :rules="[rules.required]"
              filled
              dense
              clearable
            />
          </v-col>
          <v-col cols="12" md="4" class="d-flex align-center">
            <v-switch
              v-model="form.isStigCompliant"
              inset
              class="mt-0"
              label="STIG Compliant"
            />
          </v-col>
        </v-row>

        <!-- Line 3: Jumpbox and Jumpbox STIG Compliant -->
        <v-row>
          <v-col cols="12" md="8">
            <v-select
              :items="jumpboxOptions"
              v-model="form.jumpbox"
              label="Jumpbox"
              filled
              dense
              clearable
            />
          </v-col>
          <v-col cols="12" md="4" class="d-flex align-center">
            <v-switch
              v-model="form.jumpboxStigCompliant"
              inset
              class="mt-0"
              label="Jumpbox STIG Compliant"
            />
          </v-col>
        </v-row>

          </v-tab-item>

          <!-- Kubernetes Options Tab -->
          <v-tab-item>
            <v-row>
              <v-col cols="12">
                <h3 class="text-h6 mb-4">Kubernetes Configuration</h3>
              </v-col>
            </v-row>

        <v-row>
          <v-col cols="12" md="6">
            <v-select
              :items="kubernetesTypes"
              v-model="form.kubernetesType"
              label="Kubernetes Type"
              :rules="[rules.required]"
              filled
              dense
              clearable
            />
          </v-col>

          <v-col cols="12" md="6">
            <v-select
              :items="filteredInstanceTypes"
              v-model="form.instanceType"
              :label="`Instance Type (${form.cloudProvider || 'select provider'})`"
              :placeholder="
                form.cloudProvider
                  ? 'Select instance type...'
                  : 'Must select Cloud Provider first...'
              "
              :rules="[rules.required]"
              :disabled="!form.cloudProvider"
              filled
              dense
              clearable
            />
          </v-col>

          <v-col cols="12" md="4" v-if="form.kubernetesType === 'Self-Managed Kubernetes'">
            <v-text-field
              v-model.number="form.controlPlaneNodes"
              type="number"
              label="Control Plane Nodes"
              :rules="[rules.required, rules.oddNumber]"
              min="1"
              step="2"
              filled
              dense
            />
          </v-col>

          <v-col cols="12" md="3">
            <v-text-field
              v-model.number="form.minWorkers"
              type="number"
              label="Initial Worker Nodes"
              :rules="[rules.required, rules.min1]"
              min="1"
              step="1"
              filled
              dense
            />
          </v-col>

          <v-col cols="12" md="3">
            <v-text-field
              v-model.number="form.maxWorkers"
              type="number"
              label="Max Worker Nodes"
              :rules="[rules.required, rules.maxGteMin(form.minWorkers)]"
              :min="form.minWorkers || 1"
              step="1"
              filled
              dense
            />
          </v-col>
        </v-row>

            <!-- Additional Software Section -->
            <v-divider class="my-6"></v-divider>
            <v-row>
              <v-col cols="12">
                <h3 class="text-h6 mb-4">Additional Software</h3>
              </v-col>
            </v-row>

            <v-row>
              <v-col cols="12" md="6" class="pl-4">
                <v-switch
                  v-model="form.additionalSoftware.observability"
                  inset
                  class="mt-0"
                  label="Observability"
                />
              </v-col>
              <v-col cols="12" md="6" class="pl-2">
                <v-switch
                  v-model="form.additionalSoftware.serviceMesh"
                  inset
                  class="mt-0"
                  label="Service Mesh"
                />
              </v-col>
              <v-col cols="12" md="6" class="pl-4">
                <v-switch
                  v-model="form.additionalSoftware.certificateManager"
                  inset
                  class="mt-0"
                  label="Certificate Manager"
                />
              </v-col>
              <v-col cols="12" md="6" class="pl-2">
                <v-switch
                  v-model="form.additionalSoftware.gatewayApi"
                  inset
                  class="mt-0"
                  label="Gateway API"
                />
              </v-col>
              <v-col cols="12" md="6" class="pl-4">
                <v-switch
                  v-model="form.additionalSoftware.nginxIngressProxy"
                  inset
                  class="mt-0"
                  label="Nginx Ingress Proxy"
                />
              </v-col>
            </v-row>
          </v-tab-item>
        </v-tabs-items>

        <!-- Action Buttons -->
        <v-row class="mt-6">
          <v-col cols="12" class="d-flex">
            <v-spacer></v-spacer>
            <v-btn color="primary" class="mr-2" :disabled="!isValid" @click="onSubmit">
              Generate
            </v-btn>
            <v-btn text @click="onReset">Reset</v-btn>
          </v-col>
        </v-row>

      </v-form>

      <!-- JSON Output Display -->
      <v-row v-if="jsonOutput && showJsonOutput" class="mt-6">
        <v-col cols="12">
          <v-card>
            <v-card-title class="d-flex justify-space-between align-center">
              <div class="d-flex align-center">
                <v-icon class="mr-2">mdi-code-json</v-icon>
                Generated JSON Output
              </div>
              <v-btn
                icon
                small
                @click="closeJsonOutput"
              >
                <v-icon>mdi-close</v-icon>
              </v-btn>
            </v-card-title>
            <v-card-text>
              <v-textarea
                :value="JSON.stringify(jsonOutput, null, 2)"
                readonly
                auto-grow
                rows="10"
                class="json-output"
                outlined
                dense
              ></v-textarea>
              <div class="mt-2">
                <v-btn
                  color="primary"
                  small
                  @click="copyToClipboard"
                >
                  <v-icon left>mdi-content-copy</v-icon>
                  Copy JSON
                </v-btn>
                <v-btn
                  color="secondary"
                  small
                  class="ml-2"
                  @click="downloadJSON"
                >
                  <v-icon left>mdi-download</v-icon>
                  Download JSON
                </v-btn>
                <v-btn
                  color="success"
                  small
                  class="ml-2"
                  @click="saveAsTemplate"
                  :loading="savingTemplate"
                >
                  <v-icon left>mdi-content-save</v-icon>
                  Save as Template
                </v-btn>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- Terraform Configuration Display -->
      <v-row v-if="terraformOutput && showTerraformOutput" class="mt-6">
        <v-col cols="12">
          <v-card>
            <v-card-title class="d-flex justify-space-between align-center">
              <div class="d-flex align-center">
                <v-icon class="mr-2">mdi-terraform</v-icon>
                Generated Configuration Templates
              </div>
              <v-btn
                icon
                small
                @click="closeTerraformOutput"
              >
                <v-icon>mdi-close</v-icon>
              </v-btn>
            </v-card-title>
            <v-card-text>
              <v-tabs>
                <!-- Terraform Files Tab -->
                <v-tab>
                  <v-icon class="mr-2">mdi-terraform</v-icon>
                  Terraform
                </v-tab>
                <!-- Terragrunt Files Tab -->
                <v-tab>
                  <v-icon class="mr-2">mdi-terragrunt</v-icon>
                  Terragrunt
                </v-tab>

                <!-- Terraform Tab Content -->
                <v-tab-item>
                  <v-tabs>
                    <v-tab
                      v-for="filename in Object.keys(terraformOutput.terraform || {})"
                      :key="filename"
                    >
                      {{ filename }}
                    </v-tab>
                    <v-tab-item
                      v-for="filename in Object.keys(terraformOutput.terraform || {})"
                      :key="filename"
                    >
                      <v-textarea
                        :value="terraformOutput.terraform[filename]"
                        readonly
                        auto-grow
                        rows="15"
                        class="terraform-output"
                        outlined
                        dense
                      ></v-textarea>
                    </v-tab-item>
                  </v-tabs>
                </v-tab-item>

                <!-- Terragrunt Tab Content -->
                <v-tab-item>
                  <v-tabs>
                    <v-tab
                      v-for="filename in Object.keys(terraformOutput.terragrunt || {})"
                      :key="filename"
                    >
                      {{ filename }}
                    </v-tab>
                    <v-tab-item
                      v-for="filename in Object.keys(terraformOutput.terragrunt || {})"
                      :key="filename"
                    >
                      <v-textarea
                        :value="terraformOutput.terragrunt[filename]"
                        readonly
                        auto-grow
                        rows="15"
                        class="terraform-output"
                        outlined
                        dense
                      ></v-textarea>
                    </v-tab-item>
                  </v-tabs>
                </v-tab-item>
              </v-tabs>
              <div class="mt-2">
                <v-btn
                  color="primary"
                  small
                  @click="downloadTerraformFiles"
                >
                  <v-icon left>mdi-download</v-icon>
                  Download Terraform Files
                </v-btn>
                <v-btn
                  color="secondary"
                  small
                  class="ml-2"
                  @click="downloadTerragruntFiles"
                >
                  <v-icon left>mdi-download</v-icon>
                  Download Terragrunt Files
                </v-btn>
                <v-btn
                  color="success"
                  small
                  class="ml-2"
                  @click="saveTerraformAsTemplate"
                  :loading="savingTerraformTemplate"
                >
                  <v-icon left>mdi-content-save</v-icon>
                  Save as Template
                </v-btn>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script>
import ProjectMixin from '@/components/ProjectMixin';
import DashboardMenu from '@/components/DashboardMenu.vue';
import PermissionsCheck from '@/components/PermissionsCheck';
import { USER_PERMISSIONS } from '@/lib/constants';
import axios from 'axios';
import EventBus from '@/event-bus';
import { generateTerragruntConfig } from '@/utils/terraformGenerator';

export default {
  name: 'ModuleSelector',
  mixins: [ProjectMixin, PermissionsCheck],
  components: { DashboardMenu },
  props: {
    projectId: [Number, String],
    projectType: String,
  },
  computed: {
    filteredInstanceTypes() {
      if (!this.form.cloudProvider) return [];
      return this.instanceTypesByProvider[this.form.cloudProvider] || [];
    },
    USER_PERMISSIONS() {
      return USER_PERMISSIONS;
    },
  },
  watch: {
    'form.goldenImage': function onGoldenImageChange(newValue) {
      // Enable STIG Compliant by default for Linux Golden Images
      if (newValue && (newValue.includes('Linux') || newValue.includes('Red Hat')
          || newValue.includes('Ubuntu') || newValue.includes('SUSE'))) {
        this.form.isStigCompliant = true;
      }
    },
  },
  async created() {
    await this.loadRepositories();
    await this.refreshModules();
  },
  data() {
    return {
      isValid: false,
      activeTab: 0,
      jsonOutput: null,
      terraformOutput: null,
      cloudProviders: ['AWS', 'Azure', 'Google Cloud', 'VMWare VCF'],
      kubernetesTypes: ['Self-Managed Kubernetes', 'Managed Kubernetes'],
      goldenImages: ['Red Hat Linux', 'Ubuntu Linux', 'SUSE Linux'],
      jumpboxOptions: ['None', 'Windows 11 Pro', 'Red Hat Workstation'],
      instanceTypesByProvider: {
        AWS: ['t3.medium', 'm5.large', 'c5.xlarge'],
        Azure: ['Standard_D2s_v5', 'Standard_D4s_v5', 'Standard_F4s_v2'],
        'Google Cloud': ['e2-standard-2', 'n2-standard-4', 'c2-standard-8'],
        'VMWare VCF': ['small', 'medium', 'large', 'xlarge'],
      },
      repositories: [],
      selectedRepositoryId: null,
      modules: [],
      loading: {
        repositories: false,
        modules: false,
      },
      form: {
        cloudProvider: null,
        kubernetesType: null,
        instanceType: null,
        controlPlaneNodes: 3,
        minWorkers: 3,
        maxWorkers: 6,
        goldenImage: null,
        isStigCompliant: true,
        jumpbox: null,
        jumpboxStigCompliant: true,
        additionalSoftware: {
          observability: false,
          serviceMesh: false,
          certificateManager: false,
          gatewayApi: false,
          nginxIngressProxy: false,
        },
      },
      rules: {
        required: (v) => (v !== null && v !== undefined && v !== '') || 'Required',
        min1: (v) => (Number.isFinite(Number(v)) && Number(v) >= 1) || 'Must be >= 1',
        maxGteMin: (min) => (v) => (Number(v) >= Number(min || 1)) || 'Must be >= initial',
        oddNumber: (v) => (Number.isFinite(Number(v)) && Number(v) % 2 === 1)
          || 'Must be an odd number',
      },
      savingTemplate: false,
      savingTerraformTemplate: false,
      showJsonOutput: true,
      showTerraformOutput: true,
    };
  },

  methods: {
    async loadRepositories() {
      try {
        this.loading.repositories = true;
        this.repositories = await this.loadProjectResources('repositories');
      } catch (error) {
        console.error('Failed to load repositories:', error);
        this.repositories = [];
      } finally {
        this.loading.repositories = false;
      }
    },

    async refreshModules() {
      try {
        this.loading.modules = true;
        // For now, modules are hardcoded, but this could be extended to load from API
        this.modules = [];
      } catch (error) {
        console.error('Failed to refresh modules:', error);
        this.modules = [];
      } finally {
        this.loading.modules = false;
      }
    },

    onReset() {
      // Reset form to initial values
      this.form = {
        cloudProvider: null,
        kubernetesType: null,
        instanceType: null,
        controlPlaneNodes: 3,
        minWorkers: 3,
        maxWorkers: 6,
        goldenImage: null,
        isStigCompliant: true,
        jumpbox: null,
        jumpboxStigCompliant: true,
        additionalSoftware: {
          observability: false,
          serviceMesh: false,
          certificateManager: false,
          gatewayApi: false,
          nginxIngressProxy: false,
        },
      };

      // Clear outputs
      this.jsonOutput = null;
      this.terraformOutput = null;
      this.showJsonOutput = true;
      this.showTerraformOutput = true;

      // Reset form validation
      if (this.$refs.form) {
        this.$refs.form.reset();
        this.$refs.form.resetValidation();
      }
    },

    showDrawer() {
      this.$emit('show-drawer');
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    async onSubmit() {
      if (!this.$refs.form.validate()) return;

      // Create JSON output from form data - only include selected options
      this.jsonOutput = {
        cloudProvider: this.form.cloudProvider,
        kubernetesType: this.form.kubernetesType,
        instanceType: this.form.instanceType,
        minWorkers: this.form.minWorkers,
        maxWorkers: this.form.maxWorkers,
        goldenImage: this.form.goldenImage,
        isStigCompliant: this.form.isStigCompliant,
        jumpbox: this.form.jumpbox,
        jumpboxStigCompliant: this.form.jumpboxStigCompliant,
      };

      // Only include control plane nodes if Self-Managed Kubernetes is selected
      if (this.form.kubernetesType === 'Self-Managed Kubernetes') {
        this.jsonOutput.controlPlaneNodes = this.form.controlPlaneNodes;
      }

      // Only include additional software if any are selected
      const selectedAdditionalSoftware = {};
      Object.keys(this.form.additionalSoftware).forEach((key) => {
        if (this.form.additionalSoftware[key]) {
          selectedAdditionalSoftware[key] = this.form.additionalSoftware[key];
        }
      });

      if (Object.keys(selectedAdditionalSoftware).length > 0) {
        this.jsonOutput.additionalSoftware = selectedAdditionalSoftware;
      }

      try {
        // Generate Terraform and Terragrunt configuration
        this.terraformOutput = generateTerragruntConfig(this.jsonOutput);
        // Show both outputs when new configuration is generated
        this.showJsonOutput = true;
        this.showTerraformOutput = true;
      } catch (error) {
        console.error('Error generating Terraform config:', error);
        const errorText = `Failed to generate Terraform configuration: ${
          error.message || 'Unknown error'}`;
        this.$toast?.error(errorText);
      }

      console.log(JSON.stringify(this.jsonOutput, null, 2)); // Display JSON in console
      this.$emit('selected', this.jsonOutput); // Emit the output

      // Scroll to the JSON output section
      this.$nextTick(() => {
        const jsonSection = document.querySelector('.json-output');
        if (jsonSection) {
          jsonSection.scrollIntoView({ behavior: 'smooth' });
        }
      });
    },

    copyToClipboard() {
      if (this.jsonOutput) {
        const jsonString = JSON.stringify(this.jsonOutput, null, 2);
        navigator.clipboard.writeText(jsonString).then(() => {
          this.$toast?.success('JSON copied to clipboard!');
        }).catch(() => {
          // Fallback for older browsers
          const textArea = document.createElement('textarea');
          textArea.value = jsonString;
          document.body.appendChild(textArea);
          textArea.select();
          document.execCommand('copy');
          document.body.removeChild(textArea);
          this.$toast?.success('JSON copied to clipboard!');
        });
      }
    },

    downloadJSON() {
      if (this.jsonOutput) {
        const jsonString = JSON.stringify(this.jsonOutput, null, 2);
        const blob = new Blob([jsonString], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = 'module-selector-config.json';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
        this.$toast?.success('JSON file downloaded!');
      }
    },

    async saveAsTemplate() {
      if (!this.jsonOutput) {
        this.$toast?.error('No JSON output to save');
        return;
      }

      try {
        this.savingTemplate = true;

        // Debug logging
        console.log('Saving template for projectId:', this.projectId);
        console.log('Selected repository ID:', this.selectedRepositoryId);

        // Create template data
        const providerName = this.form.cloudProvider || 'Unknown';
        const dateStr = new Date().toLocaleDateString();
        const templateName = ['Module Selector -', providerName, '-', dateStr].join(' ');
        const templateDescription = [
          'Generated from Module Selector for',
          this.form.cloudProvider,
          'with',
          this.form.kubernetesType,
        ].join(' ');

        const templateData = {
          name: templateName,
          app: 'json',
          playbook: JSON.stringify(this.jsonOutput, null, 2),
          description: templateDescription,
          arguments: [],
          environment: [],
          inventory: null,
          repository: this.selectedRepositoryId,
        };

        console.log('Template data:', templateData);

        // Save template via API
        const response = await axios({
          method: 'POST',
          url: `/api/project/${this.projectId}/templates`,
          data: templateData,
          responseType: 'json',
        });

        console.log('Template save response:', response);

        this.$toast?.success('Template saved successfully!');
        // Optionally redirect to templates page
        this.$router.push(`/project/${this.projectId}/templates`);
      } catch (error) {
        console.error('Error saving template:', error);
        const errorMessage = error.response?.data?.message || error.message;
        const errorText = ['Failed to save template:', errorMessage].join(' ');
        this.$toast?.error(errorText);
      } finally {
        this.savingTemplate = false;
      }
    },

    downloadTerraform(filter) {
      if (!this.terraformOutput) {
        this.$toast?.error('No Terraform output to download');
        return;
      }

      let filesToDownload = this.terraformOutput;

      if (filter && filesToDownload[filter]) {
        filesToDownload = { [filter]: filesToDownload[filter] };
      }

      Object.keys(filesToDownload).forEach((filename) => {
        const content = filesToDownload[filename];
        const blob = new Blob([content], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
      });

      this.$toast?.success('Terraform files downloaded!');
    },

    downloadTerraformFiles() {
      if (!this.terraformOutput || !this.terraformOutput.terraform) {
        this.$toast?.error('No Terraform files to download');
        return;
      }

      Object.keys(this.terraformOutput.terraform).forEach((filename) => {
        const content = this.terraformOutput.terraform[filename];
        const blob = new Blob([content], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
      });

      this.$toast?.success('Terraform files downloaded!');
    },

    downloadTerragruntFiles() {
      if (!this.terraformOutput || !this.terraformOutput.terragrunt) {
        this.$toast?.error('No Terragrunt files to download');
        return;
      }

      Object.keys(this.terraformOutput.terragrunt).forEach((filename) => {
        const content = this.terraformOutput.terragrunt[filename];
        const blob = new Blob([content], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
      });

      this.$toast?.success('Terragrunt files downloaded!');
    },

    closeJsonOutput() {
      this.showJsonOutput = false;
    },

    closeTerraformOutput() {
      this.showTerraformOutput = false;
    },

    async saveTerraformAsTemplate() {
      if (!this.terraformOutput) {
        this.$toast?.error('No Terraform output to save');
        return;
      }

      try {
        this.savingTerraformTemplate = true;

        // Debug logging
        console.log('Saving Terraform template for projectId:', this.projectId);
        console.log('Selected repository ID:', this.selectedRepositoryId);

        // Create template data with Terraform files
        const providerName = this.form.cloudProvider || 'Unknown';
        const dateStr = new Date().toLocaleDateString();
        const templateName = ['Terraform Config -', providerName, '-', dateStr].join(' ');
        const templateDescription = [
          'Generated Terraform configuration for',
          this.form.cloudProvider,
          'with',
          this.form.kubernetesType,
        ].join(' ');

        // Combine all Terraform files into a single playbook
        const terraformContent = Object.keys(this.terraformOutput.terraform || {})
          .map((filename) => `# ${filename}\n${this.terraformOutput.terraform[filename]}`)
          .join('\n\n');

        const templateData = {
          name: templateName,
          app: 'terraform',
          playbook: terraformContent,
          description: templateDescription,
          arguments: [],
          environment: [],
          inventory: null,
          repository: this.selectedRepositoryId,
        };

        console.log('Terraform template data:', templateData);

        // Save template via API
        const response = await axios({
          method: 'POST',
          url: `/api/project/${this.projectId}/templates`,
          data: templateData,
          responseType: 'json',
        });

        console.log('Terraform template save response:', response);

        this.$toast?.success('Terraform template saved successfully!');
        // Optionally redirect to templates page
        this.$router.push(`/project/${this.projectId}/templates`);
      } catch (error) {
        console.error('Error saving Terraform template:', error);
        const errorMessage = error.response?.data?.message || error.message;
        const errorText = ['Failed to save Terraform template:', errorMessage].join(' ');
        this.$toast?.error(errorText);
      } finally {
        this.savingTerraformTemplate = false;
      }
    },

  },
};
</script>

<style scoped>
.json-output,
.terraform-output {
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.4;
}

.json-output >>> .v-text-field__slot textarea {
  font-family: 'Courier New', monospace !important;
  font-size: 14px !important;
  line-height: 1.4 !important;
}

.terraform-output >>> .v-text-field__slot textarea {
  font-family: 'Courier New', monospace !important;
  font-size: 14px !important;
  line-height: 1.4 !important;
}
</style>
