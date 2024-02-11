<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null && integration != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? 'Create' : 'Save'"
      :title="`${itemId === 'new' ? 'New' : 'Edit'} Integration Extractor`"
      :max-width="450"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <IntegrationExtractorForm
          :integration-id="integrationId"
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
      :integration-id="integrationId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      title="Delete Integration Extractor"
      text="Are you sure you want to delete this Integration Extractor?"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />
    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title class="breadcrumbs">
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/integrations/`"
        >
          Integrations
        </router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ integration.name }}</span>
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
          :to="`/project/${projectId}/integration/${integrationId}/extractor/${item.id}`"
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
import IntegrationExtractorForm from '@/components/IntegrationExtractorForm.vue';
import IntegrationExtractorsBase from '@/components/IntegrationExtractorsBase';

export default {
  mixins: [ItemListPageBase, IntegrationExtractorsBase],
  components: { IntegrationExtractorForm },
  data() {
    return {
      integration: null,
    };
  },

  async created() {
    this.integration = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/integrations/${this.integrationId}`,
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
    integrationId() {
      if (/^-?\d+$/.test(this.$route.params.integrationId)) {
        return parseInt(this.$route.params.integrationId, 10);
      }
      return this.$route.params.integrationId;
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
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/extractors`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/extractor/${this.itemId}`;
    },
    getEventName() {
      return 'w-integration-extractor';
    },
  },
};
</script>
