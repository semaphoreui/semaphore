<template>
  <v-form
    ref="form"
    v-model="formValid"
    data-testid="inventoryHostForm-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="inventoryHostForm-error"
    >{{ formError }}</v-alert>

    <template v-if="item != null">
      <v-text-field
        v-model.trim="item.name"
        :label="$t('name')"
        :rules="[v => !!v || $t('name_required')]"
        required
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        data-testid="inventoryHostForm-name"
      ></v-text-field>

      <v-text-field
        v-model="item.description"
        :label="$t('description')"
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        data-testid="inventoryHostForm-description"
      ></v-text-field>

      <div class="mt-2 mb-1 text-medium-emphasis">{{ $t('inventory_hosts_variables') }}</div>
      <CodeMirror
        class="InventoryHostForm__editor"
        :style="{ border: '1px solid lightgray' }"
        v-model="item.variables"
        minimal
        :tab-size="2"
        :extensions="cmExtensions"
        :disabled="formSaving"
        data-testid="inventoryHostForm-variables"
      />
    </template>
  </v-form>
</template>
<style>
.InventoryHostForm__editor .cm-editor {
  height: 160px !important;
}
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import { markRaw } from 'vue';
import CodeMirror from 'vue-codemirror6';
import ItemFormBase from '@/components/ItemFormBase';
import codeEditorExtensions, { CODEMIRROR5_LOOK } from '@/lib/codeEditor';

export default {
  mixins: [ItemFormBase],

  components: { CodeMirror },

  props: {
    inventoryId: [Number, String],
  },

  data() {
    return {
      cmExtensions: markRaw([
        ...codeEditorExtensions({ language: 'yaml', tab: 'spaces' }),
        CODEMIRROR5_LOOK,
      ]),
    };
  },

  methods: {
    getNewItem() {
      return {
        name: '', description: '', enabled: true, variables: '',
      };
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/hosts`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/hosts/${this.itemId}`;
    },
  },
};
</script>
