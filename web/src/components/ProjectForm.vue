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

    <!-- Compliance Framework Section -->
    <v-divider class="my-6"></v-divider>
    <v-subheader class="px-0">
      <v-icon class="mr-2">mdi-shield-check</v-icon>
      Compliance Framework (Optional)
    </v-subheader>

    <v-switch
      v-if="itemId === 'new'"
      v-model="complianceEnabled"
      label="Enable Compliance Framework"
      class="mt-2"
      hide-details
      @change="onComplianceToggle"
    />

    <v-expand-transition>
      <div v-if="complianceEnabled && itemId === 'new'">
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

        <v-alert
          v-if="complianceEnabled && complianceFramework && complianceOS"
          type="info"
          outlined
          class="mt-4"
        >
          <div class="text-body-2">
            <strong>Selected:</strong> {{ complianceFramework }} {{ complianceOS }}<br>
            <span v-if="enableSTIG">
              <strong>Tasks:</strong> Will import Ansible compliance tasks from
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
    };
  },
  created() {
    // Set default values for new projects
    if (this.itemId === 'new' && this.item) {
      this.item.alert = true;
      this.item.path = '/local/path';
    }
    // Always ensure alerts are enabled
    if (this.item) {
      this.item.alert = true;
    }
  },
  mounted() {
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
      // Always ensure alerts are enabled for all projects
      this.item.alert = true;

      // Set compliance fields if enabled
      if (this.complianceEnabled && this.itemId === 'new') {
        this.item.compliance_framework = this.complianceFramework;
        this.item.compliance_os = this.complianceOS;
        this.item.enable_stig = this.enableSTIG;
      } else if (this.itemId === 'new') {
        // Clear compliance fields if not enabled
        this.item.compliance_framework = null;
        this.item.compliance_os = null;
        this.item.enable_stig = false;
      }
    },
    onComplianceToggle() {
      if (!this.complianceEnabled) {
        // Reset compliance fields when disabled
        this.complianceFramework = '';
        this.complianceOS = '';
        this.enableSTIG = false;
      }
    },
    onFrameworkChange() {
      // Auto-enable STIG import when framework is selected
      if (this.complianceFramework && this.complianceOS) {
        this.enableSTIG = true;
      }
    },
    onOSChange() {
      // Auto-enable STIG import when OS is selected
      if (this.complianceFramework && this.complianceOS) {
        this.enableSTIG = true;
      }
    },
  },
};
</script>
