<template>
  <v-chip size="small" variant="tonal" :color="look.color" data-testid="hostStatus">
    <v-icon v-if="look.icon" start size="small">{{ look.icon }}</v-icon>
    {{ status }}
  </v-chip>
</template>

<script>
const LOOKS = {
  ok: { color: 'success', icon: 'mdi-check-circle' },
  changed: { color: 'rgb(170,85,0)', icon: 'mdi-pencil-circle' },
  failed: { color: 'error', icon: 'mdi-close-circle' },
  unreachable: { color: 'error', icon: 'mdi-lan-disconnect' },
};

export default {
  props: {
    status: String,
  },

  computed: {
    look() {
      return LOOKS[this.status] || { color: 'grey', icon: null };
    },
  },
};
</script>
