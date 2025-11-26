<template>
  <div>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="$emit('toggle-drawer')"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('workflows') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="createWorkflow"
      >
        <v-icon left>mdi-plus</v-icon>
        {{ $t('newWorkflow') }}
      </v-btn>
    </v-toolbar>

    <v-container fluid>
      <v-row>
        <v-col
          v-for="workflow in workflows"
          :key="workflow.id"
          cols="12"
          sm="6"
          md="4"
        >
          <v-card @click="openWorkflow(workflow)">
            <v-card-title>{{ workflow.name }}</v-card-title>
            <v-card-subtitle v-if="workflow.description">
              {{ workflow.description }}
            </v-card-subtitle>
            <v-card-text>
              <div class="text--secondary">
                {{ workflow.nodes ? workflow.nodes.length : 0 }} nodes
              </div>
              <div class="text--secondary text-caption">
                Updated: {{ formatDate(workflow.updated_at) }}
              </div>
            </v-card-text>
            <v-card-actions>
              <v-btn
                text
                color="primary"
                @click.stop="editWorkflow(workflow)"
              >
                <v-icon left>mdi-pencil</v-icon>
                Edit
              </v-btn>
              <v-btn
                text
                color="success"
                @click.stop="runWorkflow(workflow)"
              >
                <v-icon left>mdi-play</v-icon>
                Run
              </v-btn>
              <v-spacer></v-spacer>
              <v-btn
                icon
                @click.stop="deleteWorkflowConfirm(workflow)"
              >
                <v-icon>mdi-delete</v-icon>
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>

      <v-row v-if="workflows.length === 0">
        <v-col>
          <v-alert type="info" outlined>
            {{ $t('noWorkflows') }}
          </v-alert>
        </v-col>
      </v-row>
    </v-container>

    <!-- Create/Edit Dialog -->
    <v-dialog v-model="dialog" max-width="600">
      <v-card>
        <v-card-title>
          {{ editingWorkflow ? $t('editWorkflow') : $t('createWorkflow') }}
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="workflowForm.name"
            :label="$t('name')"
            required
          ></v-text-field>
          <v-textarea
            v-model="workflowForm.description"
            :label="$t('description')"
            rows="3"
          ></v-textarea>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn text @click="dialog = false">{{ $t('cancel') }}</v-btn>
          <v-btn color="primary" @click="saveWorkflow">{{ $t('save') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title>{{ $t('deleteWorkflow') }}</v-card-title>
        <v-card-text>
          {{ $t('deleteWorkflowConfirm') }}
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn text @click="deleteDialog = false">{{ $t('cancel') }}</v-btn>
          <v-btn color="error" @click="deleteWorkflow">{{ $t('delete') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  data() {
    return {
      workflows: [],
      dialog: false,
      deleteDialog: false,
      editingWorkflow: null,
      workflowToDelete: null,
      workflowForm: {
        name: '',
        description: '',
      },
    };
  },
  mounted() {
    this.loadWorkflows();
  },
  methods: {
    async loadWorkflows() {
      try {
        const { data } = await axios.get(`/api/project/${this.projectId}/workflows`);
        this.workflows = data || [];
      } catch (error) {
        console.error('Failed to load workflows:', error);
      }
    },
    createWorkflow() {
      this.editingWorkflow = null;
      this.workflowForm = {
        name: '',
        description: '',
      };
      this.dialog = true;
    },
    editWorkflow(workflow) {
      this.editingWorkflow = workflow;
      this.workflowForm = {
        name: workflow.name,
        description: workflow.description || '',
      };
      this.dialog = true;
    },
    async saveWorkflow() {
      try {
        if (this.editingWorkflow) {
          await axios.put(
            `/api/project/${this.projectId}/workflows/${this.editingWorkflow.id}`,
            {
              ...this.editingWorkflow,
              ...this.workflowForm,
            },
          );
        } else {
          const { data } = await axios.post(
            `/api/project/${this.projectId}/workflows`,
            this.workflowForm,
          );
          this.$router.push(`/project/${this.projectId}/workflows/${data.id}/editor`);
        }
        this.dialog = false;
        this.loadWorkflows();
      } catch (error) {
        console.error('Failed to save workflow:', error);
      }
    },
    openWorkflow(workflow) {
      this.$router.push(`/project/${this.projectId}/workflows/${workflow.id}/editor`);
    },
    async runWorkflow(workflow) {
      try {
        await axios.post(`/api/project/${this.projectId}/workflows/${workflow.id}/run`);
        this.$router.push(`/project/${this.projectId}/workflows/${workflow.id}/runs`);
      } catch (error) {
        console.error('Failed to run workflow:', error);
      }
    },
    deleteWorkflowConfirm(workflow) {
      this.workflowToDelete = workflow;
      this.deleteDialog = true;
    },
    async deleteWorkflow() {
      try {
        await axios.delete(
          `/api/project/${this.projectId}/workflows/${this.workflowToDelete.id}`,
        );
        this.deleteDialog = false;
        this.loadWorkflows();
      } catch (error) {
        console.error('Failed to delete workflow:', error);
      }
    },
    formatDate(date) {
      return new Date(date).toLocaleDateString();
    },
  },
  computed: {
    projectId() {
      return this.$route.params.projectId;
    },
  },
};
</script>
