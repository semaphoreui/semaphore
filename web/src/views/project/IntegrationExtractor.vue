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
        class="mt-0"
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
        class="d-inline-block"
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
