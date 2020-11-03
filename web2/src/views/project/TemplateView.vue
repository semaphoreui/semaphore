<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="item != null && tasks != null">
    <YesNoDialog
      title="Delete template"
      text="Are you really want to delete this template?"
      v-model="deleteItemDialog"
      @yes="deleteItem()"
    />

    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>

      <v-toolbar-title>Task Template: {{ item.alias }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn color="error" @click="askDeleteItem()" class="mr-2">
        <v-icon left>mdi-delete</v-icon>
        Delete
      </v-btn>
      <v-btn
        color="secondary"
        class="mr-2"
        :to="`/project/${projectId}/templates/new/edit?id=${item.id}`"
      >
        <v-icon left>mdi-content-copy</v-icon>
        Copy
      </v-btn>
      <v-btn color="primary" :to="`/project/${projectId}/templates/${item.id}/edit`">
        <v-icon left>mdi-pencil</v-icon>
        Edit
      </v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="tasks"
      hide-default-footer
      class="mt-4"
    >
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

export default {
  components: { YesNoDialog },
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
      deleteItemDialog: false,
      deleteItemId: null,
    };
  },

  computed: {
    itemId() {
      return this.$route.params.templateId;
    },
    isNewItem() {
      return this.itemId === 'new';
    },
  },

  async created() {
    if (this.isNewItem) {
      await this.$router.replace({
        path: `/project/${this.projectId}/templates/new/edit`,
      });
    } else {
      await this.loadItem();
    }
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    askDeleteItem() {
      this.deleteItemDialog = true;
    },

    async deleteItem() {
      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/templates/${this.deleteItemId}`,
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
        this.deleteItemDialog = false;
      }
    },

    async loadItem() {
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
