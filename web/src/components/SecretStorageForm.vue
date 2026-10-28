<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null"
    data-testid="secretStorage-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      type="error"
      class="mb-6"
      closable
      data-testid="secretStorage-error"
    >{{ formError }} </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      data-testid="secretStorage-name"
      variant="outlined"
      density="compact"
    ></v-text-field>

    <v-text-field
      v-if="item.type !== 'aws_sm' && item.type !== 'azure_kv'"
      v-model="item.params.url"
      :label="$t('Server URL')"
      :disabled="formSaving"
      :rules="[(v) => !!v || $t('url_required')]"
      required
      data-testid="secretStorage-vaultURL"
      variant="outlined"
      density="compact"
    ></v-text-field>

    <div
      v-if="item.type === 'vault' || item.type === 'openbao'"
      data-testid="secretStorage-vaultFields"
    >
      <v-text-field
        v-model="item.params.mount"
        :label="$t('Mount')"
        hint="'secret' by default"
        :disabled="formSaving"
        data-testid="secretStorage-vaultMount"
        variant="outlined"
        density="compact"
      ></v-text-field>

      <v-text-field
        v-model="item.params.namespace"
        :label="$t('Namespace')"
        :hint="item.type === 'openbao'
          ? 'OpenBao namespaces (v2.3+)'
          : 'For Vault Enterprise and HCP Dedicated only'"
        :disabled="formSaving"
        data-testid="secretStorage-vaultNamespace"
        variant="outlined"
        density="compact"
      ></v-text-field>

      <SecretSourceToggle v-model="secretStorage" label="Token" :disabled="formSaving" />

      <v-text-field
        v-if="secretStorage === 'database'"
        class="masked-secret-input"
        v-model="item.secret"
        :label="$t('Token')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('token_required')]"
        required
        data-testid="secretStorage-vaultToken"
        variant="outlined"
        density="compact"
        append-inner-icon="mdi-lock"
      ></v-text-field>

      <v-text-field
        v-else
        v-model="item.secret"
        :label="secretStorage === 'env' ? $t('Env var name') : $t('Path to the file')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('envvar_required')]"
        required
        data-testid="secretStorage-vaultTokenSource"
        variant="outlined"
        density="compact"
      ></v-text-field>
    </div>

    <div v-else-if="item.type === 'dvls'" data-testid="secretStorage-dvlsFields">
      <v-checkbox
        class="pt-0 mb-2"
        style="margin-top: -5px"
        v-model="item.params.insecure_tls"
        label="Skip TLS certificate verification (insecure)"
        :disabled="formSaving"
        data-testid="secretStorage-dvlsInsecureTls"
      />

      <v-text-field
        v-model="item.params.vault_id"
        :label="$t('Vault ID')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('key_required')]"
        required
        data-testid="secretStorage-dvlsKey"
        variant="outlined"
        density="compact"
      ></v-text-field>

      <v-text-field
        v-model="item.params.app_key"
        :label="$t('App Key')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('key_required')]"
        required
        data-testid="secretStorage-dvlsKey"
        variant="outlined"
        density="compact"
      ></v-text-field>

      <SecretSourceToggle v-model="secretStorage" label="App secret" :disabled="formSaving" />

      <v-text-field
        v-if="secretStorage === 'database'"
        class="TextInput TextInput--no-legend masked-secret-input"
        v-model="item.secret"
        :label="$t('Secret')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('secret_required')]"
        required
        data-testid="secretStorage-dvlsSecret"
        variant="outlined"
        density="compact"
        append-inner-icon="mdi-lock"
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
        variant="outlined"
        density="compact"
      ></v-text-field>
    </div>

    <div v-else-if="item.type === 'aws_sm'" data-testid="secretStorage-awsFields">
      <v-text-field
        v-model="item.params.region"
        label="Region"
        :disabled="formSaving"
        :rules="[(v) => !!v || 'Region is required']"
        required
        placeholder="us-east-1"
        data-testid="secretStorage-awsRegion"
        variant="outlined"
        density="compact"
      ></v-text-field>

      <v-text-field
        v-model="item.params.endpoint_url"
        label="Endpoint URL (optional)"
        :disabled="formSaving"
        hint="Leave empty to use the default AWS endpoint"
        data-testid="secretStorage-awsEndpointURL"
        variant="outlined"
        density="compact"
      ></v-text-field>

      <v-checkbox
        class="pt-0 mb-2"
        style="margin-top: -5px"
        v-model="item.params.use_iam_role"
        label="Use IAM Role / Instance Profile"
        :disabled="formSaving"
        data-testid="secretStorage-awsUseIamRole"
      />

      <div
        :class="{ 'aws-credentials--inactive': useIamRole }"
        data-testid="secretStorage-awsCredentials"
      >
        <v-text-field
          v-model="item.params.access_key_id"
          label="Access Key ID"
          :disabled="formSaving || useIamRole"
          :rules="[(v) => !!v || useIamRole || 'Access Key ID is required']"
          required
          data-testid="secretStorage-awsAccessKeyId"
          variant="outlined"
          density="compact"
        ></v-text-field>

        <SecretSourceToggle
          v-model="secretStorage"
          label="Secret Key"
          :disabled="formSaving || useIamRole"
        />

        <v-text-field
          v-if="secretStorage === 'database'"
          class="TextInput TextInput--no-legend masked-secret-input"
          v-model="item.secret"
          label="Secret Access Key"
          :disabled="formSaving || useIamRole"
          :rules="[(v) => !!v || !awsSecretRequired || 'Secret Access Key is required']"
          required
          data-testid="secretStorage-awsSecretKey"
          variant="outlined"
          density="compact"
          append-inner-icon="mdi-lock"
        ></v-text-field>

        <v-text-field
          v-else
          class="TextInput TextInput--no-legend"
          v-model="item.secret"
          :label="secretStorage === 'env' ? $t('Env var name') : $t('Path to the file')"
          :disabled="formSaving || useIamRole"
          :rules="[(v) => !!v || !awsSecretRequired || $t('envvar_required')]"
          required
          data-testid="secretStorage-awsSecretKeySource"
          variant="outlined"
          density="compact"
        ></v-text-field>
      </div>
    </div>

    <div v-else-if="item.type === 'azure_kv'" data-testid="secretStorage-azureFields">
      <v-text-field
        v-model="item.params.vault_url"
        label="Vault URL"
        :disabled="formSaving"
        :rules="[(v) => !!v || 'Vault URL is required']"
        required
        placeholder="https://my-vault.vault.azure.net"
        data-testid="secretStorage-azureVaultURL"
        variant="outlined"
        density="compact"
      ></v-text-field>

      <v-text-field
        v-model="item.params.tenant_id"
        label="Tenant ID"
        :disabled="formSaving"
        :rules="[(v) => !!v || 'Tenant ID is required']"
        required
        data-testid="secretStorage-azureTenantId"
        variant="outlined"
        density="compact"
      ></v-text-field>

      <v-text-field
        v-model="item.params.client_id"
        label="Client ID"
        :disabled="formSaving"
        :rules="[(v) => !!v || 'Client ID is required']"
        required
        data-testid="secretStorage-azureClientId"
        variant="outlined"
        density="compact"
      ></v-text-field>

      <SecretSourceToggle v-model="secretStorage" label="Client Secret" :disabled="formSaving" />

      <v-text-field
        v-if="secretStorage === 'database'"
        class="TextInput TextInput--no-legend masked-secret-input"
        v-model="item.secret"
        label="Client Secret"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || 'Client Secret is required']"
        required
        data-testid="secretStorage-azureClientSecret"
        variant="outlined"
        density="compact"
        append-inner-icon="mdi-lock"
      ></v-text-field>

      <v-text-field
        v-else
        class="TextInput TextInput--no-legend"
        v-model="item.secret"
        :label="secretStorage === 'env' ? $t('Env var name') : $t('Path to the file')"
        :disabled="formSaving"
        :rules="[(v) => !!v || itemId !== 'new' || $t('envvar_required')]"
        required
        data-testid="secretStorage-azureClientSecretSource"
        variant="outlined"
        density="compact"
      ></v-text-field>
    </div>

    <v-checkbox
      v-model="item.readonly"
      :label="$t('Read only')"
      :disabled="formSaving"
      hide-details
      style="position: absolute; bottom: 15px; margin: 0; left: 25px"
      data-testid="secretStorage-readonly"
    />

    <div class="d-flex items-center justify-space-between">
      <v-checkbox
        class="mt-0"
        v-model="item.sync_enabled"
        :label="$t('Sync keys enabled')"
        :disabled="formSaving"
        data-testid="secretStorage-syncEnabled"
      />

      <v-btn
        style="margin-right: -10px"
        variant="text"
        color="primary"
        @click="syncSettingsDialog = true"
        :disabled="formSaving"
        v-if="item.sync_enabled"
        data-testid="secretStorage-syncPaths"
      >
        <v-icon start>mdi-cog-sync</v-icon>
        Sync paths
        <v-chip
          class="ml-2"
          variant="outlined"
          style="transform: translateY(-1px)"
          color="primary"
          size="small"
          data-testid="secretStorage-syncPathsCount"
        >
          {{ item.sync_paths.length }}</v-chip
        >
      </v-btn>
    </div>

    <v-dialog v-model="syncSettingsDialog" max-width="500" persistent>
      <v-card data-testid="secretStorageSyncDialog">
        <v-card-title data-testid="secretStorageSyncDialog-title">Sync paths</v-card-title>
        <v-card-text class="pt-4 pb-0">
          <v-text-field
            style="width: 140px"
            v-if="item.sync_enabled"
            v-model.number="item.sync_interval"
            min="0"
            :label="$t('Auto-sync interval')"
            persistent-hint
            :disabled="formSaving"
            suffix="minutes"
            data-testid="secretStorageSync-syncInterval"
            variant="outlined"
            density="compact"
          ></v-text-field>

          <SecretStorageSyncOptionsForm
            v-model="item.sync_paths"
            data-testid="secretStorageSync-paths"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            color="blue-darken-1"
            @click="syncSettingsDialog = false"
            data-testid="secretStorageSyncDialog-close"
          >
            {{ $t('close') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import SecretStorageSyncOptionsForm from '@/components/SecretStorageSyncOptionsForm.vue';
import SecretSourceToggle from '@/components/SecretSourceToggle.vue';

export default {
  components: { SecretStorageSyncOptionsForm, SecretSourceToggle },

  props: {
    itemType: String,
  },

  mixins: [ItemFormBase],

  data() {
    return {
      secretStorage: 'database',
      secretStorageReady: false,
      syncSettingsDialog: false,
      // IAM role state of the storage at load time.
      initialUseIamRole: false,
    };
  },

  methods: {
    getNewItem() {
      return {
        sync_enabled: false,
        sync_interval: 0,
        sync_paths: [],
        params: {},
      };
    },

    afterLoadData() {
      if (!this.item.params) {
        this.item.params = {};
      }

      if (!this.item.sync_paths) {
        this.item.sync_paths = [];
      }

      if (this.itemId === 'new') {
        this.item.type = this.itemType;
      }

      if (this.item.type === 'aws_sm' && this.item.params.use_iam_role === undefined) {
        // Storages created before the IAM role option always had an access key.
        const useIamRole = this.itemId !== 'new' && !this.item.params.access_key_id;
        this.item.params.use_iam_role = useIamRole;
      }

      this.initialUseIamRole = !!this.item.params.use_iam_role;

      this.secretStorageReady = false;
      this.secretStorage = this.item.source_storage_type || 'database';
      this.$nextTick(() => {
        this.secretStorageReady = true;
      });
    },

    beforeSave() {
      if (this.useIamRole) {
        // Credentials are taken from the environment, don't send the disabled
        // secret fields to the server.
        this.item.secret = '';
        this.item.source_storage_type = undefined;
      }
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/secret_storages`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/secret_storages/${this.itemId}`;
    },
  },

  computed: {
    useIamRole() {
      return this.item?.params?.use_iam_role;
    },

    awsSecretRequired() {
      if (this.useIamRole) {
        return false;
      }

      // Switching the IAM role off removes the previously stored credentials,
      // so a new secret must be provided.
      return this.itemId === 'new' || this.initialUseIamRole;
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
<style lang="scss" scoped>
.aws-credentials--inactive {
  position: relative;
  &::after {
    content: "";
    position: absolute;
    background: white;
    left: 0;
    right: 0;
    top: 0;
    bottom: 0;
    opacity: 0.65;
  }
}

.v-theme--dark {
  .aws-credentials--inactive {
    position: relative;
    &::after {
      background: #1E1E1E;
    }
  }
}

</style>
