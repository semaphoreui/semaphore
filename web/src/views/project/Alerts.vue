<template>
  <div v-if="items != null">
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
      <v-toolbar-title>{{ $t('alerting') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        class="mr-2"
        @click="testAllAlerts()"
        :disabled="!hasEnabledAlerts"
        data-testid="alerts-testAll"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >{{ $t('testAlerts') }}</v-btn>
      <v-btn
        color="primary"
        @click="createNew()"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >{{ $t('newAlert') }}</v-btn>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto;"
    >
      <template v-slot:item.name="{ item }">
        <v-btn
          v-if="can(USER_PERMISSIONS.manageProjectResources)"
          text
          small
          class="px-0 text-none"
          @click="editExisting(item.id)"
        >{{ item.name }}</v-btn>
        <span v-else>{{ item.name }}</span>
      </template>
      <template v-slot:item.type="{ item }">
        <code>{{ item.type }}</code>
      </template>
      <template v-slot:item.is_default="{ item }">
        <v-icon v-if="item.is_default" color="success">mdi-check</v-icon>
        <v-icon v-else color="grey">mdi-close</v-icon>
      </template>
      <template v-slot:item.enabled="{ item }">
        <v-icon v-if="item.enabled" color="success">mdi-check</v-icon>
        <v-icon v-else color="grey">mdi-close</v-icon>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn @click="testAlert(item)" :title="$t('testAlert')">
            <v-icon>mdi-send</v-icon>
          </v-btn>
          <v-btn @click="cloneItem(item.id)" :title="$t('cloneAlert')">
            <v-icon>mdi-content-copy</v-icon>
          </v-btn>
          <v-btn @click="askDeleteItem(item.id)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn @click="editExisting(item.id)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import ItemListPageBase from '@/components/ItemListPageBase';
import AlertForm from '@/components/AlertForm.vue';
import { getErrorMessage } from '@/lib/error';

export default {
  components: { AlertForm },

  mixins: [ItemListPageBase],

  data() {
    return {
      cloneSourceId: null,
    };
  },

  computed: {
    hasEnabledAlerts() {
      return (this.items || []).some((item) => item.enabled);
    },
  },

  methods: {
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
          text: this.$t('alertTestSent'),
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },
    async testAllAlerts() {
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
        let msg;
        if (err.response && err.response.status === 409) {
          msg = this.$t('alertTestAllMissing');
        } else {
          msg = getErrorMessage(err);
        }
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: msg,
        });
      }
    },
    getHeaders() {
      return [
        { text: this.$i18n.t('name'), value: 'name', width: '35%' },
        { text: this.$i18n.t('type'), value: 'type', width: '20%' },
        { text: this.$i18n.t('alertIsDefault'), value: 'is_default', width: '15%' },
        { text: this.$i18n.t('enabled'), value: 'enabled', width: '15%' },
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
