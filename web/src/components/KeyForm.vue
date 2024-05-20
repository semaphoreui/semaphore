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
        :label="$t('keyName')"
        :rules="[v => !!v || $t('name_required')]"
        required
        :disabled="formSaving"
    />

    <v-select
        v-model="item.type"
        :label="$t('type')"
        :rules="[v => (!!v || !canEditSecrets) || $t('type_required')]"
        :items="inventoryTypes"
        item-value="id"
        item-text="name"
        :required="canEditSecrets"
        :disabled="formSaving || !canEditSecrets"
    />

    <v-text-field
        v-model="item.login_password.login"
        :label="$t('loginOptional')"
        v-if="item.type === 'login_password'"
        :disabled="formSaving || !canEditSecrets"
    />

    <v-text-field
        v-model="item.login_password.password"
        :append-icon="showLoginPassword ? 'mdi-eye' : 'mdi-eye-off'"
        :label="$t('password')"
        :rules="[v => (!!v || !canEditSecrets) || $t('password_required')]"
        :type="showLoginPassword ? 'text' : 'password'"
        v-if="item.type === 'login_password'"
        :required="canEditSecrets"
        :disabled="formSaving || !canEditSecrets"
        autocomplete="new-password"
        @click:append="showLoginPassword = !showLoginPassword"
    />

    <v-text-field
      v-model="item.ssh.login"
      :label="$t('usernameOptional')"
      v-if="item.type === 'ssh'"
      :disabled="formSaving || !canEditSecrets"
    />

    <v-text-field
      v-model="item.ssh.passphrase"
      :append-icon="showSSHPassphrase ? 'mdi-eye' : 'mdi-eye-off'"
      label="Passphrase (Optional)"
      :type="showSSHPassphrase ? 'text' : 'password'"
      v-if="item.type === 'ssh'"
      :disabled="formSaving || !canEditSecrets"
      @click:append="showSSHPassphrase = !showSSHPassphrase"
    />

    <v-textarea
      outlined
      v-model="item.ssh.private_key"
      :label="$t('privateKey')"
      :disabled="formSaving || !canEditSecrets"
      :rules="[v => !canEditSecrets || !!v || $t('private_key_required')]"
      v-if="item.type === 'ssh'"
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
    };
  },

  computed: {
    canEditSecrets() {
      return this.isNew || this.item.override_secret;
    },
  },

  methods: {
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
