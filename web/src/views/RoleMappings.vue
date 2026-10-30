<template>
  <div
    v-if="items != null"
    data-testid="roleMappings-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="roleMappingDialog"
      save-button-text="Save"
      :title="$t('newRoleMapping')"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <RoleMappingForm
          item-id="new"
          :providers="providers"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      :title="$t('deleteRoleMapping')"
      :text="$t('askDeleteRoleMapping')"
      v-model="deleteItemDialog"
      test-id="roleMappingsDeleteDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-btn
        icon
        variant="text"
        class="mr-4"
        @click="returnToProjects()"
        data-testid="roleMappings-back"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="roleMappings-title"
      >{{ $t('roleMappings') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="!features.custom_roles_management"
        color="primary"
        variant="elevated"
        @click="editItem('new')"
        data-testid="roleMappings-new"
        >{{ $t('newRoleMapping') }}</v-btn
      >
    </v-toolbar>

    <v-divider />

    <v-alert
      v-if="!features.custom_roles_management"
      variant="tonal"
      color="amber-darken-3"
      class="PageAlert"
      data-testid="roleMappings-licenseAlert"
    >
      <span
        class="mr-1"
        v-html="$t('roles_only_enterprise')"
        data-testid="roleMappings-licenseAlertText"
      ></span>

      <v-btn
        variant="flat"
        color="amber-darken-3"
        href="https://semaphoreui.com/enterprise?utm_source=app&utm_content=feature_roles"
        target="_blank"
        data-testid="roleMappings-upgrade"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>
    </v-alert>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4 FixedPageSizeTable RoleMappingsTable"
      :items-per-page-options="[20]"
      :items-per-page="20"
      :show-first-last-page="false"
      :page-text="items.length === 0 ? '–' : undefined"
      data-testid="roleMappings-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.provider="{ item }">
        {{ (providers.find((p) => p.id === item.provider) || { name: item.provider }).name }}
        <v-chip
          v-if="!item.provider_configured"
          size="small"
          label
          variant="flat"
          color="amber-darken-3"
          class="ml-1"
          data-testid="roleMappings-rowProviderNotConfigured"
        >{{ $t('providerNotConfigured') }}</v-chip>
      </template>
      <template v-slot:item.role="{ item }">
        <div data-testid="roleMappings-rowRoleName">{{ item.role_name }}</div>
        <div
          class="text-medium-emphasis text-body-small"
          data-testid="roleMappings-rowRoleSlug"
        >{{ item.role }}</div>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn
          icon
          variant="text"
          @click="askDeleteItem(item.id)"
          data-testid="roleMappings-rowDelete"
        >
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </template>
    </v-data-table>
  </div>
</template>
<style lang="scss">
// The mappings table as Vuetify 2 drew it.
.RoleMappingsTable {
  // The small "Provider not configured" chip of Vuetify 2 is 24 px high (`.v-size--small`);
  // Vuetify 4 makes it 26 px.
  .v-chip.v-chip--size-small {
    --v-chip-height: 24px;
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
import ItemListPageBase from '@/components/ItemListPageBase';
import RoleMappingForm from '@/components/RoleMappingForm.vue';

export default {
  mixins: [ItemListPageBase],

  props: {
    features: Object,
  },

  components: {
    RoleMappingForm,
  },

  data() {
    return {
      providers: [],
    };
  },

  methods: {
    async beforeLoadItems() {
      this.providers = (await axios({
        method: 'get',
        url: '/api/auth/login',
        responseType: 'json',
      })).data.oidc_providers || [];
    },

    getHeaders() {
      return [
        {
          title: this.$t('roleMappingProvider'),
          key: 'provider',
        },
        {
          title: this.$t('roleMappingGroup'),
          key: 'group',
        },
        {
          title: this.$t('project'),
          key: 'project_name',
        },
        {
          title: this.$t('role'),
          key: 'role',
        },
        {
          title: this.$t('actions'),
          key: 'actions',
          sortable: false,
        },
      ];
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    getItemsUrl() {
      return '/api/role_mappings';
    },

    getSingleItemUrl() {
      return `/api/role_mappings/${this.itemId}`;
    },

    getEventName() {
      return 'i-role-mapping';
    },
  },
};
</script>
