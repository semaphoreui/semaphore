<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        {{ $t('dashboard2') }}
      </v-toolbar-title>
    </v-toolbar>

    <DashboardMenu
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
      cancel-button-text="OK"
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
            <div>{{ $t('runnerUsage') }}</div>
            <div style="position: relative;">
              <pre style="white-space: pre-wrap;
                          background: gray;
                          color: white;
                          border-radius: 10px;
                          margin-top: 5px;"
                   class="pa-2"
              >{{ runnerUsageCommand }}</pre>

              <v-btn
                style="position: absolute; right: 10px; top: 10px;"
                icon
                color="white"
                @click="copyToClipboard(runnerUsageCommand)"
              >
                <v-icon>mdi-content-copy</v-icon>
              </v-btn>
            </div>
          </div>

          <div>
            <div>{{ $t('runnerDockerCommand') }}</div>
            <div style="position: relative;">
              <pre style="white-space: pre-wrap;
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
          </div>
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
      Project Runners available only in <b>PRO</b> version.
      <v-btn
        class="ml-2"
        color="hsl(348deg, 86%, 61%)"
        href="https://semaphoreui.com/pro"
      >Upgrade</v-btn>
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
    runnerUsageCommand() {
      return `SEMAPHORE_WEB_ROOT=${this.webHost} \\
SEMAPHORE_RUNNER_TOKEN=${(this.newRunner || {}).token} \\
semaphore runner --no-config`;
    },

    runnerDockerCommand() {
      return `docker run \\
-e SEMAPHORE_WEB_ROOT=${this.webHost} \\
-e SEMAPHORE_RUNNER_TOKEN=${(this.newRunner || {}).token} \\
-d semaphoreui/runner:${this.version}`;
    },
  },

  data() {
    return {
      newRunnerTokenDialog: null,
      newRunner: null,
    };
  },

  methods: {
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
