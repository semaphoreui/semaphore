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
      :items="sshKeys"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || $t('hostConfigCredentialRequired')]"
      required
      :hint="$t('hostConfigCredentialHint')"
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

    // A mapping is applied as an ssh identity, which only an ssh key can be.
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
