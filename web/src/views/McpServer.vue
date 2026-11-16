<template>
  <div data-testid="mcp-page">
    <v-toolbar flat>
      <v-btn icon variant="text" class="mr-4" @click="returnToProjects()" data-testid="mcp-back">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title data-testid="mcp-title">{{ $t('mcp_server') }}</v-toolbar-title>
      <v-spacer />
      <v-btn
        v-if="available && tab === 'tokens'"
        variant="elevated"
        color="primary"
        data-testid="mcp-new-token"
        @click="$refs.tokens.newToken()"
      >
        {{ $t('New Token') }}
      </v-btn>
    </v-toolbar>

    <v-tabs v-if="available" v-model="tab" class="px-4" data-testid="mcp-tabs">
      <v-tab key="tokens" value="tokens" data-testid="mcp-tab-tokens">{{ $t('mcp_tokens') }}</v-tab>
      <v-tab key="project" value="project" data-testid="mcp-tab-project">
        {{ $t('mcp_tab_project') }}
      </v-tab>
      <v-tab v-if="canUseAdminMcp" key="admin" value="admin" data-testid="mcp-tab-admin">
        {{ $t('mcp_tab_admin') }}
      </v-tab>
      <v-tab v-if="canManageSystem" key="settings" value="settings" data-testid="mcp-tab-settings">
        {{ $t('mcp_tab_settings') }}
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px" />

    <template v-if="available">
      <Tokens
        v-if="tab === 'tokens'"
        ref="tokens"
        v-bind="$attrs"
        :systemInfo="systemInfo"
        :user="user"
        kind="mcp"
        embedded
      />

      <McpSettings
        v-else-if="tab === 'settings'"
        v-model="mcpSettings"
        :load-error="mcpSettingsError"
      />

      <!-- The project and the admin tabs: one layout, the set of the tab fills it. -->
      <div v-else class="pa-4" :data-testid="`mcp-connect-${mcpSet.id}`">
        <p class="mb-4" data-testid="mcp-intro">{{ $t(`mcp_intro_${mcpSet.id}`) }}</p>

        <v-alert
          v-if="plainRemoteHttp"
          type="warning"
          icon="mdi-alert"
          density="compact"
          variant="tonal"
          class="mb-4"
          data-testid="mcp-plain-http"
        >
          {{ $t('mcp_plain_http_warning') }}
        </v-alert>

        <!-- Left: how to connect a client; right: the limits of the connection. -->
        <v-row>
          <v-col cols="12" md="8">
            <div class="text-title-medium mb-5" data-testid="mcp-connection-title">
              {{ $t('mcp_connection_title') }}
            </div>

            <v-select
              v-model="clientId"
              :items="clients"
              item-title="name"
              item-value="id"
              :label="$t('mcp_client')"
              style="max-width: 400px"
              variant="outlined"
              density="compact"
              hide-details
              data-testid="mcp-client"
            />

            <div v-for="snippet in client.snippets" :key="snippet.title" class="mt-3">
              <div>{{ snippet.title }}:</div>
              <div style="position: relative">
                <pre class="McpServer__code">{{ snippet.code }}</pre>
                <CopyClipboardButton
                  :text="snippet.code"
                  class="McpServer__copy"
                  test-id="mcp-snippet-copy"
                />
              </div>
            </div>

            <div v-if="client.note" class="mt-2 text-body-small">{{ client.note }}</div>
          </v-col>

          <!-- Below md, where the columns stack, the switches go first: they change the configs
               under them. -->
          <v-col
            cols="12"
            md="4"
            class="order-first order-md-0"
            data-testid="mcp-connection-tools"
          >
            <div class="text-title-medium mb-1" data-testid="mcp-connection-tools-title">
              {{ $t('mcp_connection_tools') }}
            </div>
            <v-switch
              v-for="s in connectionSwitches"
              :key="s.key"
              :model-value="s.on"
              :label="$t(s.label)"
              :hint="s.allowed ? $t(s.tools) : $t('mcp_connection_switched_off_by_admin')"
              :disabled="!s.allowed"
              :data-testid="`mcp-connection-${mcpSet.id}-${s.key}`"
              persistent-hint
              inset
              class="McpSwitch McpServer__switch mt-2"
              @update:model-value="connection[mcpSet.id][s.key] = !!$event"
            />

          </v-col>
        </v-row>

        <!-- Every tool of the set; those tools/list of this connection would not show, for the
             switches above, are greyed out with the reason. -->
        <v-divider class="mt-8 mb-6" />
        <div data-testid="mcp-tools">
          <div class="text-title-medium mb-1" data-testid="mcp-tools-title">
            {{ $t('mcp_tools_title') }}
          </div>
          <v-progress-linear
            v-if="toolsLoading"
            indeterminate
            height="2"
            color="primary"
            class="mt-2"
            data-testid="mcp-tools-loading"
          />
          <v-alert
            v-if="toolsError"
            type="error"
            density="compact"
            variant="tonal"
            class="mt-2"
            data-testid="mcp-tools-error"
          >
            {{ $t('mcp_tools_load_error') }}: {{ toolsError }}
          </v-alert>
          <!-- Four in a row from lg; three below it, where four would push the badge under the
               name. -->
          <v-row density="compact">
            <v-col
              v-for="tool in tools || []"
              :key="tool.name"
              cols="12"
              sm="6"
              md="4"
              lg="3"
              class="py-2"
              data-testid="mcp-tool"
              :data-available="String(tool.available !== false)"
            >
              <!-- The icon has its own column, so the name and the summary share one left edge. -->
              <div class="d-flex align-start">
                <v-icon
                  :icon="toolIcon(tool.name)"
                  size="small"
                  :color="tool.available === false ? undefined : 'primary'"
                  class="McpServer__toolIcon mr-2"
                  :class="{ 'text-disabled': tool.available === false }"
                  data-testid="mcp-tool-icon"
                />
                <div class="McpServer__toolText">
                  <div class="d-flex flex-wrap align-center ga-1">
                    <div
                      class="McpServer__toolName"
                      :class="{ 'text-disabled': tool.available === false }"
                      data-testid="mcp-tool-name"
                    >{{ tool.name }}</div>
                    <v-chip
                      v-if="tool.available === false"
                      size="x-small"
                      label
                      class="McpServer__toolReason"
                      data-testid="mcp-tool-reason"
                    >{{ toolReason(tool) }}</v-chip>
                  </div>
                  <div
                    class="McpServer__toolSummary text-body-small"
                    :class="tool.available === false ? 'text-disabled' : 'text-medium-emphasis'"
                  >
                    {{ toolSummary(tool.description) }}
                  </div>
                </div>
              </div>
            </v-col>
          </v-row>
        </div>

      </div>
    </template>

    <v-alert
      v-else-if="needsProSeat"
      variant="tonal"
      color="hsl(348deg, 86%, 61%)"
      class="PageAlert"
      data-testid="mcp-needs-pro-seat"
    >
      <span v-html="$t('mcp_needs_pro_seat')"></span>
      <v-btn
        v-if="user && user.admin"
        class="ml-2 text-white"
        color="hsl(348deg, 86%, 61%)"
        to="/users"
        data-testid="mcp-assign-pro-seat"
      >
        {{ $t('mcp_assign_pro_seat') }}
      </v-btn>
      <span v-else style="font-weight: bold" data-testid="mcp-contact-admin-for-seat">
        {{ $t('mcp_contact_admin_for_seat') }}
      </span>
    </v-alert>

    <v-alert
      v-else-if="!features.mcp && !proLicenseWithoutSeat"
      variant="tonal"
      color="hsl(348deg, 86%, 61%)"
      class="PageAlert"
      data-testid="mcp-only-pro"
    >
      <span v-html="$t('mcp_only_pro')"></span>
      <v-btn
        v-if="canManageSystem"
        class="ml-2 text-white"
        color="hsl(348deg, 86%, 61%)"
        @click="upgradeToPro('mcp')"
        data-testid="mcp-upgrade"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>
      <span v-else style="font-weight: bold" data-testid="mcp-contact-admin">
        {{ $t('contact_admin_to_upgrade') }}
      </span>
    </v-alert>

    <!-- Less prominent than the Pro banner above: the feature is paid, only switched off. -->
    <v-alert
      v-else
      variant="tonal"
      color="grey-darken-1"
      class="PageAlert"
      data-testid="mcp-disabled"
    >
      <span v-html="$t('mcp_disabled_banner')"></span>
    </v-alert>

    <!-- The REST API accepts MCP tokens whatever the license, so the tokens the user still holds
         are listed for revoking under the banner that says why MCP is not available. -->
    <Tokens
      v-if="!available"
      v-bind="$attrs"
      :systemInfo="systemInfo"
      :user="user"
      kind="mcp"
      embedded
      hide-empty
      data-testid="mcp-revoke-only"
    />
  </div>
</template>

<script>
import EventBus from '@/event-bus';
import Tokens from '@/views/Tokens.vue';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';
import PageMixin from '@/components/PageMixin';
import { GLOBAL_PERMISSIONS } from '@/lib/constants';
import { hasGlobalPermission } from '@/lib/role-permissions';
import axios from 'axios';
import McpSettings, { MCP_SETTINGS_URL, MCP_SETTINGS_SWITCHES } from '@/components/McpSettings.vue';
import { getErrorMessage } from '@/lib/error';
import {
  MCP_SETS, MCP_TOOLS_URL, mcpClients, mcpConnectionHeaders, isPlainRemoteHttp, mcpToolIcon,
  mcpToolSummary, MCP_TOOL_REASONS, MCP_PERMISSION_LABELS,
} from '@/lib/mcpClients';

// The tab is the address. /mcp/connect, the old address of the project tab, redirects to
// /mcp/project.
const TAB_PATHS = {
  tokens: '/mcp',
  project: '/mcp/project',
  admin: '/mcp/admin',
  settings: '/mcp/settings',
};

// The tabs not every user sees, and the computed property that says who does.
const RESTRICTED_TABS = { admin: 'canUseAdminMcp', settings: 'canManageSystem' };

export default {
  mixins: [PageMixin],

  components: {
    Tokens,
    CopyClipboardButton,
    McpSettings,
  },

  inheritAttrs: false,

  props: {
    user: Object,
  },

  computed: {
    // The page issues MCP tokens and shows how to connect only when the user may use MCP and the
    // administrator has switched it on; otherwise it explains why not and lists the tokens the
    // user still holds, for revoking.
    available() {
      return !!this.features.mcp && !!this.systemInfo?.mcp?.enabled;
    },

    // An active Pro or Enterprise license turns on task_summary for everyone, and an Enterprise
    // one turns on MCP for everyone too: so without MCP here the license is Pro and the user has
    // no Pro seat.
    proLicenseWithoutSeat() {
      return this.systemInfo?.subscription_state === 'active'
        && !!this.features.task_summary && !this.features.mcp;
    },

    // A seat gives nothing while MCP is switched off: the page says it is off instead.
    needsProSeat() {
      return this.proLicenseWithoutSeat && !!this.systemInfo?.mcp?.enabled;
    },

    canManageSystem() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageSystem);
    },

    // Every tool of the admin set needs a global permission, and the server hides the tools the
    // user's permissions do not cover: the tab is for an admin or anyone with at least one.
    canUseAdminMcp() {
      return !!this.user && (!!this.user.admin || !!this.user.global_permissions);
    },

    // The set of MCP tools the tab connects to; the tabs without one show the project set.
    mcpSet() {
      return MCP_SETS[this.tab] || MCP_SETS.project;
    },

    // The switches of the tab's set, in the order of the settings. A group the administrator
    // switched off cannot be turned on for a connection, and its switch shows off.
    connectionSwitches() {
      const connection = this.connection[this.mcpSet.id] || {};
      return this.mcpSet.groups.map((group) => {
        const setting = MCP_SETTINGS_SWITCHES.find((s) => s.key === group.setting);
        const allowed = !!this.mcpSettings?.[group.setting];
        return {
          key: group.key,
          header: group.header,
          label: setting.label,
          tools: setting.hint,
          allowed,
          on: allowed && !!connection[group.key],
        };
      });
    },

    // A header goes out only for a group that is switched on here and allowed by the server.
    connectionHeaders() {
      return mcpConnectionHeaders(this.mcpSet, Object.fromEntries(
        this.connectionSwitches.map((s) => [s.key, s.on]),
      ));
    },

    // The server mounts its routes under the path of web_host, so the endpoint keeps that path
    // (https://host/semaphore → https://host/semaphore/api/mcp/project) and drops a trailing slash
    // that would give "//api/mcp". Without web_host the routes are at the root of the page's
    // origin.
    endpoint() {
      const host = this.systemInfo?.web_host;
      const { path } = this.mcpSet;
      try {
        return new URL(path, host ? `${host.replace(/\/+$/, '')}/` : document.baseURI).href;
      } catch (e) {
        return new URL(path, document.baseURI).href;
      }
    },

    // The query of the tool list of the tab's connection: its set and the flags of the switches
    // that are on; null on a tab without a connection.
    toolsQuery() {
      if (!this.available || !MCP_SETS[this.tab]) {
        return null;
      }
      const query = { set: this.mcpSet.id };
      this.connectionSwitches.filter((s) => s.on).forEach((s) => { query[s.key] = 'on'; });
      return query;
    },

    // A key of the query that changes only with its values, for the watcher.
    toolsQueryKey() {
      return JSON.stringify(this.toolsQuery);
    },

    clients() {
      return mcpClients(this.endpoint, {
        headers: this.connectionHeaders,
        serverName: this.mcpSet.serverName,
      });
    },

    client() {
      return this.clients.find((c) => c.id === this.clientId) || this.clients[0];
    },

    plainRemoteHttp() {
      return isPlainRemoteHttp(this.endpoint);
    },

    // The tab is the address; a restricted tab falls back to the tokens for anyone else.
    tab: {
      get() {
        const tab = Object.keys(TAB_PATHS).find((t) => TAB_PATHS[t] === this.$route.path) || 'tokens';
        return RESTRICTED_TABS[tab] && !this[RESTRICTED_TABS[tab]] ? 'tokens' : tab;
      },
      set(value) {
        const path = TAB_PATHS[value] || TAB_PATHS.tokens;
        if (this.$route.path !== path) {
          this.$router.push(path);
        }
      },
    },
  },

  data() {
    return {
      clientId: 'claude-code',
      // Every group starts off for a connection, as a client without the headers has them.
      connection: Object.fromEntries(Object.values(MCP_SETS).map((set) => [
        set.id, Object.fromEntries(set.groups.map((group) => [group.key, false])),
      ])),
      mcpSettings: null,
      mcpSettingsError: null,
      // The tools of the tab's connection, null before the first answer.
      tools: null,
      toolsLoading: false,
      toolsError: null,
      // The number of the last tool list request: an answer to an older one is dropped.
      toolsRequest: 0,
    };
  },

  watch: {
    toolsQueryKey: {
      handler() {
        this.loadTools();
      },
      immediate: true,
    },
  },

  async created() {
    await this.loadMcpSettings();
  },

  methods: {
    // Every signed-in user reads the server settings: the project and admin tabs need them for
    // their switches.
    async loadMcpSettings() {
      try {
        this.mcpSettings = (await axios({
          method: 'get',
          url: MCP_SETTINGS_URL,
          responseType: 'json',
        })).data;
        this.mcpSettingsError = null;
      } catch (err) {
        this.mcpSettingsError = getErrorMessage(err);
      }
    },

    // The server builds the list as for an MCP request of this user: the settings, the flags and
    // the global permissions of this moment.
    async loadTools() {
      this.toolsRequest += 1;
      const request = this.toolsRequest;
      const query = this.toolsQuery;
      if (!query) {
        this.tools = null;
        this.toolsLoading = false;
        this.toolsError = null;
        return;
      }
      this.toolsLoading = true;
      this.toolsError = null;
      try {
        const { data } = await axios({
          method: 'get',
          url: MCP_TOOLS_URL,
          params: query,
          responseType: 'json',
        });
        if (request === this.toolsRequest) {
          this.tools = data?.tools || [];
        }
      } catch (err) {
        if (request === this.toolsRequest) {
          this.tools = null;
          this.toolsError = getErrorMessage(err);
        }
      } finally {
        if (request === this.toolsRequest) {
          this.toolsLoading = false;
        }
      }
    },

    toolIcon: mcpToolIcon,

    toolSummary: mcpToolSummary,

    // Why a tool is not available, in words: the reason of GET /api/mcp/tools, with the missing
    // global permission by its translated name, or by its catalog ID when the page lacks one.
    toolReason(tool) {
      if (tool.reason === 'permission') {
        const label = MCP_PERMISSION_LABELS[tool.permission];
        return this.$t('mcp_tool_needs_permission', {
          permission: label ? this.$t(label) : tool.permission,
        });
      }
      const key = MCP_TOOL_REASONS[tool.reason];
      return key ? this.$t(key) : tool.reason;
    },

    returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },
  },
};
</script>

<style lang="scss">
.McpServer__code {
  overflow: auto;
  // A long one-line command wraps instead of scrolling under the copy button.
  white-space: pre-wrap;
  word-break: break-all;
  // The right side leaves room for the copy button.
  padding: 8px 48px 8px 8px;
  background: gray;
  color: white;
  border-radius: 10px;
  margin-top: 5px;
}

// A disabled switch dims the whole field, and Vuetify dims its hint once more; the hint says why
// the switch is off, so it keeps only the field's dimming.
.McpServer__switch.v-input--disabled .v-input__details {
  opacity: 1;
}

// The hint under the label text, its lines and a wrapping label — McpSwitch in McpSettings.vue.

.McpServer__toolIcon {
  opacity: .75;
  // Centres the 20px icon on the first line of the name.
  margin-top: 3px;
}

// Lets the summary shrink and clamp inside the flex row instead of widening it.
.McpServer__toolText {
  min-width: 0;
}

.McpServer__toolName {
  font-family: monospace;
  font-size: .875rem;
  font-weight: 500;
}

// Smaller than Vuetify's x-small chip: a note next to the name, not a control.
.McpServer__toolReason.v-chip {
  height: 16px;
  padding: 0 5px;
  font-size: .625rem;
}

// The first sentence of the description in at most two lines.
.McpServer__toolSummary {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.McpServer__copy {
  position: absolute;
  right: 6px;
  // Centres the 36px button on the first line of the block: padding 8px plus half of a 24px line.
  // The block's top margin collapses through the wrapper, so it does not count here.
  top: 2px;
  // The component wraps the button in an inline span, whose text baseline would push it down.
  display: flex;
}
</style>
