<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <v-dialog
    v-model="dialog"
    :max-width="maxWidth || 400"
    persistent
    :transition="false"
  >
    <v-card>
      <v-card-title class="headline">{{ title }}</v-card-title>

      <v-card-text>
        <slot
          name="form"
          :onSave="close"
          :onError="clearFlags"
          :needSave="needSave"
          :needReset="needReset"
        ></slot>
      </v-card-text>

      <v-card-actions>
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
<script>

import EventBus from '@/event-bus';

export default {
  props: {
    title: String,
    saveButtonText: String,
    value: Boolean,
    maxWidth: Number,
    eventName: String,
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
    },

    clearFlags() {
      this.needSave = false;
      this.needReset = false;
    },
  },
};
</script>
