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
  <v-form class="mt-1" v-else ref="form" lazy-validation v-model="formValid">
    <v-dialog v-model="helpDialog" hide-overlay width="300">
      <v-alert border="top" colored-border type="info" elevation="2" class="mb-0 pb-0">
        <div v-if="helpKey === 'build_version'">
          <p>
            {{ $t('definesStartVersionOfYourArtifactEachRunIncrements') }}
          </p>
          <p>
            {{ $t('forMoreInformationAboutBuildingSeeThe') }}
            <a
              href="https://docs.semaphoreui.com/user-guide/task-templates#build"
              target="_blank"
            >{{ $t('taskTemplateReference') }}</a
            >.
          </p>
        </div>
        <div v-else-if="helpKey === 'build'">
          <p>
            {{ $t('definesWhatArtifactShouldBeDeployedWhenTheTaskRun') }}
          </p>
          <p>
            {{ $t('forMoreInformationAboutDeployingSeeThe') }}
            <a
              href="https://docs.semaphoreui.com/user-guide/task-templates#build"
              target="_blank"
            >{{ $t('taskTemplateReference2') }}</a
            >.
          </p>
        </div>
        <div v-if="helpKey === 'cron'">
          <p>{{ $t('definesAutorunSchedule') }}</p>
          <p>
            {{ $t('forMoreInformationAboutCronSeeThe') }}
            <a
              href="https://pkg.go.dev/github.com/robfig/cron/v3#hdr-CRON_Expression_Format"
              target="_blank"
            >{{ $t('cronExpressionFormatReference') }}</a
            >.
          </p>
        </div>
      </v-alert>
    </v-dialog>

    <v-alert :value="formError" color="error">{{ formError }}</v-alert>

    <v-row class="mb-0">
      <v-col>
        <h2 class="mb-4">{{ $t('template_common_options') }}</h2>

        <v-card
          class="mb-6"
          :color="$vuetify.theme.dark ? '#212121' : 'white'"
          style="background: #8585850f"
        >
          <v-tabs fixed-tabs v-model="itemTypeIndex">
            <v-tab style="padding: 0" v-for="key in Object.keys(TEMPLATE_TYPE_ICONS)" :key="key">
              <v-icon small class="mr-2">{{ TEMPLATE_TYPE_ICONS[key] }}</v-icon>
              {{ $t(TEMPLATE_TYPE_TITLES[key]) }}
            </v-tab>
          </v-tabs>

          <div class="ml-4 mr-4 mt-6" v-if="item.type">
            <v-text-field
              v-if="item.type === 'build'"
              v-model="item.start_version"
              :label="$t('startVersion')"
              :rules="[(v) => !!v || $t('start_version_required')]"
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
              :rules="[(v) => !!v || $t('build_template_required')]"
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
          :rules="[(v) => !!v || $t('name_required')]"
          outlined
          dense
          required
          :disabled="formSaving"
        ></v-text-field>

        <div style="position: relative">
          <v-autocomplete
            v-model="item.repository_id"
            :label="fieldLabel('repository') + ' *'"
            :items="repositories"
            item-value="id"
            item-text="name"
            :rules="isFieldRequired('repository') ? [(v) => !!v || $t('repository_required')] : []"
            outlined
            dense
            :required="isFieldRequired('repository')"
            :disabled="formSaving"
            v-if="needField('repository')"
          >
          </v-autocomplete>
          <v-chip
            v-if="gitBranch"
            small
            label
            style="position: absolute; right: 45px; top: 8px"
            @click="setBranch = !setBranch"
          >
            {{ gitBranch }}
          </v-chip>
        </div>

        <v-card
          v-if="setBranch"
          style="background: var(--highlighted-card-bg-color)"
          class="mb-6 pt-3"
        >
          <div
            style="
              position: absolute;
              background: var(--highlighted-card-bg-color);
              width: 28px;
              height: 28px;
              transform: rotate(45deg);
              right: 55px;
              top: -14px;
              border-radius: 0;
            "
          ></div>

          <v-card-text class="pb-0">
            <div v-if="branches != null">
              <v-autocomplete
                clearable
                :items="branches"
                v-model="item.git_branch"
                :label="fieldLabel('branch')"
                outlined
                dense
                :disabled="formSaving"
                :placeholder="$t('branch')"
              ></v-autocomplete>
            </div>
            <div v-else>
              <v-text-field
                clearable
                v-model="item.git_branch"
                :label="fieldLabel('branch')"
                outlined
                dense
                :disabled="formSaving"
                :placeholder="$t('branch')"
              ></v-text-field>
            </div>
          </v-card-text>
        </v-card>

        <div v-if="needField('playbook')">
          <div v-if="playbooks != null">
            <v-autocomplete
              class="InputWithAppendedButton"
              v-model="item.playbook"
              :items="playbooks"
              :label="fieldLabel('playbook')"
              :rules="
                isFieldRequired('playbook') ? [(v) => !!v || $t('playbook_filename_required')] : []
              "
              outlined
              dense
              clearable
              :required="isFieldRequired('playbook')"
              :disabled="formSaving"
              :placeholder="$t('exampleSiteyml')"
              :loading="playbooksLoading"
            >
              <template v-slot:append-outer>
                <v-btn
                  depressed
                  @click="playbooksLoading ? cancelPlaybookLoading() : loadPlaybooks()"
                >
                  <v-icon>{{ playbooksLoading ? 'mdi-close' : 'mdi-refresh' }}</v-icon>
                </v-btn>
              </template>
            </v-autocomplete>
          </div>
          <div v-else>
            <v-text-field
              class="InputWithAppendedButton"
              v-model="item.playbook"
              :label="fieldLabel('playbook')"
              :rules="
                isFieldRequired('playbook') ? [(v) => !!v || $t('playbook_filename_required')] : []
              "
              outlined
              dense
              :required="isFieldRequired('playbook')"
              :disabled="formSaving"
              :placeholder="$t('exampleSiteyml')"
              :loading="playbooksLoading"
            >
              <template v-slot:append-outer>
                <v-btn
                  depressed
                  @click="playbooksLoading ? cancelPlaybookLoading() : loadPlaybooks()"
                >
                  <v-icon>{{ playbooksLoading ? 'mdi-close' : 'mdi-refresh' }}</v-icon>
                </v-btn>
              </template>
            </v-text-field>
          </div>
        </div>

        <v-autocomplete
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
        ></v-autocomplete>

        <v-autocomplete
          v-model="item.environment_ids"
          :label="fieldLabel('environment')"
          :items="environment"
          item-value="id"
          item-text="name"
          :rules="
            isFieldRequired('environment')
              ? [(v) => (Array.isArray(v) && v.length > 0) || $t('environment_required')]
              : []
          "
          persistent-hint
          multiple
          chips
          deletable-chips
          outlined
          small-chips
          :required="isFieldRequired('environment')"
          :disabled="formSaving"
          v-if="needField('environment')"
        ></v-autocomplete>

        <v-autocomplete
          v-model="item.key_ids"
          :label="fieldLabel('keys')"
          :items="sshKeys"
          item-value="id"
          item-text="name"
          multiple
          chips
          deletable-chips
          outlined
          small-chips
          class="mb-3"
          :disabled="formSaving"
          v-if="needField('keys')"
        ></v-autocomplete>

        <v-autocomplete
          class="mb-3"
          style="max-height: 60px"
          v-model="item.view_id"
          :label="$t('view')"
          clearable
          :items="views"
          item-value="id"
          item-text="title"
          :disabled="formSaving"
          outlined
          dense
        ></v-autocomplete>
      </v-col>

      <v-col>
        <h2 class="mb-4">{{ $t('template_advanced') }}</h2>

        <div class="mb-4">
          <v-autocomplete
            v-if="features.project_runners"
            v-model="item.runner_tag"
            :items="runnerTags"
            :label="fieldLabel('runner_tag')"
            item-value="tag"
            item-text="tag"
            outlined
            dense
            :disabled="formSaving"
            :placeholder="$t('runner_tag')"
            clearable
          ></v-autocomplete>

          <div style="position: relative">
            <v-text-field
              v-model="item.executor_image"
              :label="$t('executor_image')"
              :hint="$t('executor_image_hint')"
              persistent-hint
              placeholder="semaphoreui/job:latest"
              outlined
              dense
              clearable
              class="mb-4"
              :disabled="formSaving || !isExecutorImageAvailable"
            ></v-text-field>

            <v-chip
              v-if="!isExecutorImageAvailable"
              color="hsl(348deg, 86%, 61%)"
              text-color="white"
              small
              label
              style="position: absolute; top: -10px; right: 15px"
              @click="upgradeToPro('docker_executor')"
            >
              Upgrade to PRO
            </v-chip>
          </div>

          <SurveyVars :vars="surveyVars" @change="setSurveyVars" />

          <v-checkbox class="mt-0" v-model="item.allow_parallel_tasks">
            <template v-slot:label>
              {{ $t('allow_parallel_tasks') }}
            </template>
          </v-checkbox>

          <template v-if="systemInfo && systemInfo.jwt && systemInfo.jwt.enabled">
            <v-checkbox class="mt-0" v-model="item.jwt_params.enabled">
              <template v-slot:label>
                {{ $t('jwt_enabled') }}
                <v-chip class="ml-2" small color="error">New</v-chip>
              </template>
            </v-checkbox>

            <DropdownCard v-if="item.jwt_params.enabled" no-bottom-padding>
              <v-combobox
                v-model="item.jwt_params.audience"
                :label="$t('jwt_audience')"
                :hint="$t('jwt_audience_hint')"
                persistent-hint
                multiple
                chips
                small-chips
                deletable-chips
                outlined
                dense
                :disabled="formSaving"
              ></v-combobox>

              <v-text-field
                class="mt-4"
                v-model="item.jwt_params.ttl"
                :label="$t('jwt_ttl')"
                :hint="jwtTtlHint"
                persistent-hint
                placeholder="1h"
                outlined
                dense
                :disabled="formSaving"
              ></v-text-field>
            </DropdownCard>
          </template>

          <v-checkbox
            class="mt-0"
            :label="$t('iWantToRunATaskByTheCronOnlyForForNewCommitsOfSome')"
            v-model="cronVisible"
          />

          <DropdownCard v-if="cronVisible">
            <v-select
              v-model="cronRepositoryId"
              :label="$t('repository2')"
              :placeholder="$t('cronChecksNewCommitBeforeRun')"
              :rules="[(v) => !!v || $t('repository_required')]"
              :items="repositories"
              item-value="id"
              item-text="name"
              clearable
              :disabled="formSaving"
              outlined
              dense
            ></v-select>

            <v-select
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
          </DropdownCard>

          <v-checkbox
            class="mt-0"
            :label="$t('suppressSuccessAlerts')"
            v-model="item.suppress_success_alerts"
          />

          <div style="position: relative">
            <ArgsPicker :vars="args" @change="setArgs" title="CLI args" />

            <RichEditor
              v-model="argsJson"
              type="json_array"
              style="position: absolute; right: -23px; top: -18px; margin: 10px"
            />
          </div>
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
            {{ $t('template_app_options', { app: getAppTitle(app, true) }) }}
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
            v-if="needField('skip_galaxy_install')"
            v-model="item.task_params.skip_galaxy_install"
            :label="$t('skipGalaxyInstall')"
            class="mt-0"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('auto_approve')"
            v-model="item.task_params.auto_approve"
            v-if="needField('auto_approve')"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('terraform_override_backend')"
            v-model="item.task_params.override_backend"
            :true-value="true"
            :false-value="false"
            v-if="needField('override_backend') && features.terraform_backend"
          />

          <v-text-field
            v-model="item.task_params.backend_filename"
            :label="fieldLabel('terraform_backend_filename')"
            outlined
            dense
            :disabled="formSaving || !item.task_params.override_backend"
            placeholder="backend.tf"
            :rules="[(v) => validateBackendFilename(v) || $t('terraform_invalid_backend_filename')]"
            v-if="needField('backend_filename') && features.terraform_backend"
          ></v-text-field>
        </div>

        <h2 class="mb-4">
          {{ $t('template_app_prompts', { app: getAppTitle(app, true) }) }}
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
            v-if="needField('allow_override_skip_galaxy_install')"
            v-model="item.task_params.allow_override_skip_galaxy_install"
            :label="$t('skipGalaxyInstall')"
            class="mt-0"
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
import RichEditor from '@/components/RichEditor.vue';
import DropdownCard from '@/components/DropdownCard.vue';
import SurveyVars from './SurveyVars';

export default {
  mixins: [ItemFormBase, AppFieldsMixin, AppsMixin],

  components: {
    DropdownCard,
    RichEditor,
    TemplateVaults,
    ArgsPicker,
    SurveyVars,
  },

  props: {
    sourceItemId: Number,
    app: String,
    features: Object,
    taskType: String,
  },

  data() {
    return {
      cronFormats: [
        {
          cron: '* * * * *',
          title: '1 minute',
        },
        {
          cron: '*/5 * * * *',
          title: '5 minutes',
        },
        {
          cron: '*/10 * * * *',
          title: '10 minutes',
        },
        {
          cron: '@hourly',
          title: '1 hour',
        },
        {
          cron: '@daily',
          title: '24 hours',
        },
      ],
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
        jwt_params: { enabled: false, audience: [], ttl: '' },
      },
      systemInfo: null,
      inventory: null,
      repositories: null,
      environment: null,
      keys: null,
      views: null,
      schedules: null,
      buildTemplates: null,
      cronFormat: '* * * * *',
      cronRepositoryId: null,
      cronVisible: false,

      helpDialog: null,
      helpKey: null,

      args: [],
      runnerTags: null,
      branches: null,
      playbooks: null,
      playbooksLoading: false,
      playbooksAbort: null,
      setBranch: false,
    };
  },

  watch: {
    gitBranchOfTemplate() {
      if (this.playbooks != null) {
        this.playbooks = null;
        this.loadPlaybooks();
      }
    },

    async repositoryId() {
      this.branches = null;
      this.playbooks = null;

      await Promise.all([this.loadBranches()]);
    },

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

  async created() {
    await Promise.all([this.loadBranches()]);
  },

  computed: {
    // The image override is only honoured by the container-based executors, which
    // are themselves paid features: Docker in PRO, Kubernetes in Enterprise.
    isExecutorImageAvailable() {
      return !!(this.features?.docker_executor || this.features?.k8s_executor);
    },

    argsJson: {
      get() {
        return JSON.stringify(this.args);
      },
      set(val) {
        this.args = JSON.parse(val);
      },
    },

    // Only SSH keys can be added to the SSH agent used by ansible-galaxy.
    sshKeys() {
      return (this.keys || []).filter((key) => key.type === 'ssh');
    },

    jwtTtlHint() {
      const max = this.systemInfo?.jwt?.max_ttl;
      if (!max) {
        return this.$t('jwt_ttl_hint_no_max');
      }
      return this.$t('jwt_ttl_hint', { max });
    },

    repositoryId() {
      return this.item?.repository_id;
    },

    gitBranchOfTemplate() {
      return this.item?.git_branch;
    },

    gitBranch() {
      let ret = this.item?.git_branch;
      if (!ret) {
        const repo = this.repositories.find((x) => x.id === this.repositoryId);
        if (repo) {
          ret = repo.git_branch;
        }
      }
      return ret;
    },

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
      return ['', 'ansible', 'ansible', 'tofu', 'terraform'].includes(this.app);
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

      return (
        this.repositories != null
        && this.inventory != null
        && this.environment != null
        && this.keys != null
        && this.item != null
        && this.schedules != null
        && this.views != null
        && this.runnerTags != null
      );
    },
  },

  methods: {
    async loadBranches() {
      if (this.repositoryId == null) {
        return;
      }

      try {
        this.branches = await this.loadProjectEndpoint(
          `/repositories/${this.repositoryId}/branches`,
        );
      } catch (e) {
        this.branches = null;
      }
    },

    cancelPlaybookLoading() {
      if (this.playbooksAbort) {
        this.playbooksAbort.abort();
        this.playbooksAbort = null;
      }
    },

    async loadPlaybooks() {
      if (this.repositoryId == null) {
        this.playbooks = null;
        return;
      }

      this.cancelPlaybookLoading();
      const ctrl = new AbortController();
      this.playbooksAbort = ctrl;
      this.playbooksLoading = true;

      try {
        this.playbooks = await this.loadProjectEndpoint(
          `/repositories/${this.repositoryId}/playbooks?branch=${encodeURIComponent(this.item.git_branch || '')}`,
          { signal: ctrl.signal },
        );
      } catch (e) {
        this.playbooks = null;
      } finally {
        // ponytail: guard against a newer request having replaced this one
        if (this.playbooksAbort === ctrl || this.playbooksAbort == null) {
          this.playbooksAbort = null;
          this.playbooksLoading = false;
        }
      }
    },

    validateBackendFilename(v) {
      if (!v) {
        return true;
      }

      if (!v.endsWith('.tf')) {
        return 'File must have extension .tf';
      }

      return /^[a-zA-Z0-9_\-.]+\.tf$/.test(v);
    },

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
        jwt_params: { enabled: false, audience: [], ttl: '' },
        environment_ids: [],
        key_ids: [],
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
        this.keys,
        templates,
        this.runnerTags,
      ] = await Promise.all([
        this.loadProjectResources('repositories'),
        this.loadProjectEndpoint(`/inventory?app=${this.app}&template_id=${this.itemId}`),
        this.loadProjectEndpoint(`/inventory?app=${this.app}`),
        this.isNew ? [] : this.loadProjectEndpoint(`/templates/${this.itemId}/schedules`),
        this.loadProjectResources('views'),
        this.loadProjectResources('environment'),
        this.loadProjectResources('keys'),
        this.loadProjectResources('templates'),
        this.loadProjectResources('runner_tags'),
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

        const sourceSchedule = (
          await this.loadProjectEndpoint(`/templates/${this.sourceItemId}/schedules`)
        )[0];

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

      // The API omits executor_image when it is not set; declare it explicitly so
      // the text field stays reactive for templates without an override.
      if (this.item.executor_image === undefined) {
        this.$set(this.item, 'executor_image', null);
      }

      if (!this.item.jwt_params) {
        this.$set(this.item, 'jwt_params', { enabled: false, audience: [], ttl: '' });
      } else {
        if (!Array.isArray(this.item.jwt_params.audience)) {
          this.$set(this.item.jwt_params, 'audience', []);
        }
        if (typeof this.item.jwt_params.ttl !== 'string') {
          this.$set(this.item.jwt_params, 'ttl', '');
        }
      }

      try {
        this.systemInfo = (
          await axios({
            method: 'get',
            url: '/api/info',
            responseType: 'json',
          })
        ).data;
      } catch (e) {
        this.systemInfo = null;
      }

      if (!Array.isArray(this.item.environment_ids)) {
        this.$set(this.item, 'environment_ids', []);
      }

      if (!Array.isArray(this.item.key_ids)) {
        this.$set(this.item, 'key_ids', []);
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
