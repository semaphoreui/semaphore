<template>
  <v-form ref="form" v-model="formValid" lazy-validation>
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}</v-alert>

    <!-- Telegram Configuration -->
    <v-expansion-panels class="mb-4" multiple>
      <v-expansion-panel>
        <v-expansion-panel-header>
          <div class="d-flex align-center">
            <v-icon class="mr-3">mdi-send</v-icon>
            <span>Telegram</span>
            <v-spacer></v-spacer>
            <v-switch
              v-model="notifications.telegram.enabled"
              class="mt-0"
              hide-details
              @click.stop
            ></v-switch>
          </div>
        </v-expansion-panel-header>
        <v-expansion-panel-content>
          <v-text-field
            v-model="notifications.telegram.token"
            label="Bot Token"
            :rules="telegramTokenRules"
            :disabled="!notifications.telegram.enabled || formSaving"
            outlined
            dense
            type="password"
            hint="Get your bot token from @BotFather"
          ></v-text-field>
          
          <v-text-field
            v-model="notifications.telegram.channel"
            label="Channel/Chat ID"
            :rules="telegramChannelRules"
            :disabled="!notifications.telegram.enabled || formSaving"
            outlined
            dense
            hint="Channel username (e.g., @mychannel) or chat ID (e.g., -1001234567890)"
          ></v-text-field>
          
          <v-text-field
            v-model="notifications.telegram.chat_id"
            label="Legacy Chat ID (for backward compatibility)"
            :disabled="!notifications.telegram.enabled || formSaving"
            outlined
            dense
            hint="Legacy field for backward compatibility"
          ></v-text-field>
        </v-expansion-panel-content>
      </v-expansion-panel>

      <!-- Slack Configuration -->
      <v-expansion-panel>
        <v-expansion-panel-header>
          <div class="d-flex align-center">
            <v-icon class="mr-3">mdi-slack</v-icon>
            <span>Slack</span>
            <v-spacer></v-spacer>
            <v-switch
              v-model="notifications.slack.enabled"
              class="mt-0"
              hide-details
              @click.stop
            ></v-switch>
          </div>
        </v-expansion-panel-header>
        <v-expansion-panel-content>
          <v-text-field
            v-model="notifications.slack.webhook_url"
            label="Webhook URL"
            :rules="slackWebhookRules"
            :disabled="!notifications.slack.enabled || formSaving"
            outlined
            dense
            type="password"
            hint="Slack incoming webhook URL"
          ></v-text-field>
          
          <v-text-field
            v-model="notifications.slack.channel"
            label="Channel (optional)"
            :disabled="!notifications.slack.enabled || formSaving"
            outlined
            dense
            hint="Override default channel (e.g., #notifications)"
          ></v-text-field>
        </v-expansion-panel-content>
      </v-expansion-panel>

      <!-- Gotify Configuration -->
      <v-expansion-panel>
        <v-expansion-panel-header>
          <div class="d-flex align-center">
            <v-icon class="mr-3">mdi-bell</v-icon>
            <span>Gotify</span>
            <v-spacer></v-spacer>
            <v-switch
              v-model="notifications.gotify.enabled"
              class="mt-0"
              hide-details
              @click.stop
            ></v-switch>
          </div>
        </v-expansion-panel-header>
        <v-expansion-panel-content>
          <v-text-field
            v-model="notifications.gotify.url"
            label="Server URL"
            :rules="gotifyUrlRules"
            :disabled="!notifications.gotify.enabled || formSaving"
            outlined
            dense
            hint="Gotify server URL (e.g., https://gotify.example.com)"
          ></v-text-field>
          
          <v-text-field
            v-model="notifications.gotify.token"
            label="Application Token"
            :rules="gotifyTokenRules"
            :disabled="!notifications.gotify.enabled || formSaving"
            outlined
            dense
            type="password"
            hint="Application token from Gotify"
          ></v-text-field>
          
          <v-slider
            v-model="notifications.gotify.priority"
            label="Priority"
            :disabled="!notifications.gotify.enabled || formSaving"
            min="1"
            max="10"
            step="1"
            thumb-label
            hint="Message priority (1=lowest, 10=highest)"
          ></v-slider>
        </v-expansion-panel-content>
      </v-expansion-panel>

      <!-- Dingtalk Configuration -->
      <v-expansion-panel>
        <v-expansion-panel-header>
          <div class="d-flex align-center">
            <v-icon class="mr-3">mdi-message</v-icon>
            <span>DingTalk</span>
            <v-spacer></v-spacer>
            <v-switch
              v-model="notifications.dingtalk.enabled"
              class="mt-0"
              hide-details
              @click.stop
            ></v-switch>
          </div>
        </v-expansion-panel-header>
        <v-expansion-panel-content>
          <v-text-field
            v-model="notifications.dingtalk.webhook_url"
            label="Webhook URL"
            :rules="dingtalkWebhookRules"
            :disabled="!notifications.dingtalk.enabled || formSaving"
            outlined
            dense
            type="password"
            hint="DingTalk custom bot webhook URL"
          ></v-text-field>
          
          <v-text-field
            v-model="notifications.dingtalk.secret"
            label="Secret (optional)"
            :disabled="!notifications.dingtalk.enabled || formSaving"
            outlined
            dense
            type="password"
            hint="DingTalk bot secret for signature verification"
          ></v-text-field>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>

    <!-- Test Notifications Button -->
    <v-btn
      v-if="showTestButton"
      color="primary"
      :loading="testLoading"
      :disabled="!hasEnabledNotifications || formSaving"
      @click="testNotifications"
      class="mt-4"
    >
      <v-icon left>mdi-send</v-icon>
      Test Notifications
    </v-btn>
  </v-form>
</template>

<script>
export default {
  name: 'NotificationsForm',
  props: {
    value: {
      type: Object,
      default: () => ({}),
    },
    formSaving: {
      type: Boolean,
      default: false,
    },
    showTestButton: {
      type: Boolean,
      default: true,
    },
    testEndpoint: {
      type: String,
      default: null,
    },
  },
  data() {
    return {
      formValid: false,
      formError: null,
      testLoading: false,
      notifications: {
        telegram: {
          enabled: false,
          token: '',
          channel: '',
          chat_id: '',
        },
        slack: {
          enabled: false,
          webhook_url: '',
          channel: '',
        },
        gotify: {
          enabled: false,
          url: '',
          token: '',
          priority: 5,
        },
        dingtalk: {
          enabled: false,
          webhook_url: '',
          secret: '',
        },
      },
    };
  },
  computed: {
    hasEnabledNotifications() {
      return (
        this.notifications.telegram.enabled ||
        this.notifications.slack.enabled ||
        this.notifications.gotify.enabled ||
        this.notifications.dingtalk.enabled
      );
    },
    telegramTokenRules() {
      return [
        (v) => !this.notifications.telegram.enabled || !!v || 'Bot token is required when Telegram is enabled',
      ];
    },
    telegramChannelRules() {
      return [
        (v) => !this.notifications.telegram.enabled || !!v || !!this.notifications.telegram.chat_id || 'Channel or Chat ID is required when Telegram is enabled',
      ];
    },
    slackWebhookRules() {
      return [
        (v) => !this.notifications.slack.enabled || !!v || 'Webhook URL is required when Slack is enabled',
      ];
    },
    gotifyUrlRules() {
      return [
        (v) => !this.notifications.gotify.enabled || !!v || 'Server URL is required when Gotify is enabled',
      ];
    },
    gotifyTokenRules() {
      return [
        (v) => !this.notifications.gotify.enabled || !!v || 'Application token is required when Gotify is enabled',
      ];
    },
    dingtalkWebhookRules() {
      return [
        (v) => !this.notifications.dingtalk.enabled || !!v || 'Webhook URL is required when DingTalk is enabled',
      ];
    },
  },
  watch: {
    value: {
      handler(newValue) {
        if (newValue) {
          this.notifications = {
            telegram: {
              enabled: newValue.telegram?.enabled || false,
              token: newValue.telegram?.token || '',
              channel: newValue.telegram?.channel || '',
              chat_id: newValue.telegram?.chat_id || '',
            },
            slack: {
              enabled: newValue.slack?.enabled || false,
              webhook_url: newValue.slack?.webhook_url || '',
              channel: newValue.slack?.channel || '',
            },
            gotify: {
              enabled: newValue.gotify?.enabled || false,
              url: newValue.gotify?.url || '',
              token: newValue.gotify?.token || '',
              priority: newValue.gotify?.priority || 5,
            },
            dingtalk: {
              enabled: newValue.dingtalk?.enabled || false,
              webhook_url: newValue.dingtalk?.webhook_url || '',
              secret: newValue.dingtalk?.secret || '',
            },
          };
        }
      },
      immediate: true,
      deep: true,
    },
    notifications: {
      handler(newValue) {
        // Clean up the notifications object before emitting
        const cleanNotifications = {};
        
        if (newValue.telegram.enabled) {
          cleanNotifications.telegram = { ...newValue.telegram };
        }
        if (newValue.slack.enabled) {
          cleanNotifications.slack = { ...newValue.slack };
        }
        if (newValue.gotify.enabled) {
          cleanNotifications.gotify = { ...newValue.gotify };
        }
        if (newValue.dingtalk.enabled) {
          cleanNotifications.dingtalk = { ...newValue.dingtalk };
        }
        
        this.$emit('input', cleanNotifications);
      },
      deep: true,
    },
  },
  methods: {
    async testNotifications() {
      if (!this.testEndpoint) {
        this.$emit('test-notifications');
        return;
      }

      this.testLoading = true;
      this.formError = null;

      try {
        await this.axios.post(this.testEndpoint);
        this.$emit('success', 'Test notifications sent successfully');
      } catch (err) {
        this.formError = err.response?.data?.error || 'Failed to send test notifications';
      } finally {
        this.testLoading = false;
      }
    },
    validate() {
      return this.$refs.form.validate();
    },
  },
};
</script>