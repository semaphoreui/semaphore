<template>
  <div class="IntegrationCallCard pa-4">
    <div class="d-flex align-start">
      <div class="flex-grow-1">
        <div class="text-h6">{{ call.integration_name || call.alias }}</div>
        <div class="text-caption text--secondary IntegrationCallCard__mono">
          {{ call.method }} /api/integrations/{{ call.alias }}
        </div>
      </div>
      <CopyClipboardButton :text="json" :success-message="$t('integrationCallCopied')" />
      <v-btn icon :aria-label="$t('close')" @click="$emit('close')">
        <v-icon>mdi-close</v-icon>
      </v-btn>
    </div>

    <v-alert
      v-if="call.outcome !== 'ok'"
      dense
      text
      :type="call.outcome === 'skipped' ? 'info' : call.outcome === 'error' ? 'error' : 'warning'"
      class="mt-3 mb-0"
    >
      {{ call.outcome }}<span v-if="call.reason"> — {{ call.reason }}</span>
    </v-alert>

    <div v-for="section in sections" :key="section.key" class="mt-4">
      <div class="text-overline">{{ $t(section.key) }}</div>
      <table class="IntegrationCallCard__fields">
        <tr v-for="row in section.rows" :key="row.label">
          <th>{{ row.label }}</th>
          <td>
            <router-link v-if="row.to" :to="row.to">{{ row.value }}</router-link>
            <TaskLink v-else-if="row.taskId" :label="row.value" :task-id="row.taskId" />
            <span v-else>{{ row.value }}</span>
          </td>
        </tr>
      </table>
    </div>

    <div class="mt-4">
      <div class="d-flex align-center">
        <div class="text-overline flex-grow-1">{{ $t('integrationCallBody') }}</div>
        <CopyClipboardButton
          v-if="call.request_body"
          :text="call.request_body"
          :success-message="$t('integrationCallBodyCopied')"
        />
      </div>

      <v-progress-linear v-if="loading" indeterminate />

      <template v-else>
        <div v-if="truncated" class="text-caption text--secondary mb-1">
          {{ $t('integrationCallBodyTruncated', { stored: storedSize, size: call.request_size }) }}
        </div>

        <pre v-if="call.request_body" class="IntegrationCallCard__body">{{ body }}</pre>
        <div v-else class="text--secondary">{{ $t('integrationCallEmptyBody') }}</div>
      </template>
    </div>
  </div>
</template>
<script>
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';
import TaskLink from '@/components/TaskLink.vue';

dayjs.extend(utc);

export default {
  components: { CopyClipboardButton, TaskLink },

  props: {
    call: Object,
    projectId: Number,
    // The list leaves bodies out, so the card first shows the row it was opened
    // from and the body arrives with the single call; a missing body means
    // "not loaded yet" until then, not "empty".
    loading: Boolean,
  },

  computed: {
    json() {
      return JSON.stringify(this.call, null, 2);
    },

    // Shown as JSON when it is JSON, which most webhook bodies are; anything
    // else is shown as received.
    body() {
      try {
        return JSON.stringify(JSON.parse(this.call.request_body), null, 2);
      } catch {
        return this.call.request_body;
      }
    },

    // request_size is what arrived and request_body what was kept, capped by
    // the server, so a gap between the two means the body was cut.
    storedSize() {
      return new TextEncoder().encode(this.call.request_body || '').length;
    },

    truncated() {
      return this.call.request_size > this.storedSize;
    },

    sections() {
      const c = this.call;
      const p = this.projectId;

      let result;
      if (c.run_id && c.workflow_id) {
        result = {
          label: this.$t('run'),
          value: `#${c.run_id}`,
          to: `/project/${p}/workflows/${c.workflow_id}/runs/${c.run_id}`,
        };
      } else if (c.run_id) {
        result = { label: this.$t('run'), value: `#${c.run_id}` };
      } else if (c.task_id) {
        result = { label: this.$t('task2'), value: `#${c.task_id}`, taskId: c.task_id };
      }

      return [{
        key: 'integrationCallCaller',
        rows: [
          { label: this.$t('sourceIp'), value: c.source_ip },
          { label: this.$t('method'), value: c.method },
        ],
      }, {
        key: 'integration',
        rows: [
          {
            label: this.$t('name'),
            value: c.integration_name,
            to: c.integration_id ? `/project/${p}/integrations/${c.integration_id}` : null,
          },
          { label: this.$t('alias'), value: c.alias },
        ],
      }, {
        key: 'integrationCallWhen',
        rows: [
          { label: this.$t('integrationCallLocalTime'), value: dayjs(c.created).format('YYYY-MM-DD HH:mm:ss.SSS Z') },
          { label: this.$t('integrationCallUtc'), value: dayjs.utc(c.created).format('YYYY-MM-DD HH:mm:ss.SSS Z') },
        ],
      }, {
        key: 'integrationCallResult',
        rows: [
          { label: this.$t('status'), value: c.outcome },
          result,
          { label: this.$t('duration'), value: `${c.duration} ms` },
        ],
      }, {
        key: 'integrationCallTechnical',
        rows: [
          { label: this.$t('integrationCallId'), value: String(c.id) },
          { label: this.$t('integrationCallSize'), value: `${c.request_size} B` },
        ],
      }]
        // A call that matched no integration has no name; a row with a label and
        // nothing after it reads as broken, so empty rows are left out.
        .map((section) => ({
          ...section,
          rows: section.rows.filter((row) => row && row.value != null && row.value !== ''),
        }))
        .filter((section) => section.rows.length > 0);
    },
  },
};
</script>
<style lang="scss">
.IntegrationCallCard__mono {
  font-family: monospace;
  word-break: break-all;
}

.IntegrationCallCard__fields {
  width: 100%;
  border-collapse: collapse;

  th {
    width: 40%;
    padding: 2px 8px 2px 0;
    text-align: left;
    vertical-align: top;
    font-weight: normal;
    opacity: 0.7;
  }

  td {
    padding: 2px 0;
    word-break: break-word;
  }
}

.IntegrationCallCard__body {
  max-height: 50vh;
  overflow: auto;
  padding: 8px;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
  background: rgba(127, 127, 127, 0.12);
  border-radius: 4px;
}
</style>
