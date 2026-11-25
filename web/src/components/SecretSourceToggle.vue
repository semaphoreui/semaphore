<template>
  <div class="d-flex justify-space-between align-center mb-2" data-testid="secretSource-toggle">
    <b style="font-size: 13px; margin-left: 5px" data-testid="secretSource-label">{{ label }}</b>
    <v-btn-toggle
        data-testid="secretSource-options"
        :model-value="modelValue"
        @update:model-value="$emit('update:modelValue', $event)"
        tile
        variant="text"
        size="small"
        mandatory
        class="SecretSourceToggle__options"
    >
      <v-btn
        v-for="source in sources"
        :key="source.value"
        :value="source.value"
        size="small"
        class="ma-0"
        style="border-radius: 4px"
        :data-testid="`secretSource-${source.value}`"
      >
        {{ source.text }}
      </v-btn>
    </v-btn-toggle>
  </div>
</template>
<script>
export default {
  props: {
    /** Where the secret is taken from: 'database', 'env' or 'file'. */
    modelValue: String,
    /** Name of the secret, shown on the left of the toggle. */
    label: String,
    disabled: Boolean,
  },

  emits: ['update:modelValue'],

  data() {
    return {
      sources: [
        { value: 'database', text: 'Store in DB' },
        { value: 'env', text: 'From ENV' },
        { value: 'file', text: 'From File' },
      ],
    };
  },
};
</script>
<style lang="scss" scoped>
// Vuetify 2 `v-btn-toggle group` (`.v-btn-toggle > .v-btn.v-btn--active { color: inherit }`): the
// group takes the colour of the text around it (the secondary text of the card), the selected
// button inherits it, the others keep the button text colour (`.theme--light.v-btn`: .87, white
// in the dark theme). Vuetify 4 gives the whole group the high-emphasis colour.
.SecretSourceToggle__options {
  color: inherit;
  // The group of Vuetify 2 is as wide as its buttons (up to the width of the row) and does not
  // scroll: on a phone the label wraps and «From File» runs past the card edge, cut by the card
  // (JOURNAL@C1D0689BA2). Vuetify 4 lets the group shrink and scroll inside the row.
  min-width: auto;
  overflow: visible;

  // Each button has a thin transparent border, the left one only on the first
  // (`.v-btn-toggle > .v-btn.v-btn`, `.v-btn-toggle--group`).
  > .v-btn {
    border: thin solid transparent;
  }

  > .v-btn:not(:first-child) {
    border-inline-start-width: 0;
  }

  // The buttons that are not selected are at .8 (`.v-btn-toggle > .v-btn.v-btn`).
  > .v-btn:not(.v-btn--active) {
    opacity: .8;

    &.v-theme--light {
      color: rgba(0, 0, 0, .87);
    }

    &.v-theme--dark {
      color: #fff;
    }
  }
}
</style>
