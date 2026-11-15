<!--
The panel of the built-in help: the document of a dialog, at the right edge of the window.
EditDialog renders it inside the content of its v-dialog, next to the card: there it is above
the scrim, inside the focus trap of the dialog, and a click on it is not a click outside.
-->
<template>
  <aside
    class="HelpPanel"
    :aria-label="doc.title"
    data-testid="editDialog-docsPanel"
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

    <div class="HelpPanel__body" data-testid="editDialog-docsBody">
      <HelpContent :blocks="doc.intro" data-testid="editDialog-docsIntro" />

      <section
        v-for="section in doc.sections"
        :key="section.field"
        class="HelpPanel__section"
        :data-testid="`editDialog-docsSection-${section.field}`"
      >
        <h3 class="HelpPanel__sectionTitle">{{ section.title }}</h3>
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
    flex: 1 1 auto;
    overflow-y: auto;
    padding: 0 24px 24px;
    font-size: 14px;
    line-height: 22px;
    overflow-wrap: anywhere;
  }

  .HelpPanel__section {
    margin: 0 -12px;
    padding: 8px 12px 0;
    border-radius: 4px;
  }

  .HelpPanel__sectionTitle {
    margin: 0 0 4px;
    font-size: 16px;
    font-weight: 500;
    line-height: 24px;
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
  },

  emits: ['close'],
};
</script>
