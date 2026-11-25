<!--
The panel of the built-in help: the document of a dialog, at the right edge of the window.
EditDialog renders it inside the content of its v-dialog, next to the card: there it is above
the scrim, inside the focus trap of the dialog, and a click on it is not a click outside.

A section describes a field of the form. The panel reports the section under the pointer and the
section in focus; the host (HelpDocsHost: a dialog or a page) decides which field is the
active one and marks it.
The section of a field the form does not show now is disabled: dimmed, never the active one.
-->
<template>
  <aside
    class="HelpPanel"
    :aria-label="doc.title"
    data-testid="editDialog-docsPanel"
    @focusin="onFocusIn($event)"
  >
    <div class="HelpPanel__head">
      <h2 class="HelpPanel__title" data-testid="editDialog-docsTitle">{{ doc.title }}</h2>

      <v-btn
        icon
        variant="text"
        :title="$t('helpPanelClose')"
        :aria-label="$t('helpPanelClose')"
        data-testid="editDialog-docsClose"
        @click="$emit('close')"
      >
        <v-icon>mdi-close</v-icon>
      </v-btn>
    </div>

    <div ref="body" class="HelpPanel__body" data-testid="editDialog-docsBody">
      <HelpContent
        class="HelpPanel__intro"
        :blocks="doc.intro"
        data-testid="editDialog-docsIntro"
      />

      <!-- A section takes the focus (a click or Tab) to become the active one; the focus
           handlers are on the panel (`focusin` also covers a link inside a section). A disabled
           section takes no focus. -->
      <!-- eslint-disable-next-line vuejs-accessibility/mouse-events-have-key-events -->
      <section
        v-for="section in doc.sections"
        :key="section.field"
        ref="sections"
        class="HelpPanel__section"
        :class="{
          'HelpPanel__section--active': isActive(section.field),
          'HelpPanel__section--disabled': isMissing(section.field),
        }"
        :tabindex="isMissing(section.field) ? null : 0"
        :aria-current="isActive(section.field) ? 'true' : null"
        :aria-disabled="isMissing(section.field) ? 'true' : null"
        :data-field="section.field"
        :data-testid="`editDialog-docsSection-${section.field}`"
        @mouseenter="$emit('hover-field', isMissing(section.field) ? null : section.field)"
        @mouseleave="$emit('hover-field', null)"
      >
        <h3 class="HelpPanel__sectionTitle">{{ section.title }}</h3>
        <p
          v-if="isMissing(section.field)"
          class="HelpPanel__sectionHidden"
          :data-testid="`editDialog-docsSectionHidden-${section.field}`"
        >{{ $t('helpPanelFieldHidden') }}</p>
        <HelpContent :blocks="section.blocks" />
      </section>
    </div>
  </aside>
</template>
<style lang="scss">
  .HelpPanel {
    // Fixed to the window. In a dialog EditDialog lifts the layout containment of the overlay
    // content while the panel is open (`.item-dialog--docs`), or the panel would be laid out
    // against the dialog.
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    width: 420px;
    max-width: 100vw;
    display: flex;
    flex-direction: column;
    background: rgb(var(--v-theme-surface));
    color: rgb(var(--v-theme-on-surface));
    box-shadow: -4px 0 16px rgba(0, 0, 0, .25);
    // Above the handles a resizable dialog is dragged by.
    z-index: 2;
  }

  // A dialog with its panel open (HelpDocsHost: EditDialog, HelpDialog). The panel is fixed to
  // the window, but `.v-overlay__content` of Vuetify has
  // `contain: layout`, which makes it the containing block of its fixed descendants: the
  // containment is lifted while the panel is open. `isolation` keeps the stacking context.
  .item-dialog--docs {
    contain: none;
    isolation: isolate;
  }

  // The dialog is centred in what the panel leaves of the window: the 420 px of the panel and
  // the 24 px margin of a dialog.
  .item-dialog--docs-beside {
    margin-right: 444px;
  }

  .v-dialog--fullscreen > .item-dialog--docs-beside {
    left: 0;
    width: calc(100% - 420px);
    margin-right: 0;
  }

  .HelpPanel__head {
    display: flex;
    align-items: center;
    flex: 0 0 auto;
    padding: 12px 12px 12px 24px;
  }

  .HelpPanel__title {
    flex: 1 1 auto;
    min-width: 0;
    margin: 0;
    font-size: 20px;
    font-weight: 500;
    line-height: 32px;
    overflow-wrap: anywhere;
  }

  .HelpPanel__body {
    // The sections are placed against the body when the panel scrolls to one.
    position: relative;
    flex: 1 1 auto;
    overflow-y: auto;
    padding: 0 12px 12px;
    font-size: 14px;
    line-height: 22px;
    overflow-wrap: anywhere;
  }

  // The intro is no card: its text stands under the title of the panel, in line with the text
  // of the cards (12 px of the body and 12 px of a card).
  .HelpPanel__intro {
    padding: 0 12px;
  }

  // A section is a card: a background of its own over the surface of the panel, apart from
  // its neighbours.
  .HelpPanel__section {
    margin: 0 0 12px;
    padding: 12px;
    border-radius: 8px;
    background: rgba(var(--v-theme-on-surface), .05);
    outline: none;
    transition: background-color .15s, box-shadow .15s;

    &:last-child {
      margin-bottom: 0;
    }

    // The card ends with its padding, not with the margin of its last line.
    .HelpContent > :last-child {
      margin-bottom: 0;
    }
  }

  // The active section: the one in focus or under the pointer, or the one of the field of the
  // form in focus.
  .HelpPanel__section--active {
    background: rgba(var(--v-theme-primary), .14);
    box-shadow: inset 0 0 0 1px rgba(var(--v-theme-primary), .5);
  }

  // The section of a field the form does not show now: a paler card with its title and text
  // dimmed; the note that says why stays as it is.
  .HelpPanel__section--disabled {
    background: rgba(var(--v-theme-on-surface), .02);

    .HelpPanel__sectionTitle,
    .HelpContent {
      opacity: .45;
    }
  }

  .HelpPanel__section:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: -2px;
  }

  .HelpPanel__sectionTitle {
    margin: 0 0 4px;
    font-size: 16px;
    font-weight: 500;
    line-height: 24px;
  }

  .HelpPanel__sectionHidden {
    margin: 0 0 8px;
    font-size: 12px;
    line-height: 18px;
    opacity: .7;
  }

  // The element of the dialog or the page the active section describes (HelpDocsHost marks it):
  // a halo around the frame of a Vuetify input, or around the element itself (a button, the
  // code editor).
  [data-help-docs-field] {
    box-shadow: 0 0 0 4px rgba(var(--v-theme-primary), .24);
    transition: box-shadow .15s;
  }

  // The frame of an input is drawn by its outline, which leaves a gap for the label: the
  // outline takes the colour and the width of a focused field (a ring of `box-shadow` would
  // cross the label). The outline of an invalid field stays red.
  [data-help-docs-field].v-field--variant-outlined:not(.v-field--error) .v-field__outline {
    --v-field-border-width: 2px !important;
    color: rgb(var(--v-theme-primary));

    > * {
      color: rgb(var(--v-theme-primary));
    }
  }

  [data-help-docs-field]:not(.v-field) {
    border-radius: 4px;
    box-shadow:
      0 0 0 2px rgb(var(--v-theme-primary)),
      0 0 0 6px rgba(var(--v-theme-primary), .24);
  }

  .HelpContent {
    p,
    ul,
    pre {
      margin: 0 0 12px;
    }

    ul {
      padding-left: 20px;
    }

    a {
      color: rgb(var(--v-theme-primary));
    }

    code {
      padding: 1px 4px;
      border-radius: 3px;
      background: rgba(var(--v-theme-on-surface), .08);
      font-size: 13px;
    }

    pre {
      padding: 8px 12px;
      border-radius: 4px;
      background: rgba(var(--v-theme-on-surface), .08);
      overflow-x: auto;
      white-space: pre;

      code {
        padding: 0;
        background: none;
      }
    }
  }
</style>
<script>
import HelpContent from '@/components/HelpContent';

export default {
  components: { HelpContent },

  props: {
    // The tree of a help document (src/help).
    doc: {
      type: Object,
      required: true,
    },
    // The field whose section is the active one.
    activeField: {
      type: String,
      default: null,
    },
    // The field whose section the panel scrolls to: the field of the form in focus.
    followField: {
      type: String,
      default: null,
    },
    // Fields the form does not show now: their sections are disabled and say so.
    missingFields: {
      type: Array,
      default: () => [],
    },
  },

  emits: ['close', 'hover-field', 'focus-field'],

  watch: {
    followField(field) {
      if (field != null) {
        this.scrollToSection(field);
      }
    },
  },

  methods: {
    isMissing(field) {
      return this.missingFields.includes(field);
    },

    // The section of a field that is gone is not the active one, whatever the field in hand.
    isActive(field) {
      return field === this.activeField && !this.isMissing(field);
    },

    // The focus came to a section (its field), or to another part of the panel or into a
    // disabled section, as a link of its text (null).
    onFocusIn(ev) {
      const section = ev.target.closest('.HelpPanel__section');
      const field = section ? section.getAttribute('data-field') : null;
      this.$emit('focus-field', field != null && !this.isMissing(field) ? field : null);
    },

    // Brings the section of a field to the top of the panel, unless all of it is in view. The
    // body alone is scrolled (`scrollIntoView` would scroll every ancestor), so neither the
    // dialog nor the focus moves.
    scrollToSection(field) {
      const body = this.$refs.body;
      const section = (this.$refs.sections || [])
        .find((el) => el.getAttribute('data-field') === field);
      if (!body || !section) {
        return;
      }

      const view = body.getBoundingClientRect();
      const box = section.getBoundingClientRect();
      if (box.top >= view.top && box.bottom <= view.bottom) {
        return;
      }

      const top = Math.max(0, body.scrollTop + box.top - view.top);
      const reduced = typeof window.matchMedia === 'function'
        && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
      if (typeof body.scrollTo === 'function') {
        body.scrollTo({ top, behavior: reduced ? 'auto' : 'smooth' });
      } else {
        body.scrollTop = top;
      }
    },
  },
};
</script>
