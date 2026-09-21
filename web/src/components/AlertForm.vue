<template>
  <v-form ref="form" lazy-validation v-model="formValid" v-if="item != null && channels != null">
    <v-alert :value="formError" color="error" class="pb-2">{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      data-testid="alert-name"
      outlined
      dense
    />

    <v-select
      v-model="item.type"
      :label="$t('type')"
      :items="channels"
      item-text="title"
      item-value="type"
      :disabled="formSaving"
      data-testid="alert-type"
      outlined
      dense
      @change="onTypeChange"
    >
      <template v-slot:selection="{ item: ch }">
        <v-icon small class="mr-2">{{ ch.icon }}</v-icon>{{ ch.title }}
      </template>
      <template v-slot:item="{ item: ch }">
        <v-icon small class="mr-2">{{ ch.icon }}</v-icon>{{ ch.title }}
      </template>
    </v-select>

    <v-alert
      v-if="channel && !channel.ready"
      type="warning"
      text
      dense
      class="mb-4"
    >{{ channel.ready_error }}</v-alert>

    <template v-for="field in fields">
      <v-text-field
        v-if="field.name === 'chat_id'"
        :key="field.name"
        v-model="item.chat_id"
        :label="$t('telegramChatId')"
        :rules="fieldRules(field)"
        :required="field.required"
        :disabled="formSaving"
        outlined
        dense
      />

      <v-text-field
        v-else-if="field.name === 'thread_id'"
        :key="field.name"
        v-model="item.thread_id"
        :label="$t('telegramThreadId')"
        :hint="$t('telegramThreadIdHint')"
        persistent-hint
        class="mb-2"
        :rules="[
          (v) => !v || (/^\d+$/.test(String(v).trim()) && Number(v) >= 1) || $t('mustBe1OrGreater'),
        ]"
        :disabled="formSaving"
        outlined
        dense
      />

      <v-text-field
        v-else-if="field.name === 'url'"
        :key="field.name"
        v-model="item.url"
        :label="$t('alertWebhookUrl')"
        :hint="field.required ? undefined : $t('alertUrlOptionalHint')"
        :persistent-hint="!field.required"
        class="mb-2"
        :rules="fieldRules(field)"
        :required="field.required"
        :disabled="formSaving"
        outlined
        dense
      />

      <v-text-field
        v-else-if="field.name === 'token'"
        :key="field.name"
        v-model="item.token"
        :label="$t('alertToken')"
        :hint="isNew ? undefined : $t('alertTokenHint')"
        :persistent-hint="!isNew"
        class="mb-2"
        type="password"
        autocomplete="new-password"
        :rules="fieldRules(field)"
        :disabled="formSaving"
        outlined
        dense
      />

      <v-textarea
        v-else-if="field.name === 'recipients'"
        :key="field.name"
        v-model="item.recipients"
        :label="$t('alertRecipients')"
        :hint="$t('alertRecipientsHint')"
        persistent-hint
        class="mb-2"
        rows="2"
        auto-grow
        :disabled="formSaving"
        outlined
        dense
      />
    </template>

    <div class="mt-2 mb-1 text-body-2">{{ $t('alertSendOn') }}</div>
    <v-chip-group
      v-model="item.events"
      multiple
      column
      active-class="primary--text"
    >
      <v-chip
        v-for="event in ALERT_EVENTS"
        :key="event.value"
        :value="event.value"
        filter
        outlined
        small
        :disabled="formSaving"
      >
        <v-icon left small :color="event.color">{{ event.icon }}</v-icon>
        {{ $t(event.label) }}
      </v-chip>
    </v-chip-group>
    <div class="caption grey--text mb-3">{{ $t('alertEventsHint') }}</div>

    <v-switch
      v-model="item.is_default"
      :label="$t('alertIsDefault')"
      :hint="$t('alertIsDefaultHint')"
      persistent-hint
      class="mt-2"
      :disabled="formSaving"
      inset
      dense
    />

    <v-switch
      v-model="item.enabled"
      :label="$t('enabled')"
      class="mt-4"
      :disabled="formSaving"
      inset
      dense
      hide-details
    />

    <v-expansion-panels flat class="mt-4 alert-form__template">
      <v-expansion-panel>
        <v-expansion-panel-header class="px-0">
          <span>
            <v-icon small class="mr-2">mdi-code-braces</v-icon>
            {{ $t('alertMessageBody') }}
            <v-chip x-small class="ml-2" v-if="isCustomBody">{{ $t('alertCustomized') }}</v-chip>
          </span>
        </v-expansion-panel-header>
        <v-expansion-panel-content class="alert-form__template-content">
          <div class="caption grey--text mb-2">{{ $t('alertMessageBodyHint') }}</div>
          <v-textarea
            v-model="item.body"
            :disabled="formSaving"
            outlined
            dense
            rows="10"
            class="alert-body-editor"
            hide-details
          />
          <v-btn
            text
            small
            class="px-0 mt-1"
            :disabled="formSaving || !isCustomBody"
            @click="resetBodyToDefault"
          >{{ $t('alertResetDefaultBody') }}</v-btn>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>
  </v-form>
</template>
<script>
import axios from 'axios';
import ItemFormBase from '@/components/ItemFormBase';
import { ALERT_EVENTS, findChannel } from '@/lib/alerts';

export default {
  mixins: [ItemFormBase],

  props: {
    sourceItemId: Number,
  },

  data() {
    return {
      channels: null,
      ALERT_EVENTS,
    };
  },

  computed: {
    channel() {
      return findChannel(this.channels, this.item && this.item.type);
    },

    fields() {
      return this.channel ? this.channel.fields : [];
    },

    defaultBody() {
      return this.channel ? this.channel.default_body || '' : '';
    },

    isCustomBody() {
      const body = (this.item && this.item.body) || '';
      return body.trim() !== '' && body.trim() !== this.defaultBody.trim();
    },
  },

  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/alerts`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/alerts/${this.itemId}`;
    },

    fieldRules(field) {
      if (!field.required) {
        return [];
      }
      return [(v) => !!(v && String(v).trim()) || this.$t('isRequired')];
    },

    getNewItem() {
      return {
        name: '',
        type: null,
        enabled: true,
        is_default: true,
        events: [],
        chat_id: '',
        thread_id: '',
        url: '',
        token: '',
        recipients: '',
        body: '',
      };
    },

    async beforeLoadData() {
      this.channels = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/alerts/channels`,
        responseType: 'json',
      })).data;
    },

    async afterLoadData() {
      if (this.isNew && this.sourceItemId) {
        const source = (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/alerts/${this.sourceItemId}`,
          responseType: 'json',
        })).data;
        this.item = {
          ...this.getNewItem(),
          ...source,
          id: undefined,
          token: '',
          name: this.$t('alertCopyName', { name: source.name }),
        };
      }

      if (!this.item.type) {
        this.item.type = this.channels.length > 0 ? this.channels[0].type : null;
      }
      if (!Array.isArray(this.item.events) || this.item.events.length === 0) {
        this.$set(this.item, 'events', [...(this.channel ? this.channel.default_events : [])]);
      }
      if (!this.item.body) {
        this.$set(this.item, 'body', this.defaultBody);
      }
      if (!this.isNew) {
        this.item.token = '';
      }
    },

    onTypeChange() {
      this.$set(this.item, 'events', [...(this.channel ? this.channel.default_events : [])]);
      this.$set(this.item, 'body', this.defaultBody);
    },

    resetBodyToDefault() {
      this.item.body = this.defaultBody;
    },

    beforeSave() {
      ['chat_id', 'thread_id', 'url', 'token', 'recipients'].forEach((key) => {
        if (typeof this.item[key] === 'string' && this.item[key].trim() === '') {
          this.item[key] = null;
        }
      });

      // The built-in template is never stored: it follows server upgrades.
      if (!this.isCustomBody) {
        this.item.body = null;
      }

      // An event list equal to the channel default is stored empty so the
      // alert follows the channel default too.
      const events = Array.isArray(this.item.events) ? this.item.events : [];
      const defaults = this.channel ? this.channel.default_events : [];
      const sameAsDefault = events.length === defaults.length
        && defaults.every((e) => events.includes(e));
      this.item.events = sameAsDefault ? [] : events;
    },
  },
};
</script>
<style scoped>
.alert-body-editor >>> textarea {
  font-family: monospace;
  font-size: 13px;
  line-height: 1.4;
}

.alert-form__template >>> .v-expansion-panel {
  background: transparent !important;
}

.alert-form__template-content >>> .v-expansion-panel-content__wrap {
  padding-left: 0;
  padding-right: 0;
}
</style>
