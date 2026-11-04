<template>
  <div class="pa-4" data-testid="mcp-settings">
    <v-alert type="info" dense text data-testid="mcp-settings-rest-note">
      {{ $t('mcp_settings_rest_note') }}
    </v-alert>

    <v-alert v-if="loadError" type="error" dense text data-testid="mcp-settings-load-error">
      {{ loadError }}
    </v-alert>

    <v-progress-linear v-else-if="draft == null" indeterminate />

    <template v-else>
      <div
        v-for="section in sections"
        :key="section.key"
        :data-testid="`mcp-settings-section-${section.key}`"
      >
        <v-subheader class="px-0">{{ $t(section.title) }}</v-subheader>
        <v-switch
          v-for="s in section.switches"
          :key="s.key"
          v-model="draft[s.key]"
          :label="$t(s.label)"
          :hint="$t(s.hint)"
          :disabled="saving"
          :data-testid="`mcp-settings-${s.key}`"
          persistent-hint
          inset
          class="mt-0 mb-4"
          @change="save()"
        />
      </div>
    </template>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export const MCP_SETTINGS_URL = '/api/mcp/settings';

// One section per set of MCP tools, in the order of the page tabs; in each, the
// active actions first, then deleting.
export const MCP_SETTINGS_SECTIONS = [
  {
    key: 'project',
    title: 'mcp_tab_project',
    switches: [
      { key: 'allow_run', label: 'mcp_settings_allow_run', hint: 'mcp_settings_allow_run_hint' },
      { key: 'allow_destroy', label: 'mcp_settings_allow_destroy', hint: 'mcp_settings_allow_destroy_hint' },
    ],
  },
  {
    key: 'admin',
    title: 'mcp_tab_admin',
    switches: [
      { key: 'admin_allow_write', label: 'mcp_settings_admin_allow_write', hint: 'mcp_settings_admin_allow_write_hint' },
      { key: 'admin_allow_destroy', label: 'mcp_settings_allow_destroy', hint: 'mcp_settings_admin_allow_destroy_hint' },
    ],
  },
];

// Every switch of every section, by the setting it stores.
export const MCP_SETTINGS_SWITCHES = MCP_SETTINGS_SECTIONS.flatMap((section) => section.switches);

// Every setting as a boolean: a key the answer lacks would not be reactive in the draft.
const allSettings = (value) => Object.fromEntries(
  MCP_SETTINGS_SWITCHES.map((s) => [s.key, !!value[s.key]]),
);

/**
 * The server-wide ceiling for the MCP tools: `value` is the `GET /api/mcp/settings` answer the
 * page loaded. Only a user who may manage the system sees it; the server checks that right again
 * on every change. A switch is saved as soon as it is toggled, and `input` gives the page what
 * the server stored.
 */
export default {
  props: {
    value: Object,
    loadError: String,
  },

  data() {
    return {
      draft: null,
      saving: false,
      sections: MCP_SETTINGS_SECTIONS,
    };
  },

  watch: {
    value: {
      immediate: true,
      handler(value) {
        this.draft = value ? allSettings(value) : null;
      },
    },
  },

  methods: {
    // Sends every value, as the route requires. A failed save puts the switches back where the
    // server has them.
    async save() {
      this.saving = true;
      try {
        const stored = (await axios({
          method: 'put',
          url: MCP_SETTINGS_URL,
          responseType: 'json',
          data: allSettings(this.draft),
        })).data;
        this.draft = allSettings(stored);
        this.$emit('input', stored);
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.$t('mcp_settings_saved'),
        });
      } catch (err) {
        this.draft = this.value ? allSettings(this.value) : null;
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.saving = false;
      }
    },
  },
};
</script>
