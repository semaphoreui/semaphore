<template>
  <div class="AuditEventCard pa-4" data-testid="audit-card">
    <div class="d-flex align-start">
      <div class="flex-grow-1">
        <div class="text-h6">{{ title }}</div>
        <div class="text-caption text--secondary AuditEventCard__kind">{{ kind }}</div>
      </div>
      <CopyClipboardButton
        :text="json"
        :success-message="$t('audit_json_copied')"
        data-testid="audit-copy-json"
      />
      <v-btn icon :aria-label="$t('close')" @click="$emit('close')">
        <v-icon>mdi-close</v-icon>
      </v-btn>
    </div>

    <v-alert
      v-if="event.outcome === 'failure' || event.reason"
      dense
      text
      :type="event.outcome === 'failure' ? 'error' : 'warning'"
      class="mt-3 mb-0"
    >
      {{ event.outcome === 'failure' ? $t('audit_result_failure') : $t('audit_result_success') }}
      <span v-if="event.reason">— {{ reason }}</span>
    </v-alert>

    <div v-for="section in sections" :key="section.key" class="mt-4">
      <div class="text-overline">{{ $t(section.key) }}</div>
      <table class="AuditEventCard__fields">
        <tr v-for="row in section.rows" :key="row.label">
          <th>{{ row.label }}</th>
          <td>
            <a
              v-if="row.pivot && pivots"
              href="#"
              :data-testid="`audit-pivot-${row.pivot.name}`"
              @click.prevent="$emit('pivot', row.pivot.filter)"
            >{{ row.value }}</a>
            <span v-else-if="row.pivot" :title="$t('audit_only_pro_short')">
              {{ row.value }}
            </span>
            <span v-else>{{ row.value }}</span>
          </td>
        </tr>
      </table>
    </div>
  </div>
</template>
<script>
import dayjs from 'dayjs';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';
import utc from 'dayjs/plugin/utc';
import {
  eventKind, eventTitle, reasonLabel, targetTypeLabel, actorLabel, metadataRows,
} from '@/lib/auditEvent';

dayjs.extend(utc);

export default {
  components: { CopyClipboardButton },

  props: {
    event: Object,
    projectName: String,
    pivots: Boolean,
  },

  computed: {
    json() {
      return JSON.stringify(this.event, null, 2);
    },

    kind() {
      return eventKind(this.event);
    },

    title() {
      return eventTitle(this.event, this.$i18n);
    },

    reason() {
      return reasonLabel(this.event.reason, this.$i18n);
    },

    sections() {
      const e = this.event;
      const actor = e.actor || {};
      const target = e.target;
      const source = e.source;
      const projectId = e.scope && e.scope.project_id;
      const sections = [{
        key: 'audit_card_who',
        rows: [
          {
            label: this.$t('audit_field_name'),
            value: actorLabel(actor, this.$i18n),
            pivot: actor.type === 'user' && actor.id
              ? { name: 'user', filter: { user: Number(actor.id) } } : null,
          },
          { label: this.$t('audit_field_type'), value: this.$t(`audit_actor_type_${actor.type}`) },
          actor.auth && { label: this.$t('audit_field_auth'), value: this.$t(`audit_auth_${actor.auth}`) },
          actor.token_fingerprint && { label: this.$t('audit_field_token'), value: actor.token_fingerprint },
        ],
      }];
      if (target) {
        sections.push({
          key: 'audit_card_object',
          rows: [
            {
              label: this.$t('audit_field_name'),
              value: target.name || target.id,
              pivot: { name: 'object', filter: { target_type: target.type, target_id: target.id } },
            },
            { label: this.$t('audit_field_type'), value: targetTypeLabel(target.type, this.$i18n) },
            target.id && { label: this.$t('audit_field_id'), value: target.id },
          ],
        });
      }
      sections.push({
        key: 'audit_card_where',
        rows: [
          projectId && {
            label: this.$t('audit_field_project'),
            value: this.projectName,
            pivot: { name: 'project', filter: { project: Number(projectId) } },
          },
          source && {
            label: this.$t('audit_field_ip'),
            value: source.ip,
            pivot: { name: 'ip', filter: { ip: source.ip } },
          },
          source && source.user_agent && { label: this.$t('audit_field_user_agent'), value: source.user_agent },
        ],
      }, {
        key: 'audit_card_when',
        rows: [
          { label: this.$t('audit_field_local_time'), value: dayjs(e.timestamp).format('YYYY-MM-DD HH:mm:ss') },
          { label: this.$t('audit_field_utc'), value: dayjs.utc(e.timestamp).format('YYYY-MM-DD HH:mm:ss') },
        ],
      });
      const details = metadataRows(e.metadata).map((row) => ({ label: row.key, value: row.value }));
      if (details.length > 0) {
        sections.push({ key: 'audit_card_details', rows: details });
      }
      sections.push({
        key: 'audit_card_technical',
        rows: [
          { label: this.$t('audit_field_event_id'), value: e.event_id },
          { label: this.$t('audit_field_seq'), value: String(e.seq) },
          e.request_id && { label: this.$t('audit_field_request_id'), value: e.request_id },
          { label: this.$t('audit_field_instance'), value: e.instance_id },
          e.node_id && { label: this.$t('audit_field_node'), value: e.node_id },
        ],
      });
      return sections.map((s) => ({ ...s, rows: s.rows.filter(Boolean) }))
        .filter((s) => s.rows.length > 0);
    },
  },

};
</script>
<style lang="scss">
.AuditEventCard__kind {
  font-family: monospace;
}

.AuditEventCard__fields {
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
</style>
