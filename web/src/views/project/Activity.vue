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
      class="mt-4 CenterToScreen FixedPageSizeTable ActivityTable"
      :items-per-page-options="[20]"
      :items-per-page="20"
      :show-first-last-page="false"
      data-testid="activity-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.created="{ item }">
        {{ $filters.formatDate(item.created) }}
      </template>
    </v-data-table>
  </div>
</template>
<style lang="scss">
// Not an inline style: the utility `mt-4` of Vuetify 4 has no `!important` and would lose to it.
.ActivityTable {
  max-width: calc(var(--breakpoint-lg) - var(--nav-drawer-width));
  margin: auto;

  // The footer of Vuetify 2 (`.v-data-footer`): 8 px padding, the range 24 px from the left and
  // 32 px before the arrows (no rows-per-page select pushes them right), 36 px round arrow buttons
  // 14 px apart, in the icon colour of the theme (.54) when enabled.
  > .v-data-table-footer {
    justify-content: flex-start;
    padding: 0 8px;

    > .v-data-table-footer__info {
      min-width: 0;
      margin: 0 32px 0 24px;
      padding: 0;
    }

    .v-pagination__list > li {
      margin: 0;

      &.v-pagination__prev {
        margin-inline-end: 7px;
      }

      &.v-pagination__next {
        margin-inline-start: 7px;
      }
    }

    .v-pagination .v-btn {
      width: 36px;
      height: 36px;
    }

    // Not the .62 opacity of the `plain` variant Vuetify 4 gives the arrows (full only on hover or
    // focus): an enabled arrow of Vuetify 2 is a plain icon at .54.
    .v-pagination .v-btn.v-theme--light:not(.v-btn--disabled) {
      color: rgba(0, 0, 0, .54);
      opacity: 1;
    }
  }

  // The line over the footer: the top border of `.v-data-footer` in Vuetify 2, as wide as the
  // footer (16 px in from both sides, App.vue); Vuetify 4 draws a divider over the whole table.
  > .v-divider {
    margin: 0 16px;
  }

  // Rows as cards on a phone: the value starts 16 px after the field name, as the header of
  // `.v-data-table__mobile-row` was padded in Vuetify 2 (Vuetify 4: a 4 px grid gap); seen on the
  // long descriptions, which overflow the row.
  .v-data-table__tr--mobile > td {
    column-gap: 16px;
  }

  // "No data available" on a phone at the left edge, as the rows (see History.vue).
  @media (max-width: 599.98px) {
    .v-data-table-rows-no-data {
      text-align: start;
    }
  }
}
</style>
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
