<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div class="pb-3">
    <v-row>
      <v-col cols="12" md="6">
        <v-card
          v-if="template"
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
        >
          <v-card-title>Template info</v-card-title>
          <v-card-text>
            <v-simple-table class="TaskDetails__table">
              <template v-slot:default>
                <tbody>
                <tr>
                  <td><b>App</b></td>
                  <td>{{ getAppTitle(template.app) }}</td>
                </tr>
                <tr>
                  <td><b>Template</b></td>
                  <td>
                    <RouterLink :to="`/project/${projectId}/templates/${template.id}`">
                      {{ template.name }}
                    </RouterLink>
                  </td>
                </tr>
                </tbody>
              </template>
            </v-simple-table>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="6">
        <v-card
          v-if="item.commit_hash"
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
        >
          <v-card-title>Commit info</v-card-title>

          <v-card-text>
            <v-simple-table class="TaskDetails__table">
              <template v-slot:default>
                <tbody>
                <tr>
                  <td><b>Message</b></td>
                  <td>{{ item.commit_message }}</td>
                </tr>
                <tr>
                  <td><b>Hash</b></td>
                  <td>{{ item.commit_hash }}</td>
                </tr>
                </tbody>
              </template>
            </v-simple-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12" md="6">
        <v-card
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
          class="mb-5"
        >
          <v-card-title>Running info</v-card-title>
          <v-card-text>
            <v-simple-table class="pa-0 TaskDetails__table">
              <template v-slot:default>
                <tbody>
                <tr>
                  <td><b>Message</b></td>
                  <td>{{ item.message || '—' }}</td>
                </tr>
                <tr v-if="item.user_id != null">
                  <td><b>{{ $t('author') }}</b></td>
                  <td>{{ user?.name || '—' }}</td>
                </tr>
                <tr v-else-if="item.integration_id != null">
                  <td><b>{{ $t('integration') }}</b></td>
                  <td>
                    <router-link
                      v-if="isReady(integration)"
                      :to="`/project/${projectId}/integrations/${item.integration_id}`"
                    >{{ integration.data.name }}</router-link>
                    <span v-else>{{ originLabel(integration, item.integration_id) }}</span>
                  </td>
                </tr>
                <tr v-else-if="item.schedule_id != null">
                  <td><b>{{ $t('schedule') }}</b></td>
                  <td>
                    <!-- Schedules have no detail page, so link to the list. -->
                    <router-link
                      v-if="isReady(schedule)"
                      :to="`/project/${projectId}/schedule`"
                    >{{ schedule.data.name || $t('unnamedSchedule') }}</router-link>
                    <span v-else>{{ originLabel(schedule, item.schedule_id) }}</span>
                  </td>
                </tr>
                <tr>
                  <td><b>{{ $t('created') }}</b></td>
                  <td>{{ item.created | formatDate }}</td>
                </tr>
                <tr>
                  <td><b>{{ $t('started') }}</b></td>
                  <td>{{ item.start | formatDate }}</td>
                </tr>
                <tr>
                  <td><b>{{ $t('end') }}</b></td>
                  <td>{{ item.end | formatDate }}</td>
                </tr>
                <tr>
                  <td><b>{{ $t('duration') }}</b></td>
                  <td>{{ [item.start, item.end] | formatMilliseconds }}</td>
                </tr>
                <tr v-if="item.used_runner_name">
                  <td><b>Runner</b></td>
                  <td>{{ item.used_runner_name }}</td>
                </tr>
                </tbody>
              </template>
            </v-simple-table>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card
          v-if="item?.params"
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
          class="mb-5"
        >
          <v-card-title>Task parameters</v-card-title>
          <v-card-text>
            <v-simple-table class="pa-0 TaskDetails__table">
              <template v-slot:default>
                <tbody>
                <tr>
                  <td><b>Branch</b></td>
                  <td>
                    {{ item.get_branch || '—' }}
                  </td>
                </tr>
                <tr>
                  <td><b>Limit</b></td>
                  <td>
                    <span v-if="Array.isArray(item.params.limit) && item.params.limit.length > 0">
                      {{ item.params.limit.join(', ') }}</span>
                    <span v-else>'No'</span>
                  </td>
                </tr>
                <tr>
                  <td><b>Debug</b></td>
                  <td>
                    {{ item.params.debug ? 'Yes' : 'No' }}
                  </td>
                </tr>
                <tr>
                  <td><b>Debug level</b></td>
                  <td>{{ item.params.debug_level || '—' }}</td>
                </tr>
                <tr>
                  <td><b>Diff</b> <code>--diff</code></td>
                  <td>{{ item.params.diff ? 'Yes' : 'No' }}</td>
                </tr>
                <tr>
                  <td><b>Dry run</b> <code>--check</code></td>
                  <td>{{ item.params.dry_run ? 'Yes' : 'No' }}</td>
                </tr>
                <tr>
                  <td><b>Environment</b></td>
                  <td>
                    {{ !item.environment || item.environment === '{}' ? '—' : item.environment }}
                  </td>
                </tr>
                </tbody>
              </template>
            </v-simple-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row v-if="outputs">
      <v-col cols="12">
        <v-card
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
          class="mb-5"
        >
          <v-card-title>
            {{ $t('workflowOutputs') }}
            <v-tooltip bottom max-width="360">
              <template v-slot:activator="{ on, attrs }">
                <v-icon small class="ml-2" v-bind="attrs" v-on="on">
                  mdi-information-outline
                </v-icon>
              </template>
              <span>{{ $t('workflowOutputsHint') }}</span>
            </v-tooltip>
          </v-card-title>
          <v-card-text>
            <v-simple-table v-if="outputs.values.length" dense class="TaskDetails__table">
              <tbody>
                <tr v-for="row in outputs.values" :key="row.name">
                  <td class="TaskDetails__outputName">{{ row.name }}</td>
                  <td><pre class="TaskDetails__artifacts">{{ row.value }}</pre></td>
                </tr>
              </tbody>
            </v-simple-table>
            <div v-if="outputs.skipped.length" class="mt-3">
              <div class="text-caption text--secondary mb-1">
                {{ $t('workflowOutputsSkipped') }}
              </div>
              <v-chip
                v-for="row in outputs.skipped"
                :key="`skipped-${row.name}`"
                x-small
                label
                outlined
                class="mr-1 mb-1"
                :title="row.reason"
              >
                <v-icon x-small left>mdi-eye-off-outline</v-icon>
                {{ row.name }} — {{ row.reason }}
              </v-chip>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<style lang="scss">
.TaskDetails__table {
  background-color: transparent !important;
  .v-data-table__wrapper {
    padding-left: 0 !important;
    padding-right: 0 !important;
  }
}

.TaskDetails__artifacts {
  white-space: pre-wrap;
  word-break: break-word;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  margin: 0;
}

.TaskDetails__outputName {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  white-space: nowrap;
  width: 1%;
}

</style>

<script>

import ProjectMixin from '@/components/ProjectMixin';
import AppsMixin from '@/components/AppsMixin';

export default {
  props: {
    item: Object,
    user: Object,
    // { status: 'loading' | 'ready' | 'missing' | 'error', data } for the
    // schedule or integration that started the task; null when a person did.
    schedule: Object,
    integration: Object,
    projectId: Number,
  },

  mixins: [ProjectMixin, AppsMixin],

  data() {
    return {
      template: null,
    };
  },

  watch: {
    async item() {
      if (this.item?.template_id !== this.template?.id) {
        await this.loadData();
      }
    },
  },

  computed: {
    // The outputs document of a workflow task: {values, skipped}. A document
    // written before that shape existed is a flat map and is shown as values.
    outputs() {
      const raw = this.item?.artifacts;
      if (raw == null || raw === '') return null;
      let doc = raw;
      if (typeof raw === 'string') {
        try {
          doc = JSON.parse(raw);
        } catch (e) {
          return null;
        }
      }
      if (!doc || typeof doc !== 'object' || Array.isArray(doc)) return null;
      const hasShape = doc.values && typeof doc.values === 'object';
      const values = hasShape ? doc.values : doc;
      const skipped = hasShape && doc.skipped && typeof doc.skipped === 'object' ? doc.skipped : {};
      const rows = Object.keys(values).sort().map((name) => ({
        name,
        value: typeof values[name] === 'string' ? values[name] : JSON.stringify(values[name]),
      }));
      const skippedRows = Object.keys(skipped).sort().map((name) => ({
        name,
        reason: this.skipReason(skipped[name]),
      }));
      if (rows.length === 0 && skippedRows.length === 0) return null;
      return { values: rows, skipped: skippedRows };
    },
  },

  async created() {
    await this.loadData();
  },

  methods: {
    skipReason(reason) {
      const key = `workflowOutputSkipped_${reason}`;
      return this.$te(key) ? this.$t(key) : reason;
    },

    isReady(origin) {
      return origin != null && origin.status === 'ready';
    },

    // Only a confirmed 404 says "deleted"; while loading or after an unrelated
    // failure the id alone is all we can honestly show.
    originLabel(origin, id) {
      if (origin != null && origin.status === 'missing') {
        return this.$t('deletedOrigin', { id });
      }
      return `#${id}`;
    },

    async loadData() {
      this.template = await this.loadProjectResource('templates', this.item.template_id);
    },
  },
};
</script>
