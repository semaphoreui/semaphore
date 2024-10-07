<template>
  <div v-if="!isLoaded">
    <v-row>
      <v-col>
        <v-skeleton-loader
            type="table-heading, list-item-two-line, image, table-tfoot"
        ></v-skeleton-loader>
      </v-col>
      <v-col>
        <v-skeleton-loader
            type="table-heading, list-item-two-line, image, table-tfoot"
        ></v-skeleton-loader>
      </v-col>
    </v-row>
  </div>
  <v-form
    v-else
    ref="form"
    lazy-validation
    v-model="formValid"
  >
    <v-dialog
      v-model="helpDialog"
      hide-overlay
      width="300"
    >
      <v-alert
        border="top"
        colored-border
        type="info"
        elevation="2"
        class="mb-0 pb-0"
      >
        <div v-if="helpKey === 'build_version'">
          <p>
            {{ $t('definesStartVersionOfYourArtifactEachRunIncrements') }}
          </p>
          <p>
            {{ $t('forMoreInformationAboutBuildingSeeThe') }}
            <a href="https://docs.ansible-semaphore.com/user-guide/task-templates#build"
               target="_blank"
            >{{ $t('taskTemplateReference') }}</a>.
          </p>
        </div>
        <div v-else-if="helpKey === 'build'">
          <p>
            {{ $t('definesWhatArtifactShouldBeDeployedWhenTheTaskRun') }}
          </p>
          <p>
            {{ $t('forMoreInformationAboutDeployingSeeThe') }}
            <a href="https://docs.ansible-semaphore.com/user-guide/task-templates#build"
               target="_blank"
            >{{ $t('taskTemplateReference2') }}</a>.
          </p>
        </div>
        <div v-if="helpKey === 'cron'">
          <p>{{ $t('definesAutorunSchedule') }}</p>
          <p>
            {{ $t('forMoreInformationAboutCronSeeThe') }}
            <a href="https://pkg.go.dev/github.com/robfig/cron/v3#hdr-CRON_Expression_Format"
               target="_blank"
            >{{ $t('cronExpressionFormatReference') }}</a>.
          </p>
        </div>
      </v-alert>
    </v-dialog>

    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-row>
      <v-col cols="12" md="6" class="pb-0">
        <v-card class="mb-6" :color="$vuetify.theme.dark ? '#212121' : 'white'">
          <v-tabs
            fixed-tabs
            v-model="itemTypeIndex"
          >
            <v-tab
              style="padding: 0"
              v-for="(key) in Object.keys(TEMPLATE_TYPE_ICONS)"
              :key="key"
            >
              <v-icon small class="mr-2">{{ TEMPLATE_TYPE_ICONS[key] }}</v-icon>
              {{ $t(TEMPLATE_TYPE_TITLES[key]) }}
            </v-tab>
          </v-tabs>

          <div class="ml-4 mr-4 mt-6" v-if="item.type">
            <v-text-field
              v-if="item.type === 'build'"
              v-model="item.start_version"
              :label="$t('startVersion')"
              :rules="[v => !!v || $t('start_version_required')]"
              required
              :disabled="formSaving"
              :placeholder="$t('example000')"
              append-outer-icon="mdi-help-circle"
              @click:append-outer="showHelpDialog('build_version')"
            ></v-text-field>

            <v-autocomplete
              v-if="item.type === 'deploy'"
              v-model="item.build_template_id"
              :label="$t('buildTemplate')"
              :items="buildTemplates"
              item-value="id"
              item-text="name"
              :rules="[v => !!v || $t('build_template_required')]"
              required
              :disabled="formSaving"
              append-outer-icon="mdi-help-circle"
              @click:append-outer="showHelpDialog('build')"
            ></v-autocomplete>

            <v-checkbox
              v-if="item.type === 'deploy'"
              class="mt-0"
              :label="$t('autorun')"
              v-model="item.autorun"
            />
          </div>

        </v-card>

        <v-text-field
          v-model="item.name"
          :label="$t('name2')"
          :rules="[v => !!v || $t('name_required')]"
          outlined
          dense
          required
          :disabled="formSaving"
        ></v-text-field>

        <v-textarea
          v-model="item.description"
          :label="$t('description')"
          :disabled="formSaving"
          rows="1"
          :auto-grow="true"
          outlined
          dense
        ></v-textarea>

        <v-text-field
          v-model="item.playbook"
          :label="fieldLabel('playbook')"
          :rules="isFieldRequired('playbook') ? [v => !!v || $t('playbook_filename_required')] : []"
          outlined
          dense
          :required="isFieldRequired('playbook')"
          :disabled="formSaving"
          :placeholder="$t('exampleSiteyml')"
          v-if="needField('playbook')"
        ></v-text-field>

        <v-select
          v-model="item.inventory_id"
          :label="fieldLabel('inventory')"
          :items="inventory"
          item-value="id"
          item-text="name"
          outlined
          dense
          required
          :disabled="formSaving"
          v-if="needField('inventory')"
        ></v-select>

        <v-select
          v-model="item.repository_id"
          :label="fieldLabel('repository') + ' *'"
          :items="repositories"
          item-value="id"
          item-text="name"
          :rules="isFieldRequired('repository') ? [v => !!v || $t('repository_required')] : []"
          outlined
          dense
          :required="isFieldRequired('repository')"
          :disabled="formSaving"
          v-if="needField('repository')"
        ></v-select>

        <v-select
          v-model="item.environment_id"
          :label="fieldLabel('environment')"
          :items="environment"
          item-value="id"
          item-text="name"
          :rules="isFieldRequired('environment') ? [v => !!v || $t('environment_required')] : []"
          outlined
          dense
          :required="isFieldRequired('environment')"
          :disabled="formSaving"
          v-if="needField('environment')"
        ></v-select>

        <TemplateVaults
          v-if="itemTypeIndex === 0 && needField('vault')"
          :project-id="this.projectId"
          :vaults="item.vaults"
          @change="setTemplateVaults"
        ></TemplateVaults>

      </v-col>

      <v-col cols="12" md="6" class="pb-0">

        <TemplateVaults
          v-if="itemTypeIndex > 0 && needField('vault')"
          :project-id="this.projectId"
          :vaults="item.vaults"
          @change="setTemplateVaults"
        ></TemplateVaults>

        <SurveyVars style="margin-top: -10px;" :vars="item.survey_vars" @change="setSurveyVars"/>

        <v-select
          v-model="item.view_id"
          :label="$t('view')"
          clearable
          :items="views"
          item-value="id"
          item-text="title"
          :disabled="formSaving"
          outlined
          dense
        ></v-select>

        <v-checkbox
          class="mt-0"
          :label="$t('iWantToRunATaskByTheCronOnlyForForNewCommitsOfSome')"
          v-model="cronVisible"
        />

        <v-select
          v-if="cronVisible"
          v-model="cronRepositoryId"
          :label="$t('repository2')"
          :placeholder="$t('cronChecksNewCommitBeforeRun')"
          :rules="[v => !!v || $t('repository_required')]"
          :items="repositories"
          item-value="id"
          item-text="name"
          clearable
          :disabled="formSaving"
          outlined
          dense
        ></v-select>

        <v-select
          v-if="cronVisible"
          v-model="cronFormat"
          :label="$t('Check interval')"
          :hint="$t('New commit check interval')"
          item-value="cron"
          item-text="title"
          :items="cronFormats"
          :disabled="formSaving"
          outlined
          dense
        />

        <v-checkbox
          class="mt-0"
          :label="$t('suppressSuccessAlerts')"
          v-model="item.suppress_success_alerts"
        />

        <ArgsPicker
          :vars="args"
          @change="setArgs"
        />

        <v-checkbox
          class="mt-0"
          :label="$t('allowCliArgsInTask')"
          v-model="item.allow_override_args_in_task"
        />

      </v-col>
    </v-row>
  </v-form>
</template>
<style lang="scss">
.CodeMirror-placeholder {
  color: #a4a4a4 !important;
}
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import axios from 'axios';

import ItemFormBase from '@/components/ItemFormBase';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
import 'codemirror/addon/lint/json-lint.js';
import 'codemirror/addon/display/placeholder.js';
import ArgsPicker from '@/components/ArgsPicker.vue';
import TemplateVaults from '@/components/TemplateVaults.vue';
import { TEMPLATE_TYPE_ICONS, TEMPLATE_TYPE_TITLES } from '../lib/constants';
import SurveyVars from './SurveyVars';

export default {
  mixins: [ItemFormBase],

  components: {
    TemplateVaults,
    ArgsPicker,
    SurveyVars,
  },

  props: {
    sourceItemId: Number,
    fields: Object,
    app: String,
  },

  data() {
    return {
      cronFormats: [{
        cron: '* * * * *',
        title: '1 minute',
      }, {
        cron: '*/5 * * * *',
        title: '5 minutes',
      }, {
        cron: '*/10 * * * *',
        title: '10 minutes',
      }, {
        cron: '@hourly',
        title: '1 hour',
      }, {
        cron: '@daily',
        title: '24 hours',
      }],
      itemTypeIndex: 0,
      TEMPLATE_TYPE_ICONS,
      TEMPLATE_TYPE_TITLES,
      cmOptions: {
        tabSize: 2,
        mode: 'application/json',
        lineNumbers: true,
        line: true,
        lint: true,
        indentWithTabs: false,
      },
      item: null,
      inventory: null,
      repositories: null,
      environment: null,
      views: null,
      schedules: null,
      buildTemplates: null,
      cronFormat: '* * * * *',
      cronRepositoryId: null,
      cronVisible: false,

      helpDialog: null,
      helpKey: null,

      advancedOptions: false,
      args: [],
    };
  },

  watch: {
    needReset(val) {
      if (val) {
        if (this.item != null) {
          this.item.template_id = this.templateId;
        }
      }
    },

    sourceItemId(val) {
      this.item.template_id = val;
    },

    itemTypeIndex(val) {
      this.item.type = Object.keys(TEMPLATE_TYPE_ICONS)[val];
    },
  },

  computed: {
    isLoaded() {
      if (this.isNew && this.sourceItemId == null) {
        return true;
      }

      return this.repositories != null
        && this.inventory != null
        && this.environment != null
        && this.item != null
        && this.schedules != null
        && this.views != null;
    },
  },

  methods: {
    setArgs(args) {
      this.args = args;
    },

    fieldLabel(f) {
      return this.$t((this.fields[f] || { label: f }).label);
    },

    needField(f) {
      return this.fields[f] != null;
    },

    isFieldRequired(f) {
      return this.fields[f] != null && !this.fields[f].optional;
    },

    setSurveyVars(v) {
      this.item.survey_vars = v;
    },

    setTemplateVaults(v) {
      this.item.vaults = v;
    },

    showHelpDialog(key) {
      this.helpKey = key;
      this.helpDialog = true;
    },

    async afterLoadData() {
      if (this.sourceItemId) {
        this.item = (await axios({
          keys: 'get',
          url: `/api/project/${this.projectId}/templates/${this.sourceItemId}`,
          responseType: 'json',
        })).data;
        this.item.id = null;
      }

      this.advancedOptions = this.item.arguments != null || this.item.allow_override_args_in_task;

      this.repositories = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/repositories`,
        responseType: 'json',
      })).data;

      this.inventory = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/inventory`,
        responseType: 'json',
      })).data;

      this.environment = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/environment`,
        responseType: 'json',
      })).data;

      const template = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/templates`,
        responseType: 'json',
      })).data;
      const builds = [];
      const deploys = [];
      template.forEach((t) => {
        switch (t.type) {
          case 'build':
            if (builds.length === 0) {
              builds.push({ header: 'Build Templates' });
            }
            builds.push(t);
            break;
          case 'deploy':
            if (deploys.length === 0) {
              deploys.push({ header: 'Deploy Templates' });
            }
            deploys.push(t);
            break;
          default:
            break;
        }

        this.args = JSON.parse(this.item.arguments || '[]');
      });

      this.buildTemplates = builds;
      if (this.buildTemplates.length > 0 && deploys.length > 0) {
        this.buildTemplates.push({ divider: true });
      }
      this.buildTemplates.push(...deploys);

      this.schedules = this.isNew ? [] : (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/templates/${this.itemId}/schedules`,
        responseType: 'json',
      })).data;

      this.views = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/views`,
        responseType: 'json',
      })).data;

      if (this.schedules.length > 0) {
        const schedule = this.schedules.find((s) => s.repository_id != null);
        if (schedule != null) {
          this.cronFormat = schedule.cron_format;
          this.cronRepositoryId = schedule.repository_id;
          this.cronVisible = this.cronRepositoryId != null;
        }
      }

      this.itemTypeIndex = Object.keys(TEMPLATE_TYPE_ICONS).indexOf(this.item.type);
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/templates`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/templates/${this.itemId}`;
    },

    async beforeSave() {
      if (this.cronFormat == null || this.cronFormat === '') {
        return;
      }

      await axios({
        method: 'post',
        url: `/api/project/${this.projectId}/schedules/validate`,
        responseType: 'json',
        data: {
          cron_format: this.cronFormat,
        },
      });

      this.item.app = this.app;

      this.item.arguments = JSON.stringify(this.args);
    },

    async afterSave(newItem) {
      if (newItem || this.schedules.length === 0) {
        if (this.cronFormat != null && this.cronFormat !== '' && this.cronVisible) {
          // new schedule
          await axios({
            method: 'post',
            url: `/api/project/${this.projectId}/schedules`,
            responseType: 'json',
            data: {
              project_id: this.projectId,
              template_id: newItem ? newItem.id : this.itemId,
              cron_format: this.cronFormat,
              repository_id: this.cronRepositoryId,
              active: true,
            },
          });
        }
      } else if (this.schedules.length > 1) {
        // do nothing
      } else if (this.cronFormat == null || this.cronFormat === '' || !this.cronVisible) {
        // drop schedule
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/schedules/${this.schedules[0].id}`,
          responseType: 'json',
        });
      } else {
        // update schedule
        await axios({
          method: 'put',
          url: `/api/project/${this.projectId}/schedules/${this.schedules[0].id}`,
          responseType: 'json',
          data: {
            id: this.schedules[0].id,
            project_id: this.projectId,
            template_id: this.itemId,
            cron_format: this.cronFormat,
            repository_id: this.cronRepositoryId,
          },
        });
      }
    },
  },
};
</script>
