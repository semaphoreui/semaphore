<template>
  <div>
    <v-toolbar flat style="background: #00947C; color: white;" height="60">
      <v-app-bar-nav-icon style="color: white;" @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Support Chat</v-toolbar-title>
    </v-toolbar>
    <iframe
      class="crisp_chat"
      title="Chat"
      src="https://go.crisp.chat/chat/embed/?website_id=5867f49c-5d46-485e-85e7-eb3cdba1cdd4" />
  </div>

</template>
<style>
.crisp_chat {
  position: absolute;
  width: 100%;
  height: calc(100% - 60px);
  bottom: 0;
  display: block;
  border: 0;
}
</style>
<script>

import EventBus from '@/event-bus';

export default {
  created() {
    window.$crisp.push(['do', 'chat:hide']);
  },

  methods: {
    newTicket() {

    },
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },
  },
};

</script>
