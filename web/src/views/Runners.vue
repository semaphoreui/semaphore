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
      :title="$t('newRunnerToken')"
      hide-buttons
    >
      <template v-slot:form="{}">
        <div>
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

          <div class="mb-4">
            <div>{{ $t('Private Key') }}</div>
            <div style="position: relative">
              <code
                class="px-2 py-3 mt-2"
                style="background: gray; color: white; display: block; font-size: 14px"
                >{{ (newRunner || { private_key: '' }).private_key.substring(0, 90) + '...' }}</code
              >

              <v-btn style="position: absolute; right: 10px; top: 2px" icon color="white">
                <v-icon
                  @click="downloadFile(newRunner.private_key, 'text/plain', 'config.runner.key')"
                >
                  mdi-download
                </v-icon>
              </v-btn>

              <CopyClipboardButton
                style="position: absolute; right: 50px; top: 2px"
                :text="(newRunner || {}).private_key"
              />
            </div>
          </div>

          <h2 class="mt-11 mb-4">Variants of usage</h2>

          <v-tabs v-model="usageTab" :show-arrows="false">
            <v-tab key="config">Config file</v-tab>
            <v-tab key="setup">Setup</v-tab>
            <v-tab key="env">Env Vars</v-tab>
            <v-tab key="docker">Docker</v-tab>
          </v-tabs>

          <v-divider style="margin-top: -1px" />

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

    <v-toolbar flat v-if="!projectId">
      <v-btn icon class="mr-4" @click="returnToProjects()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>

      <v-toolbar-title>{{ $t('runners') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn color="primary" @click="editItem('new')">{{ $t('newRunner') }} </v-btn>
    </v-toolbar>

    <v-btn
      v-else
      :disabled="!features.project_runners"
      style="position: absolute; right: 15px; top: 15px"
      color="primary"
      @click="editItem('new')"
      >{{ $t('newRunner') }}
    </v-btn>

    <v-divider />

    <v-alert
      v-if="projectId && !features.project_runners"
      text
      color="hsl(348deg, 86%, 61%)"
      class="PageAlert"
    >
      <span v-html="$t('project_runners_only_pro')"></span>
      <v-btn dark v-if="isAdmin" class="ml-2" color="hsl(348deg, 86%, 61%)" @click="upgradeToPro()">
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

    <div v-if="globalFilter || defaultFilter || tagFilter" class="mt-4 ml-4 d-flex align-center">
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
      <v-chip v-if="tagFilter" small close label color="primary" @click:close="tagFilter = null">
        {{ tagFilter }}
      </v-chip>
    </div>

    <v-data-table
      :headers="headers"
      :items="filteredItems"
      class="mt-4"
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
      </template>

      <template v-slot:item.max_parallel_tasks="{ item }">
        {{ item.max_parallel_tasks || '∞' }}
      </template>

      <template v-slot:item.touched="{ item }">
        <v-chip :color="getStatusColor(item)" style="font-weight: bold">
          <span v-if="item.touched">{{ item.touched | formatDate }}</span>
          <span v-else>{{ $t('Never') }}</span>
        </v-chip>
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
import delay from '@/lib/delay';
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

    runnerConfigCommand() {
      return `{
  "web_host": "${this.webHost || window.location.origin}",
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
      if (this.tagFilter) {
        result = result.filter((item) => item.tags && item.tags.includes(this.tagFilter));
      }
      return result;
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
      globalFilter: false,
      defaultFilter: false,
      tagFilter: null,
    };
  },

  methods: {
    upgradeToPro() {
      EventBus.$emit('i-subscription', {});
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

    getStatusColor(runner) {
      if (!runner.touched) {
        return 'blue-grey lighten-3';
      }

      const d = Date.now() - new Date(runner.touched);

      if (d < 1000 * 60 * 5) {
        return 'success';
      }

      if (d < 1000 * 60 * 60) {
        return 'warning';
      }

      return 'blue-grey lighten-3';
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
          text: this.$i18n.t('activity'),
          value: 'touched',
        },
        {
          text: this.$i18n.t('actions'),
          value: 'actions',
          sortable: false,
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
