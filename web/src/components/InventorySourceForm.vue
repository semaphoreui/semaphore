<template>
  <v-form
    ref="form"
    v-model="formValid"
    data-testid="inventorySourceForm-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="inventorySourceForm-error"
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
        class="mb-2"
        data-testid="inventorySourceForm-name"
      ></v-text-field>

      <v-select
        v-model="item.repository_id"
        :label="$t('repository')"
        :items="repositories"
        item-value="id"
        item-title="name"
        :rules="[v => !!v || $t('repository_required')]"
        required
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        class="mb-2"
        data-testid="inventorySourceForm-repository"
      ></v-select>

      <v-text-field
        v-model.trim="item.params.path"
        :label="$t('path')"
        :rules="[v => !!v || $t('path_required')]"
        required
        placeholder="inventory/prod.yml"
        persistent-placeholder
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        data-testid="inventorySourceForm-path"
      ></v-text-field>
    </template>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],

  props: {
    inventoryId: [Number, String],
  },

  emits: ['error'],

  data() {
    return {
      repositories: [],
    };
  },

  methods: {
    async beforeLoadData() {
      this.repositories = await this.loadProjectResources('repositories');
    },

    getNewItem() {
      return {
        name: '', type: 'git', repository_id: null, params: { path: '' },
      };
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/sources`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.inventoryId}/sources/${this.itemId}`;
    },
  },
};
</script>
