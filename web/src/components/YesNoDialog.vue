<template>
  <v-dialog
    v-model="dialog"
    :max-width="maxWidth || 290"
    :close-on-back="false"
  >
    <v-card :data-testid="testId">
      <v-card-title
        class="text-headline-medium font-weight-medium yes-no-dialog__title"
        data-testid="yesNoDialog-title"
      >{{ title }}</v-card-title>

      <v-card-text data-testid="yesNoDialog-text">
        <slot>{{ text }}</slot>
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>

        <v-btn
          v-if="!hideNoButton"
          color="blue-darken-1"
          variant="text"
          @click="no()"
          data-testid="yesNoDialog-no"
        >
          {{ noButtonTitle || $t('cancel') }}
        </v-btn>

        <v-btn
          color="blue-darken-1"
          variant="text"
          @click="yes()"
          data-testid="yesNoDialog-yes"
        >
          {{ yesButtonTitle || $t('yes') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style lang="scss">
// A long title wraps and breaks anywhere, as `.v-card__title` of Vuetify 2 did
// (`flex-wrap: wrap; word-break: break-all`); Vuetify 4 keeps it on one line with an ellipsis.
.yes-no-dialog__title.v-card-title {
  white-space: normal;
  word-break: break-all;
}
</style>

<script>

export default {
  props: {
    modelValue: Boolean,
    title: String,
    text: String,
    yesButtonTitle: String,
    noButtonTitle: String,
    hideNoButton: Boolean,
    maxWidth: Number,
    testId: {
      type: String,
      default: 'yesNoDialog',
    },
  },

  emits: ['update:modelValue', 'yes', 'no'],

  data() {
    return {
      dialog: false,
    };
  },

  watch: {
    async dialog(val) {
      this.$emit('update:modelValue', val);
    },

    async modelValue(val) {
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
