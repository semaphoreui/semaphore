<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <v-tabs show-arrows class="pl-4">
      <v-tab
        v-if="projectType === ''"
        key="history"
        :to="`/project/${projectId}/history`"
      >{{ $t('history') }}</v-tab>
      <v-tab key="activity" :to="`/project/${projectId}/activity`">{{ $t('activity') }}</v-tab>
      <v-tab
        v-if="can(USER_PERMISSIONS.updateProject)"
        key="settings"
        :to="`/project/${projectId}/settings`"
      >
        {{ $t('settings') }}
      </v-tab>
    </v-tabs>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.created="{ item }">
        {{ item.created | formatDate }}
      </template>
    </v-data-table>
  </div>
</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';

export default {

  mixins: [ItemListPageBase],

  methods: {
    getHeaders() {
      return [
        {
          text: this.$i18n.t('time'),
          value: 'created',
          sortable: false,
          width: '20%',
        },
        {
          text: this.$i18n.t('user'),
          value: 'username',
          sortable: false,
          width: '10%',
        },
        {
          text: this.$i18n.t('description'),
          value: 'description',
          sortable: false,
          width: '70%',
        },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/events/last`;
    },
  },
};
</script>
