<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <YesNoDialog
      v-model="deleteProjectDialog"
      title="Delete project"
      text="Are you really want to delete this project?"
      @yes="deleteProject()"
    />

    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Dashboard</v-toolbar-title>
      <v-spacer></v-spacer>
      <div>
        <v-tabs centered>
          <v-tab key="history" :to="`/project/${projectId}/history`">History</v-tab>
          <v-tab key="activity" :to="`/project/${projectId}/activity`">Activity</v-tab>
          <v-tab key="settings" :to="`/project/${projectId}/settings`">Settings</v-tab>
        </v-tabs>
      </div>
    </v-toolbar>
    <div class="project-settings-form">
      <div style="height: 300px;">
        <ProjectForm :item-id="projectId" ref="form" @error="onError" @save="onSave"/>
      </div>

      <div class="text-right">
        <v-btn color="primary" @click="saveProject()">Save</v-btn>
      </div>
    </div>

    <div class="project-delete-form">
      <v-row align="center">
        <v-col class="shrink">
          <v-btn color="error" @click="deleteProjectDialog = true">Delete Project</v-btn>
        </v-col>
        <v-col class="grow">
          <div style="font-size: 14px; color: #ff5252">
            Once you delete a project, there is no going back. Please be certain.
          </div>
        </v-col>
      </v-row>
    </div>
  </div>
</template>
<style lang="scss">
  .project-settings-form {
    max-width: 400px;
    margin: 80px auto auto;
  }

  .project-delete-form {
    max-width: 400px;
    margin: 80px auto auto;
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
