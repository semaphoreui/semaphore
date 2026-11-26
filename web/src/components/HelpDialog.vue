<!--
A plain `v-dialog` with built-in help, for a dialog that is not an EditDialog. It takes the
attributes of `v-dialog` and a `help-doc`, and draws the help panel (HelpPanel.vue) inside the
content element of the dialog, next to its card, as EditDialog does:

  <HelpDialog v-model="dialog" help-doc="token-new" :max-width="480" persistent v-slot="{ docs }">
    <v-card>
      <v-card-title class="d-flex align-center">
        {{ $t('New Token') }}
        <v-spacer></v-spacer>
        <HelpButton v-if="docs.available" :open="docs.open" @click="docs.toggle()" />
      </v-card-title>
      ...
    </v-card>
  </HelpDialog>

The card stays the direct child of the content element, so the stylesheet of v-dialog applies
to it unchanged. The document is src/help/dialogs/<key>.md.
-->
<template>
  <v-dialog
    ref="dialog"
    v-bind="$attrs"
    :model-value="modelValue"
    :max-width="docsDialogMaxWidth(maxWidth)"
    :content-class="contentClasses"
    :content-props="{ onFocusin: onDocsHostFocusIn }"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <slot :docs="docsSlot"></slot>

    <HelpPanel
      v-if="docsShown"
      v-bind="docsPanelProps"
      @close="closeDocs()"
      @hover-field="docsHoverField = $event"
      @focus-field="onDocsFocusField($event)"
    />
  </v-dialog>
</template>
<script>
import HelpDocsHost from '@/components/HelpDocsHost';
import HelpPanel from '@/components/HelpPanel.vue';

export default {
  components: { HelpPanel },

  // The help panel of the dialog: its state and its tie to the fields of the card.
  mixins: [HelpDocsHost],

  inheritAttrs: false,

  props: {
    modelValue: Boolean,
    // Key of the help document of the dialog (src/help/dialogs/<key>.md).
    helpDoc: String,
    maxWidth: [Number, String],
    contentClass: String,
  },

  emits: ['update:modelValue'],

  computed: {
    // The key of the help document, for HelpDocsHost.
    docsKey() {
      return this.helpDoc;
    },

    contentClasses() {
      return [this.contentClass || '', ...this.docsDialogClasses]
        .filter((name) => name !== '').join(' ');
    },

    // What the card needs to draw its help button.
    docsSlot() {
      return {
        available: this.hasDocs,
        open: this.docsOpen,
        toggle: this.toggleDocs,
      };
    },
  },

  watch: {
    modelValue(val) {
      if (!val) {
        this.closeDocs();
      }
    },
  },

  methods: {
    // The fields are in the content element of the dialog, for HelpDocsHost. The panel is in it
    // too: its sections are anchored to fields, never named as them.
    docsRootElement() {
      const dialog = this.$refs.dialog;
      return (dialog && dialog.contentEl) || null;
    },

    // A document that has loaded after the dialog was closed must not open its panel.
    docsCanOpen() {
      return this.modelValue;
    },
  },
};
</script>
