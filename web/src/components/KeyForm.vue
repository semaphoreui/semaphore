<template>
  <v-form
      ref="form"
      lazy-validation
      v-model="formValid"
      v-if="item != null && secretStorages != null"
  >
    <v-alert
        :value="formError"
        color="error"
        class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-text-field
        v-model="item.name"
        :label="$t('keyName')"
        :rules="[v => !!v || $t('name_required')]"
        required
        :disabled="formSaving"
        outlined
        dense
    />

    <v-autocomplete
      v-if="supportStorages"
      v-model="item.source_storage_id"
      :label="$t('Storage (optional)')"
      :items="secretStorages"
      item-value="id"
      item-text="name"
      :disabled="formSaving || !canEditSecrets"
      outlined
      dense
      clearable
    />

    <v-text-field
      v-if="supportStorages && item.source_storage_id != null"
      v-model="item.source_storage_key"
      :label="$t('Source Key')"
      :disabled="formSaving || !canEditSecrets"
      outlined
      dense
    />

    <v-divider class="mb-6" />

    <v-select
        v-model="item.type"
        :label="$t('type')"
        :rules="[v => (!!v || !canEditSecrets) || $t('type_required')]"
        :items="inventoryTypes"
        item-value="id"
        item-text="name"
        :required="canEditSecrets"
        :disabled="formSaving || !canEditSecrets"
        outlined
        dense
    />

    <v-alert
      v-if="isReadOnly"
      type="info"
    >Read-only secret storage chosen.</v-alert>

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
        :rules="[v => (!!v || !canEditSecrets) || $t('password_required')]"
        :type="showLoginPassword ? 'text' : 'password'"
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
      :type="showSSHPassphrase ? 'text' : 'password'"
      v-if="!isReadOnly && item.type === 'ssh'"
      :disabled="formSaving || !canEditSecrets"
      @click:append="showSSHPassphrase = !showSSHPassphrase"
      outlined
      dense
    />

    <v-checkbox
      v-model="item.generate_ssh_key"
      label="Generate SSH Key"
      v-if="!isReadOnly && item.type === 'ssh'"
      :disabled="formSaving || !canEditSecrets"
    />

    <v-textarea
      outlined
      v-model="item.ssh.private_key"
      :label="$t('privateKey')"
      :disabled="formSaving || !canEditSecrets || item.generate_ssh_key"
      :rules="[v => !canEditSecrets || item.generate_ssh_key || !!v || $t('private_key_required')]"
      v-if="!isReadOnly && item.type === 'ssh'"
    />

    <v-checkbox
        v-model="item.override_secret"
        :label="$t('override')"
        v-if="!isNew"
    />

    <v-alert
        dense
        text
        type="info"
        v-if="item.type === 'none'"
    >
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
      inventoryTypes: [{
        id: 'ssh',
        name: `${this.$t('keyFormSshKey')}`,
      }, {
        id: 'login_password',
        name: `${this.$t('keyFormLoginPassword')}`,
      }, {
        id: 'none',
        name: `${this.$t('keyFormNone')}`,
      }],
      secretStorages: null,
      // isReadOnly: true,
    };
  },

  computed: {
    canEditSecrets() {
      return this.isNew || this.item.override_secret;
    },

    isReadOnly() {
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
    [
      this.secretStorages,
    ] = await Promise.all([
      this.loadProjectResources('secret_storages'),
    ]);
  },

  methods: {
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
