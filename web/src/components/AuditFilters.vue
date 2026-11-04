<template>
  <div class="AuditFilters px-4 pt-4" data-testid="audit-filters">
    <v-row density="compact">
      <v-col cols="12" sm="6" md="3">
        <v-text-field
          v-model="text.from"
          :label="$t('audit_filter_from')"
          type="datetime-local"
          class="AuditFilters__date"
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
          class="AuditFilters__date"
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
          :menu-props="{ contentClass: 'AuditFilters__menu', maxHeight: 304 }"
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
          :menu-props="{ contentClass: 'AuditFilters__menu', maxHeight: 304 }"
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
          v-model:search="kindSearch"
          :model-value="modelValue.kind"
          :items="kindItems"
          item-title="text"
          :label="$t('audit_filter_kinds')"
          :disabled="disabled"
          :menu-props="{ contentClass: 'AuditFilters__menu', maxHeight: 304 }"
          class="AuditFilters__kinds"
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
  ['audit_group_workflows', ['workflow']],
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
      kindSearch: '',
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

    // The group headers are left out while a search is typed, as Vuetify 2 filtered them out.
    kindItems() {
      const searching = Boolean(this.kindSearch);
      return GROUPS.flatMap(([header, categories]) => [
        ...(searching ? [] : [{ type: 'subheader', text: this.$t(header) }]),
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
<style lang="scss">
// From and To: the calendar button at the right end of the input and the field 40 px high, as in
// Vuetify 2. The input of Vuetify 4 is a flex box, which leaves the button right after the date,
// and its 24 px line makes a date input 42 px high.
.AuditFilters__date input[type="datetime-local"].v-field__input {
  display: block;
  line-height: 20px;
}

// The chips of the chosen kinds 4 px from the field and 8 px apart (`.v-chip--select`: margin 4px);
// the input of Vuetify 4 adds a column gap of 2 px.
.AuditFilters__kinds .v-autocomplete__selection {
  margin-inline: 4px 2px;
}

// The chips at full colour: Vuetify 4 draws the field text at .87 by the opacity of the whole
// input, chips included; Vuetify 2 coloured the text (.87) and left the chips alone.
.AuditFilters__kinds .v-field:not(.v-field--disabled) .v-field__input {
  opacity: 1;
  color: rgba(var(--v-theme-on-surface), var(--v-high-emphasis-opacity));
}

.AuditFilters__menu {
  // The group header of a dense select list of Vuetify 2 (`.v-list--dense .v-subheader`): 40 px
  // high, 12 px text, 8 px in.
  .v-list-subheader {
    min-height: 40px;
    padding-inline: 8px;
    font-size: .75rem;
  }

  // The matched part of an item while searching (`.v-list-item__mask`): grey text on a light grey.
  .v-list.v-theme--light .v-autocomplete__mask {
    color: rgba(0, 0, 0, .38);
    background: #eee;
  }

  .v-list.v-theme--dark .v-autocomplete__mask {
    color: rgba(255, 255, 255, .5);
    background: #494949;
  }
}
</style>
