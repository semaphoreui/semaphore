<template>
  <div v-if="items != null && keys != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('addMapping') : $t('editMapping')"
      @save="loadItems()"
      :max-width="450"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <HostConfigForm
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
      :title="$t('deleteMapping')"
      :text="$t('askDeleteMapping')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('hostConfig') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >{{ $t('addMapping') }}</v-btn>
    </v-toolbar>

    <v-divider />

    <v-alert
      v-if="items.length === 0"
      type="info"
      text
      class="mt-4"
      style="max-width: 800px; margin: auto;"
    >{{ $t('hostConfigEmpty') }}</v-alert>

    <v-data-table
      v-else
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 100px); margin: auto;"
    >
      <template v-slot:item.type="{ item }">
        <v-chip small>{{
          item.type === 'url' ? $t('hostConfigTypeUrl') : $t('hostConfigTypeHost')
        }}</v-chip>
      </template>

      <template v-slot:item.ssh_key_id="{ item }">
        {{ (keys.find((k) => k.id === item.ssh_key_id) || {}).name }}
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn @click="askDeleteItem(item.id)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn @click="editItem(item.id)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import axios from 'axios';
import ItemListPageBase from '@/components/ItemListPageBase';
import HostConfigForm from '@/components/HostConfigForm.vue';

export default {
  mixins: [ItemListPageBase],
  components: { HostConfigForm },

  data() {
    return {
      keys: null,
    };
  },

  async created() {
    this.keys = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },

  methods: {
    getHeaders() {
      return [{
        text: this.$i18n.t('hostConfigHostOrUrl'),
        value: 'host',
        width: '50%',
      },
      {
        text: this.$i18n.t('type'),
        value: 'type',
        width: '15%',
      },
      {
        text: this.$i18n.t('credential'),
        value: 'ssh_key_id',
        width: '25%',
      },
      {
        text: this.$i18n.t('actions'),
        value: 'actions',
        sortable: false,
      }];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/host_configs`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/host_configs/${this.itemId}`;
    },

    getEventName() {
      return 'i-host-configs';
    },
  },
};
</script>
