<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <v-dialog
    v-model="dialog"
    persistent
    :transition="false"
    v-if="item != null"
  >
    <v-toolbar flat color="white">
      <v-btn
        icon
        class="mr-4"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title class="breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/`"
        >Task Templates</router-link>
        <span class="breadcrumbs__separator">&gt;</span>
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/templates/${templateId}`"
        >Task Templates</router-link>
        <span class="breadcrumbs__separator">&gt;</span>
        <span class="breadcrumbs__item">{{ item.id }}</span>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        icon
        color="black"
        @click="dialog = true"
      >
        <v-icon left>mdi-close</v-icon>
      </v-btn>
    </v-toolbar>
  </v-dialog>
</template>
<script>

import axios from 'axios';

export default {
  props: {
    value: Boolean,
    projectId: Number,
    itemId: Number,
  },

  data() {
    return {
      dialog: false,
      item: null,
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

  async created() {
    await this.loadData();
  },

  methods: {
    async loadData() {
      this.item = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/tasks/${this.itemId}`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
