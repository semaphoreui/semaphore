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

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4 IntegrationsTable"
      :items-per-page="-1"
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
  </div>
</template>
<style lang="scss">
// Not an inline style: the utility `mt-4` of Vuetify 4 has no `!important` and would lose to it.
.IntegrationsTable {
  max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px);
  margin: auto;
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
