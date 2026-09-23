<template>
  <v-card outlined class="server-alert-channels" v-if="project != null && channels != null">
    <v-card-title class="text-subtitle-1 pb-1">
      <v-icon small class="mr-2">mdi-server-network</v-icon>
      {{ $t('serverAlertChannels') }}
    </v-card-title>

    <v-card-text class="pb-2">
      <div class="text-body-2 mb-3">{{ $t('serverAlertChannelsHint') }}</div>

      <div>
        <v-chip
          v-for="ch in visibleChannels"
          :key="ch.type"
          small
          class="mr-2 mb-2"
          :class="{ 'server-alert-channels__chip--clickable': isTelegram(ch) }"
          :outlined="!project.alert || isTelegramUnset(ch)"
          :color="chipColor(ch)"
          :text-color="isTelegramUnset(ch) ? 'grey' : undefined"
          :data-testid="`alerts-serverChannel-${ch.type}`"
          @click="isTelegram(ch) && toggleTelegram()"
        >
          <v-icon left small>{{ ch.icon }}</v-icon>
          {{ ch.title }}
          <span v-if="isTelegramUnset(ch)" class="ml-1 caption">
            ({{ $t('serverAlertTelegramChatMissing') }})
          </span>
          <v-icon v-if="isTelegram(ch)" right small>mdi-pencil</v-icon>
        </v-chip>

        <span v-if="visibleChannels.length === 0" class="caption grey--text">
          {{ $t('serverAlertChannelsNone') }}
        </span>
      </div>

      <HighlightedCard
          v-if="telegramEditing"
          tick-left="40px"
          style="max-width: 420px; margin-bottom: 10px !important; margin-top: 22px;"
      >
        <v-text-field
          v-model="form.alert_chat"
          :label="$t('telegramChatId')"
          persistent-hint
          :disabled="!canEdit || saving"
          data-testid="alerts-serverTelegramChat"
          autofocus
          outlined
          dense
        />
      </HighlightedCard>
    </v-card-text>

    <v-card-actions v-if="canEdit" class="px-4 pb-3">
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
<style scoped>
.server-alert-channels__chip--clickable {
  cursor: pointer;
}
</style>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import HighlightedCard from '@/components/HighlightedCard.vue';
import { getErrorMessage } from '@/lib/error';

/**
 * Server channels are the notification providers configured in config.json.
 * They are sent to a project only when the project opts in (project.alert),
 * which is what "Allow alerts for this project" did before named alerts.
 */
export default {
  components: { HighlightedCard },

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
      telegramEditing: false,
    };
  },

  computed: {
    // Channels configured server-wide, plus Telegram whenever the server has
    // a bot token: it only lacks a chat ID, which the project can set here.
    visibleChannels() {
      return (this.channels || []).filter(
        (c) => c.server_configured || (c.type === 'telegram' && c.ready),
      );
    },

    // Channels that can actually deliver for this project.
    configuredChannels() {
      return this.visibleChannels.filter((c) => c.server_configured || this.canUseProjectChat(c));
    },

    hasProjectChat() {
      return (this.form.alert_chat || '').trim() !== '';
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
    isTelegram(ch) {
      return ch.type === 'telegram';
    },

    canUseProjectChat(ch) {
      return this.isTelegram(ch) && ch.server_enabled && this.hasProjectChat;
    },

    // Telegram with no chat at all (neither server-wide nor project) can not
    // send, so the chip stays grey until a chat ID is entered.
    isTelegramUnset(ch) {
      return this.isTelegram(ch) && !ch.server_configured && !this.hasProjectChat;
    },

    chipColor(ch) {
      if (this.isTelegramUnset(ch)) {
        return undefined;
      }
      return this.project.alert ? 'primary' : undefined;
    },

    toggleTelegram() {
      this.telegramEditing = !this.telegramEditing;
    },

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
