<template>
  <v-skeleton-loader
    v-if="!isLoaded"
    data-testid="mcpInventoryForm-loading"
    :data-state="formError ? 'error' : 'loading'"
    type="list-item, list-item-two-line, list-item-two-line"
  ></v-skeleton-loader>
  <v-form
    v-else
    ref="form"
    v-model="formValid"
    data-testid="mcpInventoryForm-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="mcpInventoryForm-error"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[v => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      data-testid="mcpInventoryForm-name"
    ></v-text-field>

    <v-autocomplete
      v-if="features?.project_runners"
      v-model="item.runner_tag"
      :items="runnerTags || []"
      :label="$t('runner_tag')"
      item-value="tag"
      item-title="tag"
      variant="outlined"
      density="compact"
      clearable
      :disabled="formSaving"
      :placeholder="$t('runner_tag')"
      data-testid="mcpInventoryForm-runnerTag"
    ></v-autocomplete>

    <div class="d-flex align-center McpInventoryForm__subheader">
      <span data-testid="mcpInventoryForm-serversTitle">{{ $t('mcpServers') }}</span>
      <v-spacer />
      <v-btn
        variant="text"
        prepend-icon="mdi-plus"
        :disabled="formSaving"
        @click="addServer()"
        data-testid="mcpInventoryForm-addServer"
      >{{ $t('mcpAddServer') }}</v-btn>
    </div>

    <div
      v-if="servers.length === 0"
      class="text-medium-emphasis McpInventoryForm__empty"
      data-testid="mcpInventoryForm-noServers"
    >{{ $t('mcpNoServers') }}</div>

    <div
      v-for="(server, index) in servers"
      :key="server.key"
      class="McpInventoryForm__server"
      data-testid="mcpInventoryForm-server"
    >
      <v-switch
        v-model="server.enabled"
        class="McpInventoryForm__enabled"
        color="primary"
        density="compact"
        hide-details
        :title="$t('enabled')"
        :aria-label="$t('enabled')"
        :disabled="formSaving"
        data-testid="mcpInventoryForm-serverEnabled"
      ></v-switch>

      <v-text-field
        v-model.trim="server.name"
        class="McpInventoryForm__name"
        :label="$t('name')"
        :rules="nameRules"
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        hide-details="auto"
        autocomplete="off"
        data-testid="mcpInventoryForm-serverName"
      ></v-text-field>

      <v-text-field
        v-model.trim="server.url"
        class="McpInventoryForm__url"
        label="URL"
        :rules="urlRules"
        placeholder="https://mcp.example.com/mcp"
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        hide-details="auto"
        autocomplete="off"
        data-testid="mcpInventoryForm-serverUrl"
      ></v-text-field>

      <!-- The stored token is never shown: the field takes a new one, "new-password" keeps the
           browser from filling in the password of the login. -->
      <v-text-field
        v-model="server.token"
        class="McpInventoryForm__token"
        type="password"
        autocomplete="new-password"
        :label="$t('mcpToken')"
        :placeholder="tokenPlaceholder(server)"
        persistent-placeholder
        :readonly="server.removeToken"
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        hide-details
        :data-state="tokenState(server)"
        data-testid="mcpInventoryForm-serverToken"
      >
        <template v-if="server.hasToken" v-slot:append-inner>
          <v-btn
            icon
            size="x-small"
            variant="text"
            :title="server.removeToken ? $t('mcpTokenKeep') : $t('mcpTokenRemove')"
            :aria-label="server.removeToken ? $t('mcpTokenKeep') : $t('mcpTokenRemove')"
            @click="toggleRemoveToken(server)"
            data-testid="mcpInventoryForm-serverTokenRemove"
          >
            <v-icon>{{ server.removeToken ? 'mdi-undo' : 'mdi-key-remove' }}</v-icon>
          </v-btn>
        </template>
      </v-text-field>

      <v-btn
        class="McpInventoryForm__remove"
        icon
        size="small"
        variant="text"
        :title="$t('mcpRemoveServer')"
        :aria-label="$t('mcpRemoveServer')"
        :disabled="formSaving"
        @click="removeServer(index)"
        data-testid="mcpInventoryForm-serverRemove"
      >
        <v-icon>mdi-delete</v-icon>
      </v-btn>
    </div>

    <div
      class="text-body-small text-medium-emphasis mt-3"
      data-testid="mcpInventoryForm-hint"
    >{{ $t('mcpServersHint') }}</div>
  </v-form>
</template>
<style lang="scss">
// v-subheader of Vuetify 2, as the tables of the Variable Group form: a 48 px row of 14 px text.
.McpInventoryForm__subheader {
  height: 48px;
  font-size: .875rem;
  letter-spacing: .0071428571em;
}

.McpInventoryForm__empty {
  padding: 8px 0 16px;
}

// One server per row: the switch, the name, the URL (the widest), the token and the delete
// button; on a narrow dialog the URL and the token go under the name.
.McpInventoryForm__server {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) minmax(0, 1.8fr) minmax(0, 1.4fr) auto;
  grid-template-areas: "enabled name url token remove";
  column-gap: 8px;
  align-items: start;
  margin-bottom: 12px;
}

.McpInventoryForm__enabled {
  grid-area: enabled;
}

.McpInventoryForm__name {
  grid-area: name;
}

.McpInventoryForm__url {
  grid-area: url;
}

.McpInventoryForm__token {
  grid-area: token;
}

.McpInventoryForm__remove.v-btn {
  grid-area: remove;
  margin-top: 4px;
}

@media (max-width: 599px) {
  .McpInventoryForm__server {
    grid-template-columns: auto minmax(0, 1fr) auto;
    grid-template-areas:
      "enabled name remove"
      ". url ."
      ". token .";
    row-gap: 8px;
  }
}
</style>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import { MCP_INVENTORY } from '@/lib/constants';
import {
  isValidMcpServerName,
  isValidMcpServerUrl,
  mcpServerRows,
  mcpServersRequest,
  newMcpServerRow,
} from '@/lib/mcp';

// The form of an mcp inventory: its name and the list of remote MCP servers of the AI agents,
// saved in one request with the inventory (`mcp_servers`). A token is only ever sent, never
// received: the API tells only whether a server has one.
export default {
  mixins: [ItemFormBase],

  props: {
    features: Object,
  },

  data() {
    return {
      servers: [],
      nextServerKey: 1,
      runnerTags: null,
    };
  },

  computed: {
    isLoaded() {
      return this.item != null && (!this.features?.project_runners || this.runnerTags != null);
    },

    nameRules() {
      return [
        (v) => !!v || this.$t('name_required'),
        (v) => isValidMcpServerName(v) || this.$t('mcp_server_name_invalid'),
        (v) => this.servers.filter((s) => s.name === v).length < 2
          || this.$t('mcp_server_name_duplicate'),
      ];
    },

    urlRules() {
      return [
        (v) => !!v || this.$t('mcp_server_url_required'),
        (v) => isValidMcpServerUrl(v) || this.$t('mcp_server_url_invalid'),
      ];
    },
  },

  async created() {
    if (this.features?.project_runners) {
      this.runnerTags = await this.loadProjectResources('runner_tags');
    }
  },

  methods: {
    getNewItem() {
      return {
        name: '',
        type: MCP_INVENTORY.type,
      };
    },

    afterLoadData() {
      this.servers = mcpServerRows(this.item.mcp_servers);
      this.nextServerKey = this.servers.length + 1;
      // A new inventory starts with an empty server to fill in.
      if (this.isNew && this.servers.length === 0) {
        this.addServer();
      }
    },

    afterReset() {
      this.servers = [];
    },

    addServer() {
      this.servers.push(newMcpServerRow(this.nextServerKey));
      this.nextServerKey += 1;
    },

    removeServer(index) {
      this.servers.splice(index, 1);
    },

    toggleRemoveToken(server) {
      // eslint-disable-next-line no-param-reassign
      server.removeToken = !server.removeToken;
      // eslint-disable-next-line no-param-reassign
      server.token = '';
    },

    tokenState(server) {
      if (server.removeToken) {
        return 'remove';
      }
      if (server.token) {
        return 'new';
      }
      return server.hasToken ? 'set' : 'none';
    },

    tokenPlaceholder(server) {
      switch (this.tokenState(server)) {
        case 'remove':
          return this.$t('mcpTokenWillBeRemoved');
        case 'set':
          return this.$t('mcpTokenSet');
        default:
          return this.$t('mcpTokenOptional');
      }
    },

    beforeSave() {
      this.item.type = MCP_INVENTORY.type;
      this.item.mcp_servers = mcpServersRequest(this.servers);
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.itemId}`;
    },
  },
};
</script>
