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
    >{{ formError }}
    </v-alert>

    <v-text-field
        v-model="item.name"
        :label="$t('name')"
        :rules="[v => !!v || $t('name_required')]"
        required
        :disabled="formSaving"
        outlined
        dense
    ></v-text-field>

    <v-select
        v-model="item.type"
        :label="$t('type')"
        :rules="[v => !!v || $t('type_required')]"
        :items="secretStorageTypes"
        item-value="id"
        item-text="name"
        required
        :disabled="formSaving"
        outlined
        dense
    />

    <div v-if="item.type === 'vault'">

      <v-text-field
          v-model="item.params.url"
          :label="$t('Server URL')"
          :disabled="formSaving"
          :rules="[v => !!v || $t('url_required')]"
          required
          data-testid="secretStorage-vaultURL"
          outlined
          dense
      ></v-text-field>

      <v-text-field
          v-model="item.vault_token"
          :label="$t('Token')"
          :disabled="formSaving"
          :rules="[v => !!v || $t('token_required')]"
          required
          data-testid="secretStorage-vaultToken"
          outlined
          dense
          append-icon="mdi-lock"
      ></v-text-field>

    </div>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  props: {},

  mixins: [ItemFormBase],

  data() {
    return {
      secretStorageTypes: [{
        id: 'vault',
        name: 'Hashicorp Vault',
      }],
    };
  },

  methods: {

    getNewItem() {
      return {
        params: {},
      };
    },

    afterLoadData() {
      if (!this.item.params) {
        this.item.params = {};
      }
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/secret_storages`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/secret_storages/${this.itemId}`;
    },
  },
};
</script>
