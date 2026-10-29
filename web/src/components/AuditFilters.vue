<template>
  <div class="AuditFilters px-4 pt-4" data-testid="audit-filters">
    <v-row density="compact">
      <v-col cols="12" sm="6" md="3">
        <v-text-field
          v-model="text.from"
          :label="$t('audit_filter_from')"
          type="datetime-local"
          :disabled="disabled"
          variant="outlined"
          density="compact"
          hide-details
          data-testid="auditFilter-from"
          @update:focused="onFocused('from', $event)"
          @keydown.enter="commit('from')"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-text-field
          v-model="text.to"
          :label="$t('audit_filter_to')"
          type="datetime-local"
          :disabled="disabled"
          variant="outlined"
          density="compact"
          hide-details
          data-testid="auditFilter-to"
          @update:focused="onFocused('to', $event)"
          @keydown.enter="commit('to')"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-autocomplete
          :model-value="modelValue.user"
          :items="userItems"
          item-title="text"
          :label="$t('audit_filter_user')"
          :disabled="disabled"
          clearable
          variant="outlined"
          density="compact"
          hide-details
          data-testid="auditFilter-user"
          @update:model-value="update({ user: $event })"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-autocomplete
          :model-value="modelValue.project"
          :items="projectItems"
          item-title="text"
          :label="$t('audit_filter_project')"
          :disabled="disabled"
          clearable
          variant="outlined"
          density="compact"
          hide-details
          data-testid="auditFilter-project"
          @update:model-value="update({ project: $event })"
        />
      </v-col>
      <v-col cols="12" md="6">
        <v-autocomplete
          :model-value="modelValue.kind"
          :items="kindItems"
          item-title="text"
          :label="$t('audit_filter_kinds')"
          :disabled="disabled"
          multiple
          chips
          closable-chips
          clearable
          variant="outlined"
          density="compact"
          hide-details
          data-testid="auditFilter-kind"
          @update:model-value="update({ kind: $event })"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-select
          :model-value="modelValue.outcome"
          :items="outcomeItems"
          item-title="text"
          :label="$t('audit_filter_result')"
          :disabled="disabled"
          clearable
          variant="outlined"
          density="compact"
          hide-details
          data-testid="auditFilter-outcome"
          @update:model-value="update({ outcome: $event })"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-text-field
          v-model="text.ip"
          :label="$t('audit_filter_ip')"
          :disabled="disabled"
          clearable
          variant="outlined"
          density="compact"
          hide-details
          data-testid="auditFilter-ip"
          @update:focused="onFocused('ip', $event)"
          @keydown.enter="commit('ip')"
          @click:clear="clear('ip')"
        />
      </v-col>
    </v-row>

    <div class="d-flex align-center flex-wrap mt-2">
      <v-chip
        v-if="modelValue.target_type"
        closable
        size="small"
        class="mr-2"
        data-testid="audit-object-chip"
        @click:close="update({ target_type: null, target_id: null })"
      >
        {{ $t('audit_filter_object') }}: {{ objectLabel }}
      </v-chip>
      <v-btn
        v-if="hasFilters"
        variant="text"
        size="small"
        :disabled="disabled"
        data-testid="audit-clear-filters"
        @click="$emit('update:modelValue', {})"
      >
        {{ $t('audit_filter_clear') }}
      </v-btn>
    </div>
  </div>
</template>
<script>
import AUDIT_KINDS from '@/lib/auditKinds';
import i18n from '@/plugins/i18';
import { eventTitle, targetTypeLabel, filterParams } from '@/lib/auditEvent';

const GROUPS = [
  ['audit_group_auth', ['auth']],
  ['audit_group_iam', ['iam']],
  ['audit_group_resources', ['resource', 'secret']],
  ['audit_group_tasks', ['task']],
  ['audit_group_runners', ['runner']],
  ['audit_group_system', ['system', 'audit']],
];

export default {
  props: {
    modelValue: Object,
    disabled: Boolean,
    users: Array,
    projects: Array,
  },

  emits: ['update:modelValue'],

  // From, To and IP are applied on blur or Enter (the `change` of Vuetify 2), not on every key.
  data() {
    return {
      text: { from: this.modelValue.from, to: this.modelValue.to, ip: this.modelValue.ip },
      initial: {},
    };
  },

  watch: {
    'modelValue.from': function sync(value) {
      this.text.from = value;
    },
    'modelValue.to': function sync(value) {
      this.text.to = value;
    },
    'modelValue.ip': function sync(value) {
      this.text.ip = value;
    },
  },

  computed: {
    hasFilters() {
      return Object.keys(filterParams(this.modelValue)).length > 0;
    },

    userItems() {
      return (this.users || []).map((u) => ({ text: `${u.name} (${u.username})`, value: u.id }));
    },

    projectItems() {
      return (this.projects || []).map((p) => ({ text: p.name, value: p.id }));
    },

    outcomeItems() {
      return [
        { text: this.$t('audit_result_success'), value: 'success' },
        { text: this.$t('audit_result_failure'), value: 'failure' },
      ];
    },

    kindItems() {
      return GROUPS.flatMap(([header, categories]) => [
        { type: 'subheader', text: this.$t(header) },
        ...AUDIT_KINDS.filter((kind) => categories.includes(kind.split('.')[0])).map((kind) => {
          const [code, action] = kind.split('/');
          return {
            text: eventTitle({ event_code: code, action }, i18n.global),
            value: kind,
          };
        }),
      ]);
    },

    objectLabel() {
      return `${targetTypeLabel(this.modelValue.target_type, i18n.global)} ${this.modelValue.target_id}`;
    },
  },

  methods: {
    onFocused(key, focused) {
      if (focused) {
        this.initial[key] = this.text[key];
      } else {
        this.commit(key);
      }
    },

    commit(key) {
      if (this.text[key] !== this.initial[key]) {
        this.initial[key] = this.text[key];
        this.update({ [key]: this.text[key] });
      }
    },

    clear(key) {
      this.initial[key] = null;
      this.update({ [key]: null });
    },

    update(patch) {
      this.$emit('update:modelValue', { ...this.modelValue, ...patch });
    },
  },
};
</script>
