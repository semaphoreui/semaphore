<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? 'Create' : 'Save'"
      :title="`${itemId === 'new' ? 'New' : 'Edit'} Extract Value`"
      :max-width="450"
      :transition="false"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <WebhookExtractValueForm
          :extractor-id="extractorId"
          :item-id="itemId"
          :project-id="projectId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      object-title="extractvalue"
      :object-refs="itemRefs"
      :extractor-id="extractorId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      title="Delete Webhook ExtractValue"
      text="Are you sure you want to delete this Webhook ExtractValue?"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>ExtractValue</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >New Extracted Value</v-btn>
    </v-toolbar>
    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      >
      <template v-slot:item.name="{ item }">
        {{ item.name }}
      </template>
      <template v-slot:item.value_source="{ item }">
        <code>{{ item.value_source }}</code>
      </template>
      <template v-slot:item.body_data_type="{ item }">
        <code>{{ item.body_data_type }}</code>
      </template>
      <template v-slot:item.key="{ item }">
        <code>{{ item.key }}</code>
      </template>
      <template v-slot:item.variable="{ item }">
        <code>{{ item.variable }}</code>
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
import ItemListPageBase from '@/components/ItemListPageBase';

import WebhookExtractorsBase from '@/components/WebhookExtractorsBase';
import WebhookExtractorBase from '@/components/WebhookExtractorBase';

import WebhookExtractValueForm from '@/components/WebhookExtractValueForm.vue';

export default {
  mixins: [ItemListPageBase, WebhookExtractorsBase, WebhookExtractorBase],
  components: { WebhookExtractValueForm },

  computed: {
    projectId() {
      if (/^-?\d+$/.test(this.$route.params.projectId)) {
        return parseInt(this.$route.params.projectId, 10);
      }
      return this.$route.params.projectId;
    },

    webhookId() {
      if (/^-?\d+$/.test(this.$route.params.webhookId)) {
        return parseInt(this.$route.params.webhookId, 10);
      }
      return this.$route.params.webhookId;
    },
    extractorId() {
      if (/^-?\d+$/.test(this.$route.params.extractorId)) {
        return parseInt(this.$route.params.extractorId, 10);
      }
      return this.$route.params.extractorId;
    },
  },

  methods: {
    getHeaders() {
      return [{
        text: 'Name',
        value: 'name',
        width: '20%',
        sortable: true,
      },
      {
        text: 'Value Source',
        value: 'value_source',
        width: '10%',
        sortable: false,
      },
      {
        text: 'Body Data Type',
        value: 'body_data_type',
        width: '15%',
        sortable: false,
      },
      {
        text: 'Key',
        value: 'key',
        width: '15%',
        sortable: false,
      },
      {
        text: 'Environment Variable',
        value: 'variable',
        width: '20%',
        sortable: false,
      },
      {
        text: 'Actions',
        value: 'actions',
        sortable: false,
      }];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/webhook/${this.webhookId}/extractor/${this.extractorId}/values`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/webhook/${this.webhookId}/extractor/${this.extractorId}/value/${this.itemId}`;
    },
    getEventName() {
      return 'w-webhook-extract-value';
    },
  },
};
</script>
