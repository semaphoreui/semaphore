<template>
  <v-menu
    :model-value="modelValue"
    :target="target"
    location="bottom start"
    :close-on-content-click="false"
    max-width="300"
    transition="fade-transition"
    content-class="WorkflowQuickAddMenu"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <!-- min-width on the card: a menu opened at a point caps its own min-width at the target's
         width (0 px) in Vuetify 3. 270, not 260: in Vuetify 2 the search input (20 characters
         wide) made the menu 270 px. As in Vuetify 2, the menu has the shadow of a menu and the
         card its own (2), 8 px below the top of the menu: there the top margin of the search
         field went through the card. -->
    <v-card min-width="270" elevation="2" class="mt-2" data-testid="workflowCanvas-quickAddMenu">
      <v-text-field
        ref="search"
        data-testid="workflowCanvas-quickAddSearch"
        v-model="search"
        density="compact"
        variant="outlined"
        hide-details
        autofocus
        prepend-inner-icon="mdi-magnify"
        :placeholder="$t('workflowQuickAddSearch')"
        class="mx-2 mb-2"
        @keydown.esc="$emit('update:modelValue', false)"
        @keydown.enter="pickFirst"
      />

      <!-- prepend-gap 0: the tile keeps its own 10 px margin, as next to the title in Vuetify 2 -->
      <v-list
        density="compact"
        prepend-gap="0"
        class="WorkflowQuickAddMenu__list py-0"
        data-testid="workflowCanvas-quickAddList"
      >
        <template v-if="filteredKinds.length">
          <v-list-subheader
            class="WorkflowQuickAddMenu__subheader"
            data-testid="workflowCanvas-quickAddMenu-stepsHeader"
          >
            {{ $t('workflowQuickAddSteps') }}
          </v-list-subheader>
          <v-list-item
            v-for="k in filteredKinds"
            :key="k.kind"
            :data-testid="`workflowCanvas-quickAddMenu-${k.kind}`"
            @click="pick({ kind: k.kind })"
          >
            <template v-slot:prepend>
              <span
                class="WorkflowQuickAddMenu__tile"
                :class="`WorkflowQuickAddMenu__tile--${k.kind}`"
              >
                <v-icon size="small" :color="k.color">{{ k.icon }}</v-icon>
              </span>
            </template>
            <v-list-item-title>{{ k.text }}</v-list-item-title>
          </v-list-item>
          <v-divider />
        </template>

        <v-list-subheader
          class="WorkflowQuickAddMenu__subheader"
          data-testid="workflowCanvas-quickAddMenu-templatesHeader"
        >
          {{ $t('workflowQuickAddTemplates') }}
        </v-list-subheader>
        <v-list-item
          v-for="t in filteredTemplates"
          :key="`tpl-${t.id}`"
          data-testid="workflowCanvas-quickAddMenu-template"
          @click="pick({ kind: 'task', template_id: t.id })"
        >
          <template v-slot:prepend>
            <span class="WorkflowQuickAddMenu__tile">
              <v-icon size="small">{{ appIcon(t.app) }}</v-icon>
            </span>
          </template>
          <v-list-item-title>{{ t.name }}</v-list-item-title>
        </v-list-item>
        <v-list-item
          v-if="filteredTemplates.length === 0"
          disabled
          data-testid="workflowCanvas-quickAddMenu-noTemplates"
        >
          <v-list-item-title class="text-medium-emphasis">
            {{ $t('workflowQuickAddNoTemplates') }}
          </v-list-item-title>
        </v-list-item>
      </v-list>

      <div
        v-if="hint"
        class="WorkflowQuickAddMenu__hint"
        data-testid="workflowCanvas-quickAddHint"
      >{{ hint }}</div>
    </v-card>
  </v-menu>
</template>

<script>
import { APP_ICONS } from '@/lib/constants';

// Width of the menu (the card's min-width) and the gap Vuetify 2 kept to the window's right edge.
const MENU_WIDTH = 270;
const WINDOW_MARGIN = 12;

// "Add node" popup: opened from a node's "+" handle, the canvas "+" button or
// a right click on empty canvas. Emits `add` with { kind, template_id }.
export default {
  props: {
    modelValue: Boolean,
    x: { type: Number, default: 0 },
    y: { type: Number, default: 0 },
    templates: { type: Array, default: () => [] },
    hint: { type: String, default: null },
  },

  emits: ['update:modelValue', 'add'],

  data() {
    return { search: '' };
  },

  computed: {
    // Vuetify 2 moved a menu that did not fit to the left, up to 12 px from the window's edge;
    // Vuetify 4 flips it to the left of the point instead.
    target() {
      const maxX = document.documentElement.clientWidth - MENU_WIDTH - WINDOW_MARGIN;
      return [Math.max(0, Math.min(this.x, maxX)), this.y];
    },
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
    modelValue(open) {
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
      this.$emit('update:modelValue', false);
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
    // Thin translucent scrollbar: the default one is a solid white bar on the
    // dark menu. The thumb only shows up while the list is hovered.
    scrollbar-width: thin;
    scrollbar-color: transparent transparent;

    &::-webkit-scrollbar {
      width: 6px;
    }

    &::-webkit-scrollbar-track {
      background: transparent;
    }

    &::-webkit-scrollbar-thumb {
      background: transparent;
      border-radius: 3px;
    }

    &:hover {
      scrollbar-color: rgba(127, 127, 127, 0.4) transparent;

      &::-webkit-scrollbar-thumb {
        background: rgba(127, 127, 127, 0.4);
      }
    }
  }

  // The subheader of a dense list of Vuetify 2 (`.v-list--dense .v-subheader`: 12 px, 0 8 px),
  // which overrode the size here.
  &__subheader {
    padding-inline: 8px;
    font-size: .75rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  // Elevation 8 of Vuetify 2 (`.v-menu__content`).
  &.v-overlay__content {
    box-shadow: 0 5px 5px -3px rgba(0, 0, 0, 0.2), 0 8px 10px 1px rgba(0, 0, 0, 0.14),
      0 3px 14px 2px rgba(0, 0, 0, 0.12);
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

    // An icon without a colour: `.theme--light.v-icon` / `.theme--dark.v-icon` of Vuetify 2.
    > .v-icon:not([style*="color"]) {
      color: rgba(0, 0, 0, 0.54);
    }

    .v-theme--dark & > .v-icon:not([style*="color"]) {
      color: #fff;
    }
  }

  // The "no templates" line: Vuetify 2 did not fade a disabled list item (its text here is the
  // secondary colour of `text--secondary`); Vuetify 4 dims it to .6.
  .v-list-item--disabled {
    opacity: 1;
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
