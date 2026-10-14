<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null && keys != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-radio-group v-model="item.type" row class="mt-0" :disabled="formSaving">
      <v-radio :label="$t('hostConfigTypeHost')" value="host"></v-radio>
      <v-radio :label="$t('hostConfigTypeUrl')" value="url"></v-radio>
    </v-radio-group>

    <v-text-field
      v-model.trim="item.host"
      :label="isHost ? $t('hostConfigHost') : $t('hostConfigUrl')"
      :placeholder="isHost ? 'github.com' : 'https://github.com/acme/'"
      :hint="isHost ? $t('hostConfigHostHint') : $t('hostConfigUrlHint')"
      persistent-hint
      :rules="[v => !!v || $t('hostConfigHostRequired')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-autocomplete
      v-model="item.ssh_key_id"
      :label="$t('credential')"
      :items="credentials"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || $t('hostConfigCredentialRequired')]"
      required
      :hint="isHost ? $t('hostConfigCredentialHostHint') : $t('hostConfigCredentialUrlHint')"
      persistent-hint
      :disabled="formSaving"
      outlined
      dense
      class="mt-4"
    ></v-autocomplete>
  </v-form>
</template>
<script>
import axios from 'axios';
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      keys: null,
    };
  },

  computed: {
    isHost() {
      return (this.item || {}).type !== 'url';
    },

    // A host mapping becomes an ssh config entry, so only an ssh key fits. A URL
    // mapping rewrites the URL, so a login/password can authenticate over https.
    credentials() {
      const allowed = this.isHost ? ['ssh'] : ['ssh', 'login_password'];
      return (this.keys || []).filter((key) => allowed.includes(key.type));
    },
  },

  watch: {
    // Switching to Host narrows the list, and Vuetify only blanks the field:
    // the filtered-out credential would stay in the model and reach the API.
    'item.type': function onTypeChange() {
      if (this.keys == null || !this.item.ssh_key_id) {
        return;
      }
      if (!this.credentials.some((key) => key.id === this.item.ssh_key_id)) {
        this.item.ssh_key_id = null;
      }
    },
  },

  async created() {
    this.keys = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },

  methods: {
    getNewItem() {
      return {
        type: 'host',
      };
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/host_configs`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/host_configs/${this.itemId}`;
    },
  },
};
</script>
