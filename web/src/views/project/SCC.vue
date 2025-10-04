<template>
  <div class="scc-compliance">
    <v-toolbar flat>
      <v-toolbar-title>{{ $t('scc.sccCompliance') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn color="primary" @click="checkAvailability">
        <v-icon left>mdi-check-circle</v-icon>
        {{ $t('scc.checkAvailability') }}
      </v-btn>
    </v-toolbar>

    <v-container fluid>
      <!-- SCC Status Card -->
      <v-card class="mb-6">
        <v-card-title>
          <v-icon class="mr-2">mdi-information</v-icon>
          {{ $t('scc.sccStatus') }}
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12" md="6">
              <div class="text-h6">{{ $t('scc.availability') }}</div>
              <v-chip :color="sccAvailable ? 'success' : 'error'" class="mt-2">
                <v-icon left>{{ sccAvailable ? 'mdi-check' : 'mdi-close' }}</v-icon>
                {{ sccAvailable ? $t('scc.available') : $t('scc.notAvailable') }}
              </v-chip>
              <div v-if="sccVersion" class="text-caption mt-1">
                {{ sccVersion }}
              </div>
            </v-col>
            <v-col cols="12" md="6">
              <div class="text-h6">{{ $t('scc.scanStatistics') }}</div>
              <div class="mt-2">
                <v-chip color="primary" class="mr-2">
                  {{ totalScans }} {{ $t('scc.totalScans') }}
                </v-chip>
                <v-chip color="success" class="mr-2">
                  {{ completedScans }} {{ $t('scc.completed') }}
                </v-chip>
                <v-chip color="error">{{ failedScans }} {{ $t('scc.failed') }}</v-chip>
              </div>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- Supported OS -->
      <v-card class="mb-6">
        <v-card-title>
          <v-icon class="mr-2">mdi-server</v-icon>
          {{ $t('scc.supportedOperatingSystems') }}
        </v-card-title>
        <v-card-text>
          <v-chip-group>
            <v-chip
              v-for="os in supportedOS"
              :key="os"
              color="primary"
              outlined
              @click="loadBenchmarksForOS(os)"
            >
              {{ os }}
            </v-chip>
          </v-chip-group>
        </v-card-text>
      </v-card>

      <!-- Available Benchmarks -->
      <v-card class="mb-6">
        <v-card-title>
          <v-icon class="mr-2">mdi-shield-check</v-icon>
          {{ $t('scc.availableBenchmarks') }}
          <v-spacer></v-spacer>
          <v-text-field
            v-model="benchmarkSearch"
            append-icon="mdi-magnify"
            :label="$t('search')"
            single-line
            hide-details
            class="mt-0"
          ></v-text-field>
        </v-card-title>
        <v-card-text>
          <v-data-table
            :headers="benchmarkHeaders"
            :items="filteredBenchmarks"
            :search="benchmarkSearch"
            :loading="loadingBenchmarks"
            item-key="fileName"
            class="elevation-1"
          >
            <template v-slot:item.title="{ item }">
              <div>
                <strong>{{ item.title }}</strong>
                <div class="text-caption">{{ item.date }}</div>
              </div>
            </template>
            <template v-slot:item.os="{ item }">
              <v-chip small>{{ item.os }}</v-chip>
            </template>
            <template v-slot:item.type="{ item }">
              <v-chip small :color="item.type === 'STIG' ? 'error' : 'primary'">
                {{ item.type }}
              </v-chip>
            </template>
            <template v-slot:item.actions="{ item }">
              <v-btn
                small
                color="primary"
                @click="downloadBenchmark(item)"
                :loading="downloadingBenchmark === item.fileName"
              >
                <v-icon small left>mdi-download</v-icon>
                {{ $t('download') }}
              </v-btn>
            </template>
          </v-data-table>
        </v-card-text>
      </v-card>

      <!-- Quick Scan -->
      <v-card class="mb-6">
        <v-card-title>
          <v-icon class="mr-2">mdi-play</v-icon>
          {{ $t('scc.quickScan') }}
        </v-card-title>
        <v-card-text>
          <v-form ref="scanForm" v-model="scanFormValid">
            <v-row>
              <v-col cols="12" md="6">
                <v-select
                  v-model="scanForm.host"
                  :items="availableHosts"
                  :label="$t('scc.targetHost')"
                  outlined
                  dense
                  required
                  :rules="[v => !!v || $t('scc.hostRequired')]"
                ></v-select>
              </v-col>
              <v-col cols="12" md="6">
                <v-select
                  v-model="scanForm.benchmark"
                  :items="availableBenchmarks"
                  :label="$t('scc.benchmark')"
                  outlined
                  dense
                  required
                  :rules="[v => !!v || $t('scc.benchmarkRequired')]"
                  @change="onBenchmarkChange"
                ></v-select>
              </v-col>
            </v-row>
            <v-row>
              <v-col cols="12" md="6">
                <v-select
                  v-model="scanForm.profile"
                  :items="availableProfiles"
                  :label="$t('scc.profile')"
                  outlined
                  dense
                  required
                  :rules="[v => !!v || $t('scc.profileRequired')]"
                  :disabled="!scanForm.benchmark"
                ></v-select>
              </v-col>
              <v-col cols="12" md="6">
                <v-btn
                  color="primary"
                  :disabled="!scanFormValid"
                  :loading="runningScan"
                  @click="runScan"
                  class="mt-4"
                >
                  <v-icon left>mdi-play</v-icon>
                  {{ $t('scc.runScan') }}
                </v-btn>
              </v-col>
            </v-row>
          </v-form>
        </v-card-text>
      </v-card>

      <!-- Scan History -->
      <v-card>
        <v-card-title>
          <v-icon class="mr-2">mdi-history</v-icon>
          {{ $t('scc.scanHistory') }}
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="refreshScanHistory">
            <v-icon left>mdi-refresh</v-icon>
            {{ $t('refresh') }}
          </v-btn>
        </v-card-title>
        <v-card-text>
          <v-data-table
            :headers="scanHistoryHeaders"
            :items="scanHistory"
            :loading="loadingScanHistory"
            item-key="id"
            class="elevation-1"
          >
            <template v-slot:item.status="{ item }">
              <v-chip
                :color="getStatusColor(item.status)"
                small
              >
                {{ item.status }}
              </v-chip>
            </template>
            <template v-slot:item.started="{ item }">
              {{ formatDate(item.started) }}
            </template>
            <template v-slot:item.ended="{ item }">
              {{ item.ended ? formatDate(item.ended) : '-' }}
            </template>
            <template v-slot:item.actions="{ item }">
              <v-btn
                small
                color="primary"
                @click="viewScanResults(item)"
                :disabled="item.status !== 'completed'"
              >
                <v-icon small left>mdi-eye</v-icon>
                {{ $t('view') }}
              </v-btn>
            </template>
          </v-data-table>
        </v-card-text>
      </v-card>
    </v-container>

    <!-- Scan Results Dialog -->
    <v-dialog v-model="resultsDialog" max-width="1200px">
      <v-card>
        <v-card-title>
          {{ $t('scc.scanResults') }} - {{ selectedScan?.host }}
          <v-spacer></v-spacer>
          <v-btn icon @click="resultsDialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text>
          <div v-if="loadingResults" class="text-center">
            <v-progress-circular indeterminate></v-progress-circular>
          </div>
          <div v-else-if="scanResults.length === 0" class="text-center">
            <v-alert type="info">{{ $t('scc.noResultsFound') }}</v-alert>
          </div>
          <div v-else>
            <v-data-table
              :headers="resultsHeaders"
              :items="scanResults"
              item-key="id"
              class="elevation-1"
            >
              <template v-slot:item.result="{ item }">
                <v-chip
                  :color="getResultColor(item.result)"
                  small
                >
                  {{ item.result }}
                </v-chip>
              </template>
              <template v-slot:item.created="{ item }">
                {{ formatDate(item.created) }}
              </template>
              <template v-slot:item.actions="{ item }">
                <v-btn
                  small
                  color="primary"
                  @click="viewRuleResults(item)"
                >
                  <v-icon small left>mdi-list</v-icon>
                  {{ $t('scc.viewRules') }}
                </v-btn>
              </template>
            </v-data-table>
          </div>
        </v-card-text>
      </v-card>
    </v-dialog>

    <!-- Rule Results Dialog -->
    <v-dialog v-model="rulesDialog" max-width="1400px">
      <v-card>
        <v-card-title>
          {{ $t('scc.ruleResults') }} - {{ selectedReport?.host }}
          <v-spacer></v-spacer>
          <v-btn icon @click="rulesDialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text>
          <div v-if="loadingRules" class="text-center">
            <v-progress-circular indeterminate></v-progress-circular>
          </div>
          <div v-else-if="ruleResults.length === 0" class="text-center">
            <v-alert type="info">{{ $t('scc.noRulesFound') }}</v-alert>
          </div>
          <div v-else>
            <v-data-table
              :headers="rulesHeaders"
              :items="ruleResults"
              item-key="id"
              class="elevation-1"
            >
              <template v-slot:item.result="{ item }">
                <v-chip
                  :color="getResultColor(item.result)"
                  small
                >
                  {{ item.result }}
                </v-chip>
              </template>
              <template v-slot:item.severity="{ item }">
                <v-chip
                  :color="getSeverityColor(item.severity)"
                  small
                  outlined
                >
                  {{ item.severity }}
                </v-chip>
              </template>
            </v-data-table>
          </div>
        </v-card-text>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'SCC',

  data() {
    return {
      // SCC Status
      sccAvailable: false,
      sccVersion: '',
      totalScans: 0,
      completedScans: 0,
      failedScans: 0,
      supportedOS: [],

      // Benchmarks
      benchmarks: [],
      loadingBenchmarks: false,
      benchmarkSearch: '',
      downloadingBenchmark: null,

      // Scan Form
      scanFormValid: false,
      scanForm: {
        host: '',
        benchmark: '',
        profile: '',
      },
      availableHosts: [],
      availableBenchmarks: [],
      availableProfiles: [],
      runningScan: false,

      // Scan History
      scanHistory: [],
      loadingScanHistory: false,

      // Results
      resultsDialog: false,
      rulesDialog: false,
      selectedScan: null,
      selectedReport: null,
      scanResults: [],
      ruleResults: [],
      loadingResults: false,
      loadingRules: false,

      // Headers
      benchmarkHeaders: [
        { text: this.$t('title'), value: 'title' },
        { text: this.$t('scc.os'), value: 'os' },
        { text: this.$t('scc.type'), value: 'type' },
        { text: this.$t('scc.version'), value: 'stigVersion' },
        { text: this.$t('actions'), value: 'actions', sortable: false },
      ],
      scanHistoryHeaders: [
        { text: this.$t('scc.host'), value: 'host' },
        { text: this.$t('scc.status'), value: 'status' },
        { text: this.$t('scc.started'), value: 'started' },
        { text: this.$t('scc.ended'), value: 'ended' },
        { text: this.$t('actions'), value: 'actions', sortable: false },
      ],
      resultsHeaders: [
        { text: this.$t('scc.host'), value: 'host' },
        { text: this.$t('scc.result'), value: 'result' },
        { text: this.$t('scc.created'), value: 'created' },
        { text: this.$t('actions'), value: 'actions', sortable: false },
      ],
      rulesHeaders: [
        { text: this.$t('scc.ruleId'), value: 'ruleId' },
        { text: this.$t('scc.title'), value: 'ident' },
        { text: this.$t('scc.result'), value: 'result' },
        { text: this.$t('scc.severity'), value: 'severity' },
        { text: this.$t('scc.description'), value: 'description' },
      ],
    };
  },

  computed: {
    projectId() {
      return parseInt(this.$route.params.id, 10);
    },
    filteredBenchmarks() {
      return this.benchmarks.filter((b) => !this.benchmarkSearch
        || b.title.toLowerCase().includes(this.benchmarkSearch.toLowerCase())
        || b.os.toLowerCase().includes(this.benchmarkSearch.toLowerCase()));
    },
  },

  async created() {
    await this.initializeData();
  },

  methods: {
    async initializeData() {
      await Promise.all([
        this.checkAvailability(),
        this.loadBenchmarks(),
        this.loadScanHistory(),
        this.loadAvailableHosts(),
      ]);
    },

    async checkAvailability() {
      try {
        const response = await axios.get(`/api/project/${this.projectId}/scc/status`);
        this.sccAvailable = response.data.scc_available;
        this.sccVersion = response.data.scc_version;
        this.totalScans = response.data.total_scans || 0;
        this.completedScans = response.data.completed_scans || 0;
        this.failedScans = response.data.failed_scans || 0;
        this.supportedOS = response.data.supported_os || [];
      } catch (error) {
        console.error('Failed to check SCC availability:', error);
        this.$store.dispatch('showError', 'Failed to check SCC availability.');
      }
    },

    async loadBenchmarks() {
      this.loadingBenchmarks = true;
      try {
        const response = await axios.get('/api/scc/benchmarks');
        this.benchmarks = response.data;
        this.availableBenchmarks = this.benchmarks.map((b) => ({
          text: b.title,
          value: b.fileName,
        }));
      } catch (error) {
        console.error('Failed to load benchmarks:', error);
        this.$store.dispatch('showError', 'Failed to load benchmarks.');
      } finally {
        this.loadingBenchmarks = false;
      }
    },

    async loadBenchmarksForOS(os) {
      try {
        const response = await axios.get(`/api/scc/os/${os}/benchmarks`);
        this.benchmarks = response.data;
        this.availableBenchmarks = this.benchmarks.map((b) => ({
          text: b.title,
          value: b.fileName,
        }));
      } catch (error) {
        console.error(`Failed to load benchmarks for ${os}:`, error);
        this.$store.dispatch('showError', `Failed to load benchmarks for ${os}.`);
      }
    },

    async downloadBenchmark(benchmark) {
      this.downloadingBenchmark = benchmark.fileName;
      try {
        await axios.post('/api/scc/benchmarks/download', benchmark);
        this.$store.dispatch(
          'showSuccess',
          `Benchmark ${benchmark.title} downloaded successfully.`,
        );
      } catch (error) {
        console.error('Failed to download benchmark:', error);
        this.$store.dispatch('showError', 'Failed to download benchmark.');
      } finally {
        this.downloadingBenchmark = null;
      }
    },

    async onBenchmarkChange() {
      if (!this.scanForm.benchmark) {
        this.availableProfiles = [];
        return;
      }

      try {
        const response = await axios.get(`/api/scc/benchmarks/${this.scanForm.benchmark}/profiles`);
        this.availableProfiles = response.data.map((p) => ({
          text: p,
          value: p,
        }));
      } catch (error) {
        console.error('Failed to load profiles:', error);
        this.$store.dispatch('showError', 'Failed to load profiles.');
      }
    },

    async runScan() {
      if (!this.$refs.scanForm.validate()) return;

      this.runningScan = true;
      try {
        await axios.post(`/api/project/${this.projectId}/scc/scan`, {
          host: this.scanForm.host,
          benchmark: this.scanForm.benchmark,
          profile: this.scanForm.profile,
        });

        this.$store.dispatch('showSuccess', 'SCC scan started successfully.');
        this.scanForm = { host: '', benchmark: '', profile: '' };
        this.$refs.scanForm.reset();
        await this.loadScanHistory();
      } catch (error) {
        console.error('Failed to run scan:', error);
        this.$store.dispatch('showError', 'Failed to run scan.');
      } finally {
        this.runningScan = false;
      }
    },

    async loadScanHistory() {
      this.loadingScanHistory = true;
      try {
        const response = await axios.get(`/api/project/${this.projectId}/scc/scans`);
        this.scanHistory = response.data;
      } catch (error) {
        console.error('Failed to load scan history:', error);
        this.$store.dispatch('showError', 'Failed to load scan history.');
      } finally {
        this.loadingScanHistory = false;
      }
    },

    async refreshScanHistory() {
      await this.loadScanHistory();
    },

    async viewScanResults(scan) {
      this.selectedScan = scan;
      this.resultsDialog = true;
      this.loadingResults = true;

      try {
        const response = await axios.get(
          `/api/project/${this.projectId}/scc/scans/${scan.id}/results`,
        );
        this.scanResults = response.data;
      } catch (error) {
        console.error('Failed to load scan results:', error);
        this.$store.dispatch('showError', 'Failed to load scan results.');
      } finally {
        this.loadingResults = false;
      }
    },

    async viewRuleResults(report) {
      this.selectedReport = report;
      this.rulesDialog = true;
      this.loadingRules = true;

      try {
        const response = await axios.get(
          `/api/project/${this.projectId}/scc/reports/${report.id}/rules`,
        );
        this.ruleResults = response.data;
      } catch (error) {
        console.error('Failed to load rule results:', error);
        this.$store.dispatch('showError', 'Failed to load rule results.');
      } finally {
        this.loadingRules = false;
      }
    },

    async loadAvailableHosts() {
      try {
        // This would typically come from the inventory
        // For now, we'll use a placeholder
        this.availableHosts = [
          'localhost',
          'server1.example.com',
          'server2.example.com',
        ];
      } catch (error) {
        console.error('Failed to load hosts:', error);
      }
    },

    getStatusColor(status) {
      switch (status) {
        case 'completed': return 'success';
        case 'running': return 'primary';
        case 'failed': return 'error';
        default: return 'grey';
      }
    },

    getResultColor(result) {
      switch (result) {
        case 'pass': return 'success';
        case 'fail': return 'error';
        case 'error': return 'error';
        case 'notapplicable': return 'warning';
        default: return 'grey';
      }
    },

    getSeverityColor(severity) {
      switch (severity?.toLowerCase()) {
        case 'high': return 'error';
        case 'medium': return 'warning';
        case 'low': return 'info';
        default: return 'grey';
      }
    },

    formatDate(dateString) {
      if (!dateString) return '-';
      const date = new Date(dateString);
      return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`;
    },
  },
};
</script>

<style scoped>
.scc-compliance {
  padding: 20px;
}
</style>
