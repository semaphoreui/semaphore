<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('NewEndpoint') : $t('EditEndpoint')"
      :max-width="450"
      :transition="false"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <EndpointForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      :object-title="$t('endpoint')"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      :title="$t('DeleteEndpoint')"
      :text="$t('DeleteEndpointMsg')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <!-- Shown on the Endpoints tab of the Integrations page, the toolbar is that page's. -->
    <v-toolbar flat v-if="!embedded">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('endpoints') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        @click="editItem('new')"
        data-testid="endpoints-new"
        >{{ $t('NewEndpoint') }}
      </v-btn>
    </v-toolbar>

    <v-divider v-if="!embedded" />

    <!-- One card per endpoint, laid out like the integrations: header, meta line, its items. -->
    <div
      class="px-4 py-4 Endpoints__list"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto"
    >
      <div class="pb-4 text--secondary" style="font-size: 14px">
        {{ $t('endpointDescription') }}
      </div>

      <div v-if="items.length === 0" class="text--secondary py-6 text-center">
        {{ $t('noEndpoints') }}
      </div>

      <v-card
        v-for="item of items"
        :key="item.id"
        outlined
        class="mb-4 Endpoints__card"
        :data-testid="`endpoint-${item.id}`"
      >
        <v-card-title class="pb-2 Endpoints__cardHeader">
          <v-icon class="mr-2">mdi-webhook</v-icon>
          <span class="Endpoints__name">{{ item.name }}</span>

          <v-spacer></v-spacer>

          <template v-if="can(USER_PERMISSIONS.manageProjectResources)">
            <v-btn
              color="primary"
              small
              text
              class="mr-1"
              :disabled="endpointAliases(item.id) == null"
              @click="addAlias(item.id)"
            >
              <v-icon small left>mdi-plus</v-icon>
              {{ $t('AddAlias') }}
            </v-btn>
            <v-btn icon small class="mr-1" @click="editItem(item.id)">
              <v-icon small>mdi-pencil</v-icon>
            </v-btn>
            <v-btn icon small @click="askDeleteItem(item.id)">
              <v-icon small>mdi-delete</v-icon>
            </v-btn>
          </template>
        </v-card-title>

        <v-card-subtitle class="pb-2 Endpoints__cardMeta">
          <span class="mr-4">
            <v-icon small class="mr-1">mdi-shield-key-outline</v-icon>
            {{ $t('authMethod') }}: {{ authMethodTitle(item.auth_method) }}
          </span>
          <span v-if="item.description" class="Endpoints__description">{{ item.description }}</span>
        </v-card-subtitle>

        <v-divider />

        <v-card-text class="Endpoints__aliases">
          <v-progress-circular
            v-if="endpointAliases(item.id) == null"
            indeterminate
            size="16"
            width="2"
          />

          <div
            v-else-if="endpointAliases(item.id).length === 0"
            class="text--secondary"
          >{{ $t('noAliases') }}</div>

          <div
            v-else
            v-for="alias of endpointAliases(item.id)"
            :key="alias.id"
            class="Endpoints__alias d-flex align-center"
          >
            <code class="mr-2">{{ alias.url }}</code>

            <CopyClipboardButton :text="alias.url" :success-message="$t('aliasUrlCopied')" />

            <v-btn
              v-if="can(USER_PERMISSIONS.manageProjectResources)"
              icon
              small
              @click="deleteAlias(item.id, alias.id)"
            >
              <v-icon small>mdi-delete</v-icon>
            </v-btn>
          </div>
        </v-card-text>
      </v-card>
    </div>
  </div>
</template>
<script>
import axios from 'axios';

import { USER_PERMISSIONS } from '@/lib/constants';
import { authMethodTitle } from '@/lib/integrationAuth';

import ItemListPageBase from '@/components/ItemListPageBase';
import EndpointForm from '@/components/EndpointForm.vue';
import EndpointAliasesBase from '@/views/project/EndpointAliasesBase';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';

export default {
  mixins: [ItemListPageBase, EndpointAliasesBase],
  components: { EndpointForm, CopyClipboardButton },

  props: {
    /** Rendered inside the Integrations page: no toolbar of its own. */
    embedded: Boolean,
  },

  methods: {
    allowActions() {
      return this.can(USER_PERMISSIONS.updateProject);
    },

    authMethodTitle(method) {
      return authMethodTitle(method, (key) => this.$t(key));
    },

    // The cards replace the table; the mixin still asks for headers.
    getHeaders() {
      return [];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/endpoints`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/endpoints/${this.itemId}`;
    },

    getEventName() {
      return 'w-endpoint';
    },

    /** Loads the endpoints and the aliases of each. */
    async loadItems() {
      await ItemListPageBase.methods.loadItems.call(this);
      await Promise.all((this.items || []).map((endpoint) => this.loadAliases(endpoint.id)));
    },

    /** An endpoint used by integrations cannot be deleted; show who uses it. */
    async askDeleteItem(itemId) {
      this.itemId = itemId;

      try {
        this.itemRefs = (await axios({
          method: 'get',
          url: `${this.getSingleItemUrl()}/refs`,
          responseType: 'json',
        })).data;

        if ((this.itemRefs.integration_groups || []).length > 0) {
          this.itemRefsDialog = true;
          return;
        }
      } catch (e) {
        // Fall through to the confirmation; the server refuses the delete if it must.
      }

      this.deleteItemDialog = true;
    },
  },
};
</script>
