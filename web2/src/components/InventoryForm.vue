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
      label="Name"
      :rules="[v => !!v || 'Name is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-select
      v-model="item.ssh_key_id"
      label="User Credentials"
      :items="keys"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || 'User Credentials is required']"
      required
      :disabled="formSaving"
    ></v-select>

    <v-select
        v-model="item.become_key_id"
        label="Sudo Credentials (Optional)"
        clearable
        :items="loginPasswordKeys"
        item-value="id"
        item-text="name"
        :disabled="formSaving"
    ></v-select>

    <v-select
      v-model="item.type"
      label="Type"
      :rules="[v => !!v || 'Type is required']"
      :items="inventoryTypes"
      item-value="id"
      item-text="name"
      required
      :disabled="formSaving"
    ></v-select>

    <v-text-field
      v-model="item.inventory"
      label="Path to Inventory file"
      :rules="[v => !!v || 'Path to Inventory file is required']"
      required
      :disabled="formSaving"
      v-if="item.type === 'file'"
    ></v-text-field>

    <codemirror
        :style="{ border: '1px solid lightgray' }"
        v-model="item.inventory"
        :options="cmOptions"
        v-if="item.type === 'static'"
        placeholder="Enter inventory..."
    />

    <v-alert
        dense
        text
        class="mt-4"
        type="info"
        v-if="item.type === 'static'"
    >
      Static inventory example:
      <pre style="font-size: 14px;">[website]
172.18.8.40
172.18.8.41</pre>
    </v-alert>
  </v-form>
</template>
<style>
.CodeMirror {
  height: 200px !important;
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
      keys: null,
      inventoryTypes: [{
        id: 'static',
        name: 'Static',
      }, {
        id: 'file',
        name: 'File',
      }],
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
    this.keys = (await axios({
      keys: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
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
