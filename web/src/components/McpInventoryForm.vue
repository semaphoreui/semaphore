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

    <div class="McpInventoryForm__subheader d-flex align-center text-medium-emphasis">
      <span data-testid="mcpInventoryForm-serversTitle">{{ $t('mcpServers') }}</span>

      <v-spacer />

      <v-btn
        icon
        variant="text"
        :title="$t('mcpAddServer')"
        :aria-label="$t('mcpAddServer')"
        :disabled="formSaving"
        @click="addServer()"
        data-testid="mcpInventoryForm-addServer"
      >
        <v-icon> mdi-plus </v-icon>
      </v-btn>
    </div>

    <v-data-table
      :items="servers"
      :items-per-page="-1"
      item-value="key"
      class="elevation-1 FieldTable McpInventoryForm__table"
      hide-default-footer
      :headers="serverTableHeaders"
      hide-default-header
      data-testid="mcpInventoryForm-servers"
      :data-state="servers.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:no-data>
        <span data-testid="mcpInventoryForm-noServers">{{ $t('mcpNoServers') }}</span>
      </template>

      <template v-slot:item="props">
        <tr class="McpInventoryForm__server" data-testid="mcpInventoryForm-server">
          <td class="pa-1 McpInventoryForm__enabled">
            <div class="McpInventoryForm__control">
              <v-switch
                v-model="props.item.enabled"
                color="primary"
                density="compact"
                hide-details
                :title="$t('enabled')"
                :aria-label="$t('enabled')"
                :disabled="formSaving"
                data-testid="mcpInventoryForm-serverEnabled"
              ></v-switch>
            </div>
          </td>

          <td class="pa-1 McpInventoryForm__name">
            <v-text-field
              v-model.trim="props.item.name"
              variant="solo-inverted"
              flat
              density="comfortable"
              hide-details="auto"
              class="McpInventoryForm__field"
              :placeholder="$t('name')"
              :aria-label="$t('name')"
              :rules="nameRules"
              :disabled="formSaving"
              autocomplete="off"
              data-testid="mcpInventoryForm-serverName"
            ></v-text-field>
          </td>

          <td class="pa-1 McpInventoryForm__url">
            <v-text-field
              v-model.trim="props.item.url"
              variant="solo-inverted"
              flat
              density="comfortable"
              hide-details="auto"
              class="McpInventoryForm__field"
              placeholder="https://mcp.example.com/mcp"
              aria-label="URL"
              :rules="urlRules"
              :disabled="formSaving"
              autocomplete="off"
              data-testid="mcpInventoryForm-serverUrl"
            ></v-text-field>
          </td>

          <!-- The stored token is never shown: the field takes a new one, "new-password" keeps
               the browser from filling in the password of the login. -->
          <td class="pa-1 McpInventoryForm__token">
            <v-text-field
              v-model="props.item.token"
              variant="solo-inverted"
              flat
              density="comfortable"
              hide-details
              class="McpInventoryForm__field masked-secret-input"
              autocomplete="new-password"
              :placeholder="tokenPlaceholder(props.item)"
              :aria-label="$t('mcpToken')"
              :readonly="props.item.removeToken"
              :disabled="formSaving"
              :data-state="tokenState(props.item)"
              data-testid="mcpInventoryForm-serverToken"
            >
              <template v-if="props.item.hasToken" v-slot:append-inner>
                <v-btn
                  icon
                  size="x-small"
                  variant="text"
                  :title="props.item.removeToken ? $t('mcpTokenKeep') : $t('mcpTokenRemove')"
                  :aria-label="props.item.removeToken ? $t('mcpTokenKeep') : $t('mcpTokenRemove')"
                  @click="toggleRemoveToken(props.item)"
                  data-testid="mcpInventoryForm-serverTokenRemove"
                >
                  <v-icon>{{ props.item.removeToken ? 'mdi-undo' : 'mdi-close' }}</v-icon>
                </v-btn>
              </template>
            </v-text-field>
          </td>

          <td class="McpInventoryForm__actions">
            <div class="McpInventoryForm__control">
              <v-icon
                size="small"
                class="pa-1 McpInventoryForm__remove"
                :title="$t('mcpRemoveServer')"
                :aria-label="$t('mcpRemoveServer')"
                :disabled="formSaving"
                @click="removeServer(props.item)"
                data-testid="mcpInventoryForm-serverRemove"
              >
                mdi-delete
              </v-icon>
            </div>
          </td>
        </tr>
      </template>
    </v-data-table>
  </v-form>
</template>
<style lang="scss">
// The list of servers looks as the tables of the Variable Group form (EnvironmentForm.vue, "Extra
// variables"): a subheader with the add button, then a borderless table of `solo-inverted` fields.

// v-subheader of Vuetify 2: a 48 px row, 14 px text in the letter spacing of the card text.
.McpInventoryForm__subheader {
  height: 30px;
  font-size: .875rem;
  padding-bottom: 10px;
  font-weight: 400;
  letter-spacing: .0071428571em;
}

// The grey sheet with the rounded corners of the Vuetify 2 one.
.McpInventoryForm__table.v-table {
  background: #8585850f;
  border-radius: 4px;
}

// A row: the switch, the name, the URL (the widest), the token and the delete icon. A cell is
// aligned to the top, so a validation message under one field does not move the others.
.McpInventoryForm__server > td {
  vertical-align: top;
}

// The switch alone in its cell: without the top offset of a Vuetify 2 form control and the gap
// to a label it does not have.
.McpInventoryForm__enabled {
  width: 59px;

  .v-switch {
    flex: none;
    margin: 0 0 0 8px;
    padding-top: 0;
  }

  .v-switch.v-input .v-selection-control__wrapper {
    margin-inline-end: 0;
  }
}

.McpInventoryForm__actions {
  width: 1px;
}

.McpInventoryForm__name {
  width: 20%;
}

.McpInventoryForm__token {
  width: 30%;
}

// The switch and the delete icon, centred against the 48 px field of their row.
.McpInventoryForm__control {
  display: flex;
  align-items: center;
  height: 48px;
}

.McpInventoryForm__actions > .McpInventoryForm__control {
  margin: 4px 0;
}

// The `solo-inverted` fields of Vuetify 2 (48 px) — grey at rest, dark with white text in focus
// (light theme); the dark theme inverts it. A field without a border shows a failed rule by a
// thin ring in the error colour.
.McpInventoryForm__field {
  .v-field.v-theme--light {
    background: rgba(0, 0, 0, .06);

    &:not(.v-field--focused) > .v-field__field > .v-field__input {
      color: rgba(0, 0, 0, .87);
    }

    &.v-field--focused {
      background: #424242;
      color: #fff;
    }
  }

  .v-field.v-theme--dark {
    background: rgba(255, 255, 255, .16);

    &.v-field--focused {
      background: #fff;
      color: rgba(0, 0, 0, .87);
    }
  }

  .v-field.v-field--error {
    box-shadow: inset 0 0 0 1px rgb(var(--v-theme-error));
  }
}

// The delete icon of a row in the icon colour of the theme (`.theme--light.v-icon`: .54); its
// padding goes around the 16 px glyph (a 24 px box), as on the content-box icon of Vuetify 2.
.McpInventoryForm__remove {
  box-sizing: content-box;
}

.McpInventoryForm__remove.v-theme--light {
  color: rgba(0, 0, 0, .54);
}

.McpInventoryForm__remove.v-theme--dark {
  color: #fff;
}

// On a narrow dialog three fields do not fit in a row: the URL and the token go under the name.
@media (max-width: 599px) {
  .McpInventoryForm__server {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    grid-template-areas:
      "enabled name actions"
      ". url ."
      ". token .";
    padding-bottom: 8px;

    > td {
      width: auto;
      height: auto;
    }
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

  .McpInventoryForm__actions {
    grid-area: actions;
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
      // Columns of the servers table (its rows come from the item slot; no header is shown).
      serverTableHeaders: [
        { key: 'enabled', sortable: false },
        { key: 'name' },
        { key: 'url' },
        { key: 'token' },
        { key: 'actions', sortable: false },
      ],
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

    removeServer(server) {
      const i = this.servers.indexOf(server);
      if (i > -1) {
        this.servers.splice(i, 1);
      }
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
          // The fields of the table have no label: an empty one is named by its placeholder.
          return this.$t('mcpToken');
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
