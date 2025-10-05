<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div class="new-project-form">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('projectBuilder') }}</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <v-container fluid class="pa-4">
      <v-row justify="center">
        <v-col cols="12" style="max-width: 80%;">
          <v-card class="pa-8" elevation="3" rounded="lg">
            <v-form ref="form" v-model="isValid" lazy-validation>
              <!-- Tabbed Interface -->
              <v-tabs v-model="activeTab" class="mb-6">
                <v-tab>Project</v-tab>
                <v-tab>Cloud Provider</v-tab>
                <v-tab>Kubernetes Options</v-tab>
              </v-tabs>

              <v-tabs-items v-model="activeTab">
          <!-- Project Tab -->
          <v-tab-item>
            <v-row>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="form.projectName"
                  label="Project Name"
                  :rules="[rules.required]"
                  filled
                  dense
                  hint="Name for your project"
                  data-testid="newProject-name"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-select
                  :items="environments"
                  v-model="form.environment"
                  :label="$t('projectEnvironment')"
                  :rules="[rules.required]"
                  filled
                  dense
                  hint="Target environment for deployment"
                />
              </v-col>
            </v-row>
            <v-row>
              <v-col cols="12">
                <v-textarea
                  v-model="form.projectDescription"
                  label="Project Description"
                  filled
                  dense
                  rows="3"
                  auto-grow
                  counter="500"
                  hint="Optional description for your project"
                  data-testid="newProject-description"
                />
              </v-col>
            </v-row>
            <v-row>
              <v-col cols="12" md="6">
                <v-select
                  :items="jumpboxOptions"
                  v-model="form.jumpbox"
                  :label="$t('jumpbox')"
                  filled
                  dense
                  clearable
                />
              </v-col>
              <v-col cols="12" md="6" class="d-flex align-center">
                <v-switch
                  v-model="form.jumpboxStigCompliant"
                  inset
                  class="mt-0"
                  :label="$t('jumpboxStigCompliant')"
                />
              </v-col>
            </v-row>

            <!-- Additional Options Section -->
            <v-row class="mt-6">
              <v-col cols="12">
                <v-card class="pa-4" outlined>
                  <v-card-title class="text-h6 pa-0 mb-4">
                    {{ $t('additionalOptions') }}
                  </v-card-title>
                  <v-row>
                    <v-col cols="12" md="6">
                      <v-switch
                        v-model="form.enableAlerts"
                        inset
                        :label="$t('enableAlerts')"
                        hint="Enable alert notifications for this project"
                      />
                    </v-col>
                    <v-col cols="12" md="6">
                      <v-switch
                        v-model="form.demo"
                        inset
                        :label="$t('demo')"
                        hint="Create a demo project with sample data"
                      />
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col cols="12">
                      <v-switch
                        v-model="form.import"
                        inset
                        :label="$t('importScripts')"
                        hint="Import existing project scripts from path"
                      />
                    </v-col>
                  </v-row>
                  <v-row v-if="form.import">
                    <v-col cols="12">
                      <v-text-field
                        v-model="form.path"
                        label="Import Path"
                        filled
                        dense
                        hint="Path to existing project scripts"
                        data-testid="newProject-path"
                      />
                    </v-col>
                  </v-row>
                </v-card>
              </v-col>
            </v-row>
          </v-tab-item>

          <!-- Cloud Provider Tab -->
          <v-tab-item>
            <!-- Line 1: Cloud Provider takes entire first line -->
            <v-row>
              <v-col cols="12">
                <v-select
                  :items="cloudProviders"
                  v-model="form.cloudProvider"
                  label="Cloud Provider"
                  filled
                  dense
                  clearable
                />
              </v-col>
            </v-row>

            <!-- Cloud Provider Specific Fields -->
            <v-row v-if="form.cloudProvider">
              <v-col cols="12">
                <v-card class="pa-4" outlined>
                  <v-card-title class="text-h6 pa-0 mb-4">
                    {{ form.cloudProvider }} Configuration
                  </v-card-title>

                  <!-- Azure Configuration -->
                  <template v-if="form.cloudProvider === 'Azure'">
                    <v-row>
                      <v-col cols="12" md="6">
                        <v-text-field
                          v-model="form.azure.subscriptionId"
                          label="Subscription ID"
                          filled
                          dense
                          hint="Azure subscription ID"
                        />
                      </v-col>
                      <v-col cols="12" md="6">
                        <v-text-field
                          v-model="form.azure.resourceGroup"
                          label="Resource Group"
                          filled
                          dense
                          hint="Existing resource group or 'new' to create"
                        />
                      </v-col>
                    </v-row>
                    <v-row>
                      <v-col cols="12">
                        <v-select
                          :items="azureLocations"
                          v-model="form.azure.location"
                          label="Location/Region"
                          filled
                          dense
                          hint="Azure region for deployment"
                        />
                      </v-col>
                    </v-row>
                  </template>

                  <!-- AWS Configuration -->
                  <template v-if="form.cloudProvider === 'AWS'">
                    <v-row>
                      <v-col cols="12" md="6">
                        <v-text-field
                          v-model="form.aws.region"
                          label="AWS Region"
                          :rules="[rules.required]"
                          filled
                          dense
                          hint="AWS region for deployment"
                        />
                      </v-col>
                      <v-col cols="12" md="6">
                        <v-text-field
                          v-model="form.aws.accountId"
                          label="AWS Account ID"
                          :rules="[rules.required]"
                          filled
                          dense
                          hint="AWS account ID"
                        />
                      </v-col>
                    </v-row>
                  </template>

                  <!-- Google Cloud Configuration -->
                  <template v-if="form.cloudProvider === 'Google Cloud'">
                    <v-row>
                      <v-col cols="12" md="6">
                        <v-text-field
                          v-model="form.gcp.projectId"
                          label="GCP Project ID"
                          :rules="[rules.required]"
                          filled
                          dense
                          hint="Google Cloud project ID"
                        />
                      </v-col>
                      <v-col cols="12" md="6">
                        <v-select
                          :items="gcpRegions"
                          v-model="form.gcp.region"
                          label="GCP Region"
                          :rules="[rules.required]"
                          filled
                          dense
                          hint="Google Cloud region"
                        />
                      </v-col>
                    </v-row>
                  </template>

                  <!-- VMware VCF Configuration -->
                  <template v-if="form.cloudProvider === 'VMWare VCF'">
                    <v-row>
                      <v-col cols="12" md="6">
                        <v-text-field
                          v-model="form.vmware.vcenter"
                          label="vCenter Server"
                          :rules="[rules.required]"
                          filled
                          dense
                          hint="vCenter server hostname or IP"
                        />
                      </v-col>
                      <v-col cols="12" md="6">
                        <v-text-field
                          v-model="form.vmware.datacenter"
                          label="Datacenter"
                          :rules="[rules.required]"
                          filled
                          dense
                          hint="vCenter datacenter name"
                        />
                      </v-col>
                    </v-row>
                    <v-row>
                      <v-col cols="12" md="6">
                        <v-text-field
                          v-model="form.vmware.cluster"
                          label="Cluster"
                          :rules="[rules.required]"
                          filled
                          dense
                          hint="vCenter cluster name"
                        />
                      </v-col>
                      <v-col cols="12" md="6">
                        <v-text-field
                          v-model="form.vmware.datastore"
                          label="Datastore"
                          :rules="[rules.required]"
                          filled
                          dense
                          hint="vCenter datastore name"
                        />
                      </v-col>
                    </v-row>
                  </template>
                </v-card>
              </v-col>
            </v-row>

            <!-- Line 2: Golden Image and STIG Compliant -->
            <v-row>
              <v-col cols="12" md="8">
                <v-select
                  :items="goldenImages"
                  v-model="form.goldenImage"
                  label="Golden Image"
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
          </v-tab-item>

          <!-- Kubernetes Options Tab -->
          <v-tab-item>
            <v-row>
              <v-col cols="12" md="6">
                <v-select
                  :items="kubernetesTypes"
                  v-model="form.kubernetesType"
                  label="Kubernetes Type"
                  filled
                  dense
                  clearable
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-select
                  :items="filteredInstanceTypes"
                  v-model="form.instanceType"
                  label="Instance Type"
                  :rules="form.kubernetesType && form.kubernetesType !== 'None'
                    ? [rules.required] : []"
                  filled
                  dense
                  clearable
                />
              </v-col>
            </v-row>

            <v-row v-if="form.kubernetesType && form.kubernetesType !== 'None'">
              <v-col cols="12" md="4">
                <v-text-field
                  v-model.number="form.controlPlaneNodes"
                  label="Control Plane Nodes"
                  :rules="[rules.required, rules.min1, rules.oddNumber]"
                  filled
                  dense
                  type="number"
                />
              </v-col>
              <v-col cols="12" md="4">
                <v-text-field
                  v-model.number="form.minWorkers"
                  label="Min Workers"
                  :rules="[rules.required, rules.min1]"
                  filled
                  dense
                  type="number"
                />
              </v-col>
              <v-col cols="12" md="4">
                <v-text-field
                  v-model.number="form.maxWorkers"
                  label="Max Workers"
                  :rules="[rules.required, rules.min1, rules.maxGteMin(form.minWorkers)]"
                  filled
                  dense
                  type="number"
                />
              </v-col>
            </v-row>

            <v-row v-if="form.kubernetesType && form.kubernetesType !== 'None'">
              <v-col cols="12">
                <v-card class="pa-4" outlined>
                  <v-card-title class="text-h6 pa-0 mb-4">Additional Software</v-card-title>
                  <v-row>
                    <v-col cols="12" md="6" class="pl-4">
                      <v-switch
                        v-model="form.additionalSoftware.observability"
                        inset
                        label="Observability"
                      />
                    </v-col>
                    <v-col cols="12" md="6" class="pl-4">
                      <v-switch
                        v-model="form.additionalSoftware.serviceMesh"
                        inset
                        label="Service Mesh"
                      />
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col cols="12" md="6" class="pl-4">
                      <v-switch
                        v-model="form.additionalSoftware.certificateManager"
                        inset
                        label="Certificate Manager"
                      />
                    </v-col>
                    <v-col cols="12" md="6" class="pl-4">
                      <v-switch
                        v-model="form.additionalSoftware.gatewayApi"
                        inset
                        label="Gateway API"
                      />
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col cols="12" md="6" class="pl-4">
                      <v-switch
                        v-model="form.additionalSoftware.nginxIngressProxy"
                        inset
                        label="Nginx Ingress Proxy"
                      />
                    </v-col>
                  </v-row>
                </v-card>
              </v-col>
            </v-row>
          </v-tab-item>
        </v-tabs-items>

        <!-- Action Buttons -->
        <v-row class="mt-6">
          <v-col cols="12" class="d-flex">
            <v-btn
              v-if="activeTab > 0"
              color="grey"
              class="mr-2"
              @click="previousTab"
            >
              <v-icon left>mdi-arrow-left</v-icon>
              Back
            </v-btn>
            <v-spacer></v-spacer>
            <v-btn
              v-if="activeTab < totalTabs - 1"
              color="primary"
              class="mr-2"
              :disabled="!isCurrentTabValid"
              @click="nextTab"
            >
              Next
              <v-icon right>mdi-arrow-right</v-icon>
            </v-btn>
          <v-btn
              v-if="activeTab === totalTabs - 1"
              color="primary"
              :disabled="!isCurrentTabValid"
              @click="createProject"
            >
              {{ $t('create') }}
          </v-btn>
          </v-col>
        </v-row>
      </v-form>
    </v-container>
  </div>
</template>
<style lang="scss">
// Dark mode compatibility for New Project form
.theme--dark {
  .new-project-form {
    .v-card {
      background-color: var(--v-theme-surface, #1e1e1e) !important;
      color: var(--v-theme-on-surface, #ffffff) !important;
    }

    .v-tabs {
      .v-tab {
        color: var(--v-theme-on-surface, #ffffff) !important;

        &.v-tab--active {
          color: var(--v-theme-primary, #1976d2) !important;
        }
      }
    }

    .v-text-field {
      .v-input__control {
        .v-field {
          background-color: var(--v-theme-surface-variant, #2d2d2d) !important;
          color: var(--v-theme-on-surface, #ffffff) !important;
        }

        .v-field__input {
          color: var(--v-theme-on-surface, #ffffff) !important;
        }

        .v-label {
          color: var(--v-theme-on-surface-variant, #b3b3b3) !important;
        }
      }
    }

    .v-select {
      .v-input__control {
        .v-field {
          background-color: var(--v-theme-surface-variant, #2d2d2d) !important;
          color: var(--v-theme-on-surface, #ffffff) !important;
        }

        .v-field__input {
          color: var(--v-theme-on-surface, #ffffff) !important;
        }

        .v-label {
          color: var(--v-theme-on-surface-variant, #b3b3b3) !important;
        }
      }
    }

    .v-textarea {
      .v-input__control {
        .v-field {
          background-color: var(--v-theme-surface-variant, #2d2d2d) !important;
          color: var(--v-theme-on-surface, #ffffff) !important;
        }

        .v-field__input {
          color: var(--v-theme-on-surface, #ffffff) !important;
        }

        .v-label {
          color: var(--v-theme-on-surface-variant, #b3b3b3) !important;
        }
      }
    }

    .v-switch {
      .v-input__control {
        .v-label {
          color: var(--v-theme-on-surface, #ffffff) !important;
        }
      }
    }

    .v-btn {
      &.v-btn--variant-elevated {
        background-color: var(--v-theme-primary, #1976d2) !important;
        color: var(--v-theme-on-primary, #ffffff) !important;
      }

      &.v-btn--variant-outlined {
        border-color: var(--v-theme-outline, #666666) !important;
        color: var(--v-theme-on-surface, #ffffff) !important;
      }
    }

    .v-subheader {
      color: var(--v-theme-on-surface, #ffffff) !important;
    }

    .v-card-title {
      color: var(--v-theme-on-surface, #ffffff) !important;
    }

    .v-toolbar {
      background-color: var(--v-theme-surface, #1e1e1e) !important;
      color: var(--v-theme-on-surface, #ffffff) !important;
    }

    .v-toolbar__title {
      color: var(--v-theme-on-surface, #ffffff) !important;
    }

    .v-app-bar-nav-icon {
      color: var(--v-theme-on-surface, #ffffff) !important;
    }
  }
}

// Light mode compatibility (ensure consistency)
.theme--light {
  .new-project-form {
    .v-card {
      background-color: var(--v-theme-surface, #ffffff) !important;
      color: var(--v-theme-on-surface, #000000) !important;
    }
  }
}
</style>
<script>
import EventBus from '@/event-bus';
import axios from 'axios';

export default {
  data() {
    return {
      isValid: false,
      activeTab: 0,
      totalTabs: 3,
      environments: ['Development', 'Staging', 'Production', 'Testing', 'QA'],
      cloudProviders: ['AWS', 'Azure', 'Google Cloud', 'VMWare VCF'],
      azureLocations: [
        'East US', 'East US 2', 'West US', 'West US 2', 'Central US',
        'North Central US', 'South Central US', 'West Central US',
        'Canada East', 'Canada Central', 'Brazil South', 'Brazil Southeast',
        'North Europe', 'West Europe', 'UK South', 'UK West',
        'France Central', 'France South', 'Germany North', 'Germany West Central',
        'Switzerland North', 'Switzerland West', 'Norway East', 'Norway West',
        'Sweden Central', 'Sweden South', 'Poland Central', 'Poland North',
        'Italy North', 'South Africa North', 'South Africa West',
        'UAE North', 'UAE Central', 'Saudi Arabia North', 'Saudi Arabia Central',
        'Japan East', 'Japan West', 'Korea Central', 'Korea South',
        'India Central', 'India South', 'India West', 'India East',
        'Southeast Asia', 'East Asia', 'Australia East', 'Australia Southeast',
        'Australia Central', 'Australia Central 2', 'New Zealand North', 'New Zealand Central',
      ],
      gcpRegions: [
        'us-central1', 'us-east1', 'us-east4', 'us-west1', 'us-west2', 'us-west3', 'us-west4',
        'europe-north1', 'europe-west1', 'europe-west2', 'europe-west3', 'europe-west4',
        'europe-west6', 'asia-east1', 'asia-east2', 'asia-northeast1', 'asia-northeast2',
        'asia-south1', 'asia-southeast1', 'asia-southeast2', 'australia-southeast1',
        'australia-southeast2', 'northamerica-northeast1', 'southamerica-east1',
      ],
      kubernetesTypes: ['None', 'Self-Managed Kubernetes', 'Managed Kubernetes'],
      goldenImages: [
        'RHEL7', 'RHEL8', 'RHEL9',
        'UBUNTU18', 'UBUNTU20', 'UBUNTU22', 'UBUNTU24',
        'DEBIAN11', 'DEBIAN12',
        'AMAZON2', 'AMAZON2023',
        'SUSE15',
        'WINDOWS10', 'WINDOWS11', 'WINDOWS2016', 'WINDOWS2019', 'WINDOWS2022',
      ],
      jumpboxOptions: [
        'None', 'RHEL7', 'RHEL8', 'RHEL9',
        'UBUNTU18', 'UBUNTU20', 'UBUNTU22', 'UBUNTU24',
        'DEBIAN11', 'DEBIAN12',
        'AMAZON2', 'AMAZON2023',
        'SUSE15',
        'WINDOWS10', 'WINDOWS11', 'WINDOWS2016', 'WINDOWS2019', 'WINDOWS2022',
      ],
      instanceTypesByProvider: {
        AWS: ['t3.medium', 'm5.large', 'c5.xlarge'],
        Azure: ['Standard_B2s', 'Standard_D2s_v3', 'Standard_D4s_v3'],
        'Google Cloud': ['e2-medium', 'e2-standard-2', 'e2-standard-4'],
        'VMWare VCF': ['small', 'medium', 'large', 'xlarge'],
      },
      form: {
        projectName: null,
        projectDescription: null,
        environment: null,
        enableAlerts: true,
        demo: false,
        import: false,
        path: null,
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
        azure: {
          subscriptionId: null,
          resourceGroup: null,
          location: null,
        },
        aws: {
          region: null,
          accountId: null,
        },
        gcp: {
          projectId: null,
          region: null,
        },
        vmware: {
          vcenter: null,
          datacenter: null,
          cluster: null,
          datastore: null,
        },
      },
      rules: {
        required: (v) => (v !== null && v !== undefined && v !== '') || 'Required',
        min1: (v) => (Number.isFinite(Number(v)) && Number(v) >= 1) || 'Must be >= 1',
        maxGteMin: (min) => (v) => (Number(v) >= Number(min || 1)) || 'Must be >= initial',
        oddNumber: (v) => (Number.isFinite(Number(v)) && Number(v) % 2 === 1)
          || 'Must be an odd number',
      },
    };
  },
  computed: {
    filteredInstanceTypes() {
      if (!this.form.cloudProvider) return [];
      return this.instanceTypesByProvider[this.form.cloudProvider] || [];
    },
    isCurrentTabValid() {
      // Only validate required fields on the current tab
      if (this.activeTab === 0) { // Project tab
        return !!(this.form.projectName && this.form.environment);
      }
      if (this.activeTab === 1) { // Cloud Provider tab
        // Cloud provider is optional - always valid
        return true;
      }
      if (this.activeTab === 2) { // Kubernetes tab
        // If None is selected, no validation needed
        if (this.form.kubernetesType === 'None' || !this.form.kubernetesType) {
          return true;
        }
        // Otherwise require kubernetesType and instanceType
        return !!(this.form.kubernetesType && this.form.instanceType);
      }
      return true;
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
  props: {
    systemInfo: Object,
  },
  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    nextTab() {
      if (this.activeTab < this.totalTabs - 1) {
        this.activeTab += 1;
      }
    },

    previousTab() {
      if (this.activeTab > 0) {
        this.activeTab -= 1;
      }
    },

    async createProject() {
      try {
        // Only validate essential fields
        if (!this.form.projectName || !this.form.environment) {
          this.$toast?.error('Project Name and Environment are required.');
          return;
        }

        // Create project data object
        const projectData = {
          name: this.form.projectName,
          description: this.form.projectDescription,
          alert: this.form.enableAlerts,
          demo: this.form.demo,
          import: this.form.import,
          path: this.form.import ? this.form.path : null,
          // Include environment builder configuration
          environment_config: {
            environment: this.form.environment,
            cloudProvider: this.form.cloudProvider,
            kubernetesType: this.form.kubernetesType,
            instanceType: this.form.instanceType,
            controlPlaneNodes: this.form.controlPlaneNodes,
            minWorkers: this.form.minWorkers,
            maxWorkers: this.form.maxWorkers,
            goldenImage: this.form.goldenImage,
            isStigCompliant: this.form.isStigCompliant,
            jumpbox: this.form.jumpbox,
            jumpboxStigCompliant: this.form.jumpboxStigCompliant,
            additionalSoftware: this.form.additionalSoftware,
            cloudProviderConfig: {
              azure: this.form.azure,
              aws: this.form.aws,
              gcp: this.form.gcp,
              vmware: this.form.vmware,
            },
          },
        };

        // Create the project via API
        const response = await axios.post('/api/project', projectData);
        const createdProject = response.data;

        // Emit the project creation event to update the UI
        EventBus.$emit('i-project', {
          action: 'new',
          item: createdProject,
        });

        // Show success message
        this.$toast?.success('Project created successfully!');

        // Redirect to the newly created project
        this.$router.push(`/project/${createdProject.id}`);
      } catch (error) {
        this.$toast?.error('Failed to create project. Please try again.');
      }
    },
  },
};
</script>
