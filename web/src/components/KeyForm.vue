<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null && secretStorages != null"
  >
    <v-alert :value="formError" color="error" class="mb-6">{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('keyName')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    />

    <v-card
      class="mb-6"
      :color="$vuetify.theme.dark ? '#212121' : 'white'"
      style="background: #8585850f"
    >
      <v-tabs fixed-tabs v-model="sourceStorageTypeIndex">
        <v-tab :disabled="formSaving || !canEditSecrets || isSynced" style="padding: 0"
          >Local</v-tab
        >
        <v-tab
          v-if="isPro"
          :disabled="formSaving || !canEditSecrets || isSynced"
          style="padding: 0"
          >
            Storage
          </v-tab
        >
        <v-tab :disabled="formSaving || !canEditSecrets || isSynced" style="padding: 0">Env</v-tab>
        <v-tab :disabled="formSaving || !canEditSecrets || isSynced" style="padding: 0">File</v-tab>
      </v-tabs>

      <div
        :class="!supportStorages && sourceStorageType === 'vault' ? '' : 'ml-4 mr-4 mt-6'"
        v-if="sourceStorageType"
      >
        <v-alert
          text
          color="hsl(348deg, 86%, 61%)"
          class="PageAlert PageAlert--flat-top"
          v-if="!supportStorages && sourceStorageType === 'vault'"
        >
          <span v-html="$t('project_runners_only_pro')"></span>
          <v-btn dark class="ml-2" color="hsl(348deg, 86%, 61%)" @click="upgradeToPro()">
            {{ $t('upgrade_to_pro') }}
          </v-btn>
        </v-alert>

        <v-autocomplete
          v-if="supportStorages && sourceStorageType === 'vault'"
          v-model="item.source_storage_id"
          :label="$t('Storage')"
          :items="secretStorages"
          item-value="id"
          item-text="name"
          :disabled="formSaving || !canEditSecrets || isSynced"
          outlined
          dense
          clearable
        />

        <v-text-field
          v-if="supportStorages && sourceStorageType === 'vault' && item.source_storage_id != null"
          v-model="item.source_storage_key"
          :label="$t('Source Key')"
          :disabled="formSaving || !canEditSecrets || isSynced"
          outlined
          dense
        />

        <v-text-field
          v-if="['env', 'file'].includes(sourceStorageType)"
          v-model="item.source_storage_key"
          :label="
            sourceStorageType === 'env' ? $t('Environment variable name') : $t('Path to the file')
          "
          :rules="[(v) => !!v || $t('type_required')]"
          :disabled="formSaving || !canEditSecrets"
          outlined
          dense
        />
      </div>
    </v-card>

    <v-select
      v-model="item.type"
      :label="$t('type')"
      :rules="[(v) => !!v || !canEditSecrets || $t('type_required')]"
      :items="inventoryTypes"
      item-value="id"
      item-text="name"
      :required="canEditSecrets"
      :disabled="formSaving || !canEditSecrets"
      outlined
      dense
    />

    <v-alert v-if="isReadOnly" type="info" text>Read-only secret storage chosen.</v-alert>

    <v-text-field
      v-model="item.login_password.login"
      :label="$t('usernameOptional')"
      v-if="!isReadOnly && item.type === 'login_password'"
      :disabled="formSaving || !canEditSecrets"
      outlined
      dense
    />

    <v-text-field
      v-model="item.login_password.password"
      :append-icon="showLoginPassword ? 'mdi-eye' : 'mdi-eye-off'"
      :label="$t('password')"
      :rules="[(v) => !!v || !canEditSecrets || $t('password_required')]"
      :class="{ 'masked-secret-input': !showLoginPassword }"
      v-if="!isReadOnly && item.type === 'login_password'"
      :required="canEditSecrets"
      :disabled="formSaving || !canEditSecrets"
      autocomplete="new-password"
      @click:append="showLoginPassword = !showLoginPassword"
      outlined
      dense
    />

    <v-text-field
      v-model="item.ssh.login"
      :label="$t('usernameOptional')"
      v-if="!isReadOnly && item.type === 'ssh'"
      :disabled="formSaving || !canEditSecrets"
      outlined
      dense
    />

    <v-text-field
      v-model="item.ssh.passphrase"
      :append-icon="showSSHPassphrase ? 'mdi-eye' : 'mdi-eye-off'"
      label="Passphrase (Optional)"
      :class="{ 'masked-secret-input': !showSSHPassphrase }"
      v-if="!isReadOnly && item.type === 'ssh'"
      :disabled="formSaving || !canEditSecrets"
      @click:append="showSSHPassphrase = !showSSHPassphrase"
      outlined
      dense
    />

    <v-textarea
      outlined
      v-model="item.ssh.private_key"
      :label="$t('privateKey')"
      :disabled="formSaving || !canEditSecrets"
      :rules="[(v) => !canEditSecrets || !!v || $t('private_key_required')]"
      v-if="!isReadOnly && item.type === 'ssh'"
    />

    <v-textarea
      outlined
      v-model="item.ssh.certificate"
      :label="$t('certificate')"
      :disabled="formSaving || !canEditSecrets"
      v-if="!isReadOnly && item.type === 'ssh'"
    />

    <v-checkbox v-model="item.override_secret" :label="$t('override')" v-if="!isNew" />

    <v-alert dense text type="info" v-if="item.type === 'none'">
      {{ $t('useThisTypeOfKeyForHttpsRepositoriesAndForPlaybook') }}
    </v-alert>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],

  props: {
    supportStorages: Boolean,
  },

  data() {
    return {
      showLoginPassword: false,
      showSSHPassphrase: false,
      inventoryTypes: [
        {
          id: 'ssh',
          name: `${this.$t('keyFormSshKey')}`,
        },
        {
          id: 'login_password',
          name: `${this.$t('keyFormLoginPassword')}`,
        },
        {
          id: 'none',
          name: `${this.$t('keyFormNone')}`,
        },
      ],
      secretStorages: null,
      isSynced: false,
    };
  },

  computed: {

    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },

    sourceStorageType() {
      return this.item?.source_storage_type;
    },

    sourceStorageTypeIndex: {
      get() {
        return (
          {
            vault: 1,
            env: 2,
            file: 3,
          }[this.item.source_storage_type] || 0
        );
      },
      set(index) {
        this.item = {
          ...this.item,
          source_storage_type: [undefined, 'vault', 'env', 'file'][index],
        };
      },
    },

    canEditSecrets() {
      return this.isNew || this.item.override_secret;
    },

    isReadOnly() {
      if (!this.sourceStorageType) {
        return false;
      }

      if (['env', 'file'].includes(this.sourceStorageType)) {
        return true;
      }

      if (this.item.source_storage_id == null) {
        return false;
      }

      const storage = this.secretStorages.find((s) => s.id === this.item.source_storage_id);
      if (storage == null) {
        return false;
      }

      return storage.readonly;
    },
  },

  async created() {
    [this.secretStorages] = await Promise.all([this.loadProjectResources('secret_storages')]);
  },

  methods: {
    afterLoadData() {
      this.isSynced = JSON.parse(this.item.plain || '{}').dvls_id != null;
    },

    getNewItem() {
      return {
        ssh: {},
        login_password: {},
      };
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/keys`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/keys/${this.itemId}`;
    },
  },
};
</script>
