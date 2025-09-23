<template>
  <div>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Terraform Module Selector</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <v-container>
      <v-form ref="form" v-model="isValid" lazy-validation>
        <v-row>
          <v-col cols="12" md="6">
            <v-select
              :items="cloudProviders"
              v-model="form.cloudProvider"
              label="Cloud Provider"
              :rules="[rules.required]"
              outlined
              dense
              clearable
            />
          </v-col>

          <v-col cols="12" md="6">
            <v-select
              :items="kubernetesTypes"
              v-model="form.kubernetesType"
              label="Kubernetes Type"
              :rules="[rules.required]"
              outlined
              dense
              clearable
            />
          </v-col>

          <v-col cols="12" md="6">
            <v-select
              :items="filteredInstanceTypes"
              v-model="form.instanceType"
              :label="`Instance Type (${form.cloudProvider || 'select provider'})`"
              :rules="[rules.required]"
              :disabled="!form.cloudProvider"
              outlined
              dense
              clearable
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
              outlined
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
              outlined
              dense
            />
          </v-col>

          <v-col cols="12" md="6">
            <v-select
              :items="goldenImages"
              v-model="form.goldenImage"
              label="Golden Image"
              :rules="[rules.required]"
              outlined
              dense
              clearable
            />
          </v-col>

          <v-col cols="12" md="6" class="d-flex align-center">
            <v-switch
              v-model="form.isStigCompliant"
              inset
              class="mt-0"
              label="STIG Compliant"
            />
          </v-col>

          <v-col cols="12" md="6">
            <v-select
              :items="repositories"
              :loading="loading.repositories"
              v-model="selectedRepositoryId"
              item-text="name"
              item-value="id"
              label="Repository"
              outlined
              dense
              :hint="internalRepoHint"
              persistent-hint
            />
          </v-col>

          <v-col cols="12" md="6">
            <v-select
              :items="modules"
              :loading="loading.modules"
              v-model="form.modulePath"
              item-text="name"
              item-value="path"
              label="Module"
              :disabled="!selectedRepositoryId"
              outlined
              dense
              clearable
            />
          </v-col>

          <v-col cols="12">
            <v-alert type="info" outlined v-if="suggestedModule">
              Suggested module: <strong>{{ suggestedModule.name }}</strong>
              <div
                class="caption"
                v-if="suggestedModule.path"
              >
                Path: {{ suggestedModule.path }}
              </div>
            </v-alert>
          </v-col>

          <v-col cols="12" class="d-flex">
            <v-spacer></v-spacer>
            <v-btn color="primary" class="mr-2" :disabled="!isValid" @click="onSubmit">
              Continue
            </v-btn>
            <v-btn text @click="onReset">Reset</v-btn>
          </v-col>
        </v-row>
      </v-form>
    </v-container>
  </div>

</template>

<script>
import axios from 'axios';
import ProjectMixin from '@/components/ProjectMixin';
import EventBus from '@/event-bus';

export default {
  name: 'ModuleSelector',
  mixins: [ProjectMixin],
  async created() {
    await this.loadRepositories();
    await this.refreshModules();
  },
  data() {
    return {
      isValid: false,
      cloudProviders: ['AWS', 'Azure', 'Google Cloud'],
      kubernetesTypes: ['Self-Managed Kubernetes', 'Managed Kubernetes'],
      goldenImages: ['Red Hat Linux', 'Ubuntu Linux', 'SUSE Linux'],
      instanceTypesByProvider: {
        AWS: ['t3.medium', 'm5.large', 'c5.xlarge'],
        Azure: ['Standard_D2s_v5', 'Standard_D4s_v5', 'Standard_F4s_v2'],
        'Google Cloud': ['e2-standard-2', 'n2-standard-4', 'c2-standard-8'],
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
        minWorkers: 3,
        maxWorkers: 6,
        goldenImage: null,
        isStigCompliant: false,
        modulePath: null,
      },
      rules: {
        required: (v) => (v !== null && v !== undefined && v !== '') || 'Required',
        min1: (v) => (Number.isFinite(Number(v)) && Number(v) >= 1) || 'Must be >= 1',
        maxGteMin: (min) => (v) => (Number(v) >= Number(min || 1)) || 'Must be >= initial',
      },
    };
  },
  computed: {
    filteredInstanceTypes() {
      return this.instanceTypesByProvider[this.form.cloudProvider] || [];
    },
    suggestedModule() {
      if (!this.form.cloudProvider || !this.form.kubernetesType) return null;
      const providerKey = this.form.cloudProvider.toLowerCase().replace(/\s+/g, '-');
      const k8sKey = this.form.kubernetesType.startsWith('Self') ? 'self' : 'managed';
      return {
        name: `${providerKey}-${k8sKey}-k8s-cluster`,
        path: `modules/${providerKey}/${k8sKey}/k8s-cluster`,
      };
    },
    internalRepoHint() {
      const repo = this.repositories.find((r) => r.id === this.selectedRepositoryId);
      if (!repo) return '';
      const isLocal = (repo.git_url || '').startsWith('/');
      return isLocal ? 'Internal repository (local path)' : (repo.git_url || '');
    },
  },
  watch: {
    'form.cloudProvider': function () {
      this.form.instanceType = null;
      this.refreshModules();
    },
    'form.kubernetesType': function () {
      this.refreshModules();
    },
    selectedRepositoryId() {
      this.refreshModules();
    },
  },
  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },
    async loadRepositories() {
      try {
        this.loading.repositories = true;
        const repos = (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/repositories`,
          responseType: 'json',
        })).data;
        this.repositories = repos;
        // Auto-select internal (local path) repository if present
        const local = repos.find((r) => (r.git_url || '').startsWith('/')) || repos[0];
        this.selectedRepositoryId = local ? local.id : null;
      } finally {
        this.loading.repositories = false;
      }
    },
    async refreshModules() {
      this.modules = [];
      this.form.modulePath = null;
      if (!this.selectedRepositoryId) return;
      // Attempt to fetch modules from backend if available
      try {
        this.loading.modules = true;
        const params = new URLSearchParams();
        if (this.form.cloudProvider) params.set('provider', this.form.cloudProvider);
        if (this.form.kubernetesType) params.set('kubernetes', this.form.kubernetesType);
        const data = (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/repositories/${this.selectedRepositoryId}/modules`,
          params,
          responseType: 'json',
          validateStatus: (s) => (s >= 200 && s < 300) || s === 404,
        })).data;
        if (Array.isArray(data) && data.length > 0) {
          this.modules = data;
        } else if (this.suggestedModule) {
          // Fallback to suggested module if backend not available
          this.modules = [this.suggestedModule];
        }
      } catch (e) {
        if (this.suggestedModule) {
          this.modules = [this.suggestedModule];
        }
      } finally {
        this.loading.modules = false;
      }
    },
    onSubmit() {
      if (!this.$refs.form.validate()) return;
      this.$emit('selected', { ...this.form, suggestedModule: this.suggestedModule, repositoryId: this.selectedRepositoryId });
      this.$router.push(`/project/${this.projectId}/templates`);
    },
    onReset() {
      this.$refs.form.reset();
      this.$refs.form.resetValidation();
      this.modules = [];
      this.form.modulePath = null;
    },
  },
};
</script>

<style scoped>
</style>
