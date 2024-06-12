<!--
Modal dialog which contains slot "form" and two buttons ("Cancel" and "OK").
Should be used to wrap forms which need to be displayed in modal dialog.
Can use used in tandem with ItemFormBase.js. See KeyForm.vue for example.
-->
<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <v-dialog
    v-model="dialog"
    :max-width="maxWidth || 400"
    persistent
    :transition="false"
    :content-class="'item-dialog item-dialog--' + position"
  >
    <v-card>
      <v-card-title>
        <slot name="title">
          <v-icon v-if="icon" :color="iconColor" class="mr-3">{{ icon }}</v-icon>
          {{ title }}
        </slot>

        <v-spacer></v-spacer>
        <v-btn icon @click="close()">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <v-card-text class="pb-0" :style="{minHeight: minContentHeight + 'px'}">
        <slot
          name="form"
          :onSave="onSave"
          :onError="clearFlags"
          :needSave="needSave"
          :needReset="needReset"
        ></slot>
      </v-card-text>

      <v-card-actions v-if="!hideButtons">
        <v-spacer></v-spacer>

        <v-btn
          color="blue darken-1"
          text
          @click="close()"
        >
          {{ $t('cancel') }}
        </v-btn>

        <v-btn
          color="blue darken-1"
          text
          @click="needSave = true"
          v-if="saveButtonText != null"
        >
          {{ saveButtonText }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<style lang="scss">
  .item-dialog--top {
    align-self: flex-start;
  }
  .item-dialog--center {
  }
</style>
<script>

import EventBus from '@/event-bus';

export default {
  props: {
    position: String,
    title: String,
    icon: String,
    iconColor: String,
    saveButtonText: String,
    value: Boolean,
    maxWidth: Number,
    minContentHeight: Number,
    eventName: String,
    hideButtons: Boolean,
    dontCloseOnSave: Boolean,
  },

  data() {
    return {
      dialog: false,
      needSave: false,
      needReset: false,
    };
  },

  watch: {
    async dialog(val) {
      this.$emit('input', val);
      this.needReset = val;
      if (val) {
        window.addEventListener('keydown', this.handleEscape);
      } else {
        window.removeEventListener('keydown', this.handleEscape);
      }
    },

    async value(val) {
      this.dialog = val;
    },
  },

  methods: {
    onSave(e) {
      if (this.dontCloseOnSave) {
        this.clearFlags();
        return;
      }

      this.close(e);
    },

    close(e) {
      this.dialog = false;

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

    handleEscape(ev) {
      if (ev.key === 'Escape' && this.dialog !== false) {
        this.close();
      }
    },
  },
};
</script>
