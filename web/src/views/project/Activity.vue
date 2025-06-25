<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <DashboardMenu
      :project-id="projectId"
      :project-type="projectType"
      :can-update-project="can(USER_PERMISSIONS.updateProject)"
    />

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4 CenterToScreen"
      :footer-props="{ itemsPerPageOptions: [20] }"
      style="max-width: calc(var(--breakpoint-lg) - var(--nav-drawer-width)); margin: auto;"
    >
      <template v-slot:item.created="{ item }">
        {{ item.created | formatDate }}
      </template>
    </v-data-table>
  </div>
</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import DashboardMenu from '@/components/DashboardMenu.vue';

export default {
  components: { DashboardMenu },

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
