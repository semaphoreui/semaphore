<template>
  <v-form ref="form" lazy-validation v-model="formValid" v-if="item != null">
    <v-alert :value="formError" color="error" class="pb-2">{{ formError }} </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-text-field
      v-model="item.params.url"
      :label="$t('Server URL')"
      :disabled="formSaving"
      :rules="[(v) => !!v || $t('url_required')]"
      required
      data-testid="secretStorage-vaultURL"
      outlined
      dense
    ></v-text-field>

    <div v-if="item.type === 'vault'">
      <div class="d-flex justify-space-between align-center mb-2">
        <b style="font-size: 13px; margin-left: 5px">Token</b>
        <v-btn-toggle v-model="secretStorage" tile group mandatory>
          <v-btn value="database" small class="mr-0 mt-0" style="border-radius: 4px">
            Store in DB
          </v-btn>
          <v-btn value="env" small class="mr-0 mt-0" style="border-radius: 4px">
            From ENV
          </v-btn>
          <v-btn value="file" small class="mr-0 mt-0" style="border-radius: 4px">
            From File
          </v-btn>
        </v-btn-toggle>
      </div>

      <v-text-field
        v-if="secretStorage === 'database'"
        class="masked-secret-input"
        v-model="item.secret"
        :label="$t('Token')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('token_required')]"
        required
        data-testid="secretStorage-vaultToken"
        outlined
        dense
        append-icon="mdi-lock"
      ></v-text-field>

      <v-text-field
        v-else
        v-model="item.secret"
        :label="secretStorage === 'env' ? $t('Env var name') : $t('Path to the file')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('envvar_required')]"
        required
        data-testid="secretStorage-vaultTokenSource"
        outlined
        dense
      ></v-text-field>
    </div>

    <div v-else-if="item.type === 'dvls'">
      <v-checkbox
        class="pt-0 mb-2"
        style="margin-top: -5px"
        v-model="item.params.insecure_tls"
        label="Skip TLS certificate verification (insecure)"
        :disabled="formSaving"
      />

      <v-text-field
        v-model="item.params.vault_id"
        :label="$t('Vault ID')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('key_required')]"
        required
        data-testid="secretStorage-dvlsKey"
        outlined
        dense
      ></v-text-field>

      <v-text-field
        v-model="item.params.app_key"
        :label="$t('App Key')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('key_required')]"
        required
        data-testid="secretStorage-dvlsKey"
        outlined
        dense
      ></v-text-field>

      <div class="d-flex justify-space-between align-center">
        <b style="font-size: 13px; margin-left: 5px">App secret</b>
        <v-btn-toggle v-model="secretStorage" tile group mandatory>
          <v-btn value="database" small class="mr-0 mt-0" style="border-radius: 4px">
            Store in DB
          </v-btn>
          <v-btn value="env" small class="mr-0 mt-0" style="border-radius: 4px">
            From ENV
          </v-btn>
          <v-btn value="file" small class="mr-0 mt-0" style="border-radius: 4px">
            From File
          </v-btn>
        </v-btn-toggle>
      </div>

      <v-text-field
        v-if="secretStorage === 'database'"
        class="TextInput TextInput--no-legend masked-secret-input"
        v-model="item.secret"
        :label="$t('Secret')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('secret_required')]"
        required
        data-testid="secretStorage-dvlsSecret"
        outlined
        dense
        append-icon="mdi-lock"
      ></v-text-field>

      <v-text-field
        v-else
        class="TextInput TextInput--no-legend"
        v-model="item.secret"
        :label="secretStorage === 'env' ? $t('Env var name') : $t('Path to the file')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('envvar_required')]"
        required
        data-testid="secretStorage-dvlsEnv"
        outlined
        dense
      ></v-text-field>
    </div>

    <v-checkbox v-model="item.readonly" :label="$t('Read only')" :disabled="formSaving" />
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  props: {
    itemType: String,
  },

  mixins: [ItemFormBase],

  data() {
    return {
      secretStorage: 'database',
      secretStorageReady: false,
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

      if (this.itemId === 'new') {
        this.item.type = this.itemType;
      }

      this.secretStorageReady = false;
      this.secretStorage = this.item.source_storage_type || 'database';
      this.$nextTick(() => {
        this.secretStorageReady = true;
      });
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/secret_storages`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/secret_storages/${this.itemId}`;
    },
  },

  watch: {
    secretStorage(value, oldValue) {
      this.item.source_storage_type = value === 'database' ? undefined : value;

      if (!this.secretStorageReady || value === oldValue) {
        return;
      }

      this.item.secret = '';
    },
  },
};
</script>
