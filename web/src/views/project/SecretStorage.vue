<template>
  <div>

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('keyStore') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >New Storage</v-btn>
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

    <v-divider style="margin-top: -1px;" />
  </div>
</template>

<style scoped lang="scss">

</style>

<script>
import ItemListPageBase from '@/components/ItemListPageBase';

export default {
  mixins: [ItemListPageBase],

};
</script>
