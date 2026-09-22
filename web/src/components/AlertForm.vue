<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null && channels != null && keys != null"
  >
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

    <template v-for="field in visibleFields">
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
        :label="fieldLabel(field)"
        class="mb-2"
        :rules="fieldRules(field)"
        :required="field.required"
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

    <template v-if="channel && channel.secret">
      <div class="mt-2 mb-1 text-body-2">
        {{ $t('alertSecretSource', { label: channel.secret.label }) }}
      </div>

      <v-radio-group
        v-model="secretMode"
        class="mt-0"
        :disabled="formSaving"
        dense
        hide-details
      >
        <v-radio value="server" :disabled="!channel.ready">
          <template v-slot:label>
            <span>
              {{ $t('alertSecretServer') }}
              <span
                v-if="!channel.ready"
                class="caption grey--text ml-1"
              >({{ channel.ready_error }})</span>
            </span>
          </template>
        </v-radio>
        <v-radio :label="$t('alertSecretOwn')" value="own" />
      </v-radio-group>

      <div v-if="secretMode === 'own'" class="alert-form__override mt-3">
        <template v-for="field in overrideFields">
          <v-text-field
            v-if="field.name === 'url'"
            :key="field.name"
            v-model="item.url"
            :label="fieldLabel(field)"
            :rules="fieldRules(field)"
            :disabled="formSaving"
            class="mb-2"
            outlined
            dense
          />
          <v-switch
            v-else-if="field.kind === 'bool'"
            :key="field.name"
            v-model="item.params[field.name]"
            :label="fieldLabel(field)"
            :disabled="formSaving"
            class="mt-0 mb-2"
            inset
            dense
            hide-details
          />
          <v-text-field
            v-else-if="field.name !== 'url'"
            :key="field.name"
            v-model="item.params[field.name]"
            :label="fieldLabel(field)"
            :type="field.kind === 'number' ? 'number' : 'text'"
            :rules="fieldRules(field)"
            :required="field.required"
            :disabled="formSaving"
            class="mb-2"
            outlined
            dense
          />
        </template>

        <v-autocomplete
          v-model="item.key_id"
          :label="channel.secret.label"
          :items="secretKeys"
          item-value="id"
          item-text="name"
          :hint="$t('alertSecretKeyHint', { type: channel.secret.key_type })"
          persistent-hint
          :clearable="channel.secret.optional"
          :rules="channel.secret.optional ? [] : [(v) => !!v || $t('isRequired')]"
          :disabled="formSaving"
          class="mb-2"
          outlined
          dense
        >
          <template v-slot:no-data>
            <div class="px-4 py-2 caption">
              {{ $t('alertSecretNoKeys', { type: channel.secret.key_type }) }}
            </div>
          </template>
        </v-autocomplete>
      </div>
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
import ItemFormBase from '@/components/ItemFormBase';
import { ALERT_EVENTS, ALERT_FIELD_LABELS, findChannel } from '@/lib/alerts';

const COLUMN_FIELDS = ['chat_id', 'thread_id', 'url', 'recipients'];

export default {
  mixins: [ItemFormBase],

  props: {
    // Channel type chosen in the "New alert" menu; the type of an alert
    // never changes after creation.
    alertType: String,
    sourceItemId: Number,
  },

  data() {
    return {
      channels: null,
      keys: null,
      secretMode: 'server',
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

    // Fields shown regardless of the secret source.
    visibleFields() {
      return this.fields.filter((f) => COLUMN_FIELDS.includes(f.name) && !f.override);
    },

    // Fields of the "own server" group, shown only when overriding.
    overrideFields() {
      return this.fields.filter((f) => f.override);
    },

    secretKeys() {
      if (!this.channel || !this.channel.secret) {
        return [];
      }
      return (this.keys || []).filter((k) => k.type === this.channel.secret.key_type);
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

    fieldLabel(field) {
      const key = ALERT_FIELD_LABELS[field.name];
      return key ? this.$t(key) : field.name;
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
        type: this.alertType,
        enabled: true,
        is_default: true,
        events: [],
        chat_id: '',
        thread_id: '',
        url: '',
        recipients: '',
        key_id: null,
        params: {},
        body: '',
      };
    },

    async beforeLoadData() {
      [this.channels, this.keys] = await Promise.all([
        this.loadProjectResources('alerts/channels'),
        this.loadProjectResources('keys'),
      ]);
    },

    async afterLoadData() {
      if (this.isNew && this.sourceItemId) {
        const source = await this.loadProjectResource('alerts', this.sourceItemId);
        this.item = {
          ...this.getNewItem(),
          ...source,
          id: undefined,
          name: this.$t('alertCopyName', { name: source.name }),
        };
      }

      if (!this.item.type) {
        const fallback = this.channels.length > 0 ? this.channels[0].type : null;
        this.item.type = this.alertType || fallback;
      }
      if (!Array.isArray(this.item.events) || this.item.events.length === 0) {
        this.$set(this.item, 'events', [...(this.channel ? this.channel.default_events : [])]);
      }
      if (!this.item.params || typeof this.item.params !== 'object') {
        this.$set(this.item, 'params', {});
      }
      if (!this.item.body) {
        this.$set(this.item, 'body', this.defaultBody);
      }

      this.secretMode = this.hasOwnSecret() || !(this.channel && this.channel.ready) ? 'own' : 'server';
    },

    hasOwnSecret() {
      if (this.item.key_id) {
        return true;
      }
      return this.overrideFields.some((f) => {
        const v = f.name === 'url' ? this.item.url : this.item.params[f.name];
        return v != null && v !== '' && v !== false;
      });
    },

    resetBodyToDefault() {
      this.item.body = this.defaultBody;
    },

    beforeSave() {
      COLUMN_FIELDS.forEach((key) => {
        if (typeof this.item[key] === 'string' && this.item[key].trim() === '') {
          this.item[key] = null;
        }
      });

      if (this.secretMode !== 'own') {
        this.item.key_id = null;
        this.overrideFields.forEach((f) => {
          if (f.name === 'url') {
            this.item.url = null;
          } else {
            delete this.item.params[f.name];
          }
        });
      }

      const params = {};
      Object.keys(this.item.params || {}).forEach((k) => {
        const v = this.item.params[k];
        if (v !== null && v !== undefined && v !== '' && v !== false) {
          params[k] = v;
        }
      });
      this.item.params = Object.keys(params).length > 0 ? params : null;

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

.alert-form__override {
  border-left: 2px solid rgba(133, 133, 133, 0.3);
  padding-left: 12px;
}

.alert-form__template >>> .v-expansion-panel {
  background: transparent !important;
}

.alert-form__template-content >>> .v-expansion-panel-content__wrap {
  padding-left: 0;
  padding-right: 0;
}
</style>
