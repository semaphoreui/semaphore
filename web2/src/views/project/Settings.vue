<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <v-dialog
      v-model="deleteProjectDialog"
      max-width="290">
      <v-card>
        <v-card-title class="headline">Delete project</v-card-title>

        <v-card-text>
          Are you really want to delete this project?
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>

          <v-btn
            color="blue darken-1"
            text
            @click="deleteProjectDialog = false"
          >
            Cancel
          </v-btn>

          <v-btn
            color="blue darken-1"
            text
            @click="deleteProject()"
          >
            Yes
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

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
      <div style="height: 220px;">
        <ProjectEditForm :project-id="projectId" ref="editForm"/>
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
import ProjectEditForm from '@/components/ProjectEditForm.vue';
import { getErrorMessage } from '@/lib/error';
import axios from 'axios';

export default {
  components: { ProjectEditForm },
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

    async saveProject() {
      const item = await this.$refs.editForm.save();
      if (!item) {
        return;
      }
      EventBus.$emit('i-project', {
        action: 'edit',
        item,
      });
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
