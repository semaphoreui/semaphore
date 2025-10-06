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
      <v-toolbar-title>
        <v-icon class="mr-2">mdi-shield-check</v-icon>
        Compliance Dashboard
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        icon
        @click="refreshData"
        :loading="loading"
      >
        <v-icon>mdi-refresh</v-icon>
      </v-btn>
    </v-toolbar>

    <!-- Date Range and Project Filters -->
    <v-card class="mx-4 mt-4">
      <v-card-text>
        <v-row>
          <v-col cols="12" md="3">
            <v-select
              v-model="dateRange"
              :items="dateRanges"
              label="Date Range"
              dense
              outlined
              @change="refreshData"
            ></v-select>
          </v-col>
          <v-col cols="12" md="3">
            <v-select
              v-model="selectedProject"
              :items="projectOptions"
              label="Project Filter"
              dense
              outlined
              clearable
              @change="refreshData"
            ></v-select>
          </v-col>
          <v-col cols="12" md="3">
            <v-btn
              color="primary"
              @click="exportReport"
              :loading="exporting"
              class="mt-2"
            >
              <v-icon left>mdi-download</v-icon>
              Export Report
            </v-btn>
          </v-col>
          <v-col cols="12" md="3">
            <v-chip
              v-if="dashboardData.last_updated"
              color="success"
              text-color="white"
              class="mt-2"
            >
              <v-icon left small>mdi-clock</v-icon>
              Last updated: {{ formatDate(dashboardData.last_updated) }}
            </v-chip>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <!-- Summary Cards -->
    <v-row class="mx-4 mt-4">
      <v-col cols="12" sm="6" md="3">
        <v-card color="primary" dark>
          <v-card-text>
            <div class="d-flex align-center">
              <v-icon size="40" class="mr-3">mdi-play-circle</v-icon>
              <div>
                <div class="text-h4">{{ dashboardData.summary?.total_tasks || 0 }}</div>
                <div class="text-subtitle-1">Total Tasks</div>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card color="success" dark>
          <v-card-text>
            <div class="d-flex align-center">
              <v-icon size="40" class="mr-3">mdi-check-circle</v-icon>
              <div>
                <div class="text-h4">
                  {{ dashboardData.summary?.success_rate?.toFixed(1) || 0 }}%
                </div>
                <div class="text-subtitle-1">Success Rate</div>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card color="info" dark>
          <v-card-text>
            <div class="d-flex align-center">
              <v-icon size="40" class="mr-3">mdi-account-group</v-icon>
              <div>
                <div class="text-h4">{{ dashboardData.summary?.active_users || 0 }}</div>
                <div class="text-subtitle-1">Active Users</div>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card color="warning" dark>
          <v-card-text>
            <div class="d-flex align-center">
              <v-icon size="40" class="mr-3">mdi-shield-alert</v-icon>
              <div>
                <div class="text-h4">{{ dashboardData.summary?.security_incidents || 0 }}</div>
                <div class="text-subtitle-1">Security Incidents</div>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Charts Row -->
    <v-row class="mx-4 mt-4">
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title>
            <v-icon class="mr-2">mdi-chart-line</v-icon>
            Task Trends
          </v-card-title>
          <v-card-text>
            <ComplianceTrendsChart
              :data="dashboardData.compliance_trends?.daily_tasks"
              title="Task Execution Trends"
              color="#667eea"
              chart-type="area"
              legend-label="Tasks"
            />
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="6">
        <v-card>
          <v-card-title>
            <v-icon class="mr-2">mdi-chart-pie</v-icon>
            Success Rate Trends
          </v-card-title>
          <v-card-text>
            <ComplianceTrendsChart
              :data="dashboardData.compliance_trends?.success_rates"
              title="Success Rate Trends"
              color="#28a745"
              chart-type="line"
              legend-label="Success Rate %"
            />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Data Tables Row -->
    <v-row class="mx-4 mt-4">
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title>
            <v-icon class="mr-2">mdi-format-list-bulleted</v-icon>
            Task Compliance
          </v-card-title>
          <v-card-text>
            <TaskComplianceTable
              :data="dashboardData.task_compliance"
              @view-task="handleViewTask"
            />
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="6">
        <v-card>
          <v-card-title>
            <v-icon class="mr-2">mdi-account-multiple</v-icon>
            User Activity
          </v-card-title>
          <v-card-text>
            <UserActivityTable :data="dashboardData.user_activity" />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Project Compliance and Security Events -->
    <v-row class="mx-4 mt-4">
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title>
            <v-icon class="mr-2">mdi-folder-multiple</v-icon>
            Project Compliance
          </v-card-title>
          <v-card-text>
            <ProjectComplianceTable :data="dashboardData.project_compliance" />
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="6">
        <v-card>
          <v-card-title>
            <v-icon class="mr-2">mdi-security</v-icon>
            Security Events
          </v-card-title>
          <v-card-text>
            <SecurityEventsTable :data="dashboardData.security_events" />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Loading Overlay -->
    <v-overlay :value="loading" z-index="9999">
      <v-progress-circular
        indeterminate
        size="64"
        color="primary"
      ></v-progress-circular>
    </v-overlay>
  </div>
</template>

<script>
import axios from 'axios';
import complianceService from '@/lib/complianceService';
import ComplianceTrendsChart from '@/components/compliance/ComplianceTrendsChart.vue';
import TaskComplianceTable from '@/components/compliance/TaskComplianceTable.vue';
import UserActivityTable from '@/components/compliance/UserActivityTable.vue';
import ProjectComplianceTable from '@/components/compliance/ProjectComplianceTable.vue';
import SecurityEventsTable from '@/components/compliance/SecurityEventsTable.vue';

export default {
  name: 'ComplianceDashboard',

  components: {
    ComplianceTrendsChart,
    TaskComplianceTable,
    UserActivityTable,
    ProjectComplianceTable,
    SecurityEventsTable,
  },

  data() {
    return {
      loading: false,
      exporting: false,
      dashboardData: {},
      dateRange: 'last_month',
      selectedProject: null,
      dateRanges: [
        { text: 'Last 7 days', value: 'last_week' },
        { text: 'Last 30 days', value: 'last_month' },
        { text: 'Last 90 days', value: 'last_quarter' },
        { text: 'Last year', value: 'last_year' },
      ],
      projectOptions: [
        { text: 'All Projects', value: null },
      ],
      currentProjectId: null,
    };
  },

  computed: {
    daysParam() {
      const daysMap = {
        last_week: 7,
        last_month: 30,
        last_quarter: 90,
        last_year: 365,
      };
      return daysMap[this.dateRange] || 30;
    },
  },

  async created() {
    // Check if we're in a project context
    this.currentProjectId = this.$route.params.projectId || null;

    await this.loadProjects();

    // Set default project filter based on context
    if (this.currentProjectId) {
      // If accessed from project context, default to current project
      this.selectedProject = this.currentProjectId.toString();
    } else {
      // If accessed from global context, default to "All Projects"
      this.selectedProject = null;
    }

    await this.refreshData();
  },

  methods: {
    async loadProjects() {
      try {
        const response = await axios.get('/api/projects');
        this.projectOptions = [
          { text: 'All Projects', value: null },
          ...response.data.map((project) => ({
            text: project.name,
            value: project.id.toString(),
          })),
        ];
      } catch (error) {
        console.error('Failed to load projects:', error);
      }
    },

    async refreshData() {
      this.loading = true;
      try {
        const params = {
          days: this.daysParam,
        };

        if (this.selectedProject) {
          params.project_id = this.selectedProject;
        }

        this.dashboardData = await complianceService.getDashboardData(params);
      } catch (error) {
        console.error('Failed to load compliance data:', error);
        this.$emit('error', 'Failed to load compliance dashboard data');
      } finally {
        this.loading = false;
      }
    },

    async exportReport() {
      this.exporting = true;
      try {
        const params = {
          days: this.daysParam,
          format: 'csv',
        };

        if (this.selectedProject) {
          params.project_id = this.selectedProject;
        }

        const blob = await complianceService.exportReport(params);
        const filename = `compliance-report-${new Date().toISOString().split('T')[0]}.csv`;
        complianceService.downloadReport(blob, filename);
      } catch (error) {
        console.error('Failed to export report:', error);
        this.$emit('error', 'Failed to export compliance report');
      } finally {
        this.exporting = false;
      }
    },

    formatDate(dateString) {
      if (!dateString) return '';
      return new Date(dateString).toLocaleString();
    },

    returnToProjects() {
      if (this.currentProjectId) {
        // If accessed from project context, go back to project dashboard
        this.$router.push(`/project/${this.currentProjectId}`);
      } else {
        // If accessed from global context, go to main page
        this.$router.push('/');
      }
    },

    showDrawer() {
      this.$emit('i-show-drawer');
    },

    handleViewTask(task) {
      // Navigate to the task details screen using the t query parameter
      const route = `/project/${task.project_id}/templates/${task.template_id}/tasks?t=${
        task.task_id
      }`;
      this.$router.push(route);
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/assets/scss/compliance-dashboard.scss';

.compliance-dashboard {
  &__container {
    padding: 16px;
    background-color: #f5f5f5;
    min-height: 100vh;
  }

  &__summary-cards {
    margin-bottom: 24px;

    .v-card {
      transition: all 0.3s ease;
      border-radius: 12px;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
      }
    }
  }

  &__charts {
    margin-bottom: 24px;

    .v-card {
      border-radius: 12px;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    }

    .v-card__title {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      border-radius: 12px 12px 0 0;
      padding: 16px 20px;
    }
  }

  &__tables {
    .v-card {
      border-radius: 12px;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      overflow: hidden;
    }

    .v-card__title {
      background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
      color: white;
      padding: 16px 20px;
    }
  }

  &__filters {
    background: white;
    border-radius: 12px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    margin-bottom: 24px;
    padding: 20px;
  }
}

// Responsive design
@media (max-width: 768px) {
  .compliance-dashboard {
    &__container {
      padding: 8px;
    }

    &__summary-cards {
      .v-col {
        margin-bottom: 16px;
      }
    }
  }
}

// Animation classes
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter,
.fade-leave-to {
  opacity: 0;
}
</style>
