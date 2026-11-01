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

    <div class="px-4 pb-2 text--secondary" :class="{ 'pt-4': embedded }" style="font-size: 14px">
      {{ $t('endpointDescription') }}
    </div>

    <v-divider v-if="!embedded" />

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4 Endpoints__table"
      :items-per-page="Number.MAX_VALUE"
      hide-default-footer
      show-expand
      :expanded.sync="expanded"
      :no-data-text="$t('noEndpoints')"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto"
    >
      <template v-slot:item.name="{ item }">
        <a href="#" @click.prevent="toggleExpanded(item)" class="Endpoints__name">
          {{ item.name }}
        </a>
      </template>

      <template v-slot:item.auth_method="{ item }">
        <v-chip small label>{{ authMethodTitle(item.auth_method) }}</v-chip>
      </template>

      <template v-slot:item.aliases="{ item }">
        <span v-if="endpointAliases(item.id) != null">{{ endpointAliases(item.id).length }}</span>
        <v-progress-circular v-else indeterminate size="14" width="2" />
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

      <template v-slot:expanded-item="{ headers: cols, item }">
        <td :colspan="cols.length" class="Endpoints__aliases py-3">
          <div class="text-subtitle-2 mb-2">{{ $t('aliases') }}</div>

          <div
            class="mb-3 text--secondary"
            v-if="(endpointAliases(item.id) || []).length === 0"
          >{{ $t('noAliases') }}</div>

          <div
            v-else
            v-for="alias of endpointAliases(item.id)"
            :key="alias.id"
            class="Endpoints__alias"
          >
            <code class="mr-2">{{ alias.url }}</code>

            <CopyClipboardButton :text="alias.url" :success-message="$t('aliasUrlCopied')" />

            <v-btn
              v-if="can(USER_PERMISSIONS.manageProjectResources)"
              icon
              @click="deleteAlias(item.id, alias.id)"
            >
              <v-icon>mdi-delete</v-icon>
            </v-btn>
          </div>

          <v-btn
            v-if="can(USER_PERMISSIONS.manageProjectResources)"
            color="primary"
            small
            class="mt-1"
            @click="addAlias(item.id)"
            :disabled="endpointAliases(item.id) == null"
          >
            {{ endpointAliases(item.id) == null ? $t('LoadAlias') : $t('AddAlias') }}
          </v-btn>
        </td>
      </template>
    </v-data-table>
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

  data() {
    return {
      /** Rows whose alias block is open (v-data-table `expanded`). */
      expanded: [],
    };
  },

  methods: {
    allowActions() {
      return this.can(USER_PERMISSIONS.updateProject);
    },

    authMethodTitle(method) {
      return authMethodTitle(method, (key) => this.$t(key));
    },

    toggleExpanded(item) {
      if (this.expanded.some((e) => e.id === item.id)) {
        this.expanded = this.expanded.filter((e) => e.id !== item.id);
      } else {
        this.expanded = [...this.expanded, item];
      }
    },

    getHeaders() {
      return [
        {
          text: this.$i18n.t('name'),
          value: 'name',
          width: '30%',
          sortable: true,
        },
        {
          text: this.$i18n.t('description'),
          value: 'description',
          width: '40%',
          sortable: false,
        },
        {
          text: this.$i18n.t('authMethod'),
          value: 'auth_method',
          width: '15%',
          sortable: true,
        },
        {
          text: this.$i18n.t('aliases'),
          value: 'aliases',
          width: '10%',
          sortable: false,
        },
        {
          value: 'actions',
          sortable: false,
          width: '0%',
        },
        {
          value: 'data-table-expand',
          sortable: false,
          width: '0%',
        },
      ];
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

    /** Loads the endpoints and, for the alias count column, the aliases of each. */
    async loadItems() {
      await ItemListPageBase.methods.loadItems.call(this);
      await Promise.all((this.items || []).map((endpoint) => this.loadAliases(endpoint.id)));
    },

    /** An endpoint used by integration groups cannot be deleted; show who uses it. */
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
