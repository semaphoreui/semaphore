<!--
The panel of the built-in help: the document of a dialog, at the right edge of the window.
EditDialog renders it inside the content of its v-dialog, next to the card: there it is above
the scrim, inside the focus trap of the dialog, and a click on it is not a click outside.

A section describes a field of the form. The panel reports the section under the pointer and the
section in focus; the dialog decides which field is the active one and marks it in the form.
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
      <HelpContent :blocks="doc.intro" data-testid="editDialog-docsIntro" />

      <!-- A section takes the focus (a click or Tab) to become the active one; the focus
           handlers are on the panel (`focusin` also covers a link inside a section). -->
      <!-- eslint-disable-next-line vuejs-accessibility/mouse-events-have-key-events -->
      <section
        v-for="section in doc.sections"
        :key="section.field"
        ref="sections"
        class="HelpPanel__section"
        :class="{ 'HelpPanel__section--active': section.field === activeField }"
        tabindex="0"
        :aria-current="section.field === activeField ? 'true' : null"
        :data-field="section.field"
        :data-testid="`editDialog-docsSection-${section.field}`"
        @mouseenter="$emit('hover-field', section.field)"
        @mouseleave="$emit('hover-field', null)"
      >
        <h3 class="HelpPanel__sectionTitle">{{ section.title }}</h3>
        <p
          v-if="missingFields.includes(section.field)"
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
    // Fixed to the window, not to the dialog: EditDialog lifts the layout containment of the
    // overlay content while the panel is open (`.item-dialog--docs`).
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
    padding: 0 24px 24px;
    font-size: 14px;
    line-height: 22px;
    overflow-wrap: anywhere;
  }

  .HelpPanel__section {
    margin: 0 -12px 4px;
    padding: 8px 12px;
    border-radius: 4px;
    outline: none;
    transition: background-color .15s;

    // The background of the active section ends with the padding, not with the last line.
    .HelpContent > :last-child {
      margin-bottom: 0;
    }
  }

  // The active section: the one in focus or under the pointer, or the one of the field of the
  // form in focus.
  .HelpPanel__section--active {
    background: rgba(var(--v-theme-primary), .12);
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
    // Fields the form does not show now: their sections say so.
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
    // The focus came to a section (its field) or to another part of the panel (null).
    onFocusIn(ev) {
      const section = ev.target.closest('.HelpPanel__section');
      this.$emit('focus-field', section ? section.getAttribute('data-field') : null);
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
