<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
        v-model="editDialog"
        save-button-text="Save"
        :title="$t('Edit App')"
        @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <AppForm
            :project-id="projectId"
            :item-id="itemId"
            @save="onSave"
            @error="onError"
            :need-save="needSave"
            :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <YesNoDialog
        :title="$t('deleteUser')"
        :text="$t('askDeleteUser')"
        v-model="deleteItemDialog"
        @yes="deleteItem(itemId)"
    />

    <v-toolbar flat >
      <v-btn
          icon
          class="mr-4"
          @click="returnToProjects()"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ $t('Applications') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
          color="primary"
          @click="editItem('')"
      >{{ $t('New App') }}</v-btn>
    </v-toolbar>

    <v-data-table
        :headers="headers"
        :items="items"
        class="mt-4"
        :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.active="{ item }">
        <v-switch
            v-model="item.active"
            inset
            @change="setActive(item.id, item.active)"
        ></v-switch>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn
              v-if="!isDefaultApp(item.id)"
              icon
              class="mr-1"
              @click="askDeleteItem(item.id)"
              :disabled="item.id === userId"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn
              v-if="!isDefaultApp(item.id)"
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
import EventBus from '@/event-bus';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ItemListPageBase from '@/components/ItemListPageBase';
import EditDialog from '@/components/EditDialog.vue';
import axios from 'axios';
import AppForm from '../components/AppForm.vue';
import { DEFAULT_APPS } from '../lib/constants';

export default {
  mixins: [ItemListPageBase],

  components: {
    AppForm,
    YesNoDialog,
    EditDialog,
  },

  methods: {
    getHeaders() {
      return [{
        text: '',
        value: 'active',
      }, {
        text: 'ID',
        value: 'id',
      }, {
        text: this.$i18n.t('name'),
        width: '100%',
        value: 'title',
      }, {
        text: this.$i18n.t('actions'),
        value: 'actions',
        sortable: false,
      }];
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    getItemsUrl() {
      return '/api/apps';
    },

    getSingleItemUrl() {
      return `/api/apps/${this.itemId}`;
    },

    getEventName() {
      return 'i-app';
    },

    async setActive(appId, active) {
      await axios({
        method: 'post',
        url: `/api/apps/${appId}/active`,
        responseType: 'json',
        data: {
          active,
        },
      });
    },

    isDefaultApp(appId) {
      return DEFAULT_APPS.includes(appId);
    },
  },
};
</script>
