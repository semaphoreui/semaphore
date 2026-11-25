<!--
  A row of the task log made of an AI agent's events (lib/agentLog.js): a line of its text, a tool
  call with its arguments and result, the start and the end of the run.

  Everything the agent and its tools wrote is shown as text, never as HTML.
-->
<template>
  <div
    class="task-log-agent"
    :class="{ 'task-log-agent--sub': agent.sub !== '' }"
    :data-kind="agent.kind"
    data-testid="taskLog-agentRecord"
  >
    <span
      v-if="agent.sub !== ''"
      class="task-log-agent__sub"
      :title="agent.sub"
      data-testid="taskLog-agentSub"
    >{{ agent.sub }}</span>

    <div v-if="agent.kind === 'text'" class="task-log-agent__text" data-testid="taskLog-agentText">
      <template v-for="(segment, i) in segments" :key="i">
        <b v-if="segment.bold">{{ segment.text }}</b>
        <code v-else-if="segment.code" class="task-log-agent__code">{{ segment.text }}</code>
        <template v-else>{{ segment.text }}</template>
      </template>
    </div>

    <div
      v-else-if="agent.kind === 'thinking'"
      class="task-log-agent__thinking"
      data-testid="taskLog-agentThinking"
    >{{ thinking }}</div>

    <div v-else-if="agent.kind === 'tool'" class="task-log-agent__tool">
      <button
        type="button"
        class="task-log-agent__toolHead"
        :aria-expanded="agent.open ? 'true' : 'false'"
        data-testid="taskLog-agentTool"
        @click="$emit('toggle')"
      >
        <v-icon size="small" class="task-log-agent__chevron">
          {{ agent.open ? 'mdi-chevron-down' : 'mdi-chevron-right' }}
        </v-icon>

        <v-icon
          size="small"
          class="task-log-agent__state"
          :class="`task-log-agent__state--${toolState}`"
          :title="$t(`agentLogTool_${toolState}`)"
          data-testid="taskLog-agentToolState"
          :data-state="toolState"
        >{{ toolStateIcon }}</v-icon>

        <!-- The space each name ends with keeps the words apart in a copy of the log. -->
        <span
          v-if="agent.subagent !== ''"
          class="task-log-agent__toolName"
        >{{ `${$t('agentLogSubagent')} ` }}</span>
        <template v-else>
          <span
            v-if="agent.server !== ''"
            class="task-log-agent__toolServer"
          >{{ `${agent.server} ` }}</span>
          <span class="task-log-agent__toolName">{{ `${agent.tool} ` }}</span>
        </template>

        <span class="task-log-agent__toolSummary">{{ agent.summary }}</span>
      </button>

      <div v-if="agent.open" class="task-log-agent__details" data-testid="taskLog-agentToolDetails">
        <template v-if="hasInput">
          <div class="task-log-agent__label">{{ $t('agentLogArguments') }}</div>
          <pre class="task-log-agent__value" data-testid="taskLog-agentToolInput">{{ input }}</pre>
          <div v-if="agent.truncated > 0" class="task-log-agent__note">
            {{ $t('agentLogTruncated', { n: agent.truncated }) }}
          </div>
        </template>

        <div class="task-log-agent__label">{{ $t('agentLogResult') }}</div>
        <pre
          v-if="agent.result !== ''"
          class="task-log-agent__value"
          :class="{ 'task-log-agent__value--failed': agent.state === 'failed' }"
          data-testid="taskLog-agentToolResult"
        >{{ result }}</pre>
        <div v-else class="task-log-agent__note" data-testid="taskLog-agentToolNoResult">
          {{ $t(`agentLogTool_${toolState}`) }}
        </div>
        <div v-if="agent.resultTruncated > 0" class="task-log-agent__note">
          {{ $t('agentLogTruncated', { n: agent.resultTruncated }) }}
        </div>
      </div>
    </div>

    <div
      v-else-if="agent.kind === 'result'"
      class="task-log-agent__details"
      data-testid="taskLog-agentResult"
    >
      <pre
        class="task-log-agent__value"
        :class="{ 'task-log-agent__value--failed': agent.failed }"
      >{{ orphanResult }}</pre>
      <div v-if="agent.truncated > 0" class="task-log-agent__note">
        {{ $t('agentLogTruncated', { n: agent.truncated }) }}
      </div>
    </div>

    <div
      v-else-if="agent.kind === 'init'"
      class="task-log-agent__init"
      data-testid="taskLog-agentInit"
    >
      <span v-if="agent.model !== ''">{{ $t('agentLogModel') }} <b>{{ agent.model }}</b></span>
      <span
        v-for="server in agent.servers"
        :key="server.name"
        class="task-log-agent__server"
        :class="{ 'task-log-agent__server--down': server.status !== 'connected' }"
        data-testid="taskLog-agentServer"
      >
        <v-icon size="x-small">{{ serverIcon(server) }}</v-icon>
        {{ server.name }}<template
          v-if="server.status !== 'connected'">: {{ server.status }}</template>
      </span>
    </div>

    <div
      v-else-if="agent.kind === 'retry'"
      class="task-log-agent__retry"
      data-testid="taskLog-agentRetry"
    >
      <v-icon size="small">mdi-refresh</v-icon>
      {{ $t('agentLogRetry', { attempt: agent.attempt, max: agent.maxRetries }) }}<template
        v-if="agent.text !== ''">: {{ agent.text }}</template>
    </div>

    <div
      v-else-if="agent.kind === 'done'"
      class="task-log-agent__done"
      :class="{ 'task-log-agent__done--failed': agent.failed }"
      data-testid="taskLog-agentDone"
    >
      <div v-for="(error, i) in agent.errors" :key="i">{{ error }}</div>
      <v-icon size="small">{{ doneIcon }}</v-icon>
      {{ agent.failed ? $t('agentLogFailed') : $t('agentLogFinished') }}<template
        v-if="agent.subtype !== ''"> ({{ agent.subtype }})</template><template
        v-for="(part, i) in took" :key="i"> · {{ part }}</template>
    </div>
  </div>
</template>

<style lang="scss">
.task-log-agent {
  width: 100%;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

// The rows of a subagent stand apart from the agent's own.
.task-log-agent--sub {
  padding-left: 10px;
  border-left: 2px solid #5c6bc0;
}

.task-log-agent__sub {
  float: right;
  max-width: 40%;
  margin-left: 10px;
  padding: 0 6px;
  overflow: hidden;
  border-radius: 4px;
  background: #1f2440;
  color: #b3bcf5;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
  user-select: none;
  -webkit-user-select: none;
}

.task-log-agent__code {
  padding: 0 3px;
  border-radius: 3px;
  background: #262626;
  color: #ffcc80;
  font-family: inherit;
}

.task-log-agent__thinking {
  color: #9e9e9e;
  font-style: italic;
}

.task-log-agent__toolHead {
  display: flex;
  align-items: baseline;
  max-width: 100%;
  gap: 6px;
  padding: 0;
  border: 0;
  background: none;
  color: inherit;
  font: inherit;
  text-align: left;
  white-space: nowrap;
  cursor: pointer;

  &:hover .task-log-agent__toolName {
    text-decoration: underline;
  }

  &:focus-visible {
    outline: 1px solid #4fc3f7;
    outline-offset: 2px;
  }

  .v-icon {
    align-self: center;
    flex: none;
  }
}

.task-log-agent__chevron {
  color: #757575;
}

.task-log-agent__state--running {
  color: #9e9e9e;
}

.task-log-agent__state--done {
  color: #66bb6a;
}

.task-log-agent__state--failed,
.task-log-agent__state--interrupted {
  color: #ef5350;
}

.task-log-agent__toolServer {
  flex: none;
  color: #9e9e9e;

  &::after {
    content: '›';
  }
}

.task-log-agent__toolName {
  flex: none;
  color: #4fc3f7;
}

.task-log-agent__toolSummary {
  overflow: hidden;
  color: #9e9e9e;
  text-overflow: ellipsis;
}

.task-log-agent__details {
  margin: 2px 0 6px 9px;
  padding: 4px 10px;
  border-left: 2px solid #333;
}

.task-log-agent__label {
  color: #757575;
  font-size: 12px;
  text-transform: uppercase;
  user-select: none;
  -webkit-user-select: none;
}

.task-log-agent__value {
  max-height: 320px;
  margin: 0 0 4px;
  overflow: auto;
  color: #e0e0e0;
  font: inherit;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.task-log-agent__value--failed {
  color: #ef9a9a;
}

.task-log-agent__note {
  margin-bottom: 4px;
  color: #757575;
  font-style: italic;
}

.task-log-agent__init {
  color: #9e9e9e;

  b {
    color: #e0e0e0;
  }
}

.task-log-agent__server {
  margin-left: 12px;
  color: #81c784;
}

.task-log-agent__server--down {
  color: #ef5350;
}

.task-log-agent__retry {
  color: #ffb74d;
}

.task-log-agent__done {
  color: #81c784;
}

.task-log-agent__done--failed {
  color: #ef5350;
}
</style>

<script>
import {
  formatAgentDuration, formatAgentTokens, inlineSegments, readable, readableResult,
} from '@/lib/agentLog';

const TOOL_STATE_ICONS = {
  running: 'mdi-loading mdi-spin',
  done: 'mdi-check',
  failed: 'mdi-close',
  interrupted: 'mdi-minus-circle-outline',
};

export default {
  props: {
    // The `agent` part of a row made by AgentLog.
    agent: {
      type: Object,
      required: true,
    },
    // Whether the task is still running: a tool call without a result is running in one that
    // is, and was interrupted in one that is not.
    running: {
      type: Boolean,
      default: true,
    },
  },

  emits: ['toggle'],

  computed: {
    segments() {
      return inlineSegments(this.agent.text);
    },

    toolState() {
      return this.agent.state === 'running' && !this.running ? 'interrupted' : this.agent.state;
    },

    toolStateIcon() {
      return TOOL_STATE_ICONS[this.toolState];
    },

    hasInput() {
      return this.agent.input !== undefined && this.agent.input !== null;
    },

    input() {
      return readable(this.agent.input);
    },

    result() {
      return readableResult(this.agent.result);
    },

    orphanResult() {
      return readableResult(this.agent.text);
    },

    // The reasoning of the agent, without the marks of its headings.
    thinking() {
      return inlineSegments(this.agent.text).map((segment) => segment.text).join('');
    },

    // What the run took, as far as the CLI told: turns, time, tokens, cost.
    took() {
      const { agent } = this;
      const parts = [];
      if (agent.turns > 0) {
        parts.push(this.$t('agentLogTurns', { n: agent.turns }));
      }
      if (agent.durationMs > 0) {
        parts.push(formatAgentDuration(agent.durationMs));
      }
      if (agent.inputTokens > 0 || agent.outputTokens > 0) {
        parts.push(this.$t('agentLogTokens', {
          input: formatAgentTokens(agent.inputTokens),
          output: formatAgentTokens(agent.outputTokens),
        }));
      }
      if (agent.costUsd > 0) {
        parts.push(`$${agent.costUsd.toFixed(2)}`);
      }
      return parts;
    },

    doneIcon() {
      return this.agent.failed ? 'mdi-close-circle-outline' : 'mdi-check-circle-outline';
    },
  },

  methods: {
    serverIcon(server) {
      return server.status === 'connected' ? 'mdi-check-circle-outline' : 'mdi-alert-circle-outline';
    },
  },
};
</script>
