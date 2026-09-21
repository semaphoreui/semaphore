<template>
  <v-card outlined class="server-alert-channels" v-if="project != null && channels != null">
    <v-card-title class="text-subtitle-1 pb-1">
      <v-icon small class="mr-2">mdi-server-network</v-icon>
      {{ $t('serverAlertChannels') }}
    </v-card-title>

    <v-card-text class="pb-2">
      <div class="text-body-2 mb-3">{{ $t('serverAlertChannelsHint') }}</div>

      <div class="mb-3">
        <v-chip
          v-for="ch in configuredChannels"
          :key="ch.type"
          small
          class="mr-2 mb-2"
          :outlined="!project.alert"
          :color="project.alert ? 'primary' : undefined"
        >
          <v-icon left small>{{ ch.icon }}</v-icon>
          {{ ch.title }}
        </v-chip>

        <span v-if="configuredChannels.length === 0" class="caption grey--text">
          {{ $t('serverAlertChannelsNone') }}
        </span>
      </div>

      <v-switch
        v-model="form.alert"
        :label="$t('serverAlertChannelsEnable')"
        :disabled="!canEdit || saving || configuredChannels.length === 0"
        data-testid="alerts-serverEnabled"
        inset
        dense
        hide-details
        class="mt-0"
      />

      <v-text-field
        v-if="telegramConfigured"
        v-model="form.alert_chat"
        :label="$t('telegramChatIdOptional')"
        :hint="$t('serverAlertTelegramChatHint')"
        persistent-hint
        :disabled="!canEdit || saving"
        class="mt-4"
        style="max-width: 400px"
        data-testid="alerts-serverTelegramChat"
        outlined
        dense
      />
    </v-card-text>

    <v-card-actions v-if="canEdit" class="px-4 pb-3">
      <v-spacer />
      <v-btn
        text
        :disabled="!dirty || saving"
        @click="reset()"
      >{{ $t('cancel') }}</v-btn>
      <v-btn
        color="primary"
        :disabled="!dirty"
        :loading="saving"
        @click="save()"
        data-testid="alerts-serverSave"
      >{{ $t('save') }}</v-btn>
    </v-card-actions>
  </v-card>
</template>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

/**
 * Server channels are the notification providers configured in config.json.
 * They are sent to a project only when the project opts in (project.alert),
 * which is what "Allow alerts for this project" did before named alerts.
 */
export default {
  props: {
    projectId: Number,
    channels: Array,
    canEdit: Boolean,
  },

  data() {
    return {
      project: null,
      form: {
        alert: false,
        alert_chat: '',
      },
      saving: false,
    };
  },

  computed: {
    configuredChannels() {
      return (this.channels || []).filter((c) => c.server_configured);
    },

    telegramConfigured() {
      return (this.channels || []).some((c) => c.type === 'telegram' && c.ready);
    },

    dirty() {
      if (!this.project) {
        return false;
      }
      return this.form.alert !== Boolean(this.project.alert)
        || (this.form.alert_chat || '') !== (this.project.alert_chat || '');
    },
  },

  async created() {
    await this.load();
  },

  methods: {
    async load() {
      this.project = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}`,
        responseType: 'json',
      })).data;
      this.reset();
    },

    reset() {
      this.form.alert = Boolean(this.project.alert);
      this.form.alert_chat = this.project.alert_chat || '';
    },

    async save() {
      this.saving = true;
      try {
        const updated = {
          ...this.project,
          alert: this.form.alert,
          alert_chat: this.form.alert_chat.trim() === '' ? null : this.form.alert_chat.trim(),
        };

        await axios({
          method: 'put',
          url: `/api/project/${this.projectId}`,
          responseType: 'json',
          data: updated,
        });

        this.project = updated;
        this.reset();

        EventBus.$emit('i-project', { action: 'edit', item: updated });
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.$t('serverAlertChannelsSaved'),
        });
        this.$emit('saved', updated);
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.saving = false;
      }
    },
  },
};
</script>
