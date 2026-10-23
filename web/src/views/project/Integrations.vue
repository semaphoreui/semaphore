<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null && templates != null && workflows != null">
    <EditDialog
      v-model="editDialog"
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
      object-title="integration"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      :title="$t('DeleteIntegration')"
      :text="$t('DeleteIntegrationMsg')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('integrations') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        @click="editItem('new')"
        >{{ $t('NewIntegration') }}
      </v-btn>
    </v-toolbar>

    <v-tabs>
      <v-tab :to="`/project/${projectId}/integrations`" exact>{{ $t('integrations') }}</v-tab>
      <v-tab :to="`/project/${projectId}/integrations/history`">{{ $t('history') }}</v-tab>
    </v-tabs>

    <div class="px-4 py-3">
      <div class="mb-3 pl-1" v-if="(aliases || []).length === 0">There are no aliases.</div>

      <div v-else v-for="alias of aliases || []" :key="alias.id">
        <code class="mr-2">{{ alias.url }}</code>

        <CopyClipboardButton :text="alias.url" :success-message="$t('aliasUrlCopied')" />

        <v-btn
          v-if="can(USER_PERMISSIONS.manageProjectResources)"
          icon
          @click="deleteAlias(alias.id)"
        >
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </div>
      <v-btn
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        color="primary"
        @click="addAlias()"
        :disabled="aliases == null"
      >
        {{ aliases == null ? $t('LoadAlias') : $t('AddAlias') }}
      </v-btn>
    </div>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto"
    >
      <template v-slot:item.name="{ item }">
        <router-link :to="`/project/${projectId}/integrations/${item.id}`"
          >{{ item.name }}
        </router-link>
      </template>
      <template v-slot:item.target="{ item }">
        <router-link
          v-if="item.workflow_id"
          :to="`/project/${projectId}/workflows/${item.workflow_id}`"
          class="IntegrationsTarget IntegrationsTarget--workflow"
        >
          <v-icon small class="mr-1">mdi-graph-outline</v-icon>
          {{ (workflows.find((w) => w.id === item.workflow_id) || { name: '—' }).name }}
        </router-link>
        <router-link
          v-else
          :to="`/project/${projectId}/templates/${item.template_id}`"
          class="IntegrationsTarget IntegrationsTarget--template"
        >
          <v-icon small class="mr-1">mdi-check-all</v-icon>
          {{ (templates.find((t) => t.id === item.template_id) || { name: '—' }).name }}
        </router-link>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn @click="askDeleteItem(item.id)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn @click="editItem(item.id)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>
  </div>
</template>
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
          text: this.$i18n.t('name'),
          value: 'name',
          width: '40%',
          sortable: true,
        },
        {
          text: this.$i18n.t('integrationRunTarget'),
          value: 'target',
          width: '60%',
          sortable: false,
        },
        {
          value: 'actions',
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
