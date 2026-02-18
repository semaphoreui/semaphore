<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null && isAppsLoaded">
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
        :title="$t('Delete App')"
        :text="$t('Do you really want to delete this app?')"
        v-model="deleteItemDialog"
        @yes="deleteItem(itemId)"
    />

    <!-- Version Edit Dialog -->
    <EditDialog
        v-model="versionEditDialog"
        :save-button-text="versionItemId === 'new' ? 'Create' : 'Save'"
        :title="versionItemId === 'new' ? 'New Version' : 'Edit Version'"
        @save="loadVersions(versionsAppId)"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <AppVersionForm
            :app-id="versionsAppId"
            :item-id="versionItemId"
            @save="onSave"
            @error="onError"
            :need-save="needSave"
            :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <!-- Delete Version Dialog -->
    <YesNoDialog
        title="Delete Version"
        text="Are you sure you want to delete this version?"
        v-model="deleteVersionDialog"
        @yes="deleteVersion()"
    />

    <!-- Versions Dialog -->
    <v-dialog v-model="versionsDialog" max-width="700" persistent>
      <v-card>
        <v-card-title>
          <v-icon class="mr-2" small>{{ getAppIcon(versionsAppId) }}</v-icon>
          {{ getAppTitle(versionsAppId) }} &mdash; Versions
          <v-spacer></v-spacer>
          <v-btn icon @click="versionsDialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text class="pb-0">
          <v-data-table
              :headers="versionHeaders"
              :items="versions"
              :loading="versionsLoading"
              hide-default-footer
              :items-per-page="-1"
          >
            <template v-slot:item.active="{ item }">
              <v-icon small :color="item.active ? 'success' : 'grey'">
                {{ item.active ? 'mdi-check-circle' : 'mdi-close-circle' }}
              </v-icon>
            </template>

            <template v-slot:item.name="{ item }">
              {{ item.name || '(default)' }}
            </template>

            <template v-slot:item.path="{ item }">
              <code v-if="item.path">{{ item.path }}</code>
              <span v-else class="grey--text">—</span>
            </template>

            <template v-slot:item.actions="{ item }">
              <div style="white-space: nowrap">
                <v-btn
                    icon
                    small
                    class="mr-1"
                    @click="editVersion(item.id)"
                >
                  <v-icon small>mdi-pencil</v-icon>
                </v-btn>
                <v-btn
                    icon
                    small
                    @click="askDeleteVersion(item.id)"
                >
                  <v-icon small>mdi-delete</v-icon>
                </v-btn>
              </div>
            </template>
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
              color="primary"
              text
              @click="editVersion('new')"
          >
            <v-icon left small>mdi-plus</v-icon>
            New Version
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

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
          :disabled="!isAdmin"
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
            :disabled="!isAdmin"
            v-model="item.active"
            inset
            @change="setActive(item.id, item.active)"
        ></v-switch>
      </template>

      <template v-slot:item.title="{ item }">
        <v-icon
            class="mr-2"
            small
        >
          {{ getAppIcon(item.id) }}
        </v-icon>

        {{ getAppTitle(item.id) }}
      </template>

      <template v-slot:item.id="{ item }">
        <code>{{ item.id }}</code>
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
              icon
              class="mr-1"
              @click="editItem(item.id)"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>

          <v-btn
              icon
              class="mr-1"
              @click="openVersions(item.id)"
          >
            <v-icon>mdi-format-list-numbered</v-icon>
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
import AppForm from '../components/AppForm.vue';
import AppVersionForm from '../components/AppVersionForm.vue';
import { DEFAULT_APPS } from '../lib/constants';
import AppsMixin from '../components/AppsMixin';
import delay from '../lib/delay';
import { getErrorMessage } from '../lib/error';

export default {
  mixins: [ItemListPageBase, AppsMixin, PermissionsCheck],

  components: {
    AppForm,
    AppVersionForm,
    YesNoDialog,
    EditDialog,
  },

  data() {
    return {
      versionsDialog: false,
      versionsAppId: null,
      versions: [],
      versionsLoading: false,

      versionEditDialog: false,
      versionItemId: null,

      deleteVersionDialog: false,
      deleteVersionId: null,

      versionHeaders: [{
        text: 'Name',
        value: 'name',
      }, {
        text: 'Path',
        value: 'path',
      }, {
        text: 'Active',
        value: 'active',
        width: '80px',
      }, {
        text: 'Priority',
        value: 'priority',
        width: '80px',
      }, {
        text: 'Actions',
        value: 'actions',
        sortable: false,
        width: '100px',
      }],
    };
  },

  methods: {
    getHeaders() {
      return [{
        text: '',
        value: 'active',
      }, {
        text: this.$i18n.t('name'),
        value: 'title',
      }, {
        text: 'ID',
        value: 'id',
        width: '100%',
      }, {
        text: this.$i18n.t('actions'),
        value: 'actions',
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

    async openVersions(appId) {
      this.versionsAppId = appId;
      this.versionsDialog = true;
      await this.loadVersions(appId);
    },

    async loadVersions(appId) {
      this.versionsLoading = true;
      try {
        this.versions = (await axios({
          method: 'get',
          url: `/api/apps/${appId}/versions`,
          responseType: 'json',
        })).data;
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.versionsLoading = false;
      }
    },

    editVersion(versionId) {
      this.versionItemId = versionId;
      this.versionEditDialog = true;
    },

    askDeleteVersion(versionId) {
      this.deleteVersionId = versionId;
      this.deleteVersionDialog = true;
    },

    async deleteVersion() {
      try {
        await axios({
          method: 'delete',
          url: `/api/apps/${this.versionsAppId}/versions/${this.deleteVersionId}`,
          responseType: 'json',
        });
        await this.loadVersions(this.versionsAppId);
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
