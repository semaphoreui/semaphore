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
      <v-toolbar-title>Alert Configuration</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <v-divider />

    <v-container>
      <v-alert
        v-if="success"
        type="success"
        dismissible
        @input="success = null"
        class="mb-4"
      >
        {{ success }}
      </v-alert>

      <v-alert
        v-if="error"
        type="error"
        dismissible
        @input="error = null"
        class="mb-4"
      >
        {{ error }}
      </v-alert>

      <v-card>
        <v-tabs v-model="activeTab" background-color="transparent" color="primary">
          <v-tab key="alert-types">
            <v-icon left>mdi-alert</v-icon>
            Alert Types
          </v-tab>
          <v-tab key="integrations">
            <v-icon left>mdi-webhook</v-icon>
            Alert Integrations
          </v-tab>
        </v-tabs>

        <v-tabs-items v-model="activeTab">
          <!-- Alert Types Tab -->
          <v-tab-item key="alert-types">
            <v-card-text>
              <h3 class="text-h6 mb-4">Configure Alert Types</h3>

              <v-row>
                <v-col cols="12" md="6">
                  <v-switch
                    v-model="alertConfig.types.taskFailure"
                    inset
                    label="Task Failure"
                    class="mt-0"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-switch
                    v-model="alertConfig.types.taskSuccess"
                    inset
                    label="Task Success"
                    class="mt-0"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-switch
                    v-model="alertConfig.types.projectCreated"
                    inset
                    label="Project Created"
                    class="mt-0"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-switch
                    v-model="alertConfig.types.userLogin"
                    inset
                    label="User Login"
                    class="mt-0"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-switch
                    v-model="alertConfig.types.systemError"
                    inset
                    label="System Error"
                    class="mt-0"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-switch
                    v-model="alertConfig.types.securityEvent"
                    inset
                    label="Security Event"
                    class="mt-0"
                  />
                </v-col>
              </v-row>
            </v-card-text>
          </v-tab-item>

          <!-- Alert Integrations Tab -->
          <v-tab-item key="integrations">
            <v-card-text>
              <h3 class="text-h6 mb-4">Configure Alert Integrations</h3>

              <v-row>
                <v-col cols="12" md="6" class="pl-4">
                  <v-switch
                    v-model="alertConfig.integrations.slack"
                    inset
                    label="Slack"
                    class="mt-0"
                  />
                </v-col>
                <v-col cols="12" md="6" class="pl-4">
                  <v-switch
                    v-model="alertConfig.integrations.teams"
                    inset
                    label="Microsoft Teams"
                    class="mt-0"
                  />
                </v-col>
                <v-col cols="12" md="6" class="pl-4">
                  <v-switch
                    v-model="alertConfig.integrations.email"
                    inset
                    label="Email"
                    class="mt-0"
                  />
                </v-col>
                <v-col cols="12" md="6" class="pl-4">
                  <v-switch
                    v-model="alertConfig.integrations.webhook"
                    inset
                    label="Webhook"
                    class="mt-0"
                  />
                </v-col>
                <v-col cols="12" md="6" class="pl-4">
                  <v-switch
                    v-model="alertConfig.integrations.discord"
                    inset
                    label="Discord"
                    class="mt-0"
                  />
                </v-col>
                <v-col cols="12" md="6" class="pl-4">
                  <v-switch
                    v-model="alertConfig.integrations.pagerduty"
                    inset
                    label="PagerDuty"
                    class="mt-0"
                  />
                </v-col>
              </v-row>

              <!-- Integration Configuration -->
              <v-expansion-panels v-if="hasEnabledIntegrations" class="mt-6">
                <v-expansion-panel v-if="alertConfig.integrations.slack">
                  <v-expansion-panel-header>
                    <v-icon left>mdi-slack</v-icon>
                    Slack Configuration
                  </v-expansion-panel-header>
                  <v-expansion-panel-content>
                    <v-text-field
                      v-model="alertConfig.slack.webhookUrl"
                      label="Webhook URL"
                      outlined
                      dense
                      hint="Enter your Slack webhook URL"
                    />
                    <v-text-field
                      v-model="alertConfig.slack.channel"
                      label="Channel"
                      outlined
                      dense
                      hint="Enter the channel name (e.g., #alerts)"
                    />
                  </v-expansion-panel-content>
                </v-expansion-panel>

                <v-expansion-panel v-if="alertConfig.integrations.teams">
                  <v-expansion-panel-header>
                    <v-icon left>mdi-microsoft-teams</v-icon>
                    Microsoft Teams Configuration
                  </v-expansion-panel-header>
                  <v-expansion-panel-content>
                    <v-text-field
                      v-model="alertConfig.teams.webhookUrl"
                      label="Webhook URL"
                      outlined
                      dense
                      hint="Enter your Teams webhook URL"
                    />
                  </v-expansion-panel-content>
                </v-expansion-panel>

                <v-expansion-panel v-if="alertConfig.integrations.email">
                  <v-expansion-panel-header>
                    <v-icon left>mdi-email</v-icon>
                    Email Configuration
                  </v-expansion-panel-header>
                  <v-expansion-panel-content>
                    <v-text-field
                      v-model="alertConfig.email.recipients"
                      label="Recipients"
                      outlined
                      dense
                      hint="Enter email addresses separated by commas"
                    />
                    <v-text-field
                      v-model="alertConfig.email.subject"
                      label="Subject Template"
                      outlined
                      dense
                      hint="Enter email subject template"
                    />
                  </v-expansion-panel-content>
                </v-expansion-panel>

                <v-expansion-panel v-if="alertConfig.integrations.webhook">
                  <v-expansion-panel-header>
                    <v-icon left>mdi-webhook</v-icon>
                    Webhook Configuration
                  </v-expansion-panel-header>
                  <v-expansion-panel-content>
                    <v-text-field
                      v-model="alertConfig.webhook.url"
                      label="Webhook URL"
                      outlined
                      dense
                      hint="Enter your webhook URL"
                    />
                    <v-text-field
                      v-model="alertConfig.webhook.secret"
                      label="Secret"
                      outlined
                      dense
                      type="password"
                      hint="Enter webhook secret for authentication"
                    />
                  </v-expansion-panel-content>
                </v-expansion-panel>

                <v-expansion-panel v-if="alertConfig.integrations.discord">
                  <v-expansion-panel-header>
                    <v-icon left>mdi-discord</v-icon>
                    Discord Configuration
                  </v-expansion-panel-header>
                  <v-expansion-panel-content>
                    <v-text-field
                      v-model="alertConfig.discord.webhookUrl"
                      label="Webhook URL"
                      outlined
                      dense
                      hint="Enter your Discord webhook URL"
                    />
                    <v-text-field
                      v-model="alertConfig.discord.username"
                      label="Username"
                      outlined
                      dense
                      hint="Bot username for Discord messages"
                    />
                    <v-text-field
                      v-model="alertConfig.discord.avatarUrl"
                      label="Avatar URL"
                      outlined
                      dense
                      hint="Optional avatar URL for the bot"
                    />
                  </v-expansion-panel-content>
                </v-expansion-panel>

                <v-expansion-panel v-if="alertConfig.integrations.pagerduty">
                  <v-expansion-panel-header>
                    <v-icon left>mdi-bell-alert</v-icon>
                    PagerDuty Configuration
                  </v-expansion-panel-header>
                  <v-expansion-panel-content>
                    <v-text-field
                      v-model="alertConfig.pagerduty.integrationKey"
                      label="Integration Key"
                      outlined
                      dense
                      type="password"
                      hint="Enter your PagerDuty integration key"
                    />
                    <v-text-field
                      v-model="alertConfig.pagerduty.serviceId"
                      label="Service ID"
                      outlined
                      dense
                      hint="Enter your PagerDuty service ID"
                    />
                    <v-select
                      v-model="alertConfig.pagerduty.severity"
                      :items="pagerdutySeverityOptions"
                      label="Default Severity"
                      outlined
                      dense
                      hint="Default severity level for alerts"
                    />
                  </v-expansion-panel-content>
                </v-expansion-panel>
              </v-expansion-panels>
            </v-card-text>
          </v-tab-item>
        </v-tabs-items>

        <v-card-actions class="pa-4">
          <v-spacer></v-spacer>
          <v-btn
            color="primary"
            @click="testAlerts"
            :loading="testing"
            :disabled="!hasEnabledIntegrations"
            class="mr-2"
          >
            <v-icon left>mdi-test-tube</v-icon>
            Test Alerts
          </v-btn>
          <v-btn
            color="success"
            @click="saveAlerts"
            :loading="saving"
          >
            <v-icon left>mdi-content-save</v-icon>
            Save Configuration
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-container>
  </div>
</template>

<script>
import EventBus from '@/event-bus';

export default {
  name: 'Alerts',
  data() {
    return {
      activeTab: 0,
      saving: false,
      testing: false,
      success: null,
      error: null,
      pagerdutySeverityOptions: [
        { text: 'Critical', value: 'critical' },
        { text: 'Error', value: 'error' },
        { text: 'Warning', value: 'warning' },
        { text: 'Info', value: 'info' },
      ],
      alertConfig: {
        types: {
          taskFailure: true,
          taskSuccess: false,
          projectCreated: true,
          userLogin: false,
          systemError: true,
          securityEvent: true,
        },
        integrations: {
          slack: false,
          teams: false,
          email: true,
          webhook: false,
          discord: false,
          pagerduty: false,
        },
        slack: {
          webhookUrl: '',
          channel: '#alerts',
        },
        teams: {
          webhookUrl: '',
        },
        email: {
          recipients: '',
          subject: 'Forge Alert: {type}',
        },
        webhook: {
          url: '',
          secret: '',
        },
        discord: {
          webhookUrl: '',
          username: 'Forge Bot',
          avatarUrl: '',
        },
        pagerduty: {
          integrationKey: '',
          serviceId: '',
          severity: 'warning',
        },
      },
    };
  },
  computed: {
    hasEnabledIntegrations() {
      return Object.values(this.alertConfig.integrations).some(
        (enabled) => enabled,
      );
    },
  },
  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },
    async saveAlerts() {
      this.saving = true;
      this.error = null;
      this.success = null;

      try {
        // TODO: Implement actual API endpoint for saving alert configuration
        console.log('Saving alert configuration:', this.alertConfig);

        // Simulate API call for now
        await new Promise((resolve) => {
          setTimeout(resolve, 1000);
        });

        this.success = 'Alert configuration saved successfully!';
        setTimeout(() => {
          this.success = null;
        }, 5000);
      } catch (err) {
        console.error('Error saving alert configuration:', err);
        this.error = err.response?.data?.message || 'Failed to save alert configuration';
      } finally {
        this.saving = false;
      }
    },
    async testAlerts() {
      this.testing = true;
      this.error = null;
      this.success = null;

      try {
        // TODO: Implement actual API endpoint for testing alerts
        console.log('Testing alerts with configuration:', this.alertConfig);

        // Simulate API call for now
        await new Promise((resolve) => {
          setTimeout(resolve, 1000);
        });

        this.success = 'Test alert sent successfully!';
        setTimeout(() => {
          this.success = null;
        }, 5000);
      } catch (err) {
        console.error('Error testing alerts:', err);
        this.error = err.response?.data?.message || 'Failed to send test alert';
      } finally {
        this.testing = false;
      }
    },
    async loadConfiguration() {
      try {
        // TODO: Implement actual API endpoint for loading alert configuration
        console.log('Loading alert configuration...');

        // Simulate API call for now - keep default configuration
        await new Promise((resolve) => {
          setTimeout(resolve, 500);
        });

        // For now, just use the default configuration
        console.log('Using default alert configuration');
      } catch (err) {
        console.warn('Failed to load alert configuration:', err);
      }
    },
  },
  async created() {
    await this.loadConfiguration();
  },
};
</script>

<style scoped>
.pl-4 {
  padding-left: 16px;
}
</style>
