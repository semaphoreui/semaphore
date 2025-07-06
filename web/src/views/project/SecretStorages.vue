<template>
  <div v-if="items != null">

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('keyStore') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >New Storage
      </v-btn>
    </v-toolbar>

    <v-tabs class="pl-4">
      <v-tab
        key="keys"
        :to="`/project/${projectId}/keys`"
        data-testid="keystore-keys"
      >
        Keys
      </v-tab>

      <v-tab
        key="storages"
        :to="`/project/${projectId}/secret-storages`"
        data-testid="keystore-storages"
      >
        Storages
      </v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px;"/>

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto;"
    >
      <template v-slot:item.name="{ item }">
        {{ item.name }}
        <v-chip
          color="error"
          v-if="item.empty && item.type !== 'none'"
          small
          style="font-weight: bold;"
          class="ml-2"
        >{{ $t('empty') }}</v-chip>
      </template>
      <template v-slot:item.type="{ item }">
        <code>{{ item.type }}</code>
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

<style scoped lang="scss">

</style>

<script>
import ItemListPageBase from '@/components/ItemListPageBase';

export default {
  mixins: [ItemListPageBase],
  methods: {
    getHeaders() {
      return [{
        text: this.$i18n.t('name'),
        value: 'name',
        width: '60%',
      },
      {
        text: this.$i18n.t('type'),
        value: 'type',
        width: '40%',
      },
      {
        value: 'actions',
        sortable: false,
        width: '0%',
      },
      ];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/secret-storages`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/secret-storages/${this.itemId}`;
    },
    getEventName() {
      return 'i-secret-storage';
    },

  },
};
</script>
