<!--
Modal dialog which contains slot "form" and two buttons ("Cancel" and "OK").
Should be used to wrap forms which need to be displayed in modal dialog.
Can use used in tandem with ItemFormBase.js. See KeyForm.vue for example.
-->
<template>
  <v-dialog
    ref="dialog"
    v-model="dialog"
    :max-width="dialogMaxWidth"
    persistent
    :eager="booted"
    :close-on-back="false"
    :fullscreen="expandable && fullscreen"
    :transition="false"
    :content-class="dialogContentClass"
  >
    <template v-if="resizable && !(expandable && fullscreen)">
      <div
        v-for="side in ['left', 'right']"
        :key="side"
        class="item-dialog__resizer"
        :class="`item-dialog__resizer--${side}`"
        :data-testid="`editDialog-resize-${side}`"
        @pointerdown.prevent="startResize($event, side)"
        @pointermove="resize($event)"
        @pointerup="endResize()"
        @pointercancel="endResize()"
        @dblclick="resetWidth()"
      ></div>
    </template>

    <v-card
      ref="card"
      :data-testid="testId"
      @focusin="onDocsHostFocusIn($event)"
      :data-state="needSave ? 'loading' : 'ready'"
    >
      <v-card-title data-testid="editDialog-title">
        <slot name="title">
          <v-icon v-if="icon" :color="iconColor" class="mr-3">{{ icon }}</v-icon>
          {{ title }}
        </slot>

        <v-spacer></v-spacer>

        <div class="item-dialog__title-actions">
          <HelpButton
            v-if="hasDocs"
            class="mr-3"
            :open="docsOpen"
            data-testid="editDialog-docs"
            @click="toggleDocs()"
          />

          <v-btn
            icon
            variant="text"
            @click="toggleHelp()"
            class="mr-3"
            :style="{opacity: needHelp ? 1 : 0.3}"
            v-if="helpButton"
            data-testid="editDialog-help"
          >
            <v-icon>mdi-help-box</v-icon>
          </v-btn>

          <v-btn
            icon
            variant="text"
            @click="toggleFullscreen()"
            class="mr-3"
            v-if="expandable"
            data-testid="editDialog-fullscreen"
          >
            <v-icon>mdi-arrow-{{ fullscreen ? 'collapse' : 'expand' }}</v-icon>
          </v-btn>

          <v-btn icon variant="text" @click="close()" data-testid="editDialog-close">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </div>

      </v-card-title>

      <v-card-text
        data-testid="editDialog-body"
        :class="{
          'pb-0': !hideButtons,
          'pa-0': noBodyPaddings,
        }"
        :style="{
          minHeight: minContentHeight + 'px'
        }"
      >
        <slot
          name="form"
          :onSave="onSave"
          :onError="clearFlags"
          :needSave="needSave"
          :needReset="needReset"
          :needHelp="needHelp"
        ></slot>
      </v-card-text>

      <v-card-actions v-if="!hideButtons" data-testid="editDialog-actions">
        <v-spacer></v-spacer>

        <v-btn
          color="blue-darken-1"
          variant="text"
          @click="close()"
          data-testid="editDialog-cancel"
        >
          {{ cancelButtonText || $t('cancel') }}
        </v-btn>

        <v-btn
          color="blue-darken-1"
          variant="text"
          @click="needSave = true"
          v-if="saveButtonText != null"
          data-testid="editDialog-save"
        >
          {{ saveButtonText }}
        </v-btn>
      </v-card-actions>
    </v-card>

    <HelpPanel
      v-if="docsShown"
      v-bind="docsPanelProps"
      @close="closeDocs()"
      @hover-field="docsHoverField = $event"
      @focus-field="onDocsFocusField($event)"
    />
  </v-dialog>
</template>
<style lang="scss">
  .item-dialog--top {
    align-self: flex-start;
  }

  .item-dialog__title-actions {
    position: absolute;
    right: 12px;
  }

  // The edges a resizable dialog is dragged by. They lie outside the card but for 2 px, so that
  // a scrollbar at its right edge stays clickable.
  .item-dialog__resizer {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 12px;
    cursor: ew-resize;
    touch-action: none;
    z-index: 1;
  }

  .item-dialog__resizer--left {
    left: -10px;
  }

  .item-dialog__resizer--right {
    right: -10px;
  }

  .item-dialog {
    .v-card-title {
      display: flex;
      align-items: center;
      white-space: nowrap;
      overflow: hidden;
      margin-right: 12px;
      padding-bottom: 20px !important;
    }
  }

  .v-theme--dark {
    .item-dialog__title-actions {
        background: #1E1E1E;
    }
  }

  .v-theme--light {
    .item-dialog__title-actions {
      background: white;
    }
  }
</style>
<script>

import EventBus from '@/event-bus';
import HelpButton from '@/components/HelpButton.vue';
import HelpDocsHost, { DIALOG_MIN_WIDTH } from '@/components/HelpDocsHost';
import HelpPanel from '@/components/HelpPanel.vue';

// The narrowest a dialog can be dragged to: the default width of a dialog.
const MIN_WIDTH = DIALOG_MIN_WIDTH;

export default {
  components: { HelpButton, HelpPanel },

  // The help panel of the dialog: its state and its tie to the fields of the form.
  mixins: [HelpDocsHost],

  props: {
    testId: String,
    contentClass: String,
    position: String,
    title: String,
    icon: String,
    iconColor: String,
    modelValue: Boolean,
    maxWidth: Number,
    minContentHeight: Number,
    eventName: String,
    hideButtons: Boolean,
    dontCloseOnSave: Boolean,
    cancelButtonText: String,
    saveButtonText: String,
    expandable: Boolean,
    // The left and the right edge can be dragged to change the width. The dialog stays centred:
    // both edges move, whichever is dragged.
    resizable: Boolean,
    name: {
      type: String,
      default: 'Unnamed',
    },
    helpButton: Boolean,
    // Key of the help document of the dialog (src/help/dialogs/<key>.md). The dialog gets a
    // button that opens the document in a panel at the right edge of the window.
    helpDoc: String,
    noBodyPaddings: Boolean,
    noEscape: Boolean,
  },

  emits: ['update:modelValue', 'save', 'close'],

  data() {
    return {
      dialog: false,
      // The content is rendered on the first open and then kept, as `v-dialog` of Vuetify 2 did:
      // Vuetify 3 unmounts it on close, and the form would be created again on every open.
      booted: false,
      needSave: false,
      needReset: false,
      fullscreen: null,
      needHelp: false,
      // The width the user dragged a resizable dialog to; null until then (`maxWidth` applies).
      width: null,
    };
  },

  computed: {
    // The key of the help document, for HelpDocsHost.
    docsKey() {
      return this.helpDoc;
    },

    // A dialog wider than what the help panel leaves of the window is narrowed to it.
    dialogMaxWidth() {
      return this.docsDialogMaxWidth(this.width || this.maxWidth || MIN_WIDTH);
    },

    dialogContentClass() {
      return [
        'item-dialog',
        this.expandable ? 'item-dialog--expandable' : '',
        `item-dialog--${this.position}`,
        this.contentClass || '',
        ...this.docsDialogClasses,
      ].filter((name) => name !== '').join(' ');
    },
  },

  watch: {
    async dialog(val) {
      if (val) {
        this.booted = true;
      }
      this.needReset = val;
      this.$emit('update:modelValue', val);
      if (val) {
        window.addEventListener('keydown', this.handleEscape);
      } else {
        window.removeEventListener('keydown', this.handleEscape);
      }
    },

    async modelValue(val) {
      this.dialog = val;
    },

    fullscreen(val) {
      if (val) {
        localStorage.setItem(`EditDialog_${this.name}__fullscreen`, '1');
      } else {
        localStorage.removeItem(`EditDialog_${this.name}__fullscreen`);
      }
    },
  },

  created() {
    this.fullscreen = localStorage.getItem(`EditDialog_${this.name}__fullscreen`) === '1';

    // The drag in progress: where the pointer went down and what the dialog measured then.
    this.resizeOrigin = null;
    if (this.resizable) {
      const width = parseInt(localStorage.getItem(`EditDialog_${this.name}__width`), 10);
      this.width = width >= MIN_WIDTH ? width : null;
    }
  },

  methods: {
    toggleHelp() {
      this.needHelp = !this.needHelp;
    },

    // The fields of the form are in the card, for HelpDocsHost.
    docsRootElement() {
      return this.$refs.card ? this.$refs.card.$el : null;
    },

    // A document that has loaded after the dialog was closed must not open its panel.
    docsCanOpen() {
      return this.dialog;
    },

    onSave(e) {
      if (this.dontCloseOnSave) {
        this.clearFlags();
        return;
      }

      this.close(e);
    },

    toggleFullscreen() {
      this.fullscreen = !this.fullscreen;
    },

    close(e) {
      this.dialog = false;
      this.closeDocs();

      this.clearFlags();
      if (e) {
        this.$emit('save', e);
        if (this.eventName) {
          EventBus.$emit(this.eventName, e);
        }
      }
      this.$emit('close');
    },

    clearFlags() {
      this.needSave = false;
      this.needReset = false;
    },

    startResize(ev, side) {
      // The handles are children of the content element of the overlay, which the overlay
      // centres and whose margins the stylesheet of the dialog keeps free.
      const content = ev.currentTarget.parentElement;
      const style = window.getComputedStyle(content);
      const margins = (parseFloat(style.marginLeft) || 0) + (parseFloat(style.marginRight) || 0);

      this.resizeOrigin = {
        x: ev.clientX,
        width: content.getBoundingClientRect().width,
        direction: side === 'left' ? -1 : 1,
        maxWidth: content.parentElement.clientWidth - margins,
      };

      // The edge moves away from under the pointer when the dialog cannot follow it (the
      // limits below); the capture keeps the events coming to the handle.
      ev.currentTarget.setPointerCapture(ev.pointerId);
    },

    resize(ev) {
      if (!this.resizeOrigin) {
        return;
      }

      const {
        x, width, direction, maxWidth,
      } = this.resizeOrigin;

      // The opposite edge mirrors the dragged one, so the width changes by twice the distance
      // the pointer went.
      const dragged = width + 2 * direction * (ev.clientX - x);
      this.width = Math.round(Math.max(Math.min(dragged, maxWidth), MIN_WIDTH));
    },

    endResize() {
      if (!this.resizeOrigin) {
        return;
      }

      this.resizeOrigin = null;
      if (this.width != null) {
        localStorage.setItem(`EditDialog_${this.name}__width`, String(this.width));
      }
    },

    resetWidth() {
      this.width = null;
      localStorage.removeItem(`EditDialog_${this.name}__width`);
    },

    /**
     * Returns true if this dialog is the topmost active overlay (dialog or menu).
     * Uses the same check as Vuetify's VDialog (`globalTop` of the overlay stack),
     * so a nested dialog or an open dropdown menu makes the parent non-topmost.
     */
    isTopmost() {
      const dialog = this.$refs.dialog;
      if (!dialog || typeof dialog.globalTop !== 'boolean') {
        return true;
      }
      return dialog.globalTop;
    },

    handleEscape(ev) {
      if (ev.key !== 'Escape' || this.dialog === false || this.noEscape) {
        return;
      }

      // Only the topmost dialog must react to Escape. If this dialog opened
      // another dialog (or a menu), let that one handle the key.
      if (!this.isTopmost()) {
        return;
      }

      this.close();
    },
  },
};
</script>
