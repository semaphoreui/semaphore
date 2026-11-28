<template>
  <v-form
    ref="form"
    v-model="formValid"
    data-testid="inventoryGroupForm-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="inventoryGroupForm-error"
    >{{ formError }}</v-alert>

    <template v-if="item != null">
      <v-alert
        v-if="imported"
        type="info"
        variant="tonal"
        density="compact"
        class="mb-4"
        data-testid="inventoryGroupForm-imported"
      >{{ $t('inventory_source_imported_group', { name: sourceName }) }}</v-alert>

      <v-text-field
        v-model.trim="item.name"
        :label="$t('name')"
        :rules="[v => !!v || $t('name_required')]"
        :hint="nameHint"
        persistent-hint
        required
        :readonly="imported"
        :append-inner-icon="imported ? 'mdi-lock' : undefined"
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        class="mb-2"
        data-testid="inventoryGroupForm-name"
      ></v-text-field>

      <v-text-field
        v-model="item.description"
        :label="$t('description')"
        :readonly="imported"
        :append-inner-icon="imported ? 'mdi-lock' : undefined"
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        data-testid="inventoryGroupForm-description"
      ></v-text-field>

      <v-autocomplete
        v-model="item.children"
        :items="childOptions"
        item-title="name"
        item-value="id"
        :label="$t('inventory_groups_children')"
        multiple
        chips
        :closable-chips="!imported"
        :readonly="imported"
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        data-testid="inventoryGroupForm-children"
      ></v-autocomplete>

      <div class="mt-2 mb-1 text-medium-emphasis">{{ $t('inventory_hosts_variables') }}</div>
      <CodeMirror
        class="InventoryGroupForm__editor"
        :class="{ 'InventoryGroupForm__editor--readonly': imported }"
        :style="{ border: '1px solid lightgray' }"
        v-model="item.variables"
        minimal
        :tab-size="2"
        :extensions="cmExtensions"
        :disabled="formSaving"
        :readonly="imported"
        data-testid="inventoryGroupForm-variables"
      />
    </template>
  </v-form>
</template>
<style>
.InventoryGroupForm__editor .cm-editor {
  height: 160px !important;
}

.InventoryGroupForm__editor--readonly .cm-editor {
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
      initialChildren: [],
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

    // The rule Ansible applies to group names, with letters of any alphabet
    nameHint() {
      return /^[^\p{L}_]|[^\p{L}\p{N}_]/u.test(this.item.name)
        ? this.$t('inventory_groups_name_hint')
        : '';
    },

    // The server rejects a cycle anyway, hiding the ancestors only spares the error
    childOptions() {
      const hidden = [this.itemId];
      for (let i = 0; i < hidden.length; i += 1) {
        this.groups.forEach((group) => {
          if (group.children.includes(hidden[i]) && !hidden.includes(group.id)) {
            hidden.push(group.id);
          }
        });
      }
      return this.groups.filter((group) => !hidden.includes(group.id));
    },
  },

  methods: {
    afterLoadData() {
      this.initialChildren = [...this.item.children];
    },

    save() {
      // Save pressed before the item loaded
      if (this.item == null) {
        this.$emit('error', {});
        return null;
      }
      return ItemFormBase.methods.save.call(this, {
        children: undefined,
        add_children: this.item.children.filter((id) => !this.initialChildren.includes(id)),
        remove_children: this.initialChildren.filter((id) => !this.item.children.includes(id)),
      });
    },

    getNewItem() {
      return {
        name: '', description: '', variables: '', children: [],
      };
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/groups`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/groups/${this.itemId}`;
    },
  },
};
</script>
