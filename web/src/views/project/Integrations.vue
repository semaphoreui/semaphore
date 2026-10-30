<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div
    v-if="items != null && templates != null && workflows != null"
    data-testid="integrations-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="integrationsDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('NewIntegration') : $t('EditIntegration')"
      :max-width="450"
      :transition="false"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <IntegrationForm
          :project-id="projectId"
          :item-id="itemId"
          :features="features"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      test-id="integrationsRefsDialog"
      object-title="integration"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      test-id="integrationsDeleteDialog"
      :title="$t('DeleteIntegration')"
      :text="$t('DeleteIntegrationMsg')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="integrations-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="integrations-title"
      >{{ $t('integrations') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        variant="elevated"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        @click="editItem('new')"
        data-testid="integrations-new"
        >{{ $t('NewIntegration') }}
      </v-btn>
    </v-toolbar>

    <div
      class="px-4 py-3"
      data-testid="integrations-aliases"
      :data-state="aliases == null ? 'loading' : (aliases.length === 0 ? 'empty' : 'ready')"
    >
      <div
        class="mb-3 pl-1"
        v-if="(aliases || []).length === 0"
        data-testid="integrations-aliasesEmpty"
      >There are no aliases.</div>

      <div
        v-else
        v-for="alias of aliases || []"
        :key="alias.id"
        data-testid="integrations-alias"
      >
        <code class="mr-2" data-testid="integrations-aliasUrl">{{ alias.url }}</code>

        <CopyClipboardButton
          :text="alias.url"
          :success-message="$t('aliasUrlCopied')"
          test-id="integrations-aliasCopy"
        />

        <v-btn
          v-if="can(USER_PERMISSIONS.manageProjectResources)"
          icon
          variant="text"
          @click="deleteAlias(alias.id)"
          data-testid="integrations-aliasDelete"
        >
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </div>
      <v-btn
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        color="primary"
        @click="addAlias()"
        :disabled="aliases == null"
        data-testid="integrations-addAlias"
      >
        {{ aliases == null ? $t('LoadAlias') : $t('AddAlias') }}
      </v-btn>
    </div>

    <v-divider />

    <!-- The footer of Vuetify 2: "Rows per page:" with an underlined select (the footer of
         Vuetify 4 takes the variant of its own nested VSelect defaults, else outlined), and a dash
         for the range of an empty table. -->
    <v-defaults-provider :defaults="{ VDataTableFooter: { VSelect: { variant: 'underlined' } } }">
      <v-data-table
        :headers="headers"
        :items="items"
        class="mt-4 IntegrationsTable"
        :items-per-page="-1"
        :items-per-page-text="$t('rowsPerPage')"
        :page-text="items.length === 0 ? '–' : undefined"
        data-testid="integrations-table"
        :data-state="items.length === 0 ? 'empty' : 'ready'"
      >
        <template v-slot:item.name="{ item }">
          <router-link
            :to="`/project/${projectId}/integrations/${item.id}`"
            data-testid="integrations-rowName"
            >{{ item.name }}
          </router-link>
        </template>
        <template v-slot:item.target="{ item }">
          <router-link
            v-if="item.workflow_id"
            :to="`/project/${projectId}/workflows/${item.workflow_id}`"
            class="IntegrationsTarget IntegrationsTarget--workflow"
            data-testid="integrations-rowTarget-workflow"
          >
            <v-icon size="small" class="mr-1">mdi-graph-outline</v-icon>
            {{ (workflows.find((w) => w.id === item.workflow_id) || { name: '—' }).name }}
          </router-link>
          <router-link
            v-else
            :to="`/project/${projectId}/templates/${item.template_id}`"
            class="IntegrationsTarget IntegrationsTarget--template"
            data-testid="integrations-rowTarget-template"
          >
            <v-icon size="small" class="mr-1">mdi-check-all</v-icon>
            {{ (templates.find((t) => t.id === item.template_id) || { name: '—' }).name }}
          </router-link>
        </template>
        <template v-slot:item.actions="{ item }">
          <v-btn-group
            density="compact"
            data-testid="integrations-rowActions"
          >
            <v-btn
              @click="askDeleteItem(item.id)"
              data-testid="integrations-rowDelete"
            >
              <v-icon>mdi-delete</v-icon>
            </v-btn>
            <v-btn
              @click="editItem(item.id)"
              data-testid="integrations-rowEdit"
            >
              <v-icon>mdi-pencil</v-icon>
            </v-btn>
          </v-btn-group>
        </template>
      </v-data-table>
    </v-defaults-provider>
  </div>
</template>
<style lang="scss">
// Not an inline style: the utility `mt-4` of Vuetify 4 has no `!important` and would lose to it.
.IntegrationsTable {
  max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px);
  margin: auto;

  // The icon in front of the run target: an icon of the theme (`.theme--light.v-icon`: .54, white
  // in the dark one), not the link colour.
  .IntegrationsTarget > .v-icon.v-theme--light {
    color: rgba(0, 0, 0, .54);
  }

  .IntegrationsTarget > .v-icon.v-theme--dark {
    color: #fff;
  }

  // The line over the footer: the top border of `.v-data-footer` in Vuetify 2 — as wide as the
  // footer (16 px in from both sides, App.vue) and in the border colour (.12), not a divider over
  // the whole table at .12 of the text colour.
  > .v-divider {
    margin: 0 16px;
    border-color: rgba(var(--v-border-color), var(--v-border-opacity));
    opacity: 1;
  }

  // The footer of Vuetify 2 (`.v-data-footer`): 8 px padding; the rows-per-page block pushed to
  // the right, 14 px before the range, its select 34 px after the label, 13 px above and below
  // (a 59 px footer), as wide as its value, the value in 12 px; the range 24 px after it and
  // 32 px before the arrows; 36 px round arrow buttons 14 px apart, in the icon colour of the
  // theme (.54) when enabled. On a phone the arrows wrap to a second line at the left edge.
  > .v-data-table-footer {
    justify-content: flex-start;
    padding: 0 8px;

    > .v-data-table-footer__items-per-page {
      margin-inline: auto 14px;

      > span {
        padding: 0;
      }

      > .v-select {
        width: auto;
        margin: 13px 0 13px 34px;

        .v-field {
          --v-field-padding-start: 0;
          --v-field-padding-end: 0;
          --v-field-padding-top: 0;
          --v-field-padding-bottom: 0;
          --v-field-input-padding-top: 7px;
          --v-field-input-padding-bottom: 7px;
        }

        .v-field__input {
          min-height: 32px;
          padding-inline-end: 4px;
          font-size: 12px;
        }

        .v-field__append-inner {
          padding-top: 4px;
          align-items: flex-start;
        }
      }
    }

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

import { USER_PERMISSIONS } from '@/lib/constants';

import ItemListPageBase from '@/components/ItemListPageBase';
import IntegrationForm from '@/components/IntegrationForm.vue';
import IntegrationsBase from '@/views/project/IntegrationsBase';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';

export default {
  mixins: [ItemListPageBase, IntegrationsBase],
  components: { IntegrationForm, CopyClipboardButton },
  props: {
    /** Server feature flags; `workflows` enables the workflow target of an integration. */
    features: Object,
  },
  data() {
    return {
      templates: null,
      workflows: null,
    };
  },

  async created() {
    this.templates = (
      await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates`,
        responseType: 'json',
      })
    ).data;

    // Workflows exist only in builds with the feature; elsewhere the endpoint is 404.
    if (this.features && this.features.workflows) {
      this.workflows = (
        await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/workflows`,
          responseType: 'json',
        })
      ).data || [];
    } else {
      this.workflows = [];
    }
  },

  methods: {
    allowActions() {
      return this.can(USER_PERMISSIONS.updateProject);
    },
    getHeaders() {
      return [
        {
          title: this.$t('name'),
          key: 'name',
          width: '40%',
          sortable: true,
        },
        {
          title: this.$t('integrationRunTarget'),
          key: 'target',
          width: '60%',
          sortable: false,
        },
        {
          key: 'actions',
          sortable: false,
          width: '0%',
        },
      ];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/integrations`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.itemId}`;
    },
    getEventName() {
      return 'w-integration';
    },
  },
};
</script>
