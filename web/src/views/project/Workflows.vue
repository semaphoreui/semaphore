<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <YesNoDialog
      :title="$t('deleteWorkflow')"
      :text="$t('askDeleteWorkflow')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('workflows') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="openEditor('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >{{ $t('newWorkflow') }}
      </v-btn>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4 CenterToScreen"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-lg) - var(--nav-drawer-width) - 200px); margin: auto"
    >
      <template v-slot:item.name="{ item }">
        <a @click="openEditor(item.id)">{{ item.name }}</a>
      </template>
      <template v-slot:item.description="{ item }">
        <span>{{ item.description || '—' }}</span>
      </template>
      <template v-slot:item.nodes="{ item }">
        <span>{{ (item.nodes || []).length }}</span>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn
            @click="runWorkflow(item)"
            :title="$t('workflowRunNow')"
            :disabled="!can(USER_PERMISSIONS.runProjectTasks)"
          >
            <v-icon>mdi-play</v-icon>
          </v-btn>
          <v-btn
            @click="askDeleteItem(item.id)"
            :disabled="!can(USER_PERMISSIONS.manageProjectResources)"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn
            @click="openEditor(item.id)"
            :disabled="!can(USER_PERMISSIONS.manageProjectResources)"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  mixins: [ItemListPageBase],
  methods: {
    openEditor(id) {
      this.$router.push(
        id === 'new'
          ? `/project/${this.projectId}/workflows/new`
          : `/project/${this.projectId}/workflows/${id}/edit`,
      );
    },
    getHeaders() {
      return [
        {
          text: this.$i18n.t('name'),
          value: 'name',
        },
        {
          text: this.$i18n.t('description'),
          value: 'description',
        },
        {
          text: this.$i18n.t('workflowNodes'),
          value: 'nodes',
          sortable: false,
        },
        {
          value: 'actions',
          sortable: false,
          align: 'end',
        },
      ];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/workflows`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/workflows/${this.itemId}`;
    },
    getEventName() {
      return 'i-workflow';
    },
    async runWorkflow(workflow) {
      try {
        const run = (await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/workflows/${workflow.id}/run`,
          responseType: 'json',
        })).data;
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.$t('workflowRunStarted'),
        });
        this.$router.push(
          `/project/${this.projectId}/workflows/${workflow.id}/runs/${run.id}`,
        );
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
