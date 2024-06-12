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
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

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
    async getNoneKey() {
      return (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/keys`,
        responseType: 'json',
      })).data.filter((key) => key.type === 'none')[0];
    },

    async beforeSave() {
      let noneKey = await this.getNoneKey();

      if (!noneKey) {
        await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/keys`,
          responseType: 'json',
          data: {
            name: 'None',
            type: 'none',
            project_id: this.projectId,
          },
        });
        noneKey = await this.getNoneKey();
      }

      this.item.type = 'terraform-workspace';
      this.item.ssh_key_id = noneKey.id;
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
