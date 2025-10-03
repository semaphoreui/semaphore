<template>
  <div class="compliance-container">
    <v-container fluid>
      <!-- Header -->
      <div class="d-flex justify-space-between align-center mb-6">
        <div>
          <h1 class="text-h4 mb-2">{{ $t('compliance.title') }}</h1>
          <p class="text-body-1 text--secondary">
            {{ $t('compliance.description') }}
          </p>
        </div>
        <v-btn
          color="primary"
          @click="checkPreflight"
          :loading="preflightLoading"
        >
          <v-icon left>mdi-check-circle</v-icon>
          {{ $t('compliance.checkSystem') }}
        </v-btn>
      </div>

      <!-- Preflight Status Card -->
      <v-card v-if="preflightResult" class="mb-6">
        <v-card-title>
          <v-icon left>mdi-information</v-icon>
          {{ $t('compliance.systemStatus') }}
        </v-card-title>
        <v-card-text>
          <v-alert
            :type="preflightResult.oscap_available ? 'success' : 'error'"
            class="mb-4"
          >
            <div v-if="preflightResult.oscap_available">
              <strong>{{ $t('compliance.openscapAvailable') }}</strong>
              <div v-if="preflightResult.oscap_version">
                {{ $t('compliance.version') }}: {{ preflightResult.oscap_version }}
              </div>
            </div>
            <div v-else>
              <strong>{{ $t('compliance.openscapNotAvailable') }}</strong>
            </div>
          </v-alert>

          <!-- Errors -->
          <div v-if="preflightResult.errors && preflightResult.errors.length > 0">
            <h4 class="text-subtitle-1 mb-2">{{ $t('compliance.errors') }}</h4>
            <v-list dense>
              <v-list-item v-for="error in preflightResult.errors" :key="error">
                <v-list-item-icon>
                  <v-icon color="error">mdi-alert-circle</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>{{ error }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </div>

          <!-- Warnings -->
          <div v-if="preflightResult.warnings && preflightResult.warnings.length > 0">
            <h4 class="text-subtitle-1 mb-2">{{ $t('compliance.warnings') }}</h4>
            <v-list dense>
              <v-list-item v-for="warning in preflightResult.warnings" :key="warning">
                <v-list-item-icon>
                  <v-icon color="warning">mdi-alert</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>{{ warning }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </div>

          <!-- Info -->
          <div v-if="preflightResult.info && preflightResult.info.length > 0">
            <h4 class="text-subtitle-1 mb-2">{{ $t('compliance.info') }}</h4>
            <v-list dense>
              <v-list-item v-for="info in preflightResult.info" :key="info">
                <v-list-item-icon>
                  <v-icon color="info">mdi-information</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>{{ info }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </div>
        </v-card-text>
      </v-card>

      <!-- Quick Stats -->
      <v-row class="mb-6">
        <v-col cols="12" sm="6" md="3">
          <v-card>
            <v-card-text class="text-center">
              <v-icon size="48" color="primary" class="mb-2">mdi-file-document</v-icon>
              <div class="text-h4">{{ contents.length }}</div>
              <div class="text-body-2 text--secondary">{{ $t('compliance.scapContents') }}</div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-card>
            <v-card-text class="text-center">
              <v-icon size="48" color="success" class="mb-2">mdi-shield-check</v-icon>
              <div class="text-h4">{{ policies.length }}</div>
              <div class="text-body-2 text--secondary">{{ $t('compliance.policies') }}</div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-card>
            <v-card-text class="text-center">
              <v-icon size="48" color="info" class="mb-2">mdi-play-circle</v-icon>
              <div class="text-h4">{{ scans.length }}</div>
              <div class="text-body-2 text--secondary">{{ $t('compliance.scans') }}</div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-card>
            <v-card-text class="text-center">
              <v-icon size="48" color="warning" class="mb-2">mdi-chart-line</v-icon>
              <div class="text-h4">{{ reports.length }}</div>
              <div class="text-body-2 text--secondary">{{ $t('compliance.reports') }}</div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- Navigation Tabs -->
      <v-tabs v-model="activeTab" class="mb-6">
        <v-tab :to="`/project/${projectId}/compliance/contents`">
          <v-icon left>mdi-file-document</v-icon>
          {{ $t('compliance.contents') }}
        </v-tab>
        <v-tab :to="`/project/${projectId}/compliance/policies`">
          <v-icon left>mdi-shield-check</v-icon>
          {{ $t('compliance.policies') }}
        </v-tab>
        <v-tab :to="`/project/${projectId}/compliance/scans`">
          <v-icon left>mdi-play-circle</v-icon>
          {{ $t('compliance.scans') }}
        </v-tab>
        <v-tab :to="`/project/${projectId}/compliance/reports`">
          <v-icon left>mdi-chart-line</v-icon>
          {{ $t('compliance.reports') }}
        </v-tab>
      </v-tabs>

      <!-- Router View for sub-pages -->
      <router-view
        :projectId="projectId"
        :contents="contents"
        :policies="policies"
        :scans="scans"
        :reports="reports"
        @refresh="loadData"
      ></router-view>
    </v-container>
  </div>
</template>

<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  name: 'Compliance',

  props: {
    projectId: {
      type: Number,
      required: true,
    },
  },

  data() {
    return {
      activeTab: 0,
      contents: [],
      policies: [],
      scans: [],
      reports: [],
      preflightResult: null,
      preflightLoading: false,
    };
  },

  async mounted() {
    await this.loadData();
  },

  methods: {
    async loadData() {
      try {
        await Promise.all([
          this.loadContents(),
          this.loadPolicies(),
          this.loadScans(),
          this.loadReports(),
        ]);
      } catch (error) {
        this.$store.dispatch('showError', getErrorMessage(error));
      }
    },

    async loadContents() {
      try {
        const response = await axios.get(`/api/project/${this.projectId}/compliance/contents`);
        this.contents = response.data;
      } catch (error) {
        console.error('Failed to load contents:', error);
      }
    },

    async loadPolicies() {
      try {
        const response = await axios.get(`/api/project/${this.projectId}/compliance/policies`);
        this.policies = response.data;
      } catch (error) {
        console.error('Failed to load policies:', error);
      }
    },

    async loadScans() {
      try {
        const response = await axios.get(`/api/project/${this.projectId}/compliance/scans`);
        this.scans = response.data;
      } catch (error) {
        console.error('Failed to load scans:', error);
      }
    },

    async loadReports() {
      try {
        const response = await axios.get(`/api/project/${this.projectId}/compliance/reports`);
        this.reports = response.data;
      } catch (error) {
        console.error('Failed to load reports:', error);
      }
    },

    async checkPreflight() {
      this.preflightLoading = true;
      try {
        const response = await axios.get(`/api/project/${this.projectId}/compliance/preflight`);
        this.preflightResult = response.data;
      } catch (error) {
        this.$store.dispatch('showError', getErrorMessage(error));
      } finally {
        this.preflightLoading = false;
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.compliance-container {
  min-height: 100vh;
  background-color: #f5f5f5;
}
</style>
