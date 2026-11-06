<template>
  <div class="pa-4" data-testid="mcp-settings">
    <v-alert
      type="info"
      density="compact"
      variant="tonal"
      class="mb-4"
      data-testid="mcp-settings-scope-note"
    >
      {{ $t('mcp_settings_scope_note') }}
    </v-alert>

    <v-alert
      v-if="loadError"
      type="error"
      density="compact"
      variant="tonal"
      data-testid="mcp-settings-load-error"
    >
      {{ loadError }}
    </v-alert>

    <v-progress-linear v-else-if="draft == null" indeterminate />

    <template v-else>
      <div
        v-for="(section, i) in sections"
        :key="section.key"
        :class="{ 'mt-6': i > 0 }"
        :data-testid="`mcp-settings-section-${section.key}`"
      >
        <div class="text-title-medium mb-2" data-testid="mcp-settings-section-title">
          {{ $t(section.title) }}
        </div>
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
          class="McpSwitch McpSettings__switch mt-0 mb-4"
          @update:model-value="save()"
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
// order the user chose: changes, then deleting, then running.
export const MCP_SETTINGS_SECTIONS = [
  {
    key: 'project',
    title: 'mcp_tab_project',
    switches: [
      { key: 'allow_write', label: 'mcp_settings_admin_allow_write', hint: 'mcp_settings_allow_write_hint' },
      { key: 'allow_destroy', label: 'mcp_settings_allow_destroy', hint: 'mcp_settings_allow_destroy_hint' },
      { key: 'allow_run', label: 'mcp_settings_allow_run', hint: 'mcp_settings_allow_run_hint' },
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

// Every setting as a boolean: a key the answer lacks still gets its switch and is sent on save.
const allSettings = (value) => Object.fromEntries(
  MCP_SETTINGS_SWITCHES.map((s) => [s.key, !!value[s.key]]),
);

/**
 * The server-wide ceiling for the MCP tools: `v-model` is the `GET /api/mcp/settings` answer the
 * page loaded. Only a user who may manage the system sees it; the server checks that right again
 * on every change. A switch is saved as soon as it is toggled, and `update:modelValue` gives the
 * page what the server stored.
 */
export default {
  props: {
    modelValue: Object,
    loadError: String,
  },

  emits: ['update:modelValue'],

  data() {
    return {
      draft: null,
      saving: false,
      sections: MCP_SETTINGS_SECTIONS,
    };
  },

  watch: {
    modelValue: {
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
        this.$emit('update:modelValue', stored);
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.$t('mcp_settings_saved'),
        });
      } catch (err) {
        this.draft = this.modelValue ? allSettings(this.modelValue) : null;
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
<style lang="scss">
// A switch with its hint under the label text, here and on the connection tabs of McpServer.vue.
// The hint starts under the label, not under the switch: the inset switch and its gap take 58px.
.McpSwitch.v-input .v-input__details {
  padding-inline-start: 58px;
  // The 40px control already leaves 11px under the label text; the hint sits 5px below it.
  padding-top: 0;
  margin-top: -6px;
}

// A hint wraps on a narrow screen; Vuetify sets its line to the font size, which glues the lines.
.McpSwitch .v-messages__message {
  line-height: 1rem;
}

// A label that wraps (a phone, the narrow column of a tablet) grows the control instead of running
// under the hint: Vuetify stretches the label over the 40px control and centres its text. The label
// keeps the 10px above and below its text that centring gives one line, so a one-line label stays
// where it was, and the switch stays level with the first line.
.McpSwitch .v-selection-control {
  align-items: flex-start;
}

.McpSwitch .v-selection-control .v-label {
  height: auto;
  padding-block: 10px;
}

// The hint here has always sat 2px higher than on the connection tabs; its 1rem line would move it
// down by those 2px, so it is lifted back without moving what follows.
.McpSettings__switch .v-messages {
  position: relative;
  top: -2px;
}
</style>
