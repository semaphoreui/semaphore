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
    @keydown.esc="close()"
  >
    <v-card>
      <v-card-title>
        <slot name="title">{{ title }}</slot>
      </v-card-title>

      <v-card-text class="pb-0">
        <slot
          name="form"
          :onSave="close"
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
          Cancel
        </v-btn>

        <v-btn
          color="blue darken-1"
          text
          @click="needSave = true"
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
    saveButtonText: String,
    value: Boolean,
    maxWidth: Number,
    eventName: String,
    hideButtons: Boolean,
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
    },

    async value(val) {
      this.dialog = val;
    },
  },

  methods: {
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
  },
};
</script>
