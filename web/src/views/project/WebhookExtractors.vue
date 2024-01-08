<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null && webhook != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? 'Create' : 'Save'"
      :title="`${itemId === 'new' ? 'New' : 'Edit'} Webhook Extractor`"
      :max-width="450"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <WebhookExtractorForm
          :webhook-id="webhookId"
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
      object-title="extractor"
      :object-refs="itemRefs"
      :webhook-id="webhookId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      title="Delete Webhook Extractor"
      text="Are you sure you want to delete this Webhook Extractor?"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />
    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title class="breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/webhooks/`"
        >
          Webhooks
        </router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ webhook.name }}</span>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">Extractors</span>
      </v-toolbar-title>

      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >New Extractor</v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      >
      <template v-slot:item.name="{ item }">
        <router-link
          :to="`/project/${projectId}/webhook/${webhookId}/extractor/${item.id}`"
        >{{ item.name }}
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
/* eslint-disable vue/no-unused-components */
import axios from 'axios';

import { USER_PERMISSIONS } from '@/lib/constants';

import ItemListPageBase from '@/components/ItemListPageBase';
import WebhookExtractorForm from '@/components/WebhookExtractorForm.vue';
import WebhookExtractorsBase from '@/components/WebhookExtractorsBase';

export default {
  mixins: [ItemListPageBase, WebhookExtractorsBase],
  components: { WebhookExtractorForm },
  data() {
    return {
      webhook: null,
    };
  },

  async created() {
    this.webhook = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/webhook/${this.webhookId}`,
      responseType: 'json',
    })).data;
  },

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
  },

  methods: {
    allowActions() {
      return this.can(USER_PERMISSIONS.updateProject);
    },

    getHeaders() {
      return [{
        text: 'Name',
        value: 'name',
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
      return `/api/project/${this.projectId}/webhook/${this.webhookId}/extractors`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/webhook/${this.webhookId}/extractor/${this.itemId}`;
    },
    getEventName() {
      return 'w-webhook-extractor';
    },
  },
};
</script>
