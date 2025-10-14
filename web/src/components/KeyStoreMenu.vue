<template>
  <div>
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
        :to="`/project/${projectId}/secret_storages`"
        data-testid="keystore-storages"
      >
        Storages
      </v-tab>
    </v-tabs>
    <v-divider style="margin-top: -1px;" />
  </div>
</template>

<script>
export default {

  props: {
    projectId: Number,
  },

};
</script>
