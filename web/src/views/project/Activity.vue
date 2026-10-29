<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div
    v-if="items"
    data-testid="activity-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="activity-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title data-testid="activity-title">{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <DashboardMenu
      :project-id="projectId"
      :project-type="projectType"
      :can-update-project="can(USER_PERMISSIONS.updateProject)"
    />

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4 CenterToScreen FixedPageSizeTable"
      :items-per-page-options="[20]"
      :items-per-page="20"
      style="max-width: calc(var(--breakpoint-lg) - var(--nav-drawer-width)); margin: auto;"
      data-testid="activity-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.created="{ item }">
        {{ $filters.formatDate(item.created) }}
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
          title: this.$t('time'),
          key: 'created',
          sortable: false,
          width: '20%',
        },
        {
          title: this.$t('user'),
          key: 'username',
          sortable: false,
          width: '10%',
        },
        {
          title: this.$t('description'),
          key: 'description',
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
