<template>
  <div
    v-if="!isLoaded"
    :style="{ height: `${loaderHeight}px` }"
    class="mt-1"
    data-testid="templateForm-loading"
    :data-state="formError ? 'error' : 'loading'"
  >
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
    v-model="formValid"
    data-testid="templateForm-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-dialog v-model="helpDialog" :scrim="false" content-class="NestedDialog" width="300">
      <v-alert
        border="top"
        border-color="info"
        type="info"
        color="surface"
        elevation="1"
        class="mb-0 pb-0"
        data-testid="templateHelpDialog"
      >
        <!-- colored-border of Vuetify 2 coloured the icon along with the border -->
        <template v-slot:prepend>
          <v-icon color="info" />
        </template>
        <div v-if="helpKey === 'build_version'" data-testid="templateHelpDialog-buildVersion">
          <p>
            {{ $t('definesStartVersionOfYourArtifactEachRunIncrements') }}
          </p>
          <p>
            {{ $t('forMoreInformationAboutBuildingSeeThe') }}
            <a
              href="https://docs.semaphoreui.com/user-guide/task-templates#build"
              target="_blank"
              data-testid="templateHelpDialog-buildVersionLink"
            >{{ $t('taskTemplateReference') }}</a
            >.
          </p>
        </div>
        <div v-else-if="helpKey === 'build'" data-testid="templateHelpDialog-build">
          <p>
            {{ $t('definesWhatArtifactShouldBeDeployedWhenTheTaskRun') }}
          </p>
          <p>
            {{ $t('forMoreInformationAboutDeployingSeeThe') }}
            <a
              href="https://docs.semaphoreui.com/user-guide/task-templates#build"
              target="_blank"
              data-testid="templateHelpDialog-buildLink"
            >{{ $t('taskTemplateReference2') }}</a
            >.
          </p>
        </div>
        <div v-if="helpKey === 'cron'" data-testid="templateHelpDialog-cron">
          <p>{{ $t('definesAutorunSchedule') }}</p>
          <p>
            {{ $t('forMoreInformationAboutCronSeeThe') }}
            <a
              href="https://pkg.go.dev/github.com/robfig/cron/v3#hdr-CRON_Expression_Format"
              target="_blank"
              data-testid="templateHelpDialog-cronLink"
            >{{ $t('cronExpressionFormatReference') }}</a
            >.
          </p>
        </div>
      </v-alert>
    </v-dialog>

    <v-alert
      :model-value="!!formError"
      color="error"
      data-testid="templateForm-error"
    >{{ formError }}</v-alert>

    <v-row class="mb-0">
      <v-col class="TemplateForm__column">
        <h2
          class="mb-4"
          data-testid="templateForm-commonTitle"
        >{{ $t('template_common_options') }}</h2>

        <v-card
          class="mb-6"
          :color="$vuetify.theme.current.dark ? '#212121' : 'white'"
          style="background: #8585850f"
          data-testid="templateForm-typeCard"
        >
          <v-tabs fixed-tabs v-model="itemTypeIndex" data-testid="templateForm-typeTabs">
            <v-tab
              style="padding: 0"
              v-for="key in Object.keys(TEMPLATE_TYPE_ICONS)"
              :key="key"
              :data-testid="`templateForm-tab-${key || 'task'}`"
            >
              <v-icon size="small" class="mr-2">{{ TEMPLATE_TYPE_ICONS[key] }}</v-icon>
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
              data-testid="templateForm-startVersion"
              append-icon="mdi-help-circle"
              @click:append="showHelpDialog('build_version')"
            ></v-text-field>

            <v-autocomplete
              v-if="item.type === 'deploy'"
              v-model="item.build_template_id"
              :label="$t('buildTemplate')"
              :items="buildTemplates"
              item-value="id"
              item-title="name"
              :rules="[(v) => !!v || $t('build_template_required')]"
              required
              :disabled="formSaving"
              data-testid="templateForm-buildTemplateId"
              append-icon="mdi-help-circle"
              @click:append="showHelpDialog('build')"
            ></v-autocomplete>

            <v-checkbox
              v-if="item.type === 'deploy'"
              class="mt-0"
              :label="$t('autorun')"
              v-model="item.autorun"
              data-testid="templateForm-autorun"
            />
          </div>
        </v-card>

        <v-text-field
          v-model="item.name"
          :label="$t('name2')"
          :rules="[(v) => !!v || $t('name_required')]"
          variant="outlined"
          density="compact"
          required
          :disabled="formSaving"
          data-testid="templateForm-name"
        ></v-text-field>

        <div style="position: relative">
          <v-autocomplete
            v-model="item.repository_id"
            :label="fieldLabel('repository') + ' *'"
            :items="repositories"
            item-value="id"
            item-title="name"
            :rules="isFieldRequired('repository') ? [(v) => !!v || $t('repository_required')] : []"
            variant="outlined"
            density="compact"
            :required="isFieldRequired('repository')"
            :disabled="formSaving"
            v-if="needField('repository')"
            data-testid="templateForm-repositoryId"
          >
          </v-autocomplete>
          <v-tooltip
            location="bottom"
            :disabled="!isMonorepo"
            content-class="opacity1"
            max-width="400"
            transition="fade-transition"
          >
            <template v-slot:activator="{ props }">
              <!-- A monorepo chip gets no click listener at all, so it is neither
                   styled nor behaves as clickable; the wrapper carries the tooltip. -->
              <span
                v-bind="props"
                :tabindex="isMonorepo ? 0 : null"
                style="position: absolute; right: 45px; top: 8px"
              >
                <v-chip
                  v-if="gitBranch"
                  size="small"
                  label
                  :class="{ 'text-disabled': isMonorepo }"
                  data-testid="templateForm-branchChip"
                  v-on="isMonorepo ? {} : { click: () => { setBranch = !setBranch; } }"
                >
                  {{ gitBranch }}
                </v-chip>
              </span>
            </template>
            {{ $t('monorepoBranchTooltip') }}
          </v-tooltip>
        </div>

        <v-card
          v-if="setBranch"
          style="background: var(--highlighted-card-bg-color); overflow: visible"
          class="mb-6 pt-3"
          data-testid="templateForm-branchCard"
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
                data-testid="templateForm-gitBranch"
                v-model="item.git_branch"
                :label="fieldLabel('branch')"
                variant="outlined"
                density="compact"
                :disabled="formSaving || isMonorepo"
                :placeholder="$t('branch')"
                :hint="isMonorepo ? $t('monorepoBranchHint') : ''"
                :persistent-hint="isMonorepo"
              ></v-autocomplete>
            </div>
            <div v-else>
              <v-text-field
                clearable
                data-testid="templateForm-gitBranch"
                v-model="item.git_branch"
                :label="fieldLabel('branch')"
                variant="outlined"
                density="compact"
                :disabled="formSaving || isMonorepo"
                :placeholder="$t('branch')"
                :hint="isMonorepo ? $t('monorepoBranchHint') : ''"
                :persistent-hint="isMonorepo"
              ></v-text-field>
            </div>
          </v-card-text>
        </v-card>

        <div v-if="needField('playbook')">
          <!-- A combobox, not an autocomplete: a path which is not in the list
               must still be accepted. -->
          <v-combobox
            class="InputWithAppendedButton"
            v-model="item.playbook"
            :items="playbooks || []"
            :label="fieldLabel('playbook')"
            :rules="
              isFieldRequired('playbook') ? [(v) => !!v || $t('playbook_filename_required')] : []
            "
            variant="outlined"
            density="compact"
            clearable
            :required="isFieldRequired('playbook')"
            :disabled="formSaving"
            :placeholder="$t('exampleSiteyml')"
            :loading="playbooksLoading"
            @update:search="onPlaybookSearch"
            data-testid="templateForm-playbook"
            :data-state="playbooksLoading ? 'loading' : 'ready'"
          >
            <template v-slot:append>
              <v-btn
                variant="flat"
                @click="playbooksLoading ? cancelPlaybookLoading() : loadPlaybooks()"
                data-testid="templateForm-playbookReload"
              >
                <v-icon>{{ playbooksLoading ? 'mdi-close' : 'mdi-refresh' }}</v-icon>
              </v-btn>
            </template>
          </v-combobox>

          <div v-if="app === 'ansible'">
            <v-checkbox
              class="mt-0"
              v-model="showWorkingDirectoryField"
              :label="$t('workingDirectoryToggleLabel')"
              :disabled="formSaving"
              data-testid="templateForm-workingDirectoryToggle"
            ></v-checkbox>

            <v-text-field
              v-if="showWorkingDirectoryField"
              v-model="item.working_directory"
              :label="$t('workingDirectory')"
              :rules="[(v) => !!(v && v.trim()) || $t('working_directory_required')]"
              variant="outlined"
              density="compact"
              required
              :disabled="formSaving"
              data-testid="templateForm-workingDirectory"
            ></v-text-field>
          </div>
        </div>

        <v-autocomplete
          v-model="item.inventory_id"
          :label="fieldLabel('inventory')"
          :items="inventory"
          item-value="id"
          item-title="name"
          variant="outlined"
          density="compact"
          required
          :disabled="formSaving"
          v-if="needField('inventory')"
          data-testid="templateForm-inventoryId"
        ></v-autocomplete>

        <v-autocomplete
          v-model="item.environment_ids"
          :label="fieldLabel('environment')"
          :items="environment"
          item-value="id"
          item-title="name"
          :rules="
            isFieldRequired('environment')
              ? [(v) => (Array.isArray(v) && v.length > 0) || $t('environment_required')]
              : []
          "
          persistent-hint
          multiple
          chips
          closable-chips
          variant="outlined"
          :required="isFieldRequired('environment')"
          :disabled="formSaving"
          v-if="needField('environment')"
          data-testid="templateForm-environmentIds"
        ></v-autocomplete>

        <v-autocomplete
          class="mb-3"
          style="max-height: 60px"
          v-model="item.view_id"
          :label="$t('view')"
          clearable
          :items="views"
          item-value="id"
          item-title="title"
          :disabled="formSaving"
          variant="outlined"
          density="compact"
          data-testid="templateForm-viewId"
        ></v-autocomplete>
      </v-col>

      <v-col class="TemplateForm__column">
        <h2 class="mb-4" data-testid="templateForm-advancedTitle">{{ $t('template_advanced') }}</h2>

        <div class="mb-4">
          <v-autocomplete
            v-if="features.project_runners"
            v-model="item.runner_tag"
            :items="runnerTags"
            :label="fieldLabel('runner_tag')"
            item-value="tag"
            item-title="tag"
            variant="outlined"
            density="compact"
            :disabled="formSaving"
            :placeholder="$t('runner_tag')"
            clearable
            data-testid="templateForm-runnerTag"
          ></v-autocomplete>

          <div style="position: relative">
            <v-text-field
              v-model="item.executor_image"
              :label="$t('executor_image')"
              persistent-hint
              placeholder="semaphoreui/job:latest"
              variant="outlined"
              density="compact"
              clearable
              class="mb-4"
              :disabled="formSaving || !isExecutorImageAvailable"
              data-testid="templateForm-executorImage"
            ></v-text-field>

            <v-chip
              v-if="!isExecutorImageAvailable"
              color="hsl(348deg, 86%, 61%)"
              variant="flat"
              class="text-white"
              size="small"
              label
              style="position: absolute; top: -10px; right: 15px"
              @click="upgradeToPro('docker_executor')"
              data-testid="templateForm-executorImageUpgrade"
            >
              Upgrade to PRO
            </v-chip>
          </div>

          <SurveyVars :vars="surveyVars" @change="setSurveyVars" />

          <v-checkbox
            class="mt-0"
            v-model="item.allow_parallel_tasks"
            data-testid="templateForm-allowParallelTasks"
          >
            <template v-slot:label>
              {{ $t('allow_parallel_tasks') }}
            </template>
          </v-checkbox>

          <template v-if="systemInfo && systemInfo.jwt && systemInfo.jwt.enabled">
            <v-checkbox
              class="mt-0"
              v-model="item.jwt_params.enabled"
              data-testid="templateForm-jwtEnabled"
            >
              <template v-slot:label>
                {{ $t('jwt_enabled') }}
                <v-chip
                  class="ml-2"
                  size="small"
                  color="error"
                  data-testid="templateForm-jwtNew"
                >New</v-chip>
              </template>
            </v-checkbox>

            <DropdownCard
              v-if="item.jwt_params.enabled"
              no-bottom-padding
              data-testid="templateForm-jwtCard"
            >
              <v-combobox
                v-model="item.jwt_params.audience"
                :label="$t('jwt_audience')"
                :hint="$t('jwt_audience_hint')"
                persistent-hint
                multiple
                chips
                closable-chips
                variant="outlined"
                density="compact"
                :disabled="formSaving"
                data-testid="templateForm-jwtAudience"
              ></v-combobox>

              <v-text-field
                class="mt-4"
                v-model="item.jwt_params.ttl"
                :label="$t('jwt_ttl')"
                :hint="jwtTtlHint"
                persistent-hint
                placeholder="1h"
                variant="outlined"
                density="compact"
                :disabled="formSaving"
                data-testid="templateForm-jwtTtl"
              ></v-text-field>
            </DropdownCard>
          </template>

          <v-checkbox
            class="mt-0"
            :label="$t('iWantToRunATaskByTheCronOnlyForForNewCommitsOfSome')"
            v-model="cronVisible"
            data-testid="templateForm-cronVisible"
          />

          <DropdownCard v-if="cronVisible" data-testid="templateForm-cronCard">
            <v-select
              v-model="cronRepositoryId"
              :label="$t('repository2')"
              :placeholder="$t('cronChecksNewCommitBeforeRun')"
              :rules="[(v) => !!v || $t('repository_required')]"
              :items="repositories"
              item-value="id"
              item-title="name"
              clearable
              :disabled="formSaving"
              variant="outlined"
              density="compact"
              data-testid="templateForm-cronRepositoryId"
            ></v-select>

            <v-select
              v-model="cronFormat"
              :label="$t('checkInterval')"
              :hint="$t('newCommitCheckInterval')"
              item-value="cron"
              item-title="title"
              :items="cronFormats"
              :disabled="formSaving"
              variant="outlined"
              hide-details
              density="compact"
              data-testid="templateForm-cronFormat"
            />
          </DropdownCard>

          <v-checkbox
            class="mt-0"
            :label="$t('suppressSuccessAlerts')"
            v-model="item.suppress_success_alerts"
            data-testid="templateForm-suppressSuccessAlerts"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('suppressErrorAlerts')"
            v-model="item.suppress_error_alerts"
            data-testid="templateForm-suppressErrorAlerts"
          />

          <div style="position: relative">
            <ArgsPicker
              :vars="args"
              @change="setArgs"
              title="CLI args"
              test-id="templateForm-args"
            />

            <RichEditor
              v-model="argsJson"
              type="json_array"
              style="position: absolute; right: -23px; top: -18px; margin: 10px"
              test-id="templateForm-argsJson"
            />
          </div>
        </div>

        <h2 class="mb-4" data-testid="templateForm-promptsTitle">{{ $t('task_prompts') }}</h2>
        <div class="d-flex" style="column-gap: 20px; flex-wrap: wrap">
          <v-checkbox
            class="mt-0"
            :label="$t('allowCliArgsInTask')"
            v-model="item.allow_override_args_in_task"
            data-testid="templateForm-allowOverrideArgsInTask"
          />

          <v-tooltip
            location="bottom"
            :disabled="!isMonorepo"
            content-class="opacity1"
            max-width="400"
            transition="fade-transition"
          >
            <template v-slot:activator="{ props }">
              <!-- A disabled input emits no mouse events, so the wrapper owns the tooltip. -->
              <div v-bind="props">
                <v-checkbox
                  class="mt-0"
                  :label="$t('allow_override_branch')"
                  v-model="item.allow_override_branch_in_task"
                  :disabled="isMonorepo"
                  data-testid="templateForm-allowOverrideBranchInTask"
                />
              </div>
            </template>
            {{ $t('monorepoBranchTooltip') }}
          </v-tooltip>

          <v-checkbox
            class="mt-0"
            :label="$t('allowInventoryInTask')"
            v-model="allow_override_inventory"
            v-if="needField('allow_override_inventory')"
            data-testid="templateForm-allowOverrideInventory"
          />
        </div>
      </v-col>

      <v-col v-if="needAppBlock" class="TemplateForm__column">
        <div class="mb-3">
          <h2 class="mb-4" data-testid="templateForm-appTitle">
            {{ $t('template_app_options', { app: getAppTitle(app, true) }) }}
          </h2>

          <ArgsPicker
            v-if="needField('limit')"
            :vars="item.task_params.limit"
            @change="setLimit"
            :title="$t('limit')"
            :arg-title="$t('limit')"
            :add-arg-title="$t('addLimit')"
            test-id="templateForm-limit"
          />

          <ArgsPicker
            v-if="needField('tags')"
            :vars="item.task_params.tags"
            @change="setTags"
            :title="$t('tags')"
            :arg-title="$t('tag')"
            :add-arg-title="$t('addTag')"
            test-id="templateForm-tags"
          />

          <ArgsPicker
            v-if="needField('skip_tags')"
            :vars="item.task_params.skip_tags"
            @change="setSkipTags"
            :title="$t('skipTags')"
            :arg-title="$t('tag')"
            :add-arg-title="$t('addSkippedTag')"
            test-id="templateForm-skipTags"
          />

          <TemplateVaults
            v-if="needField('vault')"
            :project-id="projectId"
            :vaults="vaults"
            @change="setTemplateVaults"
          ></TemplateVaults>

          <v-checkbox
            class="mt-0"
            :label="$t('auto_approve')"
            v-model="item.task_params.auto_approve"
            v-if="needField('auto_approve')"
            data-testid="templateForm-autoApprove"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('terraform_override_backend')"
            v-model="item.task_params.override_backend"
            :true-value="true"
            :false-value="false"
            v-if="needField('override_backend') && features.terraform_backend"
            data-testid="templateForm-overrideBackend"
          />

          <v-text-field
            v-model="item.task_params.backend_filename"
            :label="fieldLabel('terraform_backend_filename')"
            variant="outlined"
            density="compact"
            :disabled="formSaving || !item.task_params.override_backend"
            placeholder="backend.tf"
            :rules="[(v) => validateBackendFilename(v) || $t('terraform_invalid_backend_filename')]"
            v-if="needField('backend_filename') && features.terraform_backend"
            data-testid="templateForm-backendFilename"
          ></v-text-field>
        </div>

        <h2 class="mb-4" data-testid="templateForm-appPromptsTitle">
          {{ $t('template_app_prompts', { app: getAppTitle(app, true) }) }}
        </h2>
        <div class="d-flex" style="column-gap: 20px; flex-wrap: wrap">
          <v-checkbox
            class="mt-0"
            :label="$t('allowLimitInTask')"
            v-model="item.task_params.allow_override_limit"
            v-if="needField('allow_override_limit')"
            data-testid="templateForm-allowOverrideLimit"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('tags')"
            v-model="item.task_params.allow_override_tags"
            v-if="needField('allow_override_tags')"
            data-testid="templateForm-allowOverrideTags"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('skipTags')"
            v-model="item.task_params.allow_override_skip_tags"
            v-if="needField('allow_override_skip_tags')"
            data-testid="templateForm-allowOverrideSkipTags"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('allowDebug')"
            v-model="item.task_params.allow_debug"
            v-if="needField('allow_debug')"
            data-testid="templateForm-allowDebug"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('dryRun')"
            v-model="show_dry_run"
            v-if="needField('hide_dry_run')"
            data-testid="templateForm-showDryRun"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('diff')"
            v-model="show_diff"
            v-if="needField('hide_diff')"
            data-testid="templateForm-showDiff"
          />

          <v-checkbox
            class="mt-0"
            :label="$t('auto_approve')"
            v-model="item.task_params.allow_auto_approve"
            v-if="needField('allow_auto_approve')"
            data-testid="templateForm-allowAutoApprove"
          />
        </div>

        <CollapsibleSection
          v-if="needField('skip_galaxy_install')"
          v-model="galaxyOpen"
          class="mt-2"
          :title="$t('galaxyInstallOptions')"
          :badge="galaxyBadge"
          data-testid="templateForm-galaxy"
        >
          <v-checkbox
            class="mt-0 mb-4"
            hide-details
            v-model="item.task_params.skip_galaxy_install"
            :label="$t('skipGalaxyInstall')"
            data-testid="templateForm-skipGalaxyInstall"
          />

          <ArgsPicker
            v-if="needField('galaxy_role_args')"
            :vars="item.task_params.galaxy_role_args"
            @change="setGalaxyRoleArgs"
            :title="$t('galaxyRoleArgs')"
            :arg-title="$t('arg')"
            test-id="templateForm-galaxyRoleArgs"
          />

          <ArgsPicker
            v-if="needField('galaxy_collection_args')"
            :vars="item.task_params.galaxy_collection_args"
            @change="setGalaxyCollectionArgs"
            :title="$t('galaxyCollectionArgs')"
            :arg-title="$t('arg')"
            test-id="templateForm-galaxyCollectionArgs"
          />

          <template v-if="needField('allow_override_skip_galaxy_install')">
            <div
              class="text-label-large mb-2"
              data-testid="templateForm-galaxyPromptsTitle"
            >{{ $t('prompts') }}</div>

            <v-checkbox
              class="mt-0"
              hide-details
              v-model="item.task_params.allow_override_skip_galaxy_install"
              :label="$t('skipGalaxyInstall')"
              data-testid="templateForm-allowOverrideSkipGalaxyInstall"
            />
          </template>
        </CollapsibleSection>
      </v-col>
    </v-row>
  </v-form>
</template>
<style lang="scss">
// The columns of the form have no breakpoints: they wrap when their content stops fitting.
// The narrower MD3 type of Vuetify 4 lets a column shrink to ~130px before that happens
// (two squeezed columns on a phone, truncated fields on a tablet); this basis keeps the wrap
// points of Vuetify 3: one column on a phone, the app block below on a tablet when it is wide.
.TemplateForm__column {
  flex-basis: 160px;
}

.cm-placeholder {
  color: #a4a4a4 !important;
}
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import axios from 'axios';

import ItemFormBase from '@/components/ItemFormBase';
import ArgsPicker from '@/components/ArgsPicker.vue';
import TemplateVaults from '@/components/TemplateVaults.vue';
import { TEMPLATE_TYPE_ICONS, TEMPLATE_TYPE_TITLES } from '@/lib/constants';
import AppFieldsMixin from '@/components/AppFieldsMixin';
import AppsMixin from '@/components/AppsMixin';
import RichEditor from '@/components/RichEditor.vue';
import DropdownCard from '@/components/DropdownCard.vue';
import CollapsibleSection from '@/components/CollapsibleSection.vue';
import SurveyVars from './SurveyVars';

export default {
  mixins: [ItemFormBase, AppFieldsMixin, AppsMixin],

  components: {
    DropdownCard,
    CollapsibleSection,
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
      item: {
        task_params: {},
        jwt_params: { enabled: false, audience: [], ttl: '' },
      },
      systemInfo: null,
      inventory: null,
      repositories: null,
      environment: null,
      views: null,
      schedules: null,
      buildTemplates: null,
      cronFormat: '* * * * *',
      cronRepositoryId: null,
      cronVisible: false,
      galaxyOpen: false,

      helpDialog: null,
      helpKey: null,

      args: [],
      runnerTags: null,
      branches: null,
      playbooks: null,
      playbookDir: '',
      playbooksLoading: false,
      playbooksAbort: null,
      setBranch: false,
    };
  },

  watch: {
    formError(err) {
      if (err && err.includes('ansible-galaxy')) {
        this.galaxyOpen = true;
      }
    },

    gitBranchOfTemplate() {
      if (this.playbooks != null) {
        this.playbooks = null;
        this.loadPlaybooks();
      }
    },

    // The dialog sets the app after it opens, so a loaded list can belong to the
    // previous one; the files an app can run differ.
    app() {
      if (this.playbooks != null || this.playbooksLoading) {
        this.playbooks = null;
        this.loadPlaybooks();
      }
    },

    async repositoryId() {
      this.branches = null;
      this.playbooks = null;

      await Promise.all([this.loadBranches()]);
    },

    isMonorepo(val) {
      if (val && this.item) {
        this.item.git_branch = null;
        this.item.allow_override_branch_in_task = false;
      }
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
    galaxyBadge() {
      const n = this.galaxyCustomizedCount;
      return n ? this.$t('galaxyCustomized', { n }) : null;
    },

    galaxyCustomizedCount() {
      const params = this.item.task_params || {};
      return (params.galaxy_role_args || []).length
        + (params.galaxy_collection_args || []).length
        + (params.skip_galaxy_install ? 1 : 0)
        + (params.allow_override_skip_galaxy_install ? 1 : 0);
    },

    showWorkingDirectoryField: {
      get() {
        return this.item?.working_directory != null;
      },
      set(enabled) {
        this.item.working_directory = enabled ? this.item.working_directory || '' : null;
      },
    },

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

    // A monorepo is one checkout shared by every template on the repository's
    // own branch, so a template can neither pin a branch nor let a task pick one.
    isMonorepo() {
      const repo = (this.repositories || []).find((x) => x.id === this.repositoryId);
      return !!repo?.monorepo;
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

    // Stored as hide_dry_run / hide_diff so that existing templates keep
    // showing the checkboxes; presented inverted ("Dry Run" / "Diff" prompt
    // enabled) to match the other prompt checkboxes.
    show_dry_run: {
      get() {
        return !this.item.task_params.hide_dry_run;
      },
      set(newValue) {
        this.item.task_params.hide_dry_run = !newValue;
      },
    },

    show_diff: {
      get() {
        return !this.item.task_params.hide_diff;
      },
      set(newValue) {
        this.item.task_params.hide_diff = !newValue;
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

    // The apps which run a directory list one level at a time: typing the
    // separator asks for what is inside.
    onPlaybookSearch(value) {
      if (!this.fields.playbook?.directories) {
        return;
      }

      // Vuetify resets its search to null after every change; only text the
      // user typed says anything about which directory to list.
      if (value == null) {
        return;
      }

      // Everything up to the last separator, so the listing stays put while the
      // rest of a name is typed.
      const slash = value.lastIndexOf('/');
      const dir = slash < 0 ? '' : value.slice(0, slash + 1);

      if (dir !== this.playbookDir) {
        this.playbookDir = dir;
        this.loadPlaybooks();
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
        const params = new URLSearchParams({
          branch: this.item.git_branch || '',
          app: this.app || '',
          dir: this.playbookDir,
        });

        this.playbooks = await this.loadProjectEndpoint(
          `/repositories/${this.repositoryId}/playbooks?${params}`,
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

    setGalaxyRoleArgs(args) {
      this.item.task_params.galaxy_role_args = args;
    },

    setGalaxyCollectionArgs(args) {
      this.item.task_params.galaxy_collection_args = args;
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
        this.runnerTags,
      ] = await Promise.all([
        this.loadProjectResources('repositories'),
        this.loadProjectEndpoint(`/inventory?app=${this.app}&template_id=${this.itemId}`),
        this.loadProjectEndpoint(`/inventory?app=${this.app}`),
        this.isNew ? [] : this.loadProjectEndpoint(`/templates/${this.itemId}/schedules`),
        this.loadProjectResources('views'),
        this.loadProjectResources('environment'),
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
              builds.push({ type: 'subheader', name: 'Build Templates' });
            }
            builds.push(t);
            break;
          case 'deploy':
            if (deploys.length === 0) {
              deploys.push({ type: 'subheader', name: 'Deploy Templates' });
            }
            deploys.push(t);
            break;
          default:
            break;
        }
      });

      this.buildTemplates = builds;
      if (this.buildTemplates.length > 0 && deploys.length > 0) {
        this.buildTemplates.push({ type: 'divider' });
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

      // Same reactivity concern as executor_image: keys absent in the API response
      // must be declared, otherwise galaxyCustomizedCount never updates.
      if (this.needField('skip_galaxy_install')) {
        const galaxyDefaults = {
          skip_galaxy_install: false,
          allow_override_skip_galaxy_install: false,
          galaxy_role_args: [],
          galaxy_collection_args: [],
        };
        Object.keys(galaxyDefaults).forEach((key) => {
          if (this.item.task_params[key] === undefined) {
            this.item.task_params[key] = galaxyDefaults[key];
          }
        });
      }

      // The API omits executor_image when it is not set; declare it explicitly so
      // the text field stays reactive for templates without an override.
      if (this.item.executor_image === undefined) {
        this.item.executor_image = null;
      }

      if (!this.item.jwt_params) {
        this.item.jwt_params = { enabled: false, audience: [], ttl: '' };
      } else {
        if (!Array.isArray(this.item.jwt_params.audience)) {
          this.item.jwt_params.audience = [];
        }
        if (typeof this.item.jwt_params.ttl !== 'string') {
          this.item.jwt_params.ttl = '';
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
        this.item.environment_ids = [];
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

      // The tabs of Vuetify 2 were mandatory: a type they do not know (a new template has none)
      // fell back to the first tab and wrote it back through the model; Vuetify 3 tabs leave
      // nothing selected.
      const typeIndex = Object.keys(TEMPLATE_TYPE_ICONS).indexOf(this.item.type);
      if (typeIndex < 0) {
        this.item.type = Object.keys(TEMPLATE_TYPE_ICONS)[0];
      }
      this.itemTypeIndex = Math.max(typeIndex, 0);
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
