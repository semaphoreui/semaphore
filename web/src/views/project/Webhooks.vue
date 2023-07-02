<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null && templates != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? 'Create' : 'Save'"
      :title="`${itemId === 'new' ? 'New' : 'Edit'} Webhook`"
      :max-width="450"
      :transition="false"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <WebhookForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      object-title="webhook"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      title="Delete Webhook"
      text="Are you sure you want to delete this Webhook?"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Webhook</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >New Webhook</v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      >
      <template v-slot:item.name="{ item }">
        <router-link
          :to="`/project/${projectId}/webhook/${item.id}`"
        >{{ item.name }}
        </router-link>
      </template>
      <template v-slot:item.template_id="{ item }">
        <router-link
          :to="`/project/${projectId}/templates/${item.template_id}`">
          <code>{{ templates.find((t) => t.id === item.template_id).name }}</code>
        </router-link>
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
import axios from 'axios';
import ItemListPageBase from '@/components/ItemListPageBase';
import WebhookForm from '@/components/WebhookForm.vue';

export default {
  mixins: [ItemListPageBase],
  components: { WebhookForm },
  data() {
    return {
      templates: null,
    };
  },

  async created() {
    this.templates = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/templates`,
      responseType: 'json',
    })).data;
  },

  methods: {
    getHeaders() {
      return [{
        text: 'Name',
        value: 'name',
        width: '33.33%',
        sortable: true,
      },
      {
        text: 'Template',
        value: 'template_id',
        width: '33.33%',
        sortable: true,
      },
      {
        text: 'Actions',
        value: 'actions',
        sortable: false,
      }];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/webhooks`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/webhook/${this.itemId}`;
    },
    getEventName() {
      return 'w-webhook';
    },
  },
};
</script>
