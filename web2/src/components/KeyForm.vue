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
        label="Key Name"
        :rules="[v => !!v || 'Name is required']"
        required
        :disabled="formSaving"
    />

    <v-select
        v-model="item.type"
        label="Type"
        :rules="[v => (!!v || !canEditSecrets) || 'Type is required']"
        :items="inventoryTypes"
        item-value="id"
        item-text="name"
        :required="canEditSecrets"
        :disabled="formSaving || !canEditSecrets"
    />

<!--    <v-text-field-->
<!--        v-model="item.login_password.passphrase"-->
<!--        label="Passphrase (Optional)"-->
<!--        v-if="item.type === 'ssh'"-->
<!--        :disabled="formSaving || !canEditSecrets"-->
<!--    />-->

    <v-textarea
        outlined
        v-model="item.ssh.private_key"
        label="Private Key"
        :disabled="formSaving || !canEditSecrets"
        :rules="[v => !!v || 'Private Key is required']"
        v-if="item.type === 'ssh'"
    />

    <v-text-field
        v-model="item.login_password.login"
        label="Login (Optional)"
        v-if="item.type === 'login_password'"
        :disabled="formSaving || !canEditSecrets"
    />

    <v-text-field
        v-model="item.login_password.password"
        label="Password"
        :rules="[v => (!!v || !canEditSecrets) || 'Password is required']"
        v-if="item.type === 'login_password'"
        :required="canEditSecrets"
        :disabled="formSaving || !canEditSecrets"
        autocomplete="new-password"
    />

    <v-text-field
      v-model="item.pat"
      label="Personal access token"
      v-if="item.type === 'pat'"
      :disabled="formSaving || !canEditSecrets"
    />

    <v-checkbox
        v-model="item.override_secret"
        label="Override"
        v-if="!isNew"
    />

    <v-alert
        dense
        text
        type="info"
        v-if="item.type === 'none'"
    >
      Use this type of key for HTTPS repositories and for
      playbooks which use non-SSH connections.
    </v-alert>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  data() {
    return {
      inventoryTypes: [{
        id: 'ssh',
        name: 'SSH Key',
      }, {
        id: 'login_password',
        name: 'Login with password',
      }, {
        id: 'pat',
        name: 'Personal access token',
      }, {
        id: 'none',
        name: 'None',
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
        pat: '',
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
