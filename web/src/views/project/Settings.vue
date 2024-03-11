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

    <v-tabs show-arrows class="pl-4">
      <v-tab
        v-if="projectType === ''"
        key="history"
        :to="`/project/${projectId}/history`"
      >{{ $t('history') }}</v-tab>
      <v-tab key="activity" :to="`/project/${projectId}/activity`">{{ $t('activity') }}</v-tab>
      <v-tab key="settings" :to="`/project/${projectId}/settings`">{{ $t('settings') }}</v-tab>
    </v-tabs>

    <div class="project-settings-form">
      <div style="height: 300px;">
        <ProjectForm :item-id="projectId" ref="form" @error="onError" @save="onSave"/>
      </div>

      <div class="text-right">
        <v-btn color="primary" @click="saveProject()">{{ $t('save') }}</v-btn>
      </div>
    </div>
    <div class="project-backup project-settings-button">
      <v-row align="center">
        <v-col class="shrink">
          <v-btn color="primary" @click="backupProject" >{{ $t('backup') }}
          </v-btn>
        </v-col>
        <v-col class="grow">
          <div style="font-size: 14px;">
            {{ $t('downloadTheProjectBackupFile') }}
          </div>
        </v-col>
      </v-row>
    </div>
    <div class="project-delete-form project-settings-button">
      <v-row align="center">
        <v-col class="shrink">
          <v-btn color="error" @click="deleteProjectDialog = true">{{ $t('deleteProject2') }}
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
</template>
<style lang="scss">
.project-settings-form {
  max-width: 400px;
  margin: 40px auto;
}
  .project-settings-button {
    max-width: 400px;
    margin: 20px auto auto;
  }
</style>
<script>
import EventBus from '@/event-bus';
import ProjectForm from '@/components/ProjectForm.vue';
import { getErrorMessage } from '@/lib/error';
import axios from 'axios';
import YesNoDialog from '@/components/YesNoDialog.vue';

export default {
  components: { YesNoDialog, ProjectForm },
  props: {
    projectId: Number,
    projectType: String,
  },

  data() {
    return {
      deleteProjectDialog: null,
    };
  },

  methods: {
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

    async backupProject() {
      try {
        await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/backup`,
          transformResponse: (res) => res, // Necessary to not parse json
          responseType: 'json',
        }).then((backup) => {
          const a = document.createElement('a');
          const blob = new Blob([backup.data], { type: 'application/json' });
          a.download = `backup_${this.projectId}_${Date.now()}.json`;
          a.href = URL.createObjectURL(blob);
          a.click();
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
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
