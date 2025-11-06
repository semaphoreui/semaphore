<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <YesNoDialog
      v-model="deleteProjectDialog"
      :title="$t('deleteProject')"
      :text="$t('askDeleteProj')"
      @yes="deleteProject()"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <DashboardMenu
      :project-id="projectId"
      :project-type="projectType"
      :can-update-project="true"
    />

    <div
      style="margin: auto; max-width: 600px; padding: 0 16px;"
      class="CenterToScreen"
    >
      <h2 class="mt-8 mb-1">{{ $t('general_settings') }}</h2>

      <v-divider class="mb-8" />

      <div class="project-settings-form">
        <div>
          <ProjectForm
            :item-id="projectId"
            ref="form"
            @error="onError"
            @save="onSave"
            :system-info="systemInfo"
          />
        </div>

        <div class="d-flex justify-space-between mt-4">
          <v-btn
            color="blue-grey"
            @click="sendTestNotification()"
            width="170"
            :disabled="testNotificationProgress"
            data-testid="settings-testAlerts"
          >Test Alerts</v-btn>
          <v-btn color="primary" @click="saveProject()">{{ $t('save') }}</v-btn>
        </div>

        <v-progress-linear
          v-if="testNotificationProgress"
          color="blue-grey darken-1"
          indeterminate
          rounded
          width="170"
          height="36"
          style="margin-top: -36px; width: 170px;"
        ></v-progress-linear>
      </div>

      <h2 class="mt-8 mb-1">{{ $t('danger_zone_settings') }}</h2>

      <v-divider class="mb-8" />

      <div class="project-backup project-settings-button" v-if="projectType === ''">
        <v-row align="center">
          <v-col class="shrink">

            <v-btn
              color="primary"
              @click="backupProject"
              :disabled="backupProgress"
              min-width="170"
              data-testid="settings-exportProject"
            >{{ $t('backup') }}
            </v-btn>

            <v-progress-linear
              v-if="backupProgress"
              color="primary accent-4"
              indeterminate
              rounded
              height="36"
              style="margin-top: -36px"
            ></v-progress-linear>

          </v-col>
          <v-col class="grow">
            <div style="font-size: 14px;">
              {{ $t('downloadTheProjectBackupFile') }}
            </div>
          </v-col>
        </v-row>
      </div>

      <div class="project-backup project-settings-button" v-if="projectType === ''">
        <v-row align="center">
          <v-col class="shrink">

            <v-btn
              color="blue-grey"
              @click="clearCache"
              :disabled="clearCacheProgress"
              min-width="170"
              data-testid="settings-clearCache"
            >{{ $t('clear_cache') }}</v-btn>

            <v-progress-linear
              v-if="clearCacheProgress"
              color="blue-grey darken-1"
              indeterminate
              rounded
              height="36"
              style="margin-top: -36px"
            ></v-progress-linear>

          </v-col>
          <v-col class="grow">
            <div style="font-size: 14px">
              {{ $t('clear_cache_message') }}
            </div>
          </v-col>
        </v-row>
      </div>

      <div class="project-delete-form project-settings-button">
        <v-row align="center">
          <v-col class="shrink">
            <v-btn
              color="error"
              min-width="170"
              @click="deleteProjectDialog = true"
              data-testid="settings-deleteProject"
            >{{ $t('deleteProject2') }}
            </v-btn>
          </v-col>
          <v-col class="grow">
            <div style="font-size: 14px; color: #ff5252">
              {{ $t('onceYouDeleteAProjectThereIsNoGoingBackPleaseBeCer') }}
            </div>
          </v-col>
        </v-row>
      </div>
    </div>
  </div>
</template>
<style lang="scss">
  @import '~vuetify/src/styles/styles.sass';

  .project-settings-form {
    //max-width: 600px;
    margin: 30px 0;
  }

  .project-settings-button {
    //max-width: 400px;
    margin: 30px 0;

    @media #{map-get($display-breakpoints, 'sm-and-down')} {
      padding: 0 6px;
    }
  }
</style>
<script>
import EventBus from '@/event-bus';
import ProjectForm from '@/components/ProjectForm.vue';
import { getErrorMessage } from '@/lib/error';
import axios from 'axios';
import YesNoDialog from '@/components/YesNoDialog.vue';
import delay from '@/lib/delay';
import DashboardMenu from '@/components/DashboardMenu.vue';

export default {
  components: { DashboardMenu, YesNoDialog, ProjectForm },
  props: {
    projectId: Number,
    projectType: String,
    systemInfo: Object,
  },

  data() {
    return {
      deleteProjectDialog: null,
      backupProgress: false,
      clearCacheProgress: false,
      testNotificationProgress: false,
    };
  },

  methods: {
    async sendTestNotification() {
      this.testNotificationProgress = true;
      try {
        await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/notifications/test`,
          responseType: 'json',
        });
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'Test notification sent.',
        });
      } catch (err) {
        let msg;
        if (err.response.status === 409) {
          msg = 'Please allow alerts for the project and save it.';
        } else {
          msg = getErrorMessage(err);
        }

        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: msg,
        });
      } finally {
        this.testNotificationProgress = false;
      }
    },

    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    onError(e) {
      EventBus.$emit('i-snackbar', {
        color: 'error',
        text: e.message,
      });
    },

    onSave(e) {
      EventBus.$emit('i-project', {
        action: 'edit',
        item: e.item,
      });
    },

    async saveProject() {
      await this.$refs.form.save();
    },

    async clearCache() {
      this.clearCacheProgress = true;
      await delay(1000);

      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/cache`,
          transformResponse: (res) => res, // Necessary to not parse json
          responseType: 'json',
        });

        await delay(1000);

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'Project cache cleaned.',
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.clearCacheProgress = false;
      }
    },

    async backupProject() {
      this.backupProgress = true;
      await delay(1000);

      try {
        const backup = await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/backup`,
          transformResponse: (res) => res, // Necessary to not parse json
          responseType: 'json',
        });

        const a = document.createElement('a');
        const blob = new Blob([backup.data], { type: 'application/json' });
        a.download = `backup_${this.projectId}_${Date.now()}.json`;
        a.href = URL.createObjectURL(blob);
        a.click();

        await delay(1000);

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'Project exported.',
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.backupProgress = false;
      }
    },

    async deleteProject() {
      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}`,
          responseType: 'json',
        });
        EventBus.$emit('i-project', {
          action: 'delete',
          item: {
            id: this.projectId,
          },
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },
  },
};
</script>
