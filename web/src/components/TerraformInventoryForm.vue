<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      label="Display name"
      :rules="[v => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.inventory"
      label="Workspace name"
      :rules="[v => !!v || $t('path_required')]"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-alert
        dense
        text
        class="mt-4"
        type="info"
        v-if="item.type === 'static'"
    >
      {{ $t('staticInventoryExample') }}
      <pre style="font-size: 14px;">[website]
172.18.8.40
172.18.8.41</pre>
    </v-alert>
  </v-form>
</template>
<style>
.CodeMirror {
  height: 160px !important;
}
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],

  components: {
  },

  data() {
    return {
      cmOptions: {
        tabSize: 2,
        mode: 'text/x-ini',
        lineNumbers: true,
        line: true,
        lint: true,
        indentWithTabs: false,
      },
      keys: null,
      inventoryTypes: [{
        id: 'static',
        name: 'Static',
      }, {
        id: 'static-yaml',
        name: 'Static YAML',
      }, {
        id: 'file',
        name: 'File',
      }],
    };
  },

  methods: {
    async beforeSave() {
      this.item.type = 'terraform-workspace';
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.itemId}`;
    },
  },
};
</script>
