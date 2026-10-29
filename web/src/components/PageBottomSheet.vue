<template>
  <div class="page-bottom-sheet" ref="sheet">
    <v-bottom-sheet
      :attach="this.$refs.sheet"
      :scrim="false"
      scroll-strategy="none"
      v-model="sheet"
    >
      <v-sheet class="pa-4" data-testid="tableSettings-sheet">
        <slot></slot>
      </v-sheet>
    </v-bottom-sheet>
  </div>
</template>

<style lang="scss">
.page-bottom-sheet {
  // The sheet starts where the page starts (right of the navigation drawer), not at the left
  // edge of the window: the overlay takes its static position and the width of its content,
  // which the spacer below stretches to the free width of the window. Vuetify 4: these styles sit
  // in the layer `semaphore`, above `vuetify-components` — no `!important` needed.
  .v-overlay {
    left: unset;
    right: unset;
    width: unset;
    overflow: hidden;
  }
  .v-overlay__content {
    position: relative;
  }
  .v-sheet:after {
    content: "a b c d e f g h i j k l m n o p q r s t u v w x y z \
      a b c d e f g h i j k l m n o p q r s t u v w x y z \
      a b c d e f g h i j k l m n o p q r s t u v w x y z \
      a b c d e f g h i j k l m n o p q r s t u v w x y z \
      a b c d e f g h i j k l m n o p q r s t u v w x y z \
      a b c d e f g h i j k l m n o p q r s t u v w x y z";
    display: block;
    height: 0;
    overflow: hidden;
    visibility: hidden;
    font-size: 100px;
  }
}
</style>

<script>
export default {
  props: {
    modelValue: Boolean,
  },

  emits: ['update:modelValue'],

  data() {
    return {
      sheet: false,
    };
  },

  watch: {
    async sheet(val) {
      this.$emit('update:modelValue', val);
    },

    async modelValue(val) {
      this.sheet = val;
    },
  },
};
</script>
