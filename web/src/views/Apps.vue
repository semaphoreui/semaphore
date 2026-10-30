<template>
  <div
    v-if="items != null && isAppsLoaded"
    data-testid="apps-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
        v-model="editDialog"
        test-id="appsEditDialog"
        save-button-text="Save"
        :title="$t('Edit App')"
        @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <AppForm
            :project-id="projectId"
            :item-id="itemId"
            @save="onSave"
            @error="onError"
            :need-save="needSave"
            :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <YesNoDialog
        test-id="appsDeleteDialog"
        :title="$t('Delete App')"
        :text="$t('Do you really want to delete this app?')"
        v-model="deleteItemDialog"
        @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-btn
          icon
          variant="text"
          class="mr-4"
          @click="returnToProjects()"
          data-testid="apps-back"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="apps-title"
      >{{ $t('Applications') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
          :disabled="!canManageSystem"
          color="primary"
          variant="elevated"
          @click="editItem('')"
          data-testid="apps-new"
      >{{ $t('New App') }}</v-btn>
    </v-toolbar>

    <v-data-table
        :headers="headers"
        :items="items"
        class="mt-4 FixedPageSizeTable AppsTable"
        :items-per-page-options="[20]"
        :items-per-page="20"
        :show-first-last-page="false"
        :page-text="items.length === 0 ? '–' : undefined"
        data-testid="apps-table"
        :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.active="{ item }">
        <v-switch
            :disabled="!canManageSystem"
            v-model="item.active"
            inset
            @update:model-value="setActive(item.id, $event)"
            data-testid="apps-rowActive"
        ></v-switch>
      </template>

      <template v-slot:item.title="{ item }">
        <v-icon
            class="mr-2"
            size="small"
            data-testid="apps-rowIcon"
        >
          {{ getAppIcon(item.id) }}
        </v-icon>

        {{ getAppTitle(item.id) }}
      </template>

      <template v-slot:item.id="{ item }">
        <code data-testid="apps-rowId">{{ item.id }}</code>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap" data-testid="apps-rowActions">
          <v-btn
              v-if="!isDefaultApp(item.id)"
              icon
              variant="text"
              class="mr-1"
              @click="askDeleteItem(item.id)"
              :disabled="item.id === userId"
              data-testid="apps-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn
              icon
              variant="text"
              class="mr-1"
              @click="editItem(item.id)"
              data-testid="apps-rowEdit"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>
  </div>
</template>
<style lang="scss">
// The applications table as Vuetify 2 drew it.
.AppsTable {
  // The inset switch of Vuetify 2 (`.v-input--switch--inset`): the control 48 × 24 px and 8 px
  // after it, the track (48 × 28) 4 px out to the left at .32, a white thumb when off. Vuetify 4
  // makes the control 40 px high (a row 16 px higher) and starts the track at the control.
  // The column of the switch keeps its width: a Vuetify 4 `v-input` may shrink to 0 px beside the
  // 100 % wide ID column (the switch then covered the names on a tablet).
  .v-switch--inset {
    min-width: max-content;

    .v-selection-control__wrapper {
      height: 24px;
      margin-inline: -4px 12px;
    }

    .v-switch__track {
      opacity: .32;
    }
  }

  .v-switch--inset.v-theme--light
    .v-selection-control:not(.v-selection-control--dirty) .v-switch__thumb {
    background-color: #fff;
  }

  // The icon in front of the name is an icon of the theme (`.theme--light.v-icon`: .54, white in
  // the dark theme), not the text colour.
  > .v-table__wrapper > table > tbody > tr > td > .v-icon,
  > .v-table__wrapper > table > tbody > tr > td > .v-data-table__td-value > .v-icon {
    &.v-theme--light {
      color: rgba(0, 0, 0, .54);
    }

    &.v-theme--dark {
      color: #fff;
    }
  }

  // Mobile rows: a flex row of Vuetify 2 (`.v-data-table__mobile-row`: space-between, the header
  // 16 px from the value, the value at the right end), not the two equal grid columns of
  // Vuetify 4.
  .v-data-table__tr--mobile > td {
    display: flex;
    justify-content: space-between;
    column-gap: 0;

    > .v-data-table__td-title {
      padding-right: 16px;
    }

    > .v-data-table__td-value {
      text-align: right;
    }
  }

  // The line over the footer: the top border of `.v-data-footer` in Vuetify 2 — as wide as the
  // footer (16 px in from both sides, App.vue) and in the border colour (.12), not a divider over
  // the whole table at .12 of the text colour.
  > .v-divider {
    margin: 0 16px;
    border-color: rgba(var(--v-border-color), var(--v-border-opacity));
    opacity: 1;
  }

  // The footer of Vuetify 2 (`.v-data-footer`): 8 px padding, the range 24 px from the left and
  // 32 px before the arrows (no rows-per-page select pushes them right), 36 px round arrow buttons
  // 14 px apart, in the icon colour of the theme (.54) when enabled; no first/last buttons
  // (`showFirstLastPage: false`).
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

    .v-pagination .v-btn.v-theme--light:not(.v-btn--disabled) {
      color: rgba(0, 0, 0, .54);
      opacity: 1;
    }
  }
}
</style>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ItemListPageBase from '@/components/ItemListPageBase';
import EditDialog from '@/components/EditDialog.vue';
import PermissionsCheck from '@/components/PermissionsCheck';
import { hasGlobalPermission } from '@/lib/role-permissions';
import AppForm from '../components/AppForm.vue';
import { DEFAULT_APPS, GLOBAL_PERMISSIONS } from '../lib/constants';
import AppsMixin from '../components/AppsMixin';
import delay from '../lib/delay';

export default {
  mixins: [ItemListPageBase, AppsMixin, PermissionsCheck],

  components: {
    AppForm,
    YesNoDialog,
    EditDialog,
  },

  computed: {
    canManageSystem() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageSystem);
    },
  },

  methods: {
    allowActions() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageSystem);
    },

    getHeaders() {
      return [{
        title: '',
        key: 'active',
      }, {
        title: this.$t('name'),
        key: 'title',
      }, {
        title: 'ID',
        key: 'id',
        width: '100%',
      }, {
        title: this.$t('actions'),
        key: 'actions',
        sortable: false,
      }];
    },

    async loadAppsDataFromBackend() {
      while (this.items == null) {
        // eslint-disable-next-line no-await-in-loop
        await delay(100);
      }

      return this.items;
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    getItemsUrl() {
      return '/api/apps';
    },

    getSingleItemUrl() {
      return `/api/apps/${this.itemId}`;
    },

    getEventName() {
      return 'i-app';
    },

    async setActive(appId, active) {
      await axios({
        method: 'post',
        url: `/api/apps/${appId}/active`,
        responseType: 'json',
        data: {
          active,
        },
      });
    },

    isDefaultApp(appId) {
      return DEFAULT_APPS.includes(appId);
    },
  },
};
</script>
