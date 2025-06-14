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
                  <td>{{ item.integration_id }}</td>
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
                </tbody>
              </template>
            </v-simple-table>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card v-if="item?.params">
          <v-card-title>Task parameters</v-card-title>
          <v-card-text>
            <v-simple-table class="pa-0 TaskDetails__table">
              <template v-slot:default>
                <tbody>
                <tr>
                  <td><b>Limit</b></td>
                  <td>{{ item.params.limit }}</td>
                </tr>
                <tr>
                  <td><b>Debug</b></td>
                  <td>{{ item.params.debug }}</td>
                </tr>
                <tr>
                  <td><b>Debug level</b></td>
                  <td>{{ item.params.debug_level }}</td>
                </tr>
                <tr>
                  <td><b>Diff</b> <code>--diff</code></td>
                  <td>{{ item.params.diff }}</td>
                </tr>
                <tr>
                  <td><b>Dry run</b> <code>--check</code></td>
                  <td>{{ item.params.dry_run }}</td>
                </tr>
                <tr>
                  <td><b>Environment</b></td>
                  <td>{{ item.environment }}</td>
                </tr>
                </tbody>
              </template>
            </v-simple-table>
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

</style>

<script>

import ProjectMixin from '@/components/ProjectMixin';
import AppsMixin from '@/components/AppsMixin';

export default {
  props: {
    item: Object,
    user: Object,
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

  computed: {},

  async created() {
    await this.loadData();
  },

  methods: {
    async loadData() {
      this.template = await this.loadProjectResource('templates', this.item.template_id);
    },
  },
};
</script>
