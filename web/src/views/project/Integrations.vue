<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="isLoaded">
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
          :group-id="newItemGroupId"
          :features="features"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <EditDialog
      v-model="groupEditDialog"
      :save-button-text="groupId === 'new' ? $t('create') : $t('save')"
      :title="groupId === 'new' ? $t('NewIntegrationGroup') : $t('EditIntegrationGroup')"
      :max-width="450"
      :transition="false"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <IntegrationGroupForm
          :project-id="projectId"
          :item-id="groupId"
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

    <YesNoDialog
      :title="$t('DeleteIntegrationGroup')"
      :text="$t('DeleteIntegrationGroupMsg')"
      v-model="deleteGroupDialog"
      @yes="deleteGroup(groupId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('integrations') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        v-if="tab === 'integrations' && can(USER_PERMISSIONS.manageProjectResources)"
        color="primary"
        @click="editGroup('new')"
        data-testid="integrations-new-group"
        >{{ $t('NewIntegrationGroup') }}
      </v-btn>
      <v-btn
        v-if="tab === 'endpoints' && can(USER_PERMISSIONS.manageProjectResources)"
        color="primary"
        @click="$refs.endpoints.editItem('new')"
        data-testid="endpoints-new"
        >{{ $t('NewEndpoint') }}
      </v-btn>
    </v-toolbar>

    <v-tabs class="pl-4" :value="tab === 'endpoints' ? 1 : 0">
      <v-tab :to="`/project/${projectId}/integrations`" exact data-testid="integrations-tab">
        {{ $t('integrations') }}
      </v-tab>
      <v-tab :to="`/project/${projectId}/integrations/endpoints`" data-testid="endpoints-tab">
        {{ $t('endpoints') }}
      </v-tab>
    </v-tabs>

    <v-divider />

    <Endpoints
      v-if="tab === 'endpoints'"
      ref="endpoints"
      embedded
      :project-id="projectId"
      :project-type="projectType"
      :user-permissions="userPermissions"
      :user-id="userId"
      :user="user"
    />

    <div
      v-else
      class="px-4 py-4 Integrations__groups"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto"
    >
      <div v-if="groups.length === 0" class="text--secondary py-6 text-center">
        {{ $t('noIntegrationGroups') }}
      </div>

      <v-card
        v-for="group of groups"
        :key="group.id"
        outlined
        class="mb-4 Integrations__group"
        :data-testid="`integration-group-${group.id}`"
      >
        <v-card-title class="pb-2 Integrations__groupHeader">
          <v-icon class="mr-2">mdi-folder-outline</v-icon>
          <span class="Integrations__groupName">{{ group.name }}</span>

          <v-spacer></v-spacer>

          <template v-if="can(USER_PERMISSIONS.manageProjectResources)">
            <v-btn
              color="primary"
              small
              text
              class="mr-1"
              @click="newIntegration(group.id)"
            >
              <v-icon small left>mdi-plus</v-icon>
              {{ $t('NewIntegration') }}
            </v-btn>
            <v-btn icon small class="mr-1" @click="editGroup(group.id)">
              <v-icon small>mdi-pencil</v-icon>
            </v-btn>
            <v-btn icon small @click="askDeleteGroup(group.id)">
              <v-icon small>mdi-delete</v-icon>
            </v-btn>
          </template>
        </v-card-title>

        <v-card-subtitle class="pb-2 Integrations__groupMeta">
          <router-link
            :to="`/project/${projectId}/integrations/endpoints`"
            class="mr-4 Integrations__groupEndpoint"
          >
            <v-icon small class="mr-1">mdi-webhook</v-icon>
            {{ $t('endpoint') }}: {{ endpointName(group.endpoint_id) }}
          </router-link>

          <span class="Integrations__groupAuth">
            <v-icon small class="mr-1">mdi-shield-key-outline</v-icon>
            {{ groupAuthLabel(group) }}
          </span>
        </v-card-subtitle>

        <v-divider />

        <v-data-table
          :headers="headers"
          :items="groupIntegrations(group.id)"
          :items-per-page="Number.MAX_VALUE"
          hide-default-footer
          dense
          :no-data-text="$t('noIntegrationsInGroup')"
          class="Integrations__table"
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
      </v-card>
    </div>
  </div>
</template>
<script>
import axios from 'axios';

import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import { USER_PERMISSIONS } from '@/lib/constants';
import { authMethodTitle } from '@/lib/integrationAuth';

import ItemListPageBase from '@/components/ItemListPageBase';
import IntegrationForm from '@/components/IntegrationForm.vue';
import IntegrationGroupForm from '@/components/IntegrationGroupForm.vue';
import Endpoints from '@/views/project/Endpoints.vue';

export default {
  mixins: [ItemListPageBase],
  components: { IntegrationForm, IntegrationGroupForm, Endpoints },
  props: {
    /** Server feature flags; `workflows` enables the workflow target of an integration. */
    features: Object,
    /** Which tab the route opened: 'integrations' (default) or 'endpoints'. */
    tab: {
      type: String,
      default: 'integrations',
    },
  },
  data() {
    return {
      templates: null,
      workflows: null,
      groups: null,
      endpoints: null,

      /** Group preselected in the form of a new integration. */
      newItemGroupId: null,

      groupId: null,
      groupEditDialog: null,
      deleteGroupDialog: null,
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

  computed: {
    isLoaded() {
      return this.items != null
        && this.groups != null
        && this.endpoints != null
        && this.templates != null
        && this.workflows != null;
    },
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

    /** Loads integrations together with the groups they belong to and the groups' endpoints. */
    async loadItems() {
      const [integrations, groups, endpoints] = await Promise.all([
        axios({ method: 'get', url: this.getItemsUrl(), responseType: 'json' }),
        axios({
          method: 'get',
          url: `/api/project/${this.projectId}/integration_groups`,
          responseType: 'json',
        }),
        axios({
          method: 'get',
          url: `/api/project/${this.projectId}/endpoints`,
          responseType: 'json',
        }),
      ]);
      this.items = integrations.data || [];
      this.groups = groups.data || [];
      this.endpoints = endpoints.data || [];
    },

    groupIntegrations(groupId) {
      return (this.items || []).filter((item) => item.group_id === groupId);
    },

    endpointName(endpointId) {
      const endpoint = (this.endpoints || []).find((e) => e.id === endpointId);
      return endpoint ? endpoint.name : '—';
    },

    /** "Authorization: inherited from endpoint" or the group's own method. */
    groupAuthLabel(group) {
      if (!group.auth_override) {
        return this.$t('authInherited');
      }
      return this.$t('authOwn', {
        method: authMethodTitle(group.auth_method, (key) => this.$t(key)),
      });
    },

    newIntegration(groupId) {
      this.newItemGroupId = groupId;
      this.editItem('new');
    },

    editGroup(groupId) {
      this.groupId = groupId;
      this.groupEditDialog = true;
    },

    askDeleteGroup(groupId) {
      this.groupId = groupId;
      this.deleteGroupDialog = true;
    },

    /** Deletes the group; the server deletes its integrations with it. */
    async deleteGroup(groupId) {
      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/integration_groups/${groupId}`,
          responseType: 'json',
        });
        EventBus.$emit('w-integration-group', { action: 'delete', item: { id: groupId } });
        await this.loadItems();
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },
  },
};
</script>
