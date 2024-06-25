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
          :label="$t('playbookFilename')"
          :rules="[v => !!v || $t('playbook_filename_required')]"
          outlined
          dense
          required
          :disabled="formSaving"
          :placeholder="$t('exampleSiteyml')"
        ></v-text-field>

        <v-select
          v-model="item.inventory_id"
          :label="$t('inventory2')"
          :items="inventory"
          item-value="id"
          item-text="name"
          outlined
          dense
          required
          :disabled="formSaving"
        ></v-select>

        <v-select
          v-model="item.repository_id"
          :label="$t('repository') + ' *'"
          :items="repositories"
          item-value="id"
          item-text="name"
          :rules="[v => !!v || $t('repository_required')]"
          outlined
          dense
          required
          :disabled="formSaving"
        ></v-select>

        <v-select
          v-model="item.environment_id"
          :label="$t('environment3')"
          :items="environment"
          item-value="id"
          item-text="name"
          :rules="[v => !!v || $t('environment_required')]"
          outlined
          dense
          required
          :disabled="formSaving"
        ></v-select>

        <v-select
          v-if="itemTypeIndex === 0"
          v-model="item.vault_key_id"
          :label="$t('vaultPassword')"
          clearable
          :items="loginPasswordKeys"
          item-value="id"
          item-text="name"
          :disabled="formSaving"
          outlined
          dense
        ></v-select>
      </v-col>

      <v-col cols="12" md="6" class="pb-0">

        <v-select
          v-if="itemTypeIndex > 0"
          v-model="item.vault_key_id"
          :label="$t('vaultPassword2')"
          clearable
          :items="loginPasswordKeys"
          item-value="id"
          item-text="name"
          :disabled="formSaving"
          outlined
          dense
        ></v-select>

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
          v-model="cronRepositoryIdVisible"
        />

        <v-select
          v-if="cronRepositoryIdVisible"
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
          v-if="cronRepositoryIdVisible"
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

<!--        <a @click="advancedOptions = true" v-if="!advancedOptions">-->
<!--          Advanced-->
<!--          <v-icon style="transform: translateY(-1px)">mdi-chevron-right</v-icon>-->
<!--        </a>-->

<!--        <div v-if="advancedOptions" class="mb-3">-->
<!--          <a @click="advancedOptions = false">-->
<!--            Hide-->
<!--            <v-icon style="transform: translateY(-1px)">mdi-chevron-up</v-icon>-->
<!--          </a>-->
<!--        </div>-->

        <codemirror
          :style="{ border: '1px solid lightgray' }"
          v-model="item.arguments"
          :options="cmOptions"
          :disabled="formSaving"
          :placeholder="$t('cliArgsJsonArrayExampleIMyinventoryshPrivatekeythe2')"
        />

        <v-checkbox
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

import { codemirror } from 'vue-codemirror';
import ItemFormBase from '@/components/ItemFormBase';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
import 'codemirror/addon/lint/json-lint.js';
import 'codemirror/addon/display/placeholder.js';
import { TEMPLATE_TYPE_ICONS, TEMPLATE_TYPE_TITLES } from '../lib/constants';
import SurveyVars from './SurveyVars';

export default {
  mixins: [ItemFormBase],

  components: {
    SurveyVars,
    codemirror,
  },

  props: {
    sourceItemId: Number,
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
      keys: null,
      inventory: null,
      repositories: null,
      environment: null,
      views: null,
      schedules: null,
      buildTemplates: null,
      cronFormat: '* * * * *',
      cronRepositoryId: null,
      cronRepositoryIdVisible: false,

      helpDialog: null,
      helpKey: null,

      advancedOptions: false,
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

      return this.keys != null
        && this.repositories != null
        && this.inventory != null
        && this.environment != null
        && this.item != null
        && this.schedules != null
        && this.views != null;
    },

    loginPasswordKeys() {
      if (this.keys == null) {
        return null;
      }
      return this.keys.filter((key) => key.type === 'login_password');
    },
  },

  methods: {
    setSurveyVars(v) {
      this.item.survey_vars = v;
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

      this.keys = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/keys`,
        responseType: 'json',
      })).data;

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
          this.cronRepositoryIdVisible = this.cronRepositoryId != null;
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
    },

    async afterSave(newItem) {
      if (newItem || this.schedules.length === 0) {
        if (this.cronFormat != null && this.cronFormat !== '' && this.cronRepositoryIdVisible) {
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
            },
          });
        }
      } else if (this.schedules.length > 1) {
        // do nothing
      } else if (this.cronFormat == null || this.cronFormat === '' || !this.cronRepositoryIdVisible) {
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
