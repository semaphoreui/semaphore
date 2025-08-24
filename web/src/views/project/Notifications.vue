<template>
  <div>
    <v-toolbar flat color="white">
      <v-toolbar-title>{{ $t('notifications') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        :loading="saving"
        :disabled="!formValid"
        @click="save"
      >
        <v-icon left>mdi-content-save</v-icon>
        {{ $t('save') }}
      </v-btn>
    </v-toolbar>

    <v-card>
      <v-card-text>
        <v-alert
          v-if="!project.alert"
          type="warning"
          class="mb-4"
        >
          {{ $t('notificationsDisabledForProject') }}
          <router-link :to="`/project/${project.id}/settings`">
            {{ $t('enableInProjectSettings') }}
          </router-link>
        </v-alert>

        <NotificationsForm
          ref="notificationsForm"
          v-model="notifications"
          :form-saving="saving"
          :test-endpoint="`/api/project/${project.id}/notifications/test`"
          @success="showSuccess"
          @test-notifications="testNotifications"
        />
      </v-card-text>
    </v-card>

    <v-snackbar
      v-model="snackbar.show"
      :color="snackbar.color"
      :timeout="3000"
    >
      {{ snackbar.message }}
      <template v-slot:action="{ attrs }">
        <v-btn
          color="white"
          text
          v-bind="attrs"
          @click="snackbar.show = false"
        >
          {{ $t('close') }}
        </v-btn>
      </template>
    </v-snackbar>
  </div>
</template>

<script>
import NotificationsForm from '@/components/NotificationsForm.vue';

export default {
  name: 'ProjectNotifications',
  components: {
    NotificationsForm,
  },
  data() {
    return {
      notifications: {},
      saving: false,
      formValid: false,
      snackbar: {
        show: false,
        message: '',
        color: 'success',
      },
    };
  },
  computed: {
    project() {
      return this.$store.state.project;
    },
  },
  async created() {
    await this.loadNotifications();
  },
  methods: {
    async loadNotifications() {
      try {
        const { data } = await this.axios.get(`/api/project/${this.project.id}/notifications`);
        this.notifications = data || {};
      } catch (err) {
        console.error('Failed to load notifications:', err);
        this.notifications = {};
      }
    },
    async save() {
      if (!this.$refs.notificationsForm.validate()) {
        return;
      }

      this.saving = true;
      try {
        await this.axios.put(`/api/project/${this.project.id}/notifications`, this.notifications);
        this.showSuccess(this.$t('notificationsSavedSuccessfully'));
      } catch (err) {
        this.showError(err.response?.data?.error || this.$t('failedToSaveNotifications'));
      } finally {
        this.saving = false;
      }
    },
    async testNotifications() {
      try {
        await this.axios.post(`/api/project/${this.project.id}/notifications/test`);
        this.showSuccess(this.$t('testNotificationsSentSuccessfully'));
      } catch (err) {
        this.showError(err.response?.data?.error || this.$t('failedToSendTestNotifications'));
      }
    },
    showSuccess(message) {
      this.snackbar = {
        show: true,
        message,
        color: 'success',
      };
    },
    showError(message) {
      this.snackbar = {
        show: true,
        message,
        color: 'error',
      };
    },
  },
};
</script>