<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="integration != null">
    <IntegrationExtractorCrumb
      :integration="integration"
      :group="group"
      :project-id="projectId"
    />

    <div class="px-4 pt-3 pb-4 IntegrationExtractor__meta">
      <div v-if="endpoint != null" class="mb-3">
        <v-icon small class="mr-1">mdi-webhook</v-icon>
        {{ $t('endpoint') }}:
        <router-link
          :to="`/project/${projectId}/endpoints`"
          class="IntegrationExtractor__endpoint"
        >{{ endpoint.name }}</router-link>
      </div>

      <v-alert type="info" text dense class="d-inline-block mb-0 IntegrationExtractor__hint">
        {{ $t('matchersHint') }}
      </v-alert>
    </div>

    <v-divider />

    <IntegrationMatcher
      class="mb-6"
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
import axios from 'axios';

import { USER_PERMISSIONS } from '@/lib/constants';

import PermissionsCheck from '@/components/PermissionsCheck';
import IntegrationExtractValue from './IntegrationExtractValue.vue';
import IntegrationMatcher from './IntegrationMatcher.vue';
import IntegrationExtractorCrumb from './IntegrationExtractorCrumb.vue';

export default {
  mixins: [PermissionsCheck],
  components: {
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
      /** The integration's group and the group's endpoint, for the breadcrumb and the link. */
      group: null,
      endpoint: null,
      USER_PERMISSIONS,
    };
  },

  computed: {
    integrationId() {
      if (/^-?\d+$/.test(this.$route.params.integrationId)) {
        return parseInt(this.$route.params.integrationId, 10);
      }
      return this.$route.params.integrationId;
    },
  },

  async created() {
    this.integration = (
      await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/integrations/${this.integrationId}`,
        responseType: 'json',
      })
    ).data;

    if (this.integration.group_id) {
      this.group = (
        await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/integration_groups/${this.integration.group_id}`,
          responseType: 'json',
        })
      ).data;
    }

    if (this.group && this.group.endpoint_id) {
      this.endpoint = (
        await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/endpoints/${this.group.endpoint_id}`,
          responseType: 'json',
        })
      ).data;
    }
  },
};
</script>
