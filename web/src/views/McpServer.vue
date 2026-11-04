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
        <v-alert
          v-if="plainRemoteHttp"
          type="warning"
          icon="mdi-alert"
          density="compact"
          variant="tonal"
          data-testid="mcp-plain-http"
        >
          {{ $t('mcp_plain_http_warning') }}
        </v-alert>

        <v-alert type="info" density="compact" variant="tonal" data-testid="mcp-cloud-clients">
          {{ $t('mcp_cloud_clients_info') }}
        </v-alert>

        <div>{{ $t('mcp_endpoint') }}:</div>
        <div style="position: relative">
          <pre class="McpServer__code" data-testid="mcp-endpoint">{{ endpoint }}</pre>
          <CopyClipboardButton
            :text="endpoint"
            class="McpServer__copy"
            test-id="mcp-endpoint-copy"
          />
        </div>

        <div class="mt-4">{{ $t('mcp_connection_tools') }}:</div>
        <div class="text-body-small">{{ $t('mcp_connection_tools_hint') }}</div>
        <v-switch
          v-for="s in connectionSwitches"
          :key="s.key"
          :model-value="s.on"
          :label="$t(s.label)"
          :hint="s.allowed
            ? $t('mcp_connection_header_hint', { header: s.header })
            : $t('mcp_connection_switched_off_by_admin')"
          :disabled="!s.allowed"
          :data-testid="`mcp-connection-${mcpSet.id}-${s.key}`"
          persistent-hint
          inset
          class="McpServer__switch mt-2"
          @update:model-value="connection[mcpSet.id][s.key] = !!$event"
        />

        <template v-if="mcpSet.id === 'admin'">
          <div class="mt-4 text-body-small" data-testid="mcp-admin-groups">
            <div v-for="s in connectionSwitches" :key="s.key">
              <b>{{ $t(s.label) }}</b>: {{ $t(s.tools) }}
            </div>
          </div>
          <div class="mt-1 text-body-small" data-testid="mcp-admin-hidden-tools">
            {{ $t('mcp_admin_hidden_tools') }}
          </div>
        </template>

        <v-select
          v-model="clientId"
          :items="clients"
          item-title="name"
          item-value="id"
          :label="$t('mcp_client')"
          class="mt-4"
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
    >
      <div class="px-4 pt-4">{{ $t('mcp_tokens_without_mcp') }}</div>
    </Tokens>

    <!-- The settings hold for every user who has MCP, so a system manager prepares them even
         while MCP is not available to them: switched off, or no Pro seat of their own. -->
    <template v-if="!available && settingsWithoutMcp">
      <div class="McpSubheader d-flex align-center text-medium-emphasis px-4 mt-4">
        {{ $t('mcp_tab_settings') }}
      </div>
      <div class="px-4 text-body-small">{{ $t('mcp_settings_without_mcp') }}</div>
      <McpSettings
        v-model="mcpSettings"
        :load-error="mcpSettingsError"
        data-testid="mcp-settings-without-mcp"
      />
    </template>
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
  MCP_SETS, mcpClients, mcpConnectionHeaders, isPlainRemoteHttp,
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

    // Without MCP for this user, a system manager still sees the settings when the instance has
    // MCP in its license: the user only lacks a Pro seat, or MCP is switched off. Without the
    // license nobody has MCP, and the page only offers the upgrade.
    settingsWithoutMcp() {
      return this.canManageSystem && (!!this.features.mcp || this.proLicenseWithoutSeat);
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
    };
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
