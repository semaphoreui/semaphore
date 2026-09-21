<template>
  <div v-if="items != null && channels != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('newAlert') : $t('editAlert')"
      :max-width="640"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <AlertForm
          :project-id="projectId"
          :item-id="itemId"
          :source-item-id="cloneSourceId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      object-title="alert"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      :title="$t('deleteAlert')"
      :text="$t('askDeleteAlert')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('alerts') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        class="mr-2"
        outlined
        @click="testAllAlerts()"
        :loading="testingAll"
        data-testid="alerts-testAll"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >
        <v-icon left small>mdi-send-outline</v-icon>
        {{ $t('testAlerts') }}
      </v-btn>
      <v-btn
        color="primary"
        @click="createNew()"
        data-testid="alerts-new"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >{{ $t('newAlert') }}</v-btn>
    </v-toolbar>

    <v-divider />

    <div class="alerts-page">
      <ServerAlertChannels
        class="mb-6"
        :project-id="projectId"
        :channels="channels"
        :can-edit="can(USER_PERMISSIONS.updateProject)"
      />

      <div class="text-subtitle-1 mb-1">{{ $t('projectAlerts') }}</div>
      <div class="text-body-2 mb-3">{{ $t('projectAlertsHint') }}</div>

      <v-data-table
        :headers="headers"
        :items="items"
        hide-default-footer
        :items-per-page="Number.MAX_VALUE"
        class="alerts-table"
        :no-data-text="$t('noAlertsYet')"
      >
        <template v-slot:item.name="{ item }">
          <span :class="{ 'grey--text': !item.enabled }">
            <v-icon small class="mr-2">{{ channelIcon(item.type) }}</v-icon>
            <a
              v-if="can(USER_PERMISSIONS.manageProjectResources)"
              href="#"
              @click.prevent="editExisting(item.id)"
            >{{ item.name }}</a>
            <span v-else>{{ item.name }}</span>
          </span>
        </template>

        <template v-slot:item.type="{ item }">
          {{ channelTitle(item.type) }}
        </template>

        <template v-slot:item.events="{ item }">
          <v-tooltip bottom v-for="event in ALERT_EVENTS" :key="event.value">
            <template v-slot:activator="{ on }">
              <v-icon
                small
                class="mr-1"
                v-on="on"
                :color="listensTo(item, event.value) ? event.color : 'grey lighten-1'"
              >{{ event.icon }}</v-icon>
            </template>
            <span>{{ $t(event.label) }}</span>
          </v-tooltip>
        </template>

        <template v-slot:item.is_default="{ item }">
          <v-chip
            v-if="item.is_default"
            x-small
            outlined
            color="primary"
          >{{ $t('default') }}</v-chip>
        </template>

        <template v-slot:item.enabled="{ item }">
          <v-icon small :color="item.enabled ? 'success' : 'grey'">
            {{ item.enabled ? 'mdi-check-circle' : 'mdi-pause-circle-outline' }}
          </v-icon>
        </template>

        <template v-slot:item.actions="{ item }">
          <v-btn-toggle dense :value-comparator="() => false">
            <v-btn @click="testAlert(item)" :title="$t('testAlert')" data-testid="alerts-test">
              <v-icon>mdi-send-outline</v-icon>
            </v-btn>
            <v-btn @click="cloneItem(item.id)" :title="$t('cloneAlert')">
              <v-icon>mdi-content-copy</v-icon>
            </v-btn>
            <v-btn @click="askDeleteItem(item.id)" :title="$t('deleteAlert')">
              <v-icon>mdi-delete</v-icon>
            </v-btn>
            <v-btn @click="editExisting(item.id)" :title="$t('editAlert')">
              <v-icon>mdi-pencil</v-icon>
            </v-btn>
          </v-btn-toggle>
        </template>
      </v-data-table>
    </div>
  </div>
</template>
<style lang="scss">
.alerts-page {
  max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px);
  margin: 16px auto 0;
  padding: 0 16px;
}
</style>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import ItemListPageBase from '@/components/ItemListPageBase';
import AlertForm from '@/components/AlertForm.vue';
import ServerAlertChannels from '@/components/ServerAlertChannels.vue';
import { getErrorMessage } from '@/lib/error';
import { ALERT_EVENTS, effectiveAlertEvents, findChannel } from '@/lib/alerts';

export default {
  components: { AlertForm, ServerAlertChannels },

  mixins: [ItemListPageBase],

  data() {
    return {
      channels: null,
      cloneSourceId: null,
      testingAll: false,
      ALERT_EVENTS,
    };
  },

  methods: {
    async beforeLoadItems() {
      this.channels = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/alerts/channels`,
        responseType: 'json',
      })).data;
    },

    channelIcon(type) {
      const ch = findChannel(this.channels, type);
      return ch ? ch.icon : 'mdi-bell-outline';
    },

    channelTitle(type) {
      const ch = findChannel(this.channels, type);
      return ch ? ch.title : type;
    },

    listensTo(alert, event) {
      return effectiveAlertEvents(alert, findChannel(this.channels, alert.type)).includes(event);
    },

    createNew() {
      this.cloneSourceId = null;
      this.editItem('new');
    },

    editExisting(id) {
      this.cloneSourceId = null;
      this.editItem(id);
    },

    cloneItem(id) {
      this.cloneSourceId = id;
      this.editItem('new');
    },

    async testAlert(item) {
      try {
        await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/alerts/${item.id}/test`,
        });
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.$t('alertTestSent', { name: item.name }),
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },

    async testAllAlerts() {
      this.testingAll = true;
      try {
        await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/notifications/test`,
        });
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.$t('alertTestAllSent'),
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: err.response && err.response.status === 409
            ? this.$t('alertTestAllMissing')
            : getErrorMessage(err),
        });
      } finally {
        this.testingAll = false;
      }
    },

    getHeaders() {
      return [
        { text: this.$i18n.t('name'), value: 'name', width: '30%' },
        { text: this.$i18n.t('type'), value: 'type', width: '18%' },
        {
          text: this.$i18n.t('alertSendOn'),
          value: 'events',
          sortable: false,
          width: '18%',
        },
        {
          text: '',
          value: 'is_default',
          sortable: false,
          width: '12%',
        },
        { text: this.$i18n.t('enabled'), value: 'enabled', width: '10%' },
        { value: 'actions', sortable: false, width: '0%' },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/alerts`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/alerts/${this.itemId}`;
    },

    getEventName() {
      return 'i-alerts';
    },
  },
};
</script>
