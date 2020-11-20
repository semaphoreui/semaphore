<template>
  <div class="page-bottom-sheet" ref="sheet">
    <v-bottom-sheet
      :attach="this.$refs.sheet"
      hide-overlay
      v-model="sheet"
    >
      <v-sheet class="pa-4">
        <slot></slot>
      </v-sheet>
    </v-bottom-sheet>
  </div>
</template>

<style lang="scss">
.page-bottom-sheet {
  .v-dialog__content {
    left: unset !important;
    width: unset !important;
    overflow: hidden;
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
    value: Boolean,
  },

  data() {
    return {
      sheet: false,
    };
  },

  watch: {
    async sheet(val) {
      this.$emit('input', val);
    },

    async value(val) {
      this.sheet = val;
    },
  },
};
</script>
