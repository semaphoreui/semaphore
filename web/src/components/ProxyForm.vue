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
      :rules="[v => !v || (v > 0 && v < 65536) || $t('proxyPortInvalid')]"
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
      proxyTypes: [{
        id: 'ssh',
        name: 'SSH',
      }],
    };
  },

  computed: {
    // Only SSH keys can authenticate against a jump host.
    sshKeys() {
      return (this.keys || []).filter((key) => key.type === 'ssh');
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
