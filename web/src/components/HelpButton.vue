<!--
The button that opens the built-in help of a dialog or a page (HelpDocsHost). The class, the
`data-testid` and the click handler given to it go to the button itself.
-->
<template>
  <v-btn
    icon
    variant="text"
    :title="label"
    :aria-label="label"
    :aria-pressed="open ? 'true' : 'false'"
  >
    <v-icon>mdi-help-circle-outline</v-icon>
  </v-btn>
</template>
<script>
export default {
  props: {
    // Whether the help panel is open.
    open: Boolean,
    // The help of a page, not of a dialog: the button is labelled so.
    page: Boolean,
  },

  computed: {
    label() {
      return this.$t(this.page ? 'helpPanelOpenPage' : 'helpPanelOpen');
    },
  },
};
</script>
