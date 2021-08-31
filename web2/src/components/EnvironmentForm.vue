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
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      label="Environment Name"
      :rules="[v => !!v || 'Name is required']"
      required
      :disabled="formSaving"
      class="mb-4"
    ></v-text-field>

    <codemirror
        :style="{ border: '1px solid lightgray' }"
        v-model="item.json"
        :options="cmOptions"
        placeholder="Enter environment JSON..."
    />

    <v-alert
        dense
        type="info"
        class="mt-4"
    >
      Must be valid JSON. You may use the key <code>ENV</code> to pass environment variables
      to ansible-playbook.
      Example:
      <pre style="font-size: 14px;">{
  "var_available_in_playbook_1": 1245,
  "var_available_in_playbook_2": "test",
  "ENV": {
    "VAR1": "Read by lookup('env', 'VAR1')"
  }
}</pre>
    </v-alert>
  </v-form>
</template>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';

import { codemirror } from 'vue-codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
// import 'codemirror/addon/lint/json-lint.js';
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
        mode: 'application/json',
        lineNumbers: true,
        line: true,
        lint: true,
        indentWithTabs: false,
      },
    };
  },

  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/environment`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/environment/${this.itemId}`;
    },
  },
};
</script>
