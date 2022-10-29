<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <v-dialog
    v-model="dialog"
    max-width="290"
  >
    <v-card>
      <v-card-title class="headline">{{ title }}</v-card-title>

      <v-card-text>{{ text }}</v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>

        <v-btn
          color="blue darken-1"
          text
          @click="no()"
        >
          {{ noButtonTitle || 'Cancel' }}
        </v-btn>

        <v-btn
          color="blue darken-1"
          text
          @click="yes()"
        >
          {{ yesButtonTitle || 'Yes' }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>

export default {
  props: {
    value: Boolean,
    title: String,
    text: String,
    yesButtonTitle: String,
    noButtonTitle: String,
  },

  data() {
    return {
      dialog: false,
    };
  },

  watch: {
    async dialog(val) {
      this.$emit('input', val);
    },

    async value(val) {
      this.dialog = val;
    },
  },

  methods: {
    async yes() {
      this.$emit('yes');
      this.dialog = false;
    },
    async no() {
      this.$emit('no');
      this.dialog = false;
    },
  },
};
</script>
