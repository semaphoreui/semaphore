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
      <v-alert
        v-if="imported"
        type="info"
        variant="tonal"
        density="compact"
        class="mb-4"
        data-testid="inventoryHostForm-imported"
      >{{ $t('inventory_source_imported', { name: sourceName }) }}</v-alert>

      <v-text-field
        v-model.trim="item.name"
        :label="$t('name')"
        :rules="[v => !!v || $t('name_required')]"
        required
        :readonly="imported"
        :append-inner-icon="imported ? 'mdi-lock' : undefined"
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

      <v-autocomplete
        v-model="item.groups"
        :items="groupOptions"
        item-title="name"
        item-value="id"
        :item-props="(group) => (group.source_id == null ? {} : { disabled: true })"
        :label="$t('inventory_tab_groups')"
        multiple
        chips
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        data-testid="inventoryHostForm-groups"
      >
        <template v-slot:chip="{ props, item: option }">
          <v-chip
            v-bind="props"
            :closable="option.source_id == null"
            :prepend-icon="option.source_id == null ? undefined : 'mdi-lock'"
            :aria-label="option.source_id == null
              ? undefined
              : $t('inventory_source_group_locked', { name: option.name })"
          ></v-chip>
        </template>
      </v-autocomplete>

      <div class="mt-2 mb-1 text-medium-emphasis">{{ $t('inventory_hosts_variables') }}</div>
      <CodeMirror
        class="InventoryHostForm__editor"
        :class="{ 'InventoryHostForm__editor--readonly': imported }"
        :style="{ border: '1px solid lightgray' }"
        v-model="item.variables"
        minimal
        :tab-size="2"
        :extensions="cmExtensions"
        :disabled="formSaving"
        :readonly="imported"
        data-testid="inventoryHostForm-variables"
      />
    </template>
  </v-form>
</template>
<style>
.InventoryHostForm__editor .cm-editor {
  height: 160px !important;
}

.InventoryHostForm__editor--readonly .cm-editor {
  filter: brightness(0.9);
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
    groups: Array,
    sources: { type: Object, default: () => ({}) },
  },

  emits: ['error'],

  data() {
    return {
      initialGroups: [],
      cmExtensions: markRaw([
        ...codeEditorExtensions({ language: 'yaml', tab: 'spaces' }),
        CODEMIRROR5_LOOK,
      ]),
    };
  },

  computed: {
    imported() {
      return this.item != null && this.item.source_id != null;
    },

    sourceName() {
      return this.sources[this.item.source_id] || this.$t('inventory_source_unknown');
    },

    // Groups of a source are not offered, but those the host is in stay as locked chips
    groupOptions() {
      return this.groups.filter((group) => group.source_id == null
        || this.initialGroups.includes(group.id));
    },
  },

  methods: {
    afterLoadData() {
      this.item.groups = this.item.groups || [];
      this.initialGroups = [...this.item.groups];
    },

    save() {
      // Save pressed before the item loaded
      if (this.item == null) {
        this.$emit('error', {});
        return null;
      }
      const manual = (id) => !this.groups
        .some((group) => group.id === id && group.source_id != null);
      const added = this.item.groups.filter((id) => !this.initialGroups.includes(id));
      const removed = this.initialGroups.filter((id) => !this.item.groups.includes(id));
      return ItemFormBase.methods.save.call(this, {
        groups: undefined,
        add_groups: added.filter(manual),
        remove_groups: removed.filter(manual),
      });
    },

    getNewItem() {
      return {
        name: '', description: '', enabled: true, variables: '', groups: [],
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
