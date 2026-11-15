<!--
A page with built-in help: wraps the content of a page and shows the help document of the page
in a panel at the right edge of the window (HelpPanel.vue), as EditDialog does for a dialog.

  <HelpPage help-doc="project-settings" v-slot="{ docs }">
    <v-toolbar flat>
      ...
      <HelpButton v-if="docs.available" page :open="docs.open" @click="docs.toggle()" />
    </v-toolbar>
    ...
  </HelpPage>

The document is src/help/pages/<key>.md. Its sections are anchored to the `data-testid` of the
fields and the buttons of the page.
-->
<template>
  <div class="HelpPage" :class="{ 'HelpPage--docs-beside': docsBeside }">
    <!-- The described elements are looked up, and their focus is followed, in the content
         alone: the panel is not a part of it. -->
    <div ref="content" class="HelpPage__content" @focusin="onDocsHostFocusIn($event)">
      <slot :docs="docsSlot"></slot>
    </div>

    <HelpPanel
      v-if="docsShown"
      v-bind="docsPanelProps"
      @close="closeDocs()"
      @hover-field="docsHoverField = $event"
      @focus-field="onDocsFocusField($event)"
    />
  </div>
</template>
<style lang="scss">
  // The content of the page leaves the panel its 420 px of the window.
  .HelpPage--docs-beside {
    padding-right: 420px;
  }
</style>
<script>
import HelpDocsHost, { DOCS_WIDTH } from '@/components/HelpDocsHost';
import HelpPanel from '@/components/HelpPanel.vue';

// The narrowest window the content of a page and the help panel share: the side menu (260 px),
// the panel (420 px) and 420 px for the content. In a narrower window the panel lies over the
// page.
const BESIDE_WIDTH = 260 + DOCS_WIDTH + 420;

export default {
  components: { HelpPanel },

  // The help panel of the page: its state and its tie to the elements of the page.
  mixins: [HelpDocsHost],

  props: {
    // Key of the help document of the page (src/help/pages/<key>.md).
    helpDoc: String,
  },

  computed: {
    // The key of the help document, for HelpDocsHost.
    docsKey() {
      return this.helpDoc;
    },

    docsBeside() {
      return this.docsShown && this.$vuetify.display.width >= BESIDE_WIDTH;
    },

    // What the page needs to draw its help button.
    docsSlot() {
      return {
        available: this.hasDocs,
        open: this.docsOpen,
        toggle: this.toggleDocs,
      };
    },
  },

  methods: {
    // The elements of the page are in its content, for HelpDocsHost.
    docsRootElement() {
      return this.$refs.content;
    },
  },
};
</script>
