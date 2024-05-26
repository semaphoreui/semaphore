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
      :label="$t('environmentName')"
      :rules="[v => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      class="mb-4"
    ></v-text-field>

    <v-subheader class="pl-0">
      {{ $t('extraVariables') }}
    </v-subheader>

    <codemirror
      :style="{ border: '1px solid lightgray' }"
      v-model="json"
      :options="cmOptions"
      :placeholder="$t('enterExtraVariablesJson')"
    />

    <div class="mt-4">
      <div class="d-flex flex-row justify-space-between">
        <div>
          <div style="line-height: 1.1;" class="pl-1">
            Avoid host key checking by the tools Ansible uses to connect to the host.
          </div>
          <code>"ANSIBLE_HOST_KEY_CHECKING": false</code>
        </div>
        <v-btn
          color="primary"
          @click="setExtraVar('ANSIBLE_HOST_KEY_CHECKING', false)"
        >Set variable</v-btn>
      </div>
    </div>

    <v-alert
      dense
      text
      type="info"
      class="mt-4"
    >
      {{ $t('environmentAndExtraVariablesMustBeValidJsonExample') }}
      <pre style="font-size: 14px;">{
  "var_available_in_playbook_1": 1245,
  "var_available_in_playbook_2": "test"
}</pre>
    </v-alert>

    <div class="mt-4" v-if="!advancedOptions">
      <a @click="advancedOptions = true">
        {{ $t('advanced') }}
        <v-icon style="transform: translateY(-1px)">mdi-chevron-right</v-icon>
      </a>
    </div>

    <div class="mt-4" v-else>
      <a @click="advancedOptions = false">
        {{ $t('hide') }}
        <v-icon style="transform: translateY(-1px)">mdi-chevron-up</v-icon>
      </a>
    </div>

    <div v-if="advancedOptions">

      <v-subheader class="pl-0">
        {{ $t('environmentVariables') }}
      </v-subheader>

      <codemirror
        :style="{ border: '1px solid lightgray' }"
        v-model="env"
        :options="cmOptions"
        :placeholder="$t('enterEnvJson')"
      />

    </div>

  </v-form>
</template>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';

import { codemirror } from 'vue-codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
import 'codemirror/addon/display/placeholder.js';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  mixins: [ItemFormBase],
  components: {
    codemirror,
  },

  created() {
  },

  data() {
    return {
      images: [
        'dind-runner:latest',
      ],
      advancedOptions: false,
      json: '{}',
      env: '{}',

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
    setExtraVar(name, value) {
      try {
        const obj = JSON.parse(this.json || '{}');
        obj[name] = value;
        this.json = JSON.stringify(obj, null, 2);
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },

    beforeSave() {
      this.item.json = this.json;
      this.item.env = this.env;
    },

    afterLoadData() {
      this.json = this.item?.json || '{}';
      this.env = this.item?.env || '{}';
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/environment`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/environment/${this.itemId}`;
    },
  },
};
</script>
