<template>
  <div
    v-if="integration != null"
    data-testid="integration-page"
    :data-state="aliases == null ? 'loading' : 'ready'"
  >
    <IntegrationExtractorCrumb
      :integration="integration"
      :integration-id="integrationId"
      :project-id="projectId"
    />

    <div class="px-4 pt-3 pb-2">
      <v-switch
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        class="mt-0 IntegrationExtractorSwitch"
        v-model="integration.searchable"
        :label="$t('globalAlias')"
        @update:model-value="updateIntegration()"
        data-testid="integration-searchable"
      />
    </div>

    <div v-if="integration.searchable" class="px-4">
      <v-alert
        type="info"
        variant="tonal"
        class="IntegrationExtractorInfo"
        data-testid="integration-matchersInfo"
      >
        Matchers allow the integration to be found by a project alias.
      </v-alert>
    </div>

    <div
      v-else
      class="px-4 pb-6"
      data-testid="integration-aliases"
      :data-state="aliases == null ? 'loading' : (aliases.length === 0 ? 'empty' : 'ready')"
    >
      <div
        class="mb-3 pl-1"
        v-if="(aliases || []).length === 0"
        data-testid="integration-aliasesEmpty"
      >There is no aliases.</div>

      <div
        v-else
        v-for="alias of aliases || []"
        :key="alias.id"
        data-testid="integration-alias"
      >
        <code class="mr-2" data-testid="integration-aliasUrl">{{ alias.url }}</code>

        <CopyClipboardButton
          :text="alias.url"
          success-message="The alias URL  has been copied to the clipboard."
          test-id="integration-aliasCopy"
        />

        <v-btn
          icon
          variant="text"
          @click="deleteAlias(alias.id)"
          v-if="can(USER_PERMISSIONS.manageProjectResources)"
          data-testid="integration-aliasDelete"
        >
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </div>

      <v-btn
        color="primary"
        @click="addAlias()"
        :disabled="aliases == null"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        data-testid="integration-addAlias"
      >
        {{ aliases == null ? $t('LoadAlias') : $t('AddAlias') }}
      </v-btn>
    </div>

    <v-divider />

    <IntegrationMatcher
      class="mb-6"
      v-if="integration.searchable"
      :is-admin="isAdmin"
      :userPermissions="userPermissions"
      :project-id="projectId"
    />

    <IntegrationExtractValue
      :is-admin="isAdmin"
      :userPermissions="userPermissions"
      :project-id="projectId"
    />
  </div>
</template>
<style lang="scss">
// The info alert as wide as its text, the icon in front of it (Vuetify 2 `d-inline-block` on a
// flex wrapper). The alert of Vuetify 4 is a grid: `d-inline-block` would drop the grid and put
// the icon over the text, so it is an inline grid instead.
.IntegrationExtractorInfo {
  display: inline-grid;
}

// The switch label in the secondary text colour, as every `.v-label` of Vuetify 2
// (`.theme--light.v-label`: .6, .7 dark); Vuetify 4 writes it in the text colour, .87.
.IntegrationExtractorSwitch .v-label {
  color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
  opacity: 1;
}

// The two tables of the page (IntegrationMatcher, IntegrationExtractValue) with the footer of
// Vuetify 2 (`.v-data-footer`): 8 px padding; the rows-per-page block pushed to the right, 14 px
// before the range, its select 34 px after the label, 13 px above and below (a 59 px footer), as
// wide as its value, the value in 12 px; the range 24 px after it and 32 px before the arrows;
// 36 px round arrow buttons 14 px apart, in the icon colour of the theme (.54) when enabled; the
// line over it as wide as the footer (16 px in from both sides, App.vue).
.IntegrationExtractorTable {
  > .v-divider {
    margin: 0 16px;
  }

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
import IntegrationsBase from '@/views/project/IntegrationsBase';
import axios from 'axios';

import { USER_PERMISSIONS } from '@/lib/constants';

import CopyClipboardButton from '@/components/CopyClipboardButton.vue';
import PermissionsCheck from '@/components/PermissionsCheck';
import IntegrationExtractValue from './IntegrationExtractValue.vue';
import IntegrationMatcher from './IntegrationMatcher.vue';
import IntegrationExtractorCrumb from './IntegrationExtractorCrumb.vue';

export default {
  mixins: [IntegrationsBase, PermissionsCheck],
  components: {
    CopyClipboardButton,
    IntegrationMatcher,
    IntegrationExtractValue,
    IntegrationExtractorCrumb,
  },
  props: {
    projectId: Number,
  },

  data() {
    return {
      integration: null,
      USER_PERMISSIONS,
    };
  },

  async created() {
    this.integration = (
      await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/integrations/${this.integrationId}`,
        responseType: 'json',
      })
    ).data;
  },

  methods: {
    allowActions() {
      return true;
    },
    async updateIntegration() {
      await axios({
        method: 'put',
        url: `/api/project/${this.projectId}/integrations/${this.integrationId}`,
        responseType: 'json',
        data: this.integration,
      });
    },
  },
};
</script>
