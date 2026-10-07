<template>
  <div class="AuditFilters px-4 pt-4" data-testid="audit-filters">
    <v-row dense>
      <v-col cols="12" sm="6" md="3">
        <v-text-field
          :value="value.from"
          :label="$t('audit_filter_from')"
          type="datetime-local"
          :disabled="disabled"
          outlined
          dense
          hide-details
          @change="update({ from: $event })"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-text-field
          :value="value.to"
          :label="$t('audit_filter_to')"
          type="datetime-local"
          :disabled="disabled"
          outlined
          dense
          hide-details
          @change="update({ to: $event })"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-autocomplete
          :value="value.user"
          :items="userItems"
          :label="$t('audit_filter_user')"
          :disabled="disabled"
          clearable
          outlined
          dense
          hide-details
          @change="update({ user: $event })"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-autocomplete
          :value="value.project"
          :items="projectItems"
          :label="$t('audit_filter_project')"
          :disabled="disabled"
          clearable
          outlined
          dense
          hide-details
          @change="update({ project: $event })"
        />
      </v-col>
      <v-col cols="12" md="6">
        <v-autocomplete
          :value="value.kind"
          :items="kindItems"
          :label="$t('audit_filter_kinds')"
          :disabled="disabled"
          multiple
          small-chips
          deletable-chips
          clearable
          outlined
          dense
          hide-details
          @change="update({ kind: $event })"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-select
          :value="value.outcome"
          :items="outcomeItems"
          :label="$t('audit_filter_result')"
          :disabled="disabled"
          clearable
          outlined
          dense
          hide-details
          @change="update({ outcome: $event })"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-text-field
          :value="value.ip"
          :label="$t('audit_filter_ip')"
          :disabled="disabled"
          clearable
          outlined
          dense
          hide-details
          @change="update({ ip: $event })"
          @click:clear="update({ ip: null })"
        />
      </v-col>
    </v-row>

    <div class="d-flex align-center flex-wrap mt-2">
      <v-chip
        v-if="value.target_type"
        close
        small
        class="mr-2"
        data-testid="audit-object-chip"
        @click:close="update({ target_type: null, target_id: null })"
      >
        {{ $t('audit_filter_object') }}: {{ objectLabel }}
      </v-chip>
      <v-btn
        v-if="hasFilters"
        text
        small
        :disabled="disabled"
        data-testid="audit-clear-filters"
        @click="$emit('input', {})"
      >
        {{ $t('audit_filter_clear') }}
      </v-btn>
    </div>
  </div>
</template>
<script>
import AUDIT_KINDS from '@/lib/auditKinds';
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
    value: Object,
    disabled: Boolean,
    users: Array,
    projects: Array,
  },

  computed: {
    hasFilters() {
      return Object.keys(filterParams(this.value)).length > 0;
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
          const [category, rest] = kind.split('.');
          const [code, action] = rest.split('/');
          return {
            text: eventTitle({ category, event_code: code, action }, this.$i18n),
            value: kind,
          };
        }),
      ]);
    },

    objectLabel() {
      return `${targetTypeLabel(this.value.target_type, this.$i18n)} ${this.value.target_id}`;
    },
  },

  methods: {
    update(patch) {
      this.$emit('input', { ...this.value, ...patch });
    },
  },
};
</script>
