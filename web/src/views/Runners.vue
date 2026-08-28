<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <v-toolbar flat v-if="projectId">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        {{ $t('dashboard2') }}
      </v-toolbar-title>
    </v-toolbar>

    <EditDialog
      v-model="editDialog"
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
      :save-button-text="null"
      :title="isUnregisteredRunner ? $t('runnerRegistrationToken') : $t('newRunnerToken')"
      hide-buttons
    >
      <template v-slot:form="{}">
        <div v-if="isUnregisteredRunner">
          <div class="mb-4">
            <div>{{ $t('registrationToken') }}</div>
            <div style="position: relative">
              <code
                class="pa-2 mt-2"
                style="background: gray; color: white; display: block; font-size: 14px"
              >{{ (newRunner || {}).registration_token }}</code
              >

              <CopyClipboardButton
                :text="(newRunner || {}).registration_token"
                style="position: absolute; right: 10px; top: 2px"
              />
            </div>
          </div>

          <v-alert color="warning" dense text>
            {{ $t('registrationTokenHint') }}
          </v-alert>

          <v-text-field
            v-model="checkIntervalSeconds"
            type="number"
            min="1"
            class="mt-6"
            style="max-width: 320px"
            :label="$t('runnerCheckInterval')"
            :hint="$t('runnerCheckIntervalHint')"
            :rules="[checkIntervalRule]"
            persistent-hint
            dense
          />

          <h2 class="mt-8 mb-4">{{ $t('howToRegister') }}</h2>

          <v-tabs v-model="registerTab" :show-arrows="false">
            <v-tab key="env">Env Vars</v-tab>
            <v-tab key="config">Config file</v-tab>
            <v-tab key="docker">Docker</v-tab>
          </v-tabs>

          <v-divider style="margin-top: -1px"/>

          <v-tabs-items v-model="registerTab">
            <v-tab-item key="env">
              <div class="mt-3">Register and start the runner:</div>
              <div style="position: relative">
                <pre
                  class="pa-2"
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
                  style="position: absolute; right: 10px; top: 10px"
                />
              </div>
            </v-tab-item>

            <v-tab-item key="config">
              <div class="mt-3">Config file content:</div>
              <div style="position: relative">
                <pre
                  class="pa-2"
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
                  style="position: absolute; right: 10px; top: 10px"
                />
              </div>

              <div class="mt-3">Register and start the runner:</div>
              <div style="position: relative">
                <pre
                  class="pa-2"
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
                  style="position: absolute; right: 10px; top: 10px"
                />
              </div>
            </v-tab-item>

            <v-tab-item key="docker">
              <div class="mt-3">Register and start the runner:</div>
              <div style="position: relative">
                <pre
                  class="pa-2"
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
                  style="position: absolute; right: 10px; top: 10px"
                />
              </div>
            </v-tab-item>
          </v-tabs-items>
        </div>

        <div v-else>
          <div class="mb-4">
            <div>{{ $t('runnerToken') }}</div>
            <div style="position: relative">
              <code
                class="pa-2 mt-2"
                style="background: gray; color: white; display: block; font-size: 14px"
              >{{ (newRunner || {}).token }}</code
              >

              <CopyClipboardButton
                style="position: absolute; right: 10px; top: 2px"
                :text="(newRunner || {}).token"
              />
            </div>
          </div>

          <v-text-field
            v-model="checkIntervalSeconds"
            type="number"
            min="1"
            class="mt-6"
            style="max-width: 320px"
            :label="$t('runnerCheckInterval')"
            :hint="$t('runnerCheckIntervalHint')"
            :rules="[checkIntervalRule]"
            persistent-hint
            dense
          />

          <h2 class="mt-11 mb-4">Variants of usage</h2>

          <v-tabs v-model="usageTab" :show-arrows="false">
            <v-tab key="config">Config file</v-tab>
            <v-tab key="setup">Setup</v-tab>
            <v-tab key="env">Env Vars</v-tab>
            <v-tab key="docker">Docker</v-tab>
          </v-tabs>

          <v-divider style="margin-top: -1px"/>

          <v-tabs-items v-model="usageTab">
            <v-tab-item key="config">
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
                >{{ runnerConfigCommand }}</pre
                >

                <CopyClipboardButton
                  style="position: absolute; right: 10px; top: 10px"
                  :text="runnerConfigCommand"
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
                >
semaphore runner start --config /path/to/config/file</pre
                >
              </div>
            </v-tab-item>
            <v-tab-item key="setup">
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
                >{{ runnerSetupCommand }}</pre
                >

                <CopyClipboardButton
                  style="position: absolute; right: 10px; top: 10px"
                  :text="runnerSetupCommand"
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
                >
semaphore runner start --config ./config.runner.json</pre
                >
              </div>
            </v-tab-item>
            <v-tab-item key="env">
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
                >{{ runnerEnvCommand }}</pre
                >

                <CopyClipboardButton
                  style="position: absolute; right: 10px; top: 10px"
                  :text="runnerEnvCommand"
                />
              </div>
            </v-tab-item>

            <v-tab-item key="docker">
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
                >{{ runnerDockerCommand }}</pre
                >

                <CopyClipboardButton
                  style="position: absolute; right: 10px; top: 10px"
                  :text="runnerDockerCommand"
                />
              </div>
            </v-tab-item>
          </v-tabs-items>
        </div>
      </template>
    </EditDialog>

    <YesNoDialog
      :title="$t('deleteRunner')"
      :text="$t('askDeleteRunner', { runner: itemId })"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <YesNoDialog
      v-model="resetRegistrationDialog"
      :text="$t('askResetRunnerRegistration')"
      :title="$t('regenerateRegistrationToken')"
      @yes="regenerateRegistrationToken(resetRegistrationRunner)"
    />

    <v-toolbar flat v-if="!projectId">
      <v-btn icon class="mr-4" @click="returnToProjects()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>

      <v-toolbar-title>{{ $t('runners') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn color="primary" @click="editItem('new')">{{ $t('newRunner') }}</v-btn>
    </v-toolbar>

    <v-btn
      v-else
      :disabled="!features.project_runners"
      style="position: absolute; right: 15px; top: 15px"
      color="primary"
      @click="editItem('new')"
    >{{ $t('newRunner') }}
    </v-btn>

    <v-divider/>

    <v-alert
      v-if="projectId && !features.project_runners"
      text
      color="hsl(348deg, 86%, 61%)"
      class="PageAlert"
    >
      <span v-html="$t('project_runners_only_pro')"></span>
      <v-btn
        dark
        v-if="isAdmin"
        class="ml-2"
        color="hsl(348deg, 86%, 61%)"
        @click="upgradeToPro('project_runners')"
      >
        {{ $t('upgrade_to_pro') }}
      </v-btn>
      <span v-else style="font-weight: bold">
        {{ $t('contact_admin_to_upgrade') }}
      </span>
    </v-alert>

    <v-alert
      style="border-radius: 0"
      type="info"
      text
      v-if="!systemInfo.use_remote_runner && projectId == null"
    >
      Global runners
      <a
        target="_blank"
        href="https://docs.semaphoreui.com/administration-guide/runners/#set-up-a-server"
      >disabled</a
      >.
    </v-alert>

    <div
      v-if="globalFilter || defaultFilter || tagFilter || unregisteredFilter"
      class="mt-4 ml-4 d-flex align-center"
    >
      <v-chip
        v-if="globalFilter"
        class="mr-2"
        small
        close
        color="info"
        @click:close="globalFilter = false"
      >
        {{ $t('global') }}
      </v-chip>
      <v-chip
        v-if="defaultFilter"
        class="mr-2"
        small
        close
        color="warning"
        @click:close="defaultFilter = false"
      >
        {{ $t('default') }}
      </v-chip>

      <v-chip
        v-if="unregisteredFilter"
        class="mr-2"
        close
        small
        @click:close="unregisteredFilter = false"
      >
        {{ $t('unregistered') }}
      </v-chip>

      <v-chip v-if="tagFilter" small close label color="primary" @click:close="tagFilter = null">
        {{ tagFilter }}
      </v-chip>
    </div>

    <v-data-table
      :headers="headers"
      :items="filteredItems"
      class="mt-4"
      :style="projectId && !features.project_runners ? 'opacity: 0.4' : ''"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.active="{ item }">
        <v-switch
          v-if="item.project_id != null || projectId == null"
          v-model="item.active"
          inset
          @change="setActive(item.id, item.active)"
          :disabled="item.project_id == null && !isAdmin"
        />
      </template>

      <template v-slot:item.name="{ item }">
        {{ item.name || '&mdash;' }}
        <v-chip
          v-if="item.is_default"
          class="ml-2"
          small
          color="warning"
          style="cursor: pointer"
          @click="defaultFilter = !defaultFilter"
        >
          {{ $t('default') }}
        </v-chip>

        <v-chip
          v-if="item.project_id == null"
          class="ml-2"
          small
          color="info"
          style="cursor: pointer"
          @click="globalFilter = !globalFilter"
        >
          {{ $t('global') }}
        </v-chip>

        <v-chip
          v-if="!item.registered"
          class="ml-2"
          small
          style="cursor: pointer"
          @click="unregisteredFilter = !unregisteredFilter"
        >
          {{ $t('unregistered') }}
        </v-chip>
      </template>

      <template v-slot:item.max_parallel_tasks="{ item }">
        {{ item.max_parallel_tasks || '∞' }}
      </template>

      <template v-slot:item.status="{ item }">
        <v-tooltip bottom>
          <template v-slot:activator="{ on, attrs }">
            <v-chip
              small
              v-bind="attrs"
              v-on="on"
              :color="item.status === 'online' ? 'success' : 'blue-grey lighten-3'"
              style="font-weight: bold"
            >
              {{ item.status === 'online' ? $t('online') : $t('offline') }}
            </v-chip>
          </template>
          <div style="font-weight: bold">{{ $t('lastActivity') }}</div>
          <div style="font-size: 12px; line-height: 1.2">
            <span v-if="item.touched">{{ item.touched | formatDate }}</span>
            <span v-else>{{ $t('Never') }}</span>
          </div>
        </v-tooltip>
      </template>

      <template v-slot:item.project_id="{ item }">
        {{ item.project_id ? `#${item.project_id}` : '&mdash;' }}
      </template>

      <template v-slot:item.tags="{ item }">
        <div v-if="item.tags && item.tags.length" style="white-space: normal">
          <v-chip
            v-for="t in item.tags"
            :key="t"
            x-small
            label
            class="mr-1 mb-1"
            style="cursor: pointer"
            :color="tagFilter === t ? 'primary' : undefined"
            :dark="tagFilter === t"
            @click="tagFilter = tagFilter === t ? null : t"
          >
            {{ t }}
          </v-chip>
        </div>
        <span v-else>&mdash;</span>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-tooltip
            v-if="item.project_id != null || projectId == null"
            :max-width="200"
            bottom
          >
            <template v-slot:activator="{ on, attrs }">
              <v-btn
                :disabled="item.project_id == null && !isAdmin"
                class="mr-1"
                icon
                v-bind="attrs"
                @click="askRegenerateRegistrationToken(item)"
                v-on="on"
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
            class="mr-1"
            @click="askDeleteItem(item.id)"
            :disabled="item.project_id == null && !isAdmin"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn
            v-if="item.project_id != null || projectId == null"
            icon
            class="mr-1"
            @click="editItem(item.id)"
            :disabled="item.project_id == null && !isAdmin"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>

          <v-tooltip v-if="item.project_id != null || projectId == null" bottom :max-width="150">
            <template v-slot:activator="{ on, attrs }">
              <v-btn
                v-bind="attrs"
                v-on="on"
                icon
                class="mr-1"
                @click="clearCache(item)"
                :disabled="item.project_id == null && !isAdmin"
              >
                <v-icon>mdi-broom</v-icon>
              </v-btn>
            </template>
            <div style="font-weight: bold">
              {{ $t('clear_cache') }}
            </div>

            <div v-if="item.cleaning_requested" style="font-size: 12px; line-height: 1.2">
              <span v-if="item.touched < item.cleaning_requested">
                Already requested {{ item.cleaning_requested | formatDate }}.
              </span>
              <span v-else> Last cleaned {{ item.cleaning_requested | formatDate }}. </span>
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

export default {
  mixins: [ItemListPageBase, PageMixin],

  components: {
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
      const n = parseInt(this.checkIntervalSeconds, 10);
      return Number.isInteger(n) && n > 0 ? n : 1;
    },

    isUnregisteredRunner() {
      return !!(this.newRunner || {}).registration_token;
    },

    runnerRegisterEnvCommand() {
      return `SEMAPHORE_WEB_ROOT=${this.webHost} \\
SEMAPHORE_RUNNER_REGISTRATION_TOKEN=${(this.newRunner || {}).registration_token} \\
SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=${this.checkInterval} \\
semaphore runner register --config ./config.runner.json

semaphore runner start --config ./config.runner.json`;
    },

    runnerRegisterConfigContent() {
      return `{
  "web_host": "${this.webHost || window.location.origin}",
  "runner": {
    "check_interval_seconds": ${this.checkInterval}
  }
}`;
    },

    runnerRegisterConfigCommand() {
      return `SEMAPHORE_RUNNER_REGISTRATION_TOKEN=${(this.newRunner || {}).registration_token} \\
semaphore runner register --config ./config.runner.json

semaphore runner start --config ./config.runner.json`;
    },

    runnerRegisterDockerCommand() {
      return `docker run \\
-e SEMAPHORE_WEB_ROOT=${this.webHost} \\
-e SEMAPHORE_RUNNER_REGISTRATION_TOKEN=${(this.newRunner || {}).registration_token} \\
-e SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=${this.checkInterval} \\
-d semaphoreui/runner:${this.version}`;
    },

    runnerConfigCommand() {
      return `{
  "web_host": "${this.webHost || window.location.origin}",
  "runner": {
    "token": "${(this.newRunner || {}).token}",
    "check_interval_seconds": ${this.checkInterval}
  }
}`;
    },

    runnerSetupCommand() {
      return `cat << EOF > /tmp/config.runner.stdin
${this.webHost}
${this.checkInterval}
no
yes
${(this.newRunner || {}).token}
./
EOF

semaphore runner setup --config ./config.runner.json < /tmp/config.runner.stdin`;
    },

    runnerEnvCommand() {
      return `SEMAPHORE_WEB_ROOT=${this.webHost} \\
SEMAPHORE_RUNNER_TOKEN=${(this.newRunner || {}).token} \\
SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=${this.checkInterval} \\
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
      return `docker run \\
-e SEMAPHORE_WEB_ROOT=${this.webHost} \\
-e SEMAPHORE_RUNNER_TOKEN=${(this.newRunner || {}).token} \\
-e SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=${this.checkInterval} \\
-d semaphoreui/runner:${this.version}`;
    },
  },

  data() {
    return {
      newRunnerTokenDialog: null,
      newRunner: null,
      // Mirrors the runner-side default in util.RunnerConfig.CheckIntervalSeconds.
      // Only shapes the snippets below; the server does not store it.
      checkIntervalSeconds: 1,
      usageTab: null,
      registerTab: null,
      globalFilter: false,
      defaultFilter: false,
      tagFilter: null,
      unregisteredFilter: false,
      resetRegistrationDialog: false,
      resetRegistrationRunner: null,
    };
  },

  methods: {
    checkIntervalRule(v) {
      const n = parseInt(v, 10);
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
          value: 'active',
        },
        {
          text: this.$i18n.t('name'),
          value: 'name',
          width: '50%',
        },
        ...(this.projectId
          ? []
          : [
            {
              text: this.$i18n.t('project'),
              value: 'project_id',
            },
          ]),
        {
          text: this.$i18n.t('tag'),
          value: 'tags',
          sortable: false,
        },
        {
          text: this.$i18n.t('status'),
          value: 'status',
        },
        {
          text: this.$i18n.t('actions'),
          value: 'actions',
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
