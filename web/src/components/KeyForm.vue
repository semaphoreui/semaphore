<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null && secretStorages != null"
    class="KeyForm"
    data-testid="key-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="mb-6"
      data-testid="key-error"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('keyName')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      data-testid="key-name"
    />

    <v-card
      class="mb-6"
      :color="$vuetify.theme.current.dark ? '#212121' : 'white'"
      data-testid="key-source"
    >
      <v-tabs
        fixed-tabs
        v-model="sourceStorageTypeIndex"
        class="KeyForm__sourceTabs"
        data-testid="key-sourceTabs"
      >
        <v-tab
          :disabled="formSaving || !canEditSecrets || isSynced"
          style="padding: 0"
          data-testid="key-tab-local"
          >Local</v-tab
        >
        <v-tab
          v-if="isPro"
          :disabled="formSaving || !canEditSecrets || isSynced"
          style="padding: 0"
          data-testid="key-tab-storage"
          >
            Storage
          </v-tab
        >
        <v-tab
          :disabled="formSaving || !canEditSecrets || isSynced"
          style="padding: 0"
          data-testid="key-tab-env"
        >Env</v-tab>
        <v-tab
          :disabled="formSaving || !canEditSecrets || isSynced"
          style="padding: 0"
          data-testid="key-tab-file"
        >File</v-tab>
      </v-tabs>

      <div
        :class="!supportStorages && sourceStorageType === 'vault' ? '' : 'ml-4 mr-4 mt-6'"
        v-if="sourceStorageType"
        data-testid="key-sourceBody"
      >
        <v-alert
          variant="tonal"
          color="hsl(348deg, 86%, 61%)"
          class="PageAlert PageAlert--flat-top mb-0"
          v-if="!supportStorages && sourceStorageType === 'vault'"
          data-testid="key-storageUpgradeAlert"
        >
          <span v-html="$t('project_runners_only_pro')"></span>
          <v-btn
            class="ml-2"
            color="hsl(348deg, 86%, 61%)"
            @click="upgradeToPro()"
            data-testid="key-storageUpgrade"
          >
            {{ $t('upgrade_to_pro') }}
          </v-btn>
        </v-alert>

        <v-autocomplete
          v-if="supportStorages && sourceStorageType === 'vault'"
          v-model="item.source_storage_id"
          :label="$t('Storage')"
          :items="secretStorages"
          item-value="id"
          item-title="name"
          :disabled="formSaving || !canEditSecrets || isSynced"
          variant="outlined"
          density="compact"
          clearable
          :menu-props="storageMenu"
          data-testid="key-sourceStorageId"
        />

        <v-text-field
          v-if="supportStorages && sourceStorageType === 'vault' && item.source_storage_id != null"
          v-model="item.source_storage_key"
          :label="$t('Source Key')"
          :disabled="formSaving || !canEditSecrets || isSynced"
          variant="outlined"
          density="compact"
          data-testid="key-sourceStorageKey"
        />

        <v-text-field
          v-if="['env', 'file'].includes(sourceStorageType)"
          v-model="item.source_storage_key"
          :label="
            sourceStorageType === 'env' ? $t('Environment variable name') : $t('Path to the file')
          "
          :rules="[(v) => !!v || $t('type_required')]"
          :disabled="formSaving || !canEditSecrets"
          variant="outlined"
          density="compact"
          :data-testid="`key-sourceStorageKey-${sourceStorageType}`"
        />
      </div>
    </v-card>

    <v-select
      v-model="item.type"
      :label="$t('type')"
      :rules="[(v) => !!v || !canEditSecrets || $t('type_required')]"
      :items="inventoryTypes"
      item-value="id"
      item-title="name"
      :required="canEditSecrets"
      :disabled="formSaving || !canEditSecrets"
      variant="outlined"
      density="compact"
      data-testid="key-type"
    />

    <v-alert
      v-if="isReadOnly"
      type="info"
      variant="tonal"
      data-testid="key-readOnlyAlert"
    >Read-only secret storage chosen.</v-alert>

    <v-text-field
      v-model="item.login_password.login"
      :label="$t('usernameOptional')"
      v-if="!isReadOnly && item.type === 'login_password'"
      :disabled="formSaving || !canEditSecrets"
      variant="outlined"
      density="compact"
      data-testid="key-loginPasswordLogin"
    />

    <v-text-field
      v-model="item.login_password.password"
      :append-inner-icon="showLoginPassword ? 'mdi-eye' : 'mdi-eye-off'"
      :label="$t('password')"
      :rules="[(v) => !!v || !canEditSecrets || $t('password_required')]"
      :class="{ 'masked-secret-input': !showLoginPassword }"
      v-if="!isReadOnly && item.type === 'login_password'"
      :required="canEditSecrets"
      :disabled="formSaving || !canEditSecrets"
      autocomplete="new-password"
      @click:append-inner="showLoginPassword = !showLoginPassword"
      variant="outlined"
      density="compact"
      data-testid="key-loginPasswordPassword"
    />

    <v-text-field
      v-model="item.ssh.login"
      :label="$t('usernameOptional')"
      v-if="!isReadOnly && item.type === 'ssh'"
      :disabled="formSaving || !canEditSecrets"
      variant="outlined"
      density="compact"
      data-testid="key-sshLogin"
    />

    <v-text-field
      v-model="item.ssh.passphrase"
      :append-inner-icon="showSSHPassphrase ? 'mdi-eye' : 'mdi-eye-off'"
      label="Passphrase (Optional)"
      :class="{ 'masked-secret-input': !showSSHPassphrase }"
      v-if="!isReadOnly && item.type === 'ssh'"
      :disabled="formSaving || !canEditSecrets"
      @click:append-inner="showSSHPassphrase = !showSSHPassphrase"
      variant="outlined"
      density="compact"
      data-testid="key-sshPassphrase"
    />

    <v-checkbox
      v-model="item.generate_ssh_key"
      label="Generate SSH Key"
      v-if="!isReadOnly && item.type === 'ssh'"
      :disabled="formSaving || !canEditSecrets"
      class="mt-0 mb-3"
      hide-details
      data-testid="key-generateSshKey"
    />

    <v-textarea
      variant="outlined"
      v-model="item.ssh.private_key"
      :label="$t('privateKey')"
      :disabled="formSaving || !canEditSecrets || item.generate_ssh_key"
      :rules="
        [
          (v) => !canEditSecrets || item.generate_ssh_key || !!v || $t('private_key_required')
        ]"
      v-if="!isReadOnly && item.type === 'ssh'"
      class="KeyForm__privateKey"
      data-testid="key-sshPrivateKey"
    />

    <div
      v-if="!item.override_secret && item.type === 'ssh' && !isNew && hasGeneratedPublicKey"
      class="mb-4"
      data-testid="key-publicKey"
    >
      <div class="pb-1">Public key:</div>
      <div style="position: relative">
        <pre
          style="
            overflow: hidden;
            background: gray;
            color: white;
            border-radius: 10px;
            margin-top: 0;
            white-space: normal;
            height: 35px;
          "
          :style="{height: showPublicKey ? 'auto' : '35px' }"
          class="pa-2"
          data-testid="key-publicKeyValue"
          >{{ publicKey }}</pre
        >

        <v-btn
          style="position: absolute; right: 40px; top: 0; transform: scale(0.9);"
          variant="text"
          @click="showPublicKey = !showPublicKey"
          data-testid="key-publicKeyToggle"
        >
          {{ showPublicKey ? 'Hide' : 'Show' }}
        </v-btn>

        <CopyClipboardButton
          style="position: absolute; right: 0; top: 0; transform: scale(0.9);"
          :text="publicKey"
          test-id="key-publicKeyCopy"
        />
      </div>
    </div>

    <v-checkbox
        style="position: absolute; bottom: 18px;"
        v-model="item.override_secret"
        :label="$t('override')"
        v-if="!isNew"
        hide-details
        class="mt-0"
        data-testid="key-overrideSecret"
    />

    <v-alert
      density="compact"
      variant="tonal"
      type="info"
      v-if="item.type === 'none'"
      data-testid="key-noneInfo"
    >
      {{ $t('useThisTypeOfKeyForHttpsRepositoriesAndForPlaybook') }}
    </v-alert>
  </v-form>
</template>
<style lang="scss">
// The source card is white (#212121 in the dark theme) as in Vuetify 2, where the colour class
// of `color` hid the inline grey background the card had (removed: in Vuetify 4 it won).

// The source tabs share the card width equally, as `fixed-tabs` of Vuetify 2 did
// (`.v-tabs--fixed-tabs .v-tab { flex: 1 1 auto; width: 100% }`); Vuetify 4 sizes them by text.
.KeyForm__sourceTabs .v-tab {
  flex: 1 1 auto;
  width: 100%;
}

// The outlined textarea of Vuetify 2 (`.v-textarea textarea`): lines of 28 px, 10 px above the
// text, none below — 150 px for 5 rows (Vuetify 4: 24 px lines between 16 px paddings, 152 px).
.KeyForm__privateKey textarea.v-field__input {
  padding-top: 10px;
  padding-bottom: 0;
  line-height: 28px;
}

// The message of a disabled field (the stale "Private key is required" under Generate SSH Key)
// in the disabled text colour, not dimmed again (Vuetify 2: no error state while disabled; the
// grey outline — App.vue).
.KeyForm .v-input--disabled > .v-input__details {
  opacity: 1;

  .v-messages {
    color: rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
  }
}

// A disabled checkbox (Generate SSH Key, Override) in the disabled text colour
// (`.v-input--is-disabled` of Vuetify 2), not the medium emphasis dimmed to .38 again.
.KeyForm .v-selection-control--disabled {
  opacity: 1;

  .v-label,
  .v-selection-control__input > .v-icon {
    color: rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
    opacity: 1;
  }
}
</style>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';

export default {
  components: {
    CopyClipboardButton,
  },

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
      showPublicKey: false,
      // The list of the storage autocomplete 1 px under the field and at most 304 px high, as the
      // `offset-y` menu of Vuetify 2 (`menuProps`; Vuetify 4: flush with the field, 310 px).
      storageMenu: { maxHeight: 304, offset: 1 },
    };
  },

  computed: {
    hasGeneratedPublicKey() {
      return this.publicKey !== '';
    },

    publicKey: {
      get() {
        try {
          const plain = JSON.parse(this.item?.plain || '{}');
          return plain.public_key || '';
        } catch (e) {
          return '';
        }
      },
    },

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

    beforeSave() {
      // The checkbox is hidden for non-ssh types and read-only storages but
      // keeps its value, and the server rejects generate_ssh_key in both cases.
      // Generation only makes sense when the secret is being overridden.
      if (this.item.type !== 'ssh' || this.isReadOnly || (!this.isNew && !this.item.override_secret)) {
        this.item.generate_ssh_key = false;
      }
    },

    getNewItem() {
      return {
        ssh: {},
        login_password: {},
        generate_ssh_key: false,
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
