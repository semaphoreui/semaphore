<template>
  <v-form ref="form" lazy-validation v-model="formValid" v-if="item != null">
    <v-alert :value="formError" color="error" class="pb-2">{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    />

    <v-select
      v-model="item.type"
      :label="$t('type')"
      :items="alertTypes"
      item-text="text"
      item-value="value"
      :disabled="formSaving"
      outlined
      dense
      @change="onTypeChange"
    />

    <v-text-field
      v-if="item.type === 'telegram'"
      v-model="item.chat_id"
      :label="$t('telegramChatId')"
      :rules="[(v) => !!v || $t('telegramChatId')]"
      required
      :disabled="formSaving"
      outlined
      dense
    />

    <v-text-field
      v-if="item.type === 'telegram'"
      v-model="item.thread_id"
      :label="$t('telegramThreadId')"
      :rules="[(v) => !v || /^-?\\d+$/.test(String(v)) || $t('telegramThreadId')]"
      :disabled="formSaving"
      outlined
      dense
    />

    <v-text-field
      v-if="['slack', 'teams', 'rocketchat', 'dingtalk', 'gotify'].includes(item.type)"
      v-model="item.url"
      :label="$t('alertWebhookUrl')"
      :rules="webhookUrlRules"
      :disabled="formSaving"
      outlined
      dense
    />

    <v-text-field
      v-if="item.type === 'gotify'"
      v-model="item.token"
      :label="$t('alertToken')"
      :hint="isNew ? undefined : $t('alertTokenHint')"
      :persistent-hint="!isNew"
      :rules="gotifyTokenRules"
      type="password"
      class="masked-secret-input"
      autocomplete="new-password"
      :disabled="formSaving"
      outlined
      dense
    />

    <v-textarea
      v-if="item.type === 'email'"
      v-model="item.recipients"
      :label="$t('alertRecipients')"
      :hint="$t('alertRecipientsHint')"
      persistent-hint
      :disabled="formSaving"
      outlined
      dense
      rows="2"
      class="mb-4"
    />

    <v-textarea
      v-model="item.body"
      :label="$t('alertMessageBody')"
      :disabled="formSaving"
      outlined
      dense
      rows="12"
      class="mt-4 alert-body-editor"
    />
    <div class="mb-2">
      <v-btn
        text
        x-small
        class="px-0"
        :disabled="formSaving || !hasDefaultBody"
        @click="resetBodyToDefault"
      >{{ $t('alertResetDefaultBody') }}</v-btn>
    </div>

    <v-checkbox
      class="mt-4"
      v-model="item.is_default"
      :label="$t('alertIsDefault')"
      :hint="$t('alertIsDefaultHint')"
      persistent-hint
      :disabled="formSaving"
    />

    <v-checkbox
      class="mt-4"
      v-model="item.enabled"
      :label="$t('enabled')"
      :disabled="formSaving"
    />
  </v-form>
</template>
<script>
import axios from 'axios';
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  props: {
    sourceItemId: Number,
  },
  data() {
    return {
      defaultBodies: {},
    };
  },
  computed: {
    alertTypes() {
      return [
        { value: 'telegram', text: 'Telegram' },
        { value: 'slack', text: 'Slack' },
        { value: 'email', text: this.$t('email') },
        { value: 'teams', text: 'Microsoft Teams' },
        { value: 'rocketchat', text: 'Rocket.Chat' },
        { value: 'dingtalk', text: 'DingTalk' },
        { value: 'gotify', text: 'Gotify' },
      ];
    },
    hasDefaultBody() {
      return Boolean(this.defaultBodyFor(this.item && this.item.type));
    },
    gotifyTokenRules() {
      if (this.isNew && this.item && this.item.url && String(this.item.url).trim()) {
        return [(v) => !!v || this.$t('alertToken')];
      }
      return [];
    },
    webhookUrlRules() {
      if (this.item && ['slack', 'teams', 'rocketchat', 'dingtalk'].includes(this.item.type)) {
        return [(v) => !!v || this.$t('alertWebhookUrl')];
      }
      return [];
    },
  },
  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/alerts`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/alerts/${this.itemId}`;
    },
    defaultBodyFor(type) {
      return (this.defaultBodies && this.defaultBodies[type]) || '';
    },
    applyDefaultBody() {
      if (!this.item) {
        return;
      }
      this.item.body = this.defaultBodyFor(this.item.type);
    },
    resetBodyToDefault() {
      this.applyDefaultBody();
    },
    onTypeChange() {
      this.applyDefaultBody();
    },
    getNewItem() {
      return {
        name: '',
        type: 'telegram',
        enabled: true,
        is_default: true,
        chat_id: '',
        thread_id: '',
        url: '',
        token: '',
        recipients: '',
        body: '',
      };
    },
    async beforeLoadData() {
      const res = await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/alerts/defaults`,
        responseType: 'json',
      });
      this.defaultBodies = res.data || {};
    },
    beforeSave() {
      const optional = ['chat_id', 'thread_id', 'url', 'token', 'recipients', 'body'];
      optional.forEach((key) => {
        if (typeof this.item[key] === 'string' && this.item[key].trim() === '') {
          this.item[key] = null;
        }
      });
      const def = this.defaultBodyFor(this.item.type).trim();
      if (typeof this.item.body === 'string' && this.item.body.trim() === def) {
        this.item.body = null;
      }
      if (this.item.type !== 'telegram') {
        this.item.chat_id = null;
        this.item.thread_id = null;
      }
      if (!['slack', 'teams', 'rocketchat', 'dingtalk', 'gotify'].includes(this.item.type)) {
        this.item.url = null;
      }
      if (this.item.type !== 'gotify') {
        this.item.token = null;
      }
      if (this.item.type !== 'email') {
        this.item.recipients = null;
      }
    },
    async afterLoadData() {
      if (!this.isNew) {
        this.item.token = '';
      } else if (this.sourceItemId) {
        const source = await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/alerts/${this.sourceItemId}`,
          responseType: 'json',
        });
        this.item = {
          ...source.data,
          id: undefined,
          token: '',
          name: `${source.data.name} copy`,
        };
      }
      if (!this.item.body) {
        this.applyDefaultBody();
      }
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
</style>
