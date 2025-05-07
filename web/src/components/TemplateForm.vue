<template>
  <div v-if="!isLoaded" :style="{ height: `${loaderHeight}px` }" class="mt-1">
    <v-row>
      <v-col>
        <v-skeleton-loader
          type="
            table-heading,
            image,
            list-item-two-line,
            list-item-two-line,
            list-item-two-line"
        ></v-skeleton-loader>
      </v-col>
      <v-col>
        <v-skeleton-loader
          type="
            table-heading,
            list-item-two-line,
            list-item-two-line,
            list-item-two-line,
            article"
        ></v-skeleton-loader>
      </v-col>
      <v-col v-if="needAppBlock">
        <v-skeleton-loader
          type="
            table-heading,
            list-item-two-line,
            article,
            list-item-two-line,
            article"
        ></v-skeleton-loader>
      </v-col>
    </v-row>
  </div>
  <v-form
    class="mt-1"
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
            <a href="https://docs.semaphoreui.com/user-guide/task-templates#build"
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
            <a href="https://docs.semaphoreui.com/user-guide/task-templates#build"
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
    >{{ formError }}
    </v-alert>

    <v-row class="mb-0">
      <v-col>
        <h2 class="mb-4">{{ $t('template_common_options') }}</h2>

        <v-card
          class="mb-6"
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
        >
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

        <v-text-field
          v-model="item.playbook"
          :label="fieldLabel('playbook')"
          :rules="
              isFieldRequired('playbook')
              ? [v => !!v || $t('playbook_filename_required')]
              : []"
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

        <v-select
          class="mb-3"
          style="max-height: 60px;"
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
      </v-col>

      <v-col>
        <h2 class="mb-4">{{ $t('template_advanced') }}</h2>

        <div class="mb-4">
          <v-text-field
            v-model="item.git_branch"
            :label="fieldLabel('branch')"
            outlined
            dense
            :disabled="formSaving"
            :placeholder="$t('branch')"
          ></v-text-field>

          <v-text-field
            v-if="premiumFeatures.project_runners"
            v-model="item.runner_tag"
            :label="fieldLabel('runner_tag')"
            outlined
            dense
            :disabled="formSaving"
            :placeholder="$t('runner_tag')"
          ></v-text-field>

          <SurveyVars
            :vars="surveyVars"
            @change="setSurveyVars"
          />

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
            :label="$t('checkInterval')"
            :hint="$t('newCommitCheckInterval')"
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
            title="CLI args"
          />
        </div>

        <h2 class="mb-4">{{ $t('task_prompts') }}</h2>
        <div class="d-flex" style="column-gap: 20px; flex-wrap: wrap">
          <v-checkbox
            class="mt-0"
            :label="$t('allowCliArgsInTask')"
            v-model="item.allow_override_args_in_task"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('allow_override_branch')"
            v-model="item.allow_override_branch_in_task"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('allowInventoryInTask')"
            v-model="allow_override_inventory"
            v-if="needField('allow_override_inventory')"
          />
        </div>
      </v-col>

      <v-col v-if="needAppBlock">
        <div class="mb-3">
          <h2 class="mb-4">
            {{ $t('template_app_options', {app: getAppTitle(app, true)}) }}
          </h2>

          <ArgsPicker
            v-if="needField('limit')"
            :vars="item.task_params.limit"
            @change="setLimit"
            :title="$t('limit')"
            :arg-title="$t('limit')"
            :add-arg-title="$t('addLimit')"
          />

          <ArgsPicker
            v-if="needField('tags')"
            :vars="item.task_params.tags"
            @change="setTags"
            :title="$t('tags')"
            :arg-title="$t('tag')"
            :add-arg-title="$t('addTag')"
          />

          <ArgsPicker
            v-if="needField('skip_tags')"
            :vars="item.task_params.skip_tags"
            @change="setSkipTags"
            :title="$t('skipTags')"
            :arg-title="$t('tag')"
            :add-arg-title="$t('addSkippedTag')"
          />

          <TemplateVaults
            v-if="needField('vault')"
            :project-id="this.projectId"
            :vaults="vaults"
            @change="setTemplateVaults"
          ></TemplateVaults>

          <v-checkbox
            class="mt-0"
            :label="$t('auto_approve')"
            v-model="item.task_params.auto_approve"
            v-if="needField('auto_approve')"
          />

        </div>

        <h2 class="mb-4">
          {{ $t('template_app_prompts', {app: getAppTitle(app, true)}) }}
        </h2>
        <div class="d-flex" style="column-gap: 20px; flex-wrap: wrap">
          <v-checkbox
            class="mt-0"
            :label="$t('allowLimitInTask')"
            v-model="item.task_params.allow_override_limit"
            v-if="needField('allow_override_limit')"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('tags')"
            v-model="item.task_params.allow_override_tags"
            v-if="needField('allow_override_tags')"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('skipTags')"
            v-model="item.task_params.allow_override_skip_tags"
            v-if="needField('allow_override_skip_tags')"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('allowDebug')"
            v-model="item.task_params.allow_debug"
            v-if="needField('allow_debug')"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('auto_approve')"
            v-model="item.task_params.allow_auto_approve"
            v-if="needField('allow_auto_approve')"
          />

        </div>
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
import { TEMPLATE_TYPE_ICONS, TEMPLATE_TYPE_TITLES } from '@/lib/constants';
import AppFieldsMixin from '@/components/AppFieldsMixin';
import AppsMixin from '@/components/AppsMixin';
import SurveyVars from './SurveyVars';

export default {
  mixins: [ItemFormBase, AppFieldsMixin, AppsMixin],

  components: {
    TemplateVaults,
    ArgsPicker,
    SurveyVars,
  },

  props: {
    sourceItemId: Number,
    app: String,
    premiumFeatures: Object,
    taskType: String,
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
      item: {
        task_params: {},
      },
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

      args: [],
    };
  },

  watch: {
    needReset(val) {
      if (val) {
        if (this.item != null) {
          this.item.template_id = this.templateId;
        }
        this.inventory = null;
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
    allow_override_inventory: {
      get() {
        return this.item.task_params.allow_override_inventory;
      },
      set(newValue) {
        this.item.task_params.allow_override_inventory = newValue;
      },
    },

    loaderHeight() {
      switch (this.taskType) {
        case 'build':
          if (['', 'ansible', 'terraform', 'tofu'].includes(this.app)) {
            return 626;
          }
          return 560;
        case 'deploy':
          if (['', 'ansible', 'terraform', 'tofu'].includes(this.app)) {
            return 676;
          }
          return 610;
        default:
          if (['', 'ansible', 'terraform', 'tofu'].includes(this.app)) {
            return 564;
          }
          return 514;
      }
    },

    appBlockTitle() {
      switch (this.app) {
        case '':
        case 'ansible':
          return this.$t('ansible_playbook_options');
        default:
          return this.app;
      }
    },

    needAppBlock() {
      return ['', 'ansible', 'ansible', 'tofu'].includes(this.app);
    },

    surveyVars() {
      // if (this.sourceItemId != null && this.item.survey_vars === undefined) {
      //   throw new Error();
      // }
      return this.item.survey_vars;
    },

    vaults() {
      // if (this.sourceItemId != null && this.item.vaults === undefined) {
      //   throw new Error();
      // }
      return this.item.vaults;
    },

    isLoaded() {
      // if (this.isNew && this.sourceItemId == null) {
      //   return true;
      // }

      return this.repositories != null
        && this.inventory != null
        && this.environment != null
        && this.item != null
        && this.schedules != null
        && this.views != null;
    },

  },

  methods: {
    setSkipTags(tags) {
      this.item.task_params.skip_tags = tags;
    },

    setTags(tags) {
      this.item.task_params.tags = tags;
    },

    setLimit(limit) {
      this.item.task_params.limit = limit;
    },

    setArgs(args) {
      this.args = args;
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

    getNewItem() {
      return {
        task_params: {},
      };
    },

    async loadRelativeData() {
      let templates;
      let inventory1;
      let inventory2;

      [
        this.repositories,
        inventory1,
        inventory2,
        this.schedules,
        this.views,
        this.environment,
        templates,
      ] = await Promise.all([
        this.loadProjectResources('repositories'),
        this.loadProjectEndpoint(`/inventory?app=${this.app}&template_id=${this.itemId}`),
        this.loadProjectEndpoint(`/inventory?app=${this.app}`),
        this.isNew ? [] : this.loadProjectEndpoint(`/templates/${this.itemId}/schedules`),
        this.loadProjectResources('views'),
        this.loadProjectResources('environment'),
        this.loadProjectResources('templates'),
      ]);

      this.inventory = [...inventory1, ...inventory2];

      const builds = [];
      const deploys = [];

      templates.forEach((t) => {
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
    },

    async afterLoadData() {
      if (this.sourceItemId) {
        const item = await this.loadProjectResource('templates', this.sourceItemId);

        item.id = null;

        if (item.vaults) {
          for (let i = 0; i < item.vaults.length; i += 1) {
            item.vaults[i].id = null;
          }
        }

        const sourceSchedule = (await this.loadProjectEndpoint(`/templates/${this.sourceItemId}/schedules`))[0];

        if (sourceSchedule != null) {
          this.cronFormat = sourceSchedule.cron_format;
          this.cronRepositoryId = sourceSchedule.repository_id;
          this.cronVisible = this.cronRepositoryId != null;
        }

        this.item = item;
      }

      if (!this.item.task_params) {
        this.item.task_params = {};
      }

      this.args = JSON.parse(this.item.arguments || '[]');

      await this.loadRelativeData();

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
