<template>
  <div
    v-if="items != null && isAppsLoaded"
    data-testid="apps-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
        v-model="editDialog"
        test-id="appsEditDialog"
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
        test-id="appsDeleteDialog"
        :title="$t('Delete App')"
        :text="$t('Do you really want to delete this app?')"
        v-model="deleteItemDialog"
        @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-btn
          icon
          class="mr-4"
          @click="returnToProjects()"
          data-testid="apps-back"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="apps-title"
      >{{ $t('Applications') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
          :disabled="!canManageSystem"
          color="primary"
          variant="elevated"
          @click="editItem('')"
          data-testid="apps-new"
      >{{ $t('New App') }}</v-btn>
    </v-toolbar>

    <v-data-table
        :headers="headers"
        :items="items"
        class="mt-4"
        :items-per-page-options="[20]"
        :items-per-page="20"
        :mobile-breakpoint="600"
        data-testid="apps-table"
        :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.active="{ item }">
        <!-- min-width: the grid of a Vuetify 3 v-input is 0 px at its narrowest, and the column
             of the switch (beside the 100 % wide «ID») shrank to it: the switch covered
             the names -->
        <v-switch
            :disabled="!canManageSystem"
            v-model="item.active"
            inset
            style="min-width: max-content"
            @update:model-value="setActive(item.id, $event)"
            data-testid="apps-rowActive"
        ></v-switch>
      </template>

      <template v-slot:item.title="{ item }">
        <v-icon
            class="mr-2"
            size="small"
            data-testid="apps-rowIcon"
        >
          {{ getAppIcon(item.id) }}
        </v-icon>

        {{ getAppTitle(item.id) }}
      </template>

      <template v-slot:item.id="{ item }">
        <code data-testid="apps-rowId">{{ item.id }}</code>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap" data-testid="apps-rowActions">
          <v-btn
              v-if="!isDefaultApp(item.id)"
              icon
              variant="text"
              class="mr-1"
              @click="askDeleteItem(item.id)"
              :disabled="item.id === userId"
              data-testid="apps-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn
              icon
              variant="text"
              class="mr-1"
              @click="editItem(item.id)"
              data-testid="apps-rowEdit"
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
import EventBus from '@/event-bus';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ItemListPageBase from '@/components/ItemListPageBase';
import EditDialog from '@/components/EditDialog.vue';
import PermissionsCheck from '@/components/PermissionsCheck';
import { hasGlobalPermission } from '@/lib/role-permissions';
import AppForm from '../components/AppForm.vue';
import { DEFAULT_APPS, GLOBAL_PERMISSIONS } from '../lib/constants';
import AppsMixin from '../components/AppsMixin';
import delay from '../lib/delay';

export default {
  mixins: [ItemListPageBase, AppsMixin, PermissionsCheck],

  components: {
    AppForm,
    YesNoDialog,
    EditDialog,
  },

  computed: {
    canManageSystem() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageSystem);
    },
  },

  methods: {
    allowActions() {
      return hasGlobalPermission(this.user, GLOBAL_PERMISSIONS.manageSystem);
    },

    getHeaders() {
      return [{
        title: '',
        key: 'active',
      }, {
        title: this.$t('name'),
        key: 'title',
      }, {
        title: 'ID',
        key: 'id',
        width: '100%',
      }, {
        title: this.$t('actions'),
        key: 'actions',
        sortable: false,
      }];
    },

    async loadAppsDataFromBackend() {
      while (this.items == null) {
        // eslint-disable-next-line no-await-in-loop
        await delay(100);
      }

      return this.items;
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
