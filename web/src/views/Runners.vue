<template>
  <div
    v-if="items != null"
    data-testid="runners-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <v-toolbar flat v-if="projectId" data-testid="runners-projectToolbar">
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="runners-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title data-testid="runners-title">
        {{ $t('dashboard2') }}
      </v-toolbar-title>
    </v-toolbar>

    <EditDialog
      v-model="editDialog"
      test-id="runnerDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('newRunner') : $t('editRunner')"
      @save="loadItemsAndShowRunnerDetails($event)"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <RunnerForm
          :project-id="projectId || itemProjectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :is-admin="true"
          :is-tags-available="features.project_runners"
        />
      </template>
    </EditDialog>

    <EditDialog
      :max-width="600"
      v-model="newRunnerTokenDialog"
      test-id="runnerTokenDialog"
      :save-button-text="null"
      :title="isUnregisteredRunner ? $t('runnerRegistrationToken') : $t('newRunnerToken')"
      hide-buttons
    >
      <template #form>
        <div v-if="isUnregisteredRunner" data-testid="runnerToken-register">
          <div class="mb-4">
            <div>{{ $t('registrationToken') }}</div>
            <div style="position: relative">
              <code
                class="pa-2 mt-2"
                data-testid="runnerToken-registrationToken"
                style="background: gray; color: white; display: block; font-size: 14px"
              >{{ (newRunner || {}).registration_token }}</code
              >

              <CopyClipboardButton
                :text="(newRunner || {}).registration_token"
                test-id="runnerToken-registrationTokenCopy"
                style="position: absolute; right: 10px; top: 2px"
              />
            </div>
          </div>

          <v-alert
            color="warning"
            density="compact"
            variant="tonal"
            data-testid="runnerToken-registrationHint"
          >
            {{ $t('registrationTokenHint') }}
          </v-alert>

          <h2
            class="mt-8 mb-4"
            data-testid="runnerToken-howToRegister"
          >{{ $t('howToRegister') }}</h2>

          <v-checkbox
              v-model="advancedOptions"
              data-testid="runnerToken-advancedOptions"
              label="Advanced options"
            />

          <HighlightedCard
              v-if="advancedOptions"
              tick-left="80px"
              data-testid="runnerToken-advancedCard"
              style="width: 350px;"
          >
            <v-text-field
                v-model="checkIntervalSeconds"
                data-testid="runnerToken-checkInterval"
                type="number"
                min="1"
                :label="$t('runnerCheckInterval')"
                :hint="$t('runnerCheckIntervalHint')"
                :rules="[checkIntervalRule]"
                persistent-hint
                density="compact"
                variant="outlined"
            />
          </HighlightedCard>

          <v-tabs v-model="registerTab" :show-arrows="false" data-testid="runnerToken-registerTabs">
            <v-tab value="env" data-testid="runnerToken-tab-env">Env Vars</v-tab>
            <v-tab value="config" data-testid="runnerToken-tab-config">Config file</v-tab>
            <v-tab value="docker" data-testid="runnerToken-tab-docker">Docker</v-tab>
          </v-tabs>

          <v-divider style="margin-top: -1px"/>

          <v-window v-model="registerTab">
            <v-window-item value="env">
              <div class="mt-3">Register and start the runner:</div>
              <div style="position: relative">
                <pre
                  class="pa-2"
                  data-testid="runnerToken-registerEnvCommand"
                  style="
                    overflow: auto;
                    background: gray;
                    color: white;
                    border-radius: 10px;
                    margin-top: 5px;
                  "
                >{{ runnerRegisterEnvCommand }}</pre
                >

                <CopyClipboardButton
                  :text="runnerRegisterEnvCommand"
                  test-id="runnerToken-registerEnvCommandCopy"
                  style="position: absolute; right: 10px; top: 10px"
                />
              </div>
            </v-window-item>

            <v-window-item value="config">
              <div class="mt-3">Config file content:</div>
              <div style="position: relative">
                <pre
                  class="pa-2"
                  data-testid="runnerToken-registerConfigContent"
                  style="
                    overflow: auto;
                    background: gray;
                    color: white;
                    border-radius: 10px;
                    margin-top: 5px;
                  "
                >{{ runnerRegisterConfigContent }}</pre
                >

                <CopyClipboardButton
                  :text="runnerRegisterConfigContent"
                  test-id="runnerToken-registerConfigContentCopy"
                  style="position: absolute; right: 10px; top: 10px"
                />
              </div>

              <div class="mt-3">Register and start the runner:</div>
              <div style="position: relative">
                <pre
                  class="pa-2"
                  data-testid="runnerToken-registerConfigCommand"
                  style="
                    overflow: auto;
                    background: gray;
                    color: white;
                    border-radius: 10px;
                    margin-top: 5px;
                  "
                >{{ runnerRegisterConfigCommand }}</pre
                >

                <CopyClipboardButton
                  :text="runnerRegisterConfigCommand"
                  test-id="runnerToken-registerConfigCommandCopy"
                  style="position: absolute; right: 10px; top: 10px"
                />
              </div>
            </v-window-item>

            <v-window-item value="docker">
              <div class="mt-3">Register and start the runner:</div>
              <div style="position: relative">
                <pre
                  class="pa-2"
                  data-testid="runnerToken-registerDockerCommand"
                  style="
                    overflow: auto;
                    background: gray;
                    color: white;
                    border-radius: 10px;
                    margin-top: 5px;
                  "
                >{{ runnerRegisterDockerCommand }}</pre
                >

                <CopyClipboardButton
                  :text="runnerRegisterDockerCommand"
                  test-id="runnerToken-registerDockerCommandCopy"
                  style="position: absolute; right: 10px; top: 10px"
                />
              </div>
            </v-window-item>
          </v-window>
        </div>

        <div v-else data-testid="runnerToken-usage">
          <div class="mb-4">
            <div>{{ $t('runnerToken') }}</div>
            <div style="position: relative">
              <code
                class="pa-2 mt-2"
                data-testid="runnerToken-token"
                style="background: gray; color: white; display: block; font-size: 14px"
              >{{ (newRunner || {}).token }}</code
              >

              <CopyClipboardButton
                style="position: absolute; right: 10px; top: 2px"
                :text="(newRunner || {}).token"
                test-id="runnerToken-tokenCopy"
              />
            </div>
          </div>

          <v-checkbox
              v-model="advancedOptions"
              data-testid="runnerToken-advancedOptions"
              label="Advanced options"
            />

          <HighlightedCard
              v-if="advancedOptions"
              tick-left="80px"
              data-testid="runnerToken-advancedCard"
              style="width: 350px;"
          >
            <v-text-field
                v-model="checkIntervalSeconds"
                data-testid="runnerToken-checkInterval"
                type="number"
                min="1"
                :label="$t('runnerCheckInterval')"
                :hint="$t('runnerCheckIntervalHint')"
                :rules="[checkIntervalRule]"
                persistent-hint
                density="compact"
                variant="outlined"
            />
          </HighlightedCard>

          <h2 class="mt-11 mb-4" data-testid="runnerToken-variantsOfUsage">Variants of usage</h2>

          <v-tabs v-model="usageTab" :show-arrows="false" data-testid="runnerToken-usageTabs">
            <v-tab value="config" data-testid="runnerToken-tab-config">Config file</v-tab>
            <v-tab value="setup" data-testid="runnerToken-tab-setup">Setup</v-tab>
            <v-tab value="env" data-testid="runnerToken-tab-env">Env Vars</v-tab>
            <v-tab value="docker" data-testid="runnerToken-tab-docker">Docker</v-tab>
          </v-tabs>

          <v-divider style="margin-top: -1px"/>

          <v-window v-model="usageTab">
            <v-window-item value="config">
              <div class="mt-3">Config file content:</div>
              <div style="position: relative">
                <pre
                  style="
                    overflow: auto;
                    background: gray;
                    color: white;
                    border-radius: 10px;
                    margin-top: 5px;
                  "
                  class="pa-2"
                  data-testid="runnerToken-configCommand"
                >{{ runnerConfigCommand }}</pre
                >

                <CopyClipboardButton
                  style="position: absolute; right: 10px; top: 10px"
                  :text="runnerConfigCommand"
                  test-id="runnerToken-configCommandCopy"
                />
              </div>

              <div class="mt-3">Launching the runner:</div>
              <div>
                <pre
                  style="
                    overflow: auto;
                    background: gray;
                    color: white;
                    border-radius: 10px;
                    margin-top: 5px;
                  "
                  class="pa-2"
                  data-testid="runnerToken-configStartCommand"
                >
semaphore runner start --config /path/to/config/file</pre
                >
              </div>
            </v-window-item>
            <v-window-item value="setup">
              <div class="mt-3">Config file creation:</div>
              <div style="position: relative">
                <pre
                  style="
                    overflow: auto;
                    background: gray;
                    color: white;
                    border-radius: 10px;
                    margin-top: 5px;
                  "
                  class="pa-2"
                  data-testid="runnerToken-setupCommand"
                >{{ runnerSetupCommand }}</pre
                >

                <CopyClipboardButton
                  style="position: absolute; right: 10px; top: 10px"
                  :text="runnerSetupCommand"
                  test-id="runnerToken-setupCommandCopy"
                />
              </div>

              <div class="mt-3">
                <div>Launching the runner:</div>
                <pre
                  style="
                    overflow: auto;
                    background: gray;
                    color: white;
                    border-radius: 10px;
                    margin-top: 5px;
                  "
                  class="pa-2"
                  data-testid="runnerToken-setupStartCommand"
                >
semaphore runner start --config ./config.runner.json</pre
                >
              </div>
            </v-window-item>
            <v-window-item value="env">
              <div class="mt-3">Launching the runner:</div>
              <div style="position: relative">
                <pre
                  style="
                    overflow: auto;
                    background: gray;
                    color: white;
                    border-radius: 10px;
                    margin-top: 5px;
                  "
                  class="pa-2"
                  data-testid="runnerToken-envCommand"
                >{{ runnerEnvCommand }}</pre
                >

                <CopyClipboardButton
                  style="position: absolute; right: 10px; top: 10px"
                  :text="runnerEnvCommand"
                  test-id="runnerToken-envCommandCopy"
                />
              </div>
            </v-window-item>

            <v-window-item value="docker">
              <div class="mt-3">Launching the runner:</div>
              <div style="position: relative">
                <pre
                  style="
                    overflow: auto;
                    background: gray;
                    color: white;
                    border-radius: 10px;
                    margin-top: 5px;
                  "
                  class="pa-2"
                  data-testid="runnerToken-dockerCommand"
                >{{ runnerDockerCommand }}</pre
                >

                <CopyClipboardButton
                  style="position: absolute; right: 10px; top: 10px"
                  :text="runnerDockerCommand"
                  test-id="runnerToken-dockerCommandCopy"
                />
              </div>
            </v-window-item>
          </v-window>
        </div>
      </template>
    </EditDialog>

    <YesNoDialog
      :title="$t('deleteRunner')"
      :text="$t('askDeleteRunner', { runner: itemId })"
      v-model="deleteItemDialog"
      test-id="runnersDeleteDialog"
      @yes="deleteItem(itemId)"
    />

    <YesNoDialog
      v-model="resetRegistrationDialog"
      test-id="runnersRegenerateDialog"
      :text="$t('askResetRunnerRegistration')"
      :title="$t('regenerateRegistrationToken')"
      @yes="regenerateRegistrationToken(resetRegistrationRunner)"
    />

    <v-toolbar flat v-if="!projectId" data-testid="runners-toolbar">
      <v-btn icon class="mr-4" @click="returnToProjects()" data-testid="runners-back">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>

      <v-toolbar-title
        class="flex-0-1"
        data-testid="runners-title"
      >{{ $t('runners') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        variant="elevated"
        @click="editItem('new')"
        data-testid="runners-new"
      >{{ $t('newRunner') }}</v-btn>
    </v-toolbar>

    <v-btn
      v-else
      :disabled="!features.project_runners"
      style="position: absolute; right: 15px; top: 15px"
      color="primary"
      @click="editItem('new')"
      data-testid="runners-new"
    >{{ $t('newRunner') }}
    </v-btn>

    <v-divider/>

    <v-alert
      v-if="projectId && !features.project_runners"
      variant="tonal"
      color="hsl(348deg, 86%, 61%)"
      class="PageAlert"
      data-testid="runners-licenseAlert"
    >
      <span v-html="$t('project_runners_only_pro')" data-testid="runners-licenseAlertText"></span>
      <v-btn
        v-if="canManageSystem"
        class="ml-2 text-white"
        color="hsl(348deg, 86%, 61%)"
        @click="upgradeToPro('project_runners')"
        data-testid="runners-upgrade"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>
      <span v-else style="font-weight: bold" data-testid="runners-contactAdmin">
        {{ $t('contact_admin_to_upgrade') }}
      </span>
    </v-alert>

    <v-alert
      style="border-radius: 0"
      type="info"
      variant="tonal"
      v-if="!systemInfo.use_remote_runner && projectId == null"
      data-testid="runners-remoteRunnersDisabled"
    >
      Global runners
      <a
        target="_blank"
        data-testid="runners-remoteRunnersDocs"
        href="https://docs.semaphoreui.com/administration-guide/runners/#set-up-a-server"
      >disabled</a
      >.
    </v-alert>

    <div
      v-if="globalFilter || defaultFilter || tagFilter || unregisteredFilter"
      class="mt-4 ml-4 d-flex align-center"
      data-testid="runners-filters"
    >
      <v-chip
        v-if="globalFilter"
        class="mr-2"
        size="small"
        closable
        color="info"
        variant="flat"
        @click:close="globalFilter = false"
        data-testid="runners-filter-global"
      >
        {{ $t('global') }}
      </v-chip>
      <v-chip
        v-if="defaultFilter"
        class="mr-2"
        size="small"
        closable
        color="warning"
        variant="flat"
        @click:close="defaultFilter = false"
        data-testid="runners-filter-default"
      >
        {{ $t('default') }}
      </v-chip>

      <v-chip
        v-if="unregisteredFilter"
        class="mr-2"
        closable
        size="small"
        @click:close="unregisteredFilter = false"
        data-testid="runners-filter-unregistered"
      >
        {{ $t('unregistered') }}
      </v-chip>

      <v-chip
        v-if="tagFilter"
        size="small"
        closable
        label
        color="primary"
        variant="flat"
        @click:close="tagFilter = null"
        data-testid="runners-filter-tag"
      >
        {{ tagFilter }}
      </v-chip>
    </div>

    <v-data-table
      :headers="headers"
      :items="filteredItems"
      class="mt-4 FixedPageSizeTable"
      :style="projectId && !features.project_runners ? 'opacity: 0.4' : ''"
      :items-per-page-options="[20]"
      :items-per-page="20"
      data-testid="runners-table"
      :data-state="filteredItems.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.active="{ item }">
        <v-switch
          v-if="item.project_id != null || projectId == null"
          v-model="item.active"
          inset
          @update:model-value="setActive(item.id, $event)"
          :disabled="item.project_id == null && !canManageSystem"
          data-testid="runners-rowActive"
        />
      </template>

      <template v-slot:item.name="{ item }">
        {{ item.name || '&mdash;' }}
        <v-chip
          v-if="item.is_default"
          class="ml-2"
          size="small"
          color="warning"
          variant="flat"
          style="cursor: pointer"
          @click="defaultFilter = !defaultFilter"
          data-testid="runners-rowDefault"
        >
          {{ $t('default') }}
        </v-chip>

        <v-chip
          v-if="item.project_id == null"
          class="ml-2"
          size="small"
          color="info"
          variant="flat"
          style="cursor: pointer"
          @click="globalFilter = !globalFilter"
          data-testid="runners-rowGlobal"
        >
          {{ $t('global') }}
        </v-chip>

        <v-chip
          v-if="!item.registered"
          class="ml-2"
          size="small"
          style="cursor: pointer"
          @click="unregisteredFilter = !unregisteredFilter"
          data-testid="runners-rowUnregistered"
        >
          {{ $t('unregistered') }}
        </v-chip>
      </template>

      <template v-slot:item.max_parallel_tasks="{ item }">
        {{ item.max_parallel_tasks || '∞' }}
      </template>

      <template v-slot:item.status="{ item }">
        <v-tooltip location="bottom">
          <template v-slot:activator="{ props }">
            <v-chip
              size="small"
              v-bind="props"
              :color="item.status === 'online' ? 'success' : 'blue-grey-lighten-3'"
              variant="flat"
              style="font-weight: bold"
              :data-testid="`runners-rowStatus-${item.status === 'online' ? 'online' : 'offline'}`"
            >
              {{ item.status === 'online' ? $t('online') : $t('offline') }}
            </v-chip>
          </template>
          <div style="font-weight: bold">{{ $t('lastActivity') }}</div>
          <div style="font-size: 12px; line-height: 1.2" data-testid="runners-rowLastActivity">
            <span v-if="item.touched">{{ $filters.formatDate(item.touched) }}</span>
            <span v-else>{{ $t('Never') }}</span>
          </div>
        </v-tooltip>
      </template>

      <template v-slot:item.project_id="{ item }">
        {{ item.project_id ? `#${item.project_id}` : '&mdash;' }}
      </template>

      <template v-slot:item.tags="{ item }">
        <div
          v-if="item.tags && item.tags.length"
          style="white-space: normal"
          data-testid="runners-rowTags"
        >
          <v-chip
            v-for="t in item.tags"
            :key="t"
            size="x-small"
            label
            class="mr-1 mb-1"
            style="cursor: pointer"
            :color="tagFilter === t ? 'primary' : undefined"
            :variant="tagFilter === t ? 'flat' : 'tonal'"
            @click="tagFilter = tagFilter === t ? null : t"
            data-testid="runners-rowTag"
          >
            {{ t }}
          </v-chip>
        </div>
        <span v-else data-testid="runners-rowNoTags">&mdash;</span>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap" data-testid="runners-rowActions">
          <v-tooltip
            v-if="item.project_id != null || projectId == null"
            :max-width="200"
            location="bottom"
          >
            <template v-slot:activator="{ props }">
              <v-btn
                :disabled="item.project_id == null && !canManageSystem"
                class="mr-1"
                icon
                variant="text"
                @click="askRegenerateRegistrationToken(item)"
                data-testid="runners-rowRegenerateToken"
                v-bind="props"
              >
                <v-icon>mdi-sync</v-icon>
              </v-btn>
            </template>
            <div style="font-weight: bold">
              {{ $t('regenerateRegistrationToken') }}
            </div>
            <div v-if="item.registered" style="font-size: 12px; line-height: 1.2">
              {{ $t('askResetRunnerRegistration') }}
            </div>
          </v-tooltip>

          <v-btn
            v-if="item.project_id != null || projectId == null"
            icon
            variant="text"
            class="mr-1"
            @click="askDeleteItem(item.id)"
            data-testid="runners-rowDelete"
            :disabled="item.project_id == null && !canManageSystem"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn
            v-if="item.project_id != null || projectId == null"
            icon
            variant="text"
            class="mr-1"
            @click="editItem(item.id)"
            data-testid="runners-rowEdit"
            :disabled="item.project_id == null && !canManageSystem"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>

          <v-tooltip
            v-if="item.project_id != null || projectId == null"
            location="bottom"
            :max-width="150"
          >
            <template v-slot:activator="{ props }">
              <v-btn
                v-bind="props"
                icon
                variant="text"
                class="mr-1"
                @click="clearCache(item)"
                data-testid="runners-rowClearCache"
                :disabled="item.project_id == null && !canManageSystem"
              >
                <v-icon>mdi-broom</v-icon>
              </v-btn>
            </template>
            <div style="font-weight: bold">
              {{ $t('clear_cache') }}
            </div>

            <div
              v-if="item.cleaning_requested"
              style="font-size: 12px; line-height: 1.2"
              data-testid="runners-rowCleaningInfo"
            >
              <span v-if="item.touched < item.cleaning_requested">
                Already requested {{ $filters.formatDate(item.cleaning_requested) }}.
              </span>
              <span v-else> Last cleaned {{ $filters.formatDate(item.cleaning_requested) }}. </span>
            </div>
          </v-tooltip>
        </div>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import EventBus from '@/event-bus';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ItemListPageBase from '@/components/ItemListPageBase';
import EditDialog from '@/components/EditDialog.vue';
import RunnerForm from '@/components/RunnerForm.vue';
import axios from 'axios';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';
import PageMixin from '@/components/PageMixin';
import HighlightedCard from '@/components/HighlightedCard.vue';
import { GLOBAL_PERMISSIONS } from '@/lib/constants';
import { hasGlobalPermission } from '@/lib/role-permissions';

export default {
  mixins: [ItemListPageBase, PageMixin],

  components: {
    HighlightedCard,
    CopyClipboardButton,
    RunnerForm,
    YesNoDialog,
    EditDialog,
  },

  props: {
    projectId: Number,
  },

  watch: {
    async projectId() {
      await this.loadItems();
    },
  },

  computed: {
    canManageSystem() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageSystem);
    },

    webHost() {
      return this.systemInfo?.web_host || window.location.origin;
    },

    version() {
      return (this.systemInfo?.version || '').split('-')[0];
    },

    itemProjectId() {
      return this.getProjectIdOfItem(this.itemId);
    },

    // Vue keeps an emptied number input as "", which would be interpolated
    // straight into the snippets below and produce invalid JSON. Every snippet
    // reads this instead of the raw field.
    checkInterval() {
      const n = Number(this.checkIntervalSeconds);
      return Number.isInteger(n) && n > 0 ? n : 5;
    },

    isUnregisteredRunner() {
      return !!(this.newRunner || {}).registration_token;
    },

    runnerRegisterEnvCommand() {
      const advancedOptions = this.advancedOptions
        ? `SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=${this.checkInterval} \\\n`
        : '';
      return `SEMAPHORE_WEB_ROOT=${this.webHost} \\
SEMAPHORE_RUNNER_REGISTRATION_TOKEN=${(this.newRunner || {}).registration_token} \\
${advancedOptions}semaphore runner register --config ./config.runner.json

${advancedOptions}semaphore runner start --config ./config.runner.json`;
    },

    runnerRegisterConfigContent() {
      const advancedOptions = this.advancedOptions
        ? `,\n  "runner": {
    "check_interval_seconds": ${this.checkInterval}
  }` : '';

      return `{
  "web_host": "${this.webHost || window.location.origin}"${advancedOptions}
}`;
    },

    runnerRegisterConfigCommand() {
      return `SEMAPHORE_RUNNER_REGISTRATION_TOKEN=${(this.newRunner || {}).registration_token} \\
semaphore runner register --config ./config.runner.json

semaphore runner start --config ./config.runner.json`;
    },

    runnerRegisterDockerCommand() {
      const advancedOptions = this.advancedOptions
        ? `-e SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=${this.checkInterval} \\\n` : '';
      return `docker run \\
-e SEMAPHORE_WEB_ROOT=${this.webHost} \\
-e SEMAPHORE_RUNNER_REGISTRATION_TOKEN=${(this.newRunner || {}).registration_token} \\
${advancedOptions}-d semaphoreui/runner:${this.version}`;
    },

    runnerConfigCommand() {
      const advancedOptions = this.advancedOptions
        ? `,\n    "check_interval_seconds": ${this.checkInterval}` : '';

      return `{
  "web_host": "${this.webHost || window.location.origin}",
  "runner": {
    "token": "${(this.newRunner || {}).token}"${advancedOptions}
  }
}`;
    },

    runnerSetupCommand() {
      return `cat << EOF > /tmp/config.runner.stdin
${this.webHost}
no
yes
${(this.newRunner || {}).token}
./
EOF

semaphore runner setup --config ./config.runner.json < /tmp/config.runner.stdin`;
    },

    runnerEnvCommand() {
      const advancedOptions = this.advancedOptions
        ? `SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=${this.checkInterval} \\\n`
        : '';
      return `SEMAPHORE_WEB_ROOT=${this.webHost} \\
${advancedOptions}SEMAPHORE_RUNNER_TOKEN=${(this.newRunner || {}).token} \\
semaphore runner start --no-config`;
    },

    filteredItems() {
      if (!this.items) {
        return [];
      }
      let result = this.items;
      if (this.globalFilter) {
        result = result.filter((item) => item.project_id == null);
      }
      if (this.defaultFilter) {
        result = result.filter((item) => item.is_default);
      }
      if (this.unregisteredFilter) {
        result = result.filter((item) => !item.registered);
      }
      if (this.tagFilter) {
        result = result.filter((item) => item.tags && item.tags.includes(this.tagFilter));
      }
      return result;
    },

    runnerDockerCommand() {
      const advancedOptions = this.advancedOptions
        ? `-e SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=${this.checkInterval} \\\n`
        : '';
      return `docker run \\
-e SEMAPHORE_WEB_ROOT=${this.webHost} \\
-e SEMAPHORE_RUNNER_TOKEN=${(this.newRunner || {}).token} \\
${advancedOptions}-d semaphoreui/runner:${this.version}`;
    },
  },

  data() {
    return {
      newRunnerTokenDialog: null,
      newRunner: null,
      // Mirrors the runner-side default in util.RunnerConfig.CheckIntervalSeconds.
      // Only shapes the snippets below; the server does not store it.
      checkIntervalSeconds: 5,
      usageTab: null,
      registerTab: null,
      globalFilter: false,
      defaultFilter: false,
      tagFilter: null,
      unregisteredFilter: false,
      resetRegistrationDialog: false,
      resetRegistrationRunner: null,
      advancedOptions: null,
    };
  },

  methods: {
    checkIntervalRule(v) {
      const n = Number(v);
      return (Number.isInteger(n) && n > 0) || this.$t('runnerCheckIntervalInvalid');
    },

    async clearCache(runner) {
      const projectId = this.getProjectIdOfItem(runner.id);

      const url = projectId
        ? `/api/project/${projectId}/runners/${runner.id}/cache`
        : `/api/runners/${runner.id}/cache`;

      try {
        await axios({
          method: 'delete',
          url,
          responseType: 'json',
        });
        await this.loadItems();
      } catch (e) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: `Cannot clear cache: ${e.message}`,
        });
      }
    },

    askRegenerateRegistrationToken(runner) {
      // Regenerating for an already-registered runner is destructive (it resets the
      // runner's token and forces it offline), so confirm first. For an unregistered
      // runner there is nothing to lose, so do it right away.
      if (runner.registered) {
        this.resetRegistrationRunner = runner;
        this.resetRegistrationDialog = true;
      } else {
        this.regenerateRegistrationToken(runner);
      }
    },

    async regenerateRegistrationToken(runner) {
      const projectId = this.getProjectIdOfItem(runner.id);

      const url = projectId
        ? `/api/project/${projectId}/runners/${runner.id}/registration-token`
        : `/api/runners/${runner.id}/registration-token`;

      try {
        const { data } = await axios({
          method: 'post',
          url,
          responseType: 'json',
        });

        // Reuse the same dialog shown right after creating an unregistered runner.
        this.newRunner = { ...runner, registration_token: data.registration_token };
        this.registerTab = null;
        this.newRunnerTokenDialog = true;

        // The runner's registration state may have changed (registered -> unregistered).
        await this.loadItems();
      } catch (e) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: `Cannot regenerate registration token: ${e.message}`,
        });
      }
    },

    getProjectIdOfItem(itemId) {
      if (!itemId || itemId === 'new') {
        return null;
      }

      const item = this.items.find((x) => x.id === itemId);
      if (item) {
        return item.project_id;
      }

      return null;
    },

    async loadItemsAndShowRunnerDetails(e) {
      if (e.item.token || e.item.registration_token) {
        // A registered runner returns an auth token; show the details dialog with
        // the connection instructions.
        this.newRunnerTokenDialog = true;
        this.newRunner = e.item;
      } else if (e.action === 'new') {
        // An unregistered runner is created with no token at all, so there is
        // nothing to show — just confirm it was created.
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.$t('runnerCreated'),
        });
      }
      return this.loadItems();
    },

    async setActive(runnerId, active) {
      const projectId = this.getProjectIdOfItem(runnerId);

      const url = projectId
        ? `/api/project/${projectId}/runners/${runnerId}/active`
        : `/api/runners/${runnerId}/active`;

      await axios({
        method: 'post',
        url,
        responseType: 'json',
        data: {
          active,
        },
      });
    },

    getHeaders() {
      return [
        {
          key: 'active',
        },
        {
          title: this.$t('name'),
          key: 'name',
          width: '50%',
        },
        ...(this.projectId
          ? []
          : [
            {
              title: this.$t('project'),
              key: 'project_id',
            },
          ]),
        {
          title: this.$t('tag'),
          key: 'tags',
          sortable: false,
        },
        {
          title: this.$t('status'),
          key: 'status',
        },
        {
          title: this.$t('actions'),
          key: 'actions',
          sortable: false,
          align: 'end',
        },
      ];
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    getItemsUrl() {
      if (this.projectId) {
        return `/api/project/${this.projectId}/runners`;
      }

      return '/api/runners';
    },

    getSingleItemUrl() {
      const projectId = this.getProjectIdOfItem(this.itemId);

      if (projectId) {
        return `/api/project/${projectId}/runners/${this.itemId}`;
      }

      return `/api/runners/${this.itemId}`;
    },

    getEventName() {
      return 'i-runner';
    },
  },
};
</script>
