<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="item != null && tasks != null">
    <ItemDialog
      v-model="editDialog"
      save-button-text="Save"
      title="Edit Template"
      @save="loadData()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TemplateForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </ItemDialog>

    <ItemDialog
      v-model="copyDialog"
      save-button-text="Create"
      title="New Template"
      @save="loadData()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TemplateForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </ItemDialog>

    <YesNoDialog
      title="Delete template"
      text="Are you really want to delete this template?"
      v-model="deleteDialog"
      @yes="remove()"
    />

    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title class="breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/templates/`"
        >Task Templates</router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ item.alias }}</span>
      </v-toolbar-title>
      <v-spacer></v-spacer>

      <v-btn
        icon
        color="error"
        @click="deleteDialog = true"
      >
        <v-icon>mdi-delete</v-icon>
      </v-btn>

      <v-btn
        icon
        color="black"
        @click="copyDialog = true"
      >
        <v-icon>mdi-content-copy</v-icon>
      </v-btn>

      <v-btn
        icon
        color="black"
        @click="editDialog = true"
      >
        <v-icon>mdi-pencil</v-icon>
      </v-btn>
    </v-toolbar>

    <v-container class="pa-0">
      <v-row>
        <v-col>
          <v-list two-line subheader>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-title>Playbook</v-list-item-title>
                <v-list-item-subtitle>{{ item.playbook }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>

            <v-list-item>
              <v-list-item-content>
                <v-list-item-title>SSH Key</v-list-item-title>
                <v-list-item-subtitle>{{ item.ssh_key_id }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list two-line subheader>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-title>Inventory</v-list-item-title>
                <v-list-item-subtitle>{{ item.inventory_id }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>

            <v-list-item>
              <v-list-item-content>
                <v-list-item-title>Environment</v-list-item-title>
                <v-list-item-subtitle>{{ item.environment_id }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>

            <v-list-item>
              <v-list-item-content>
                <v-list-item-title>Repository</v-list-item-title>
                <v-list-item-subtitle>{{ item.repository_id }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
      </v-row>
    </v-container>

    <h4 class="ml-4 mt-4">Task History</h4>
    <v-data-table
      :headers="headers"
      :items="tasks"
      hide-default-footer
      class="mt-2"
    >
      <template v-slot:item.id="{ item }">
        <a @click="showTaskLog(item.id)">#{{ item.id }}</a>
      </template>
      <template v-slot:item.status="{ item }">
        {{ item.status }}
      </template>
    </v-data-table>
  </div>
</template>
<style lang="scss">

</style>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ItemDialog from '@/components/ItemDialog.vue';
import TemplateForm from '@/components/TemplateForm.vue';

export default {
  components: {
    YesNoDialog, ItemDialog, TemplateForm,
  },
  props: {
    projectId: Number,
  },
  data() {
    return {
      headers: [
        {
          text: 'Task ID',
          value: 'id',
          sortable: false,
        },
        {
          text: 'Status',
          value: 'status',
          sortable: false,
        },
        {
          text: 'User',
          value: 'user_name',
          sortable: false,
        },
        {
          text: 'Start',
          value: 'start',
          sortable: false,
        },
        {
          text: 'Duration',
          value: 'start',
          sortable: false,
        },
      ],
      tasks: null,
      item: null,
      deleteDialog: null,
      editDialog: null,
      copyDialog: null,
      taskLogDialog: null,
      taskId: null,
    };
  },

  computed: {
    itemId() {
      return this.$route.params.templateId;
    },
    isNew() {
      return this.itemId === 'new';
    },
  },

  async created() {
    if (this.isNew) {
      await this.$router.replace({
        path: `/project/${this.projectId}/templates/new/edit`,
      });
    } else {
      await this.loadData();
    }
  },

  methods: {
    showTaskLog(taskId) {
      EventBus.$emit('i-show-task', {
        taskId,
      });
    },

    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async remove() {
      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/templates/${this.itemId}`,
          responseType: 'json',
        });

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: `Template "${this.item.alias}" deleted`,
        });

        await this.$router.push({
          path: `/project/${this.projectId}/templates`,
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.deleteDialog = false;
      }
    },

    async loadData() {
      this.item = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates/${this.itemId}`,
        responseType: 'json',
      })).data;

      this.tasks = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates/${this.itemId}/tasks/last`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
