<template>
  <v-toolbar flat>
    <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
    <v-toolbar-title>Support</v-toolbar-title>
    <v-spacer></v-spacer>
    <v-btn
      color="primary"
      @click="newTicket()"
      class="mr-2"
    >New Ticket
    </v-btn>
  </v-toolbar>
</template>
<script>

import EventBus from '@/event-bus';

export default {
  methods: {
    newTicket() {

    },
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

  },
};

</script>
