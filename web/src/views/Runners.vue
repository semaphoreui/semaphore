<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">

    <v-toolbar flat v-if="projectId">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        {{ $t('dashboard2') }}
      </v-toolbar-title>
    </v-toolbar>
    <v-divider />

    <DashboardMenu
      v-if="projectId"
      :project-id="projectId"
      project-type=""
      :can-update-project="can(USER_PERMISSIONS.updateProject)"
    />

    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('newRunner') : $t('editRunner')"
      @save="loadItemsAndShowRunnerDetails($event)"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <RunnerForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :is-admin="true"
        />
      </template>
    </EditDialog>

    <EditDialog
      :max-width="600"
      v-model="newRunnerTokenDialog"
      :save-button-text="null"
      :title="$t('newRunnerToken')"
      hide-buttons
    >
      <template v-slot:form="{}">
        <div>
          <div class="mb-4">
            <div>{{ $t('runnerToken') }}</div>
            <div style="position: relative;">
              <code
                class="pa-2 mt-2"
                style="background: gray; color: white; display: block; font-size: 14px;"
              >{{ (newRunner || {}).token }}</code>

              <v-btn
                style="position: absolute; right: 10px; top: 2px;"
                icon
                color="white"
                @click="copyToClipboard((newRunner || {}).token)"
              >
                <v-icon>mdi-content-copy</v-icon>
              </v-btn>
            </div>
          </div>

          <div class="mb-4">
            <div>{{ $t('Private Key') }}</div>
            <div style="position: relative;">
              <code
                class="px-2 py-3 mt-2"
                style="background: gray; color: white; display: block; font-size: 14px;"
              >{{ (newRunner || {private_key: ''}).private_key.substring(0, 90) + '...' }}</code>

              <v-btn
                style="position: absolute; right: 10px; top: 2px;"
                icon
                color="white"
              >
                <v-icon
                  @click="downloadFile(newRunner.private_key, 'text/plain', 'config.runner.key')"
                >
                  mdi-download
                </v-icon>
              </v-btn>

              <v-btn
                style="position: absolute; right: 50px; top: 2px;"
                icon
                color="white"
                @click="copyToClipboard((newRunner || {}).private_key)"
              >
                <v-icon>mdi-content-copy</v-icon>
              </v-btn>
            </div>
          </div>

          <h2 class="mt-11 mb-4">Variants of usage</h2>

          <v-tabs v-model="usageTab" :show-arrows="false">
            <v-tab key="config">Config file</v-tab>
            <v-tab key="setup">Setup</v-tab>
            <v-tab key="env">Env Vars</v-tab>
            <v-tab key="docker">Docker</v-tab>
          </v-tabs>

          <v-divider style="margin-top: -1px;"/>

          <v-tabs-items v-model="usageTab">
            <v-tab-item key="config">
              <div class="mt-3">Config file content:</div>
              <div style="position: relative;">
                <pre style="overflow: auto;
                            background: gray;
                            color: white;
                            border-radius: 10px;
                            margin-top: 5px;"
                     class="pa-2"
                >{{ runnerConfigCommand }}</pre>

                <v-btn
                  style="position: absolute; right: 10px; top: 10px;"
                  icon
                  color="white"
                  @click="copyToClipboard(runnerConfigCommand)"
                >
                  <v-icon>mdi-content-copy</v-icon>
                </v-btn>
              </div>

              <div class="mt-3">Launching the runner:</div>
              <div>
                <pre style="overflow: auto;
                  background: gray;
                  color: white;
                  border-radius: 10px;
                  margin-top: 5px;"
                     class="pa-2"
                >semaphore runner start --config /path/to/config/file</pre>
              </div>
            </v-tab-item>
            <v-tab-item key="setup">
              <div class="mt-3">Config file creation:</div>
              <div style="position: relative;">
                <pre style="overflow: auto;
                            background: gray;
                            color: white;
                            border-radius: 10px;
                            margin-top: 5px;"
                     class="pa-2"
                >{{ runnerSetupCommand }}</pre>

                <v-btn
                  style="position: absolute; right: 10px; top: 10px;"
                  icon
                  color="white"
                  @click="copyToClipboard(runnerSetupCommand)"
                >
                  <v-icon>mdi-content-copy</v-icon>
                </v-btn>
              </div>

              <div class="mt-3">
                <div>Launching the runner:</div>
                <pre style="overflow: auto;
                  background: gray;
                  color: white;
                  border-radius: 10px;
                  margin-top: 5px;"
                     class="pa-2"
                >semaphore runner start --config ./config.runner.json</pre>
              </div>
            </v-tab-item>
            <v-tab-item key="env">
              <div class="mt-3">Launching the runner:</div>
              <div style="position: relative;">
                <pre style="overflow: auto;
                            background: gray;
                            color: white;
                            border-radius: 10px;
                            margin-top: 5px;"
                     class="pa-2"
                >{{ runnerEnvCommand }}</pre>

                <v-btn
                  style="position: absolute; right: 10px; top: 10px;"
                  icon
                  color="white"
                  @click="copyToClipboard(runnerEnvCommand)"
                >
                  <v-icon>mdi-content-copy</v-icon>
                </v-btn>
              </div>
            </v-tab-item>

            <v-tab-item key="docker">
              <div class="mt-3">Launching the runner:</div>
              <div style="position: relative;">
                <pre style="overflow: auto;
                            background: gray;
                            color: white;
                            border-radius: 10px;
                            margin-top: 5px;"
                     class="pa-2"
                >{{ runnerDockerCommand }}</pre>

                <v-btn
                  style="position: absolute; right: 10px; top: 10px;"
                  icon
                  color="white"
                  @click="copyToClipboard(runnerDockerCommand)"
                >
                  <v-icon>mdi-content-copy</v-icon>
                </v-btn>
              </div>
            </v-tab-item>
          </v-tabs-items>
        </div>
      </template>
    </EditDialog>

    <YesNoDialog
      :title="$t('deleteRunner')"
      :text="$t('askDeleteRunner', {runner: itemId})"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat v-if="!projectId">
      <v-btn
        icon
        class="mr-4"
        @click="returnToProjects()"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>

      <v-toolbar-title>{{ $t('runners') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >{{ $t('newRunner') }}
      </v-btn>
    </v-toolbar>

    <v-alert
      v-if="!premiumFeatures.project_runners"
      type="info"
      text
      color="hsl(348deg, 86%, 61%)"
      style="border-radius: 0;"
    >
      <span v-if="projectId">
        Project-level runners are only available in the <b>PRO</b> version.
      </span>

      <span v-else>
        The open-source version has limited functionality;
        full functionality is in the <b>PRO</b> version.
      </span>
      <v-btn
        class="ml-2 pr-2"
        color="hsl(348deg, 86%, 61%)"
        href="https://semaphoreui.com/pro"
      >
        Learn more
        <v-icon>mdi-chevron-right</v-icon>
      </v-btn>
    </v-alert>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.active="{ item }">
        <v-switch
          v-model="item.active"
          inset
          @change="setActive(item.id, item.active)"
        ></v-switch>
      </template>

      <template v-slot:item.name="{ item }">{{ item.name || '&mdash;' }}</template>

      <template v-slot:item.webhook="{ item }">{{ item.webhook || '&mdash;' }}</template>

      <template v-slot:item.max_parallel_tasks="{ item }">
        {{ item.max_parallel_tasks || '∞' }}
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn
            icon
            class="mr-1"
            @click="askDeleteItem(item.id)"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn
            icon
            class="mr-1"
            @click="editItem(item.id)"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
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
import DashboardMenu from '@/components/DashboardMenu.vue';
import delay from '@/lib/delay';

export default {
  mixins: [ItemListPageBase],

  components: {
    DashboardMenu,
    RunnerForm,
    YesNoDialog,
    EditDialog,
  },

  props: {
    webHost: String,
    version: String,
    projectId: Number,
    premiumFeatures: Object,
  },

  computed: {
    runnerConfigCommand() {
      return `{
  "web_host": "${this.webHost}",
  "runner": {
    "token": "${(this.newRunner || {}).token}",
    "private_key_file": "/path/to/private/key"
  }
}`;
    },

    runnerSetupCommand() {
      return `cat << EOF > /tmp/config.runner.stdin
${this.webHost}
no
yes
${(this.newRunner || {}).token}
yes
/path/to/private/key
./
EOF

semaphore runner setup --config ./config.runner.json < /tmp/config.runner.stdin`;
    },

    runnerEnvCommand() {
      return `SEMAPHORE_WEB_ROOT=${this.webHost} \\
SEMAPHORE_RUNNER_TOKEN=${(this.newRunner || {}).token} \\
SEMAPHORE_RUNNER_PRIVATE_KEY_FILE=/path/to/private/key \\
semaphore runner start --no-config`;
    },

    runnerDockerCommand() {
      return `docker run \\
-e SEMAPHORE_WEB_ROOT=${this.webHost} \\
-e SEMAPHORE_RUNNER_TOKEN=${(this.newRunner || {}).token} \\
-e SEMAPHORE_RUNNER_PRIVATE_KEY_FILE=/config.runner.key \\
-v "/path/to/private/key:/config.runner.key" \\
-d semaphoreui/runner:${this.version}`;
    },
  },

  data() {
    return {
      newRunnerTokenDialog: null,
      newRunner: null,
      usageTab: null,
    };
  },

  methods: {
    async downloadFile(content, type, name) {
      const a = document.createElement('a');
      const blob = new Blob([content], { type });
      a.download = name;
      a.href = URL.createObjectURL(blob);
      a.click();

      await delay(1000);
    },

    async loadItemsAndShowRunnerDetails(e) {
      if (e.item.token) {
        this.newRunnerTokenDialog = true;
        this.newRunner = e.item;
      }
      return this.loadItems();
    },

    async copyToClipboard(text) {
      try {
        await window.navigator.clipboard.writeText(text);
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'The command has been copied to the clipboard.',
        });
      } catch (e) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: `Can't copy the command: ${e.message}`,
        });
      }
    },

    async setActive(runnerId, active) {
      await axios({
        method: 'post',
        url: `/api/runners/${runnerId}/active`,
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
        }, {
          text: this.$i18n.t('name'),
          value: 'name',
          width: '50%',
        },
        {
          text: this.$i18n.t('webhook'),
          value: 'webhook',
        },
        {
          text: this.$i18n.t('maxNumberOfParallelTasks'),
          value: 'max_parallel_tasks',
        }, {
          text: this.$i18n.t('actions'),
          value: 'actions',
          sortable: false,
        }];
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
      if (this.projectId) {
        return `/api/project/${this.projectId}/runners/${this.itemId}`;
      }

      return `/api/runners/${this.itemId}`;
    },

    getEventName() {
      return 'i-runner';
    },
  },
};
</script>
