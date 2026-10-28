<template>
  <div class="AuditFilters px-4 pt-4" data-testid="audit-filters">
    <v-row dense>
      <v-col cols="12" sm="6" md="3">
        <v-text-field
          :model-value="modelValue.from"
          :label="$t('audit_filter_from')"
          type="datetime-local"
          :disabled="disabled"
          variant="outlined"
          density="compact"
          hide-details
          data-testid="auditFilter-from"
          @change="update({ from: $event })"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-text-field
          :model-value="modelValue.to"
          :label="$t('audit_filter_to')"
          type="datetime-local"
          :disabled="disabled"
          variant="outlined"
          density="compact"
          hide-details
          data-testid="auditFilter-to"
          @change="update({ to: $event })"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-autocomplete
          :model-value="modelValue.user"
          :items="userItems"
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
          :label="$t('audit_filter_kinds')"
          :disabled="disabled"
          multiple
          small-chips
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
          :model-value="modelValue.ip"
          :label="$t('audit_filter_ip')"
          :disabled="disabled"
          clearable
          variant="outlined"
          density="compact"
          hide-details
          data-testid="auditFilter-ip"
          @change="update({ ip: $event })"
          @click:clear="update({ ip: null })"
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
        { header: this.$t(header) },
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
    update(patch) {
      this.$emit('update:modelValue', { ...this.modelValue, ...patch });
    },
  },
};
</script>
