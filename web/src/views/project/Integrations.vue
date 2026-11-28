<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div
    v-if="isLoaded"
    data-testid="integrations-page"
    :data-state="groups.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="integrationsDialog"
      help-doc="integration"
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
      test-id="integrationGroupsDialog"
      help-doc="integration-group"
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

    <YesNoDialog
      test-id="integrationGroupsDeleteDialog"
      :title="$t('DeleteIntegrationGroup')"
      :text="$t('DeleteIntegrationGroupMsg')"
      v-model="deleteGroupDialog"
      @yes="deleteGroup(groupId)"
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
        v-if="tab === 'integrations' && can(USER_PERMISSIONS.manageProjectResources)"
        color="primary"
        variant="elevated"
        @click="editGroup('new')"
        data-testid="integrations-newGroup"
        >{{ $t('NewIntegrationGroup') }}
      </v-btn>
      <v-btn
        v-if="tab === 'endpoints' && can(USER_PERMISSIONS.manageProjectResources)"
        color="primary"
        variant="elevated"
        @click="$refs.endpoints.editItem('new')"
        data-testid="endpoints-new"
        >{{ $t('NewEndpoint') }}
      </v-btn>
    </v-toolbar>

    <v-tabs>
      <v-tab
        :to="`/project/${projectId}/integrations`"
        exact
        data-testid="integrations-tab"
      >{{ $t('integrations') }}</v-tab>
      <v-tab
        :to="`/project/${projectId}/integrations/endpoints`"
        data-testid="integrations-endpointsTab"
      >{{ $t('endpoints') }}</v-tab>
      <v-tab
        :to="`/project/${projectId}/integrations/history`"
        data-testid="integrations-historyTab"
      >{{ $t('history') }}</v-tab>
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
      data-testid="integrations-groups"
    >
      <div
        v-if="groups.length === 0"
        class="text-medium-emphasis py-6 text-center"
        data-testid="integrations-groupsEmpty"
      >
        {{ $t('noIntegrationGroups') }}
      </div>

      <v-card
        v-for="group of groups"
        :key="group.id"
        flat
        border
        class="mb-4 Integrations__group"
        :data-testid="`integration-group-${group.id}`"
      >
        <v-card-title class="d-flex align-center pb-2 Integrations__groupHeader">
          <v-icon class="mr-2">mdi-folder-outline</v-icon>
          <span
            class="Integrations__groupName"
            data-testid="integrations-groupName"
          >{{ group.name }}</span>

          <v-spacer></v-spacer>

          <template v-if="can(USER_PERMISSIONS.manageProjectResources)">
            <v-btn
              color="primary"
              size="small"
              variant="text"
              class="mr-1"
              @click="newIntegration(group.id)"
              data-testid="integrations-new"
            >
              <v-icon size="small" start>mdi-plus</v-icon>
              {{ $t('NewIntegration') }}
            </v-btn>
            <v-btn
              icon
              size="small"
              variant="text"
              class="mr-1"
              @click="editGroup(group.id)"
              data-testid="integrations-groupEdit"
            >
              <v-icon size="small">mdi-pencil</v-icon>
            </v-btn>
            <v-btn
              icon
              size="small"
              variant="text"
              @click="askDeleteGroup(group.id)"
              data-testid="integrations-groupDelete"
            >
              <v-icon size="small">mdi-delete</v-icon>
            </v-btn>
          </template>
        </v-card-title>

        <v-card-subtitle class="pb-2 Integrations__groupMeta">
          <router-link
            :to="`/project/${projectId}/integrations/endpoints`"
            class="mr-4 Integrations__groupEndpoint"
            data-testid="integrations-groupEndpoint"
          >
            <v-icon size="small" class="mr-1">mdi-webhook</v-icon>
            {{ $t('endpoint') }}: {{ endpointName(group.endpoint_id) }}
          </router-link>

          <span class="Integrations__groupAuth" data-testid="integrations-groupAuth">
            <v-icon size="small" class="mr-1">mdi-shield-key-outline</v-icon>
            {{ groupAuthLabel(group) }}
          </span>
        </v-card-subtitle>

        <v-divider />

        <v-data-table
          :headers="headers"
          :items="groupIntegrations(group.id)"
          :items-per-page="-1"
          hide-default-footer
          density="compact"
          :no-data-text="$t('noIntegrationsInGroup')"
          class="Integrations__table"
          data-testid="integrations-table"
          :data-state="groupIntegrations(group.id).length === 0 ? 'empty' : 'ready'"
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
      </v-card>
    </div>
  </div>
</template>
<style lang="scss">
.Integrations__groups {
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
}
</style>
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
          title: this.$t('name'),
          key: 'name',
          width: '40%',
          sortable: true,
        },
        {
          title: this.$t('template'),
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
