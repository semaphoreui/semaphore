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
      :items="proxyTypes"
      item-value="id"
      item-text="name"
      :disabled="formSaving"
      outlined
      dense
    ></v-select>

    <v-text-field
      v-model.trim="item.host"
      :label="$t('proxyHost')"
      :rules="[v => !!v || $t('proxyHostRequired')]"
      required
      :disabled="formSaving"
      outlined
      dense
      placeholder="bastion.example.org"
    ></v-text-field>

    <v-text-field
      v-model.number="item.port"
      :label="$t('proxyPort')"
      type="number"
      :rules="[v => v === '' || v == null || (v > 0 && v < 65536) || $t('proxyPortInvalid')]"
      :disabled="formSaving"
      outlined
      dense
      placeholder="22"
    ></v-text-field>

    <v-text-field
      v-model.trim="item.user"
      :label="$t('proxyUser')"
      :disabled="formSaving"
      outlined
      dense
      placeholder="ansible-proxy"
    ></v-text-field>

    <v-autocomplete
      v-model="item.ssh_key_id"
      :label="$t('accessKey')"
      :items="sshKeys"
      item-value="id"
      item-text="name"
      :hint="$t('proxyKeyHint')"
      persistent-hint
      clearable
      :disabled="formSaving"
      outlined
      dense
    ></v-autocomplete>

    <v-autocomplete
      v-if="isSshProxy"
      v-model="item.requires_proxy_id"
      :label="$t('requiresProxy')"
      :items="otherProxies"
      item-value="id"
      item-text="name"
      :hint="$t('requiresProxyHint')"
      persistent-hint
      clearable
      :disabled="formSaving"
      outlined
      dense
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
      proxies: null,
      proxyTypes: [{
        id: 'ssh',
        name: 'SSH (jump host)',
      }, {
        id: 'socks5',
        name: 'SOCKS5',
      }, {
        id: 'http',
        name: 'HTTP',
      }, {
        id: 'https',
        name: 'HTTPS',
      }],
    };
  },

  computed: {
    isSshProxy() {
      return (this.item || {}).type === 'ssh';
    },

    // A jump host authenticates with an ssh key; a SOCKS or HTTP proxy uses a
    // login and a password.
    sshKeys() {
      const wanted = this.isSshProxy ? 'ssh' : 'login_password';
      return (this.keys || []).filter((key) => key.type === wanted);
    },

    // A proxy can not require itself, which the API rejects as well.
    otherProxies() {
      return (this.proxies || []).filter((proxy) => proxy.id !== this.itemId);
    },
  },

  async created() {
    this.keys = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;

    this.proxies = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/proxies`,
      responseType: 'json',
    })).data;
  },

  methods: {
    beforeSave() {
      // v-model.number keeps the raw string when it can not be parsed, so a
      // cleared field submits "" which does not decode into the nullable int
      // of the API.
      if (this.item.port === '' || this.item.port == null) {
        this.item.port = null;
      }
    },

    getNewItem() {
      return {
        type: 'ssh',
      };
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/proxies`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/proxies/${this.itemId}`;
    },
  },
};
</script>
