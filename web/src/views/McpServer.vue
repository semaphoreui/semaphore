<template>
  <div>
    <v-toolbar flat>
      <v-btn icon class="mr-4" @click="returnToProjects()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ $t('mcp_server') }}</v-toolbar-title>
      <v-spacer />
      <v-btn
        v-if="available && tab === 'tokens'"
        color="primary"
        data-testid="mcp-new-token"
        @click="$refs.tokens.newToken()"
      >
        {{ $t('New Token') }}
      </v-btn>
    </v-toolbar>

    <v-tabs v-if="available" v-model="tab" class="px-4">
      <v-tab key="tokens" href="#tokens" data-testid="mcp-tab-tokens">{{ $t('mcp_tokens') }}</v-tab>
      <v-tab key="connect" href="#connect" data-testid="mcp-tab-connect">
        {{ $t('mcp_connection') }}
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

      <div v-else class="pa-4" data-testid="mcp-connect">
        <v-alert
          v-if="plainRemoteHttp"
          type="warning"
          icon="mdi-alert"
          dense
          text
          data-testid="mcp-plain-http"
        >
          {{ $t('mcp_plain_http_warning') }}
        </v-alert>

        <v-alert type="info" dense text>
          {{ $t('mcp_cloud_clients_info') }}
        </v-alert>

        <div>{{ $t('mcp_endpoint') }}:</div>
        <div style="position: relative">
          <pre class="McpServer__code">{{ endpoint }}</pre>
          <CopyClipboardButton :text="endpoint" class="McpServer__copy" />
        </div>

        <v-select
          v-model="clientId"
          :items="clients"
          item-text="name"
          item-value="id"
          :label="$t('mcp_client')"
          class="mt-4"
          style="max-width: 400px"
          outlined
          dense
          hide-details
          data-testid="mcp-client"
        />

        <div v-for="snippet in client.snippets" :key="snippet.title" class="mt-3">
          <div>{{ snippet.title }}:</div>
          <div style="position: relative">
            <pre class="McpServer__code">{{ snippet.code }}</pre>
            <CopyClipboardButton :text="snippet.code" class="McpServer__copy" />
          </div>
        </div>

        <div v-if="client.note" class="mt-2 text-caption">{{ client.note }}</div>

      </div>
    </template>

    <v-alert
      v-else-if="!features.mcp"
      text
      color="hsl(348deg, 86%, 61%)"
      class="PageAlert"
      data-testid="mcp-only-pro"
    >
      <span v-html="$t('mcp_only_pro')"></span>
      <v-btn
        dark
        v-if="canManageSystem"
        class="ml-2"
        color="hsl(348deg, 86%, 61%)"
        @click="upgradeToPro('mcp')"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>
      <span v-else style="font-weight: bold">
        {{ $t('contact_admin_to_upgrade') }}
      </span>
    </v-alert>

    <!-- Less prominent than the Pro banner above: the feature is paid, only switched off. -->
    <v-alert v-else text color="grey darken-1" class="PageAlert" data-testid="mcp-disabled">
      <span v-html="$t('mcp_disabled_banner')"></span>
    </v-alert>
  </div>
</template>

<script>
import EventBus from '@/event-bus';
import Tokens from '@/views/Tokens.vue';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';
import PageMixin from '@/components/PageMixin';
import { GLOBAL_PERMISSIONS } from '@/lib/constants';
import { hasGlobalPermission } from '@/lib/role-permissions';
import { mcpClients, isPlainRemoteHttp } from '@/lib/mcpClients';

const TOKENS_PATH = '/mcp';
const CONNECT_PATH = '/mcp/connect';

export default {
  mixins: [PageMixin],

  components: {
    Tokens,
    CopyClipboardButton,
  },

  inheritAttrs: false,

  props: {
    user: Object,
  },

  computed: {
    // The page talks to the MCP token routes only when the user may use MCP
    // and the administrator has switched it on; otherwise it explains why not.
    available() {
      return !!this.features.mcp && !!this.systemInfo?.mcp?.enabled;
    },

    canManageSystem() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageSystem);
    },

    endpoint() {
      return `${this.systemInfo?.web_host || window.location.origin}/api/mcp`;
    },

    clients() {
      return mcpClients(this.endpoint);
    },

    client() {
      return this.clients.find((c) => c.id === this.clientId) || this.clients[0];
    },

    plainRemoteHttp() {
      return isPlainRemoteHttp(this.endpoint);
    },

    // The tab is the address: /mcp/connect shows how to connect, /mcp the tokens.
    tab: {
      get() {
        return this.$route.path === CONNECT_PATH ? 'connect' : 'tokens';
      },
      set(value) {
        const path = value === 'connect' ? CONNECT_PATH : TOKENS_PATH;
        if (this.$route.path !== path) {
          this.$router.push(path);
        }
      },
    },
  },

  data() {
    return {
      clientId: 'claude-code',
    };
  },

  methods: {
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
