<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? 'Create' : 'Save'"
      :title="`${itemId === 'new' ? 'New' : 'Edit'} Matcher`"
      :max-width="450"
      :transition="false"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <WebhookMatcherForm
          :extractor-id="extractorId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      object-title="matcher"
      :object-refs="itemRefs"
      :extractor-id="extractorId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      title="Delete Webhook Matcher"
      text="Are you sure you want to delete this Webhook Matcher?"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Matcher</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >New Matcher</v-btn>
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
      <template v-slot:item.match_type="{ item }">
        <code>{{ item.match_type }}</code>
      </template>
      <template v-slot:item.method="{ item }">
        <code>{{ item.method }}</code>
      </template>
      <template v-slot:item.body_data_type="{ item }">
        {{ item.body_data_type || "N/A" }}
      </template>
      <template v-slot:item.key="{ item }">
        <code>{{ item.key }}</code>
      </template>
      <template v-slot:item.value="{ item }">
        <code>{{ item.value }}</code>
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
import { USER_PERMISSIONS } from '@/lib/constants';
import ItemListPageBase from '@/components/ItemListPageBase';

import WebhookExtractorsBase from '@/components/WebhookExtractorsBase';
import WebhookExtractorBase from '@/components/WebhookExtractorBase';

import WebhookMatcherForm from '@/components/WebhookMatcherForm.vue';

export default {
  mixins: [ItemListPageBase, WebhookExtractorsBase, WebhookExtractorBase],
  components: { WebhookMatcherForm },
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
    allowActions() {
      return this.can(USER_PERMISSIONS.updateProject);
    },
    getHeaders() {
      return [{
        text: 'Name',
        value: 'name',
        sortable: true,
      },
      {
        text: 'Match Type',
        value: 'match_type',
        sortable: true,
      },
      {
        text: 'Method',
        value: 'method',
        sortable: true,
      },
      {
        text: 'Body Data Type',
        value: 'body_data_type',
        sortable: false,
      },
      {
        text: 'Key',
        value: 'key',
        sortable: true,
      },
      {
        text: 'Value',
        value: 'value',
        sortable: true,
      },
      {
        text: 'Actions',
        value: 'actions',
        sortable: false,
      }];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/webhook/${this.webhookId}/extractor/${this.extractorId}/matchers`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/webhook/${this.webhookId}/extractor/${this.extractorId}/matcher/${this.itemId}`;
    },
    getEventName() {
      return 'w-webhook-matcher';
    },
  },
};
</script>
