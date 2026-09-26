<template>
  <v-menu
    :value="value"
    :position-x="x"
    :position-y="y"
    absolute
    :close-on-content-click="false"
    :close-on-click="true"
    min-width="260"
    max-width="300"
    transition="fade-transition"
    content-class="WorkflowQuickAddMenu"
    @input="$emit('input', $event)"
  >
    <v-card>
      <v-text-field
        ref="search"
        v-model="search"
        dense
        outlined
        hide-details
        autofocus
        prepend-inner-icon="mdi-magnify"
        :placeholder="$t('workflowQuickAddSearch')"
        class="ma-2"
        @keydown.esc="$emit('input', false)"
        @keydown.enter="pickFirst"
      />

      <v-list dense class="WorkflowQuickAddMenu__list py-0">
        <template v-if="filteredKinds.length">
          <v-subheader class="WorkflowQuickAddMenu__subheader">
            {{ $t('workflowQuickAddSteps') }}
          </v-subheader>
          <v-list-item
            v-for="k in filteredKinds"
            :key="k.kind"
            @click="pick({ kind: k.kind })"
          >
            <span
              class="WorkflowQuickAddMenu__tile"
              :class="`WorkflowQuickAddMenu__tile--${k.kind}`"
            >
              <v-icon small :color="k.color">{{ k.icon }}</v-icon>
            </span>
            <v-list-item-title>{{ k.text }}</v-list-item-title>
          </v-list-item>
          <v-divider />
        </template>

        <v-subheader class="WorkflowQuickAddMenu__subheader">
          {{ $t('workflowQuickAddTemplates') }}
        </v-subheader>
        <v-list-item
          v-for="t in filteredTemplates"
          :key="`tpl-${t.id}`"
          @click="pick({ kind: 'task', template_id: t.id })"
        >
          <span class="WorkflowQuickAddMenu__tile">
            <v-icon small>{{ appIcon(t.app) }}</v-icon>
          </span>
          <v-list-item-title>{{ t.name }}</v-list-item-title>
        </v-list-item>
        <v-list-item v-if="filteredTemplates.length === 0" disabled>
          <v-list-item-title class="text--secondary">
            {{ $t('workflowQuickAddNoTemplates') }}
          </v-list-item-title>
        </v-list-item>
      </v-list>

      <div v-if="hint" class="WorkflowQuickAddMenu__hint">{{ hint }}</div>
    </v-card>
  </v-menu>
</template>

<script>
import { APP_ICONS } from '@/lib/constants';

// "Add node" popup: opened from a node's "+" handle, the canvas "+" button or
// a right click on empty canvas. Emits `add` with { kind, template_id }.
export default {
  props: {
    value: Boolean,
    x: { type: Number, default: 0 },
    y: { type: Number, default: 0 },
    templates: { type: Array, default: () => [] },
    hint: { type: String, default: null },
  },

  data() {
    return { search: '' };
  },

  computed: {
    kinds() {
      return [
        {
          kind: 'task', icon: 'mdi-cog', color: null, text: this.$t('workflowNodeKindTask'),
        },
        {
          kind: 'approval', icon: 'mdi-account-check', color: '#ab47bc', text: this.$t('workflowNodeKindApproval'),
        },
        {
          kind: 'delay', icon: 'mdi-timer-outline', color: '#ff9800', text: this.$t('workflowNodeKindDelay'),
        },
        {
          kind: 'note', icon: 'mdi-note-text-outline', color: null, text: this.$t('workflowNodeKindNote'),
        },
      ];
    },
    query() {
      return this.search.trim().toLowerCase();
    },
    filteredKinds() {
      if (!this.query) return this.kinds;
      return this.kinds.filter((k) => k.text.toLowerCase().includes(this.query));
    },
    filteredTemplates() {
      const list = this.templates || [];
      if (!this.query) return list;
      return list.filter((t) => (t.name || '').toLowerCase().includes(this.query));
    },
  },

  watch: {
    value(open) {
      if (open) {
        this.search = '';
        this.$nextTick(() => {
          const field = this.$refs.search;
          if (field && field.focus) field.focus();
        });
      }
    },
  },

  methods: {
    appIcon(app) {
      return (APP_ICONS[app] && APP_ICONS[app].icon) || 'mdi-cog';
    },
    pick(choice) {
      this.$emit('add', choice);
      this.$emit('input', false);
    },
    pickFirst() {
      if (this.filteredTemplates.length > 0 && this.query) {
        this.pick({ kind: 'task', template_id: this.filteredTemplates[0].id });
      } else if (this.filteredKinds.length > 0) {
        this.pick({ kind: this.filteredKinds[0].kind });
      }
    },
  },
};
</script>

<style lang="scss">
.WorkflowQuickAddMenu {
  &__list {
    max-height: 320px;
    overflow-y: auto;
  }

  &__subheader {
    height: 28px;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  &__tile {
    flex: 0 0 26px;
    width: 26px;
    height: 26px;
    margin-right: 10px;
    border-radius: 7px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: rgba(127, 127, 127, 0.12);

    &--approval {
      background: rgba(171, 71, 188, 0.14);
    }

    &--delay {
      background: rgba(255, 152, 0, 0.16);
    }
  }

  &__hint {
    padding: 6px 12px 8px;
    font-size: 11px;
    line-height: 14px;
    opacity: 0.7;
    border-top: 1px solid rgba(127, 127, 127, 0.2);
  }
}
</style>
