<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">

    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? 'New Notification' : 'Edit Notification'"
      :max-width="500"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <NotificationForm
          :project-id="projectId"
          :item-id="itemId"
          :item-app="itemApp"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      title="Delete Notification"
      text="Are you sure you want to delete this notification?"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('notifications') }}</v-toolbar-title>
      <v-spacer></v-spacer>

      <v-menu offset-y>
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            class="pr-2"
            v-bind="attrs"
            v-on="on"
            color="primary"
            v-if="can(USER_PERMISSIONS.manageProjectResources)"
          >
            {{ $t('newNotification') }}
            <v-icon>mdi-chevron-down</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item
            v-for="provider in notificationProviders"
            :key="provider.id"
            link
            @click="itemApp = provider.id; editItem('new');"
          >
            <v-list-item-icon>
              <v-icon>{{ provider.icon }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>{{ provider.name }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
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
        <v-icon class="mr-3" small>
          {{ getProviderIcon(item.type) }}
        </v-icon>
        {{ item.name }}
      </template>

      <template v-slot:item.type="{ item }">
        <code>{{ item.type }}</code>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn @click="askDeleteItem(item.id)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn @click="itemApp = item.type; editItem(item.id)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>
  </div>
</template>

<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import AppsMixin from '@/components/AppsMixin';
import NotificationForm from '@/components/NotificationForm.vue';

export default {
  mixins: [ItemListPageBase, AppsMixin],
  components: { NotificationForm },

  props: {
    features: Object,
  },

  data() {
    return {
      itemApp: '',
      item: {},
      notificationProviders: [
        { id: 'email', name: 'Email', icon: 'mdi-email' },
        { id: 'slack', name: 'Slack', icon: 'mdi-slack' },
      ],
    };
  },

  methods: {
    getProviderIcon(type) {
      if (!type) return 'mdi-bell';
      const provider = this.notificationProviders.find(p => p.id === type.toLowerCase());
      return provider ? provider.icon : 'mdi-bell';
    },

    getHeaders() {
      return [
        {
          text: this.$i18n.t('name'),
          value: 'name',
          width: '50%',
        },
        {
          text: this.$i18n.t('type'),
          value: 'type',
          width: '30%',
        },
        {
          value: 'actions',
          sortable: false,
          width: '20%',
        },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/notifications`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/notifications/${this.itemId}`;
    },
    getEventName() {
      return 'i-notifications';
    },
  },
};
</script>
