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
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[v => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-select
      v-model="item.ssh_key_id"
      :label="$t('userCredentials')"
      :items="keys"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || $t('user_credentials_required')]"
      required
      :disabled="formSaving"
    ></v-select>

    <v-select
        v-model="item.become_key_id"
        :label="$t('sudoCredentialsOptional')"
        clearable
        :items="loginPasswordKeys"
        item-value="id"
        item-text="name"
        :disabled="formSaving"
    ></v-select>

    <v-select
      v-model="item.type"
      :label="$t('type')"
      :rules="[v => !!v || $t('type_required')]"
      :items="inventoryTypes"
      item-value="id"
      item-text="name"
      required
      :disabled="formSaving"
    ></v-select>

    <v-text-field
      v-model.trim="item.inventory"
      :label="$t('pathToInventoryFile')"
      :rules="[v => !!v || $t('path_required')]"
      required
      :disabled="formSaving"
      v-if="item.type === 'file'"
    ></v-text-field>

    <v-select
      v-model="item.repository_id"
      :label="$t('repository') + ' (Optional)'"
      clearable
      :items="repositories"
      item-value="id"
      item-text="name"
      :disabled="formSaving"
      v-if="item.type === 'file'"
    ></v-select>

    <codemirror
        :style="{ border: '1px solid lightgray' }"
        v-model.trim="item.inventory"
        :options="cmOptions"
        v-if="item.type === 'static' || item.type === 'static-yaml'"
        :placeholder="$t('enterInventory')"
    />

    <v-alert
        dense
        text
        class="mt-4"
        type="info"
        v-if="item.type === 'static'"
    >
      {{ $t('staticInventoryExample') }}
      <pre style="font-size: 14px;">[website]
172.18.8.40
172.18.8.41</pre>
    </v-alert>

    <v-alert
        dense
        text
        class="mt-4"
        type="info"
        v-if="item.type === 'static-yaml'"
    >
      {{ $t('staticYamlInventoryExample') }}
      <pre style="font-size: 14px;">all:
  children:
    website:
      hosts:
        172.18.8.40:
        172.18.8.41:</pre>
    </v-alert>
  </v-form>
</template>
<style>
.CodeMirror {
  height: 160px !important;
}
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

import { codemirror } from 'vue-codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
import 'codemirror/addon/display/placeholder.js';

export default {
  mixins: [ItemFormBase],

  components: {
    codemirror,
  },

  data() {
    return {
      cmOptions: {
        tabSize: 2,
        mode: 'text/x-ini',
        lineNumbers: true,
        line: true,
        lint: true,
        indentWithTabs: false,
      },
      inventoryTypes: [{
        id: 'static',
        name: 'Static',
      }, {
        id: 'static-yaml',
        name: 'Static YAML',
      }, {
        id: 'file',
        name: 'File',
      }],
      keys: null,
      repositories: null,
    };
  },

  computed: {
    loginPasswordKeys() {
      if (this.keys == null) {
        return null;
      }
      return this.keys.filter((key) => key.type === 'login_password');
    },
  },

  async created() {
    [this.keys, this.repositories] = (await Promise.all([
      await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/keys`,
        responseType: 'json',
      }),
      await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/repositories`,
        responseType: 'json',
      }),
    ])).map((x) => x.data);
  },

  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.itemId}`;
    },
  },
};
</script>
