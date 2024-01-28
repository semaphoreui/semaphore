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
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      class="mb-4"
    ></v-text-field>

    <v-header>
      {{ $t("extraVariables") }}
    </v-header>

    <v-select
      v-model="extraVarsFormat"
      label="format"
      :items="formats"
      required
      :disabled="formSaving || formatSelectExtraVarsDisabled"
      @change="setExtraVarsWindow"
    ></v-select>

    <codemirror
      v-if="extraVarsFormat === 'yaml'"
      :style="{ border: '1px solid lightgray' }"
      :value="codeWindows.json"
      :options="cmOptions.json"
      @input="updateYamlInputExtraVars"
      :placeholder="$t('enterExtraVariablesYaml')"
    />

    <codemirror
      v-if="extraVarsFormat === 'json'"
      :style="{ border: '1px solid lightgray' }"
      :value="codeWindows.json"
      :options="cmOptions.json"
      @input="updateJsonInputExtraVars"
      :placeholder="$t('enterExtraVariablesJson')"
    />

    <v-spacer></v-spacer>

    <v-header>
      {{ $t("environmentVariables") }}
    </v-header>

    <v-select
      v-model="envFormat"
      label="format"
      :items="formats"
      required
      :disabled="formSaving || formatSelectEnvDisabled"
      @change="setEnvironmentWindow"
    ></v-select>

    <codemirror
      v-if="envFormat === 'yaml'"
      :style="{ border: '1px solid lightgray' }"
      :value="codeWindows.env"
      :options="cmOptions.env"
      @input="updateYamlInputEnv"
      :placeholder="$t('enterEnvYaml')"
    />

    <codemirror
      v-if="envFormat === 'json'"
      :style="{ border: '1px solid lightgray' }"
      :value="codeWindows.env"
      :options="cmOptions.env"
      @input="updateJsonInputEnv"
      :placeholder="$t('enterEnvJson')"
    />

    <v-alert
        v-if="envFormat == 'json' || extraVarsFormat == 'json'"
        dense
        text
        type="info"
        class="mt-4"
    >
      {{ $t("environmentAndExtraVariablesMustBeValidJsonExample") }}
      <pre style="font-size: 14px">{
  "var_available_in_playbook_1": 1245,
  "var_available_in_playbook_2": "test"
}</pre>
    </v-alert>

    <v-alert
        v-if="envFormat == 'yaml' || extraVarsFormat == 'yaml'"
        dense
        text
        type="info"
        class="mt-4"
    >
      {{ $t("environmentAndExtraVariablesMustBeValidYamlExample") }}
      <pre style="font-size: 14px">
var_available_in_playbook_1: 1245
var_available_in_playbook_2: "test"
</pre >
    </v-alert>
  </v-form>
</template>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';

import { codemirror } from 'vue-codemirror';
import { safeLoad, safeDump } from 'js-yaml';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
import 'codemirror/mode/yaml/yaml.js';
import 'codemirror/addon/display/placeholder.js';

export default {
  mixins: [ItemFormBase],
  components: {
    codemirror,
  },

  data() {
    const jsonFormat = 'json';
    const yamlFormat = 'yaml';
    const formats = [jsonFormat, yamlFormat];
    const modes = {
      json: 'application/json',
      yaml: 'text/x-yaml',
    };
    return {
      codeWindows: {
        env: '',
        json: '',
      },
      extraVarsFormat: jsonFormat,
      envFormat: jsonFormat,
      formats,
      modes,
      formatSelectEnvDisabled: false,
      formatSelectExtraVarsDisabled: false,
      cmOptions: {
        json: {
          tabSize: 2,
          mode: modes.json,
          lineNumbers: true,
          line: true,
          lint: true,
          indentWithTabs: false,
        },
        env: {
          tabSize: 2,
          mode: modes.json,
          lineNumbers: true,
          line: true,
          lint: true,
          indentWithTabs: false,
        },
      },
    };
  },

  watch: {
    formSaving(formSaving) {
      if (formSaving) {
        this.yamlToJson('env');
        this.yamlToJson('json');
        this.item.json = this.codeWindows.json;
        this.item.env = this.codeWindows.env;
      } else if (this.formError != null) {
        this.formError = this.parseErrorString(this.formError);
        if (this.envFormat === 'yaml') {
          this.jsonToYaml('env');
        }
        if (this.extraVarsFormat === 'yaml') {
          this.jsonToYaml('json');
        }
      }
    },
  },

  methods: {
    setFormat(format, windowName) {
      switch (format) {
        case 'json':
          this.cmOptions[windowName].mode = this.modes.json;
          this.yamlToJson(windowName);
          break;
        case 'yaml':
          this.cmOptions[windowName].mode = this.modes.yaml;
          this.jsonToYaml(windowName);
          break;
        default:
          this.cmOptions[windowName].mode = this.modes.json;
          this.yamlToJson(windowName);
          break;
      }
    },

    setEnvironmentWindow() {
      this.setFormat(this.envFormat, 'env');
    },

    setExtraVarsWindow() {
      this.setFormat(this.extraVarsFormat, 'json');
    },

    updateJsonInputExtraVars(updatedString) {
      this.formatSelectExtraVarsDisabled = !this.isJson(updatedString);
      this.codeWindows.json = updatedString;
    },

    updateYamlInputExtraVars(updatedString) {
      this.formatSelectExtraVarsDisabled = !this.isYaml(updatedString);
      this.codeWindows.json = updatedString;
    },

    updateJsonInputEnv(updatedString) {
      this.formatSelectEnvDisabled = !this.isJson(updatedString);
      this.codeWindows.env = updatedString;
    },

    updateYamlInputEnv(updatedString) {
      this.formatSelectEnvDisabled = !this.isYaml(updatedString);
      this.codeWindows.env = updatedString;
    },

    yamlToJson(windowName) {
      const yamlValue = this.codeWindows[windowName];
      console.log(yamlValue);
      if (this.isYaml(yamlValue)) {
        // yamlValue = yamlValue === '\n' ? '{}' : yamlValue;
        let jsonObject = safeLoad(yamlValue);
        console.log(jsonObject);
        jsonObject = !jsonObject ? {} : jsonObject;
        const jsonString = JSON.stringify(jsonObject, null, 2);
        this.codeWindows[windowName] = jsonString;
      }
    },

    jsonToYaml(windowName) {
      const jsonString = this.codeWindows[windowName];
      console.log(jsonString);
      if (this.isJson(jsonString)) {
        const jsonObject = JSON.parse(jsonString);
        let yamlValue = safeDump(jsonObject);
        console.log(yamlValue);
        yamlValue = yamlValue === '{}\n' ? '' : yamlValue;
        console.log(yamlValue);
        this.codeWindows[windowName] = yamlValue;
      }
    },

    isJson(str) {
      try {
        JSON.parse(str);
      } catch (e) {
        return false;
      }
      return true;
    },

    isYaml(str) {
      try {
        safeLoad(str);
      } catch (e) {
        return false;
      }
      return true;
    },

    parseErrorString(errorString) {
      if (errorString == null) {
        return errorString;
      }
      if (errorString.includes('Environment')) {
        if (this.envFormat === 'yaml') {
          return errorString.replace('JSON', 'YAML');
        }
      }
      if (this.extraVarsFormat === 'yaml') {
        return errorString.replace('JSON', 'YAML');
      }
      return errorString;
    },

    async afterLoadData() {
      this.codeWindows.json = this.item.json === undefined ? '{}' : this.item.json;
      this.codeWindows.env = this.item.env === undefined ? '{}' : this.item.env;
      this.setEnvironmentWindow();
      this.setExtraVarsWindow();
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
