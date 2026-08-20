<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null && (!supportStorages || secretStorages != null)"
    class="pb-3"
  >
    <v-alert :value="formError" color="error" data-testid="varGroup-error"
      >{{ formError }}
    </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('environmentName')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-row v-if="supportStorages && isNew">
      <v-col>
        <v-autocomplete
          v-model="item.secret_storage_id"
          :label="$t('Secret storage (optional)')"
          :items="secretStorages"
          :disabled="formSaving || !isNew"
          item-value="id"
          item-text="name"
          outlined
          dense
          clearable
        />
      </v-col>
      <v-col>
        <v-text-field
          v-model="item.secret_storage_key_prefix"
          :label="$t('Secret key prefix')"
          :disabled="formSaving || !item.secret_storage_id || !isNew"
          outlined
          dense
        />
      </v-col>
    </v-row>

    <v-tabs grow v-model="tab">
      <v-tab key="variables">Variables</v-tab>
      <v-tab key="secrets">Secrets</v-tab>
    </v-tabs>

    <v-divider style="margin-top: -1px" class="mb-7" />

    <v-tabs-items v-model="tab">
      <v-tab-item key="variables">
        <v-subheader class="px-0">
          {{ $t('extraVariables') }}

          <v-tooltip v-if="needHelp" bottom color="black" open-delay="300" max-width="400">
            <template v-slot:activator="{ on, attrs }">
              <v-icon class="ml-1" v-bind="attrs" v-on="on">mdi-help-box </v-icon>
            </template>
            <div>
              <div><code>--extra-vars</code> for Ansible</div>
              <div><code>-var</code> for Terraform/OpenTofu</div>
            </div>
          </v-tooltip>

          <v-spacer />

          <v-btn-toggle v-model="extraVarsEditMode" tile group>
            <v-btn value="table" small class="mr-0" style="border-radius: 4px"> Table </v-btn>
            <v-btn value="json" small class="mr-0" style="border-radius: 4px"> JSON </v-btn>
            <v-btn value="yaml" small class="mr-0" style="border-radius: 4px"> YAML </v-btn>
          </v-btn-toggle>

          <v-btn icon @click="addExtraVar()" data-testid="varGroup-addVar">
            <v-icon> mdi-plus </v-icon>
          </v-btn>
        </v-subheader>

        <div v-if="extraVarsEditMode === 'json'" style="position: relative">
          <codemirror
            :class="{
              EnvironmentEditor: true,
            }"
            :style="{ border: '1px solid lightgray' }"
            v-model="json"
            :options="cmOptions"
            :placeholder="$t('enterExtraVariablesJson')"
          />

          <RichEditor
            v-model="json"
            type="json"
            v-if="extraVarsEditMode === 'json'"
            style="position: absolute; right: 0; top: 0; margin: 10px"
          />
        </div>
        <div v-else-if="extraVarsEditMode === 'yaml'" style="position: relative">
          <codemirror
            :class="{
              EnvironmentEditor: true,
            }"
            :style="{ border: '1px solid lightgray' }"
            v-model="yaml"
            :options="cmYamlOptions"
            :placeholder="$t('enterExtraVariablesYaml')"
          />

          <RichEditor
            v-model="yaml"
            type="yaml"
            v-if="extraVarsEditMode === 'yaml'"
            style="position: absolute; right: 0; top: 0; margin: 10px"
          />
        </div>
        <div v-else-if="extraVarsEditMode === 'table'">
          <v-data-table
            v-if="extraVars != null"
            :items="extraVars"
            :items-per-page="-1"
            class="elevation-1 FieldTable"
            hide-default-footer
            :no-data-text="$t('noValues')"
            style="background: #8585850f"
          >
            <template v-slot:item="props">
              <tr>
                <td class="pa-1">
                  <v-text-field
                    solo-inverted
                    flat
                    hide-details
                    v-model="props.item.name"
                    class="v-text-field--solo--no-min-height"
                    :placeholder="$t('name')"
                  ></v-text-field>
                </td>
                <td class="pa-1" style="width: 130px">
                  <v-select
                    solo-inverted
                    flat
                    hide-details
                    v-model="props.item.type"
                    :items="extraVarTypes"
                    class="v-text-field--solo--no-min-height"
                    data-testid="varGroup-varType"
                  ></v-select>
                </td>
                <td class="pa-1">
                  <div class="d-flex align-center">
                    <v-text-field
                      solo-inverted
                      flat
                      hide-details
                      v-model="props.item.value"
                      class="v-text-field--solo--no-min-height"
                      :placeholder="extraVarValuePlaceholder(props.item.type)"
                    ></v-text-field>
                    <RichEditor
                      v-if="props.item.type === 'list' || props.item.type === 'dict'"
                      v-model="props.item.value"
                      :type="props.item.type === 'list' ? 'json_array' : 'json'"
                      class="ml-1 EnvVarExpandBtn"
                    />
                  </div>
                </td>
                <td style="width: 38px">
                  <v-icon small class="pa-1" @click="removeExtraVar(props.item)">
                    mdi-delete
                  </v-icon>
                </td>
              </tr>
            </template>
          </v-data-table>

          <v-alert color="warning" v-else>
            Oops! This JSON structure is a little too complex to display as a table.
          </v-alert>
        </div>

        <div>
          <v-subheader class="px-0 mt-4">
            {{ $t('environmentVariables') }}

            <v-spacer />

            <v-btn icon @click="addEnvVar()" data-testid="varGroup-addEnv">
              <v-icon> mdi-plus </v-icon>
            </v-btn>
          </v-subheader>
          <v-data-table
            :items="env"
            :items-per-page="-1"
            class="elevation-1 FieldTable"
            hide-default-footer
            :no-data-text="$t('noValues')"
            style="background: #8585850f"
          >
            <template v-slot:item="props">
              <tr>
                <td class="pa-1">
                  <v-text-field
                    solo-inverted
                    flat
                    hide-details
                    v-model="props.item.name"
                    class="v-text-field--solo--no-min-height"
                    :placeholder="$t('name')"
                  ></v-text-field>
                </td>
                <td class="pa-1">
                  <v-text-field
                    solo-inverted
                    flat
                    hide-details
                    v-model="props.item.value"
                    class="v-text-field--solo--no-min-height"
                    :placeholder="$t('Value')"
                  ></v-text-field>
                </td>
                <td style="width: 38px">
                  <v-icon small class="pa-1" @click="removeEnvVar(props.item)"> mdi-delete </v-icon>
                </td>
              </tr>
            </template>
          </v-data-table>
        </div>
      </v-tab-item>

      <v-tab-item key="secrets">
        <div
          v-if="!isNew && secretStorage"
          class="px-4 py-3"
          style="
            background: rgba(133, 133, 133, 0.06);
            border-color: rgb(33, 33, 33);
            border-radius: 6px;
          "
        >
          <div style="font-weight: bold; font-size: 20px">
            <v-icon small class="mr-1">{{ getIcon(secretStorage.type) }}</v-icon>
            {{ secretStorage.name }}
          </div>
          <pre>Source path pattern: <b>{{ item.secret_storage_key_prefix }}*</b></pre>

          <div class="d-flex items-center justify-space-between mt-2">
            <v-checkbox
              class="mt-0 mb-2"
              v-model="item.sync_enabled"
              :label="$t('Sync keys enabled')"
              :disabled="formSaving"
              hide-details
            />

            <div class="d-flex align-center">
              <v-btn
                style="margin-right: -10px"
                text
                color="primary"
                @click="syncSettingsDialog = true"
                :disabled="formSaving"
                v-if="item.sync_enabled"
              >
                <v-icon left>mdi-cog-sync</v-icon>
                Sync paths
                <v-chip
                  class="ml-2"
                  outlined
                  style="transform: translateY(-1px)"
                  color="primary"
                  small
                >
                  {{ (item.sync_paths || []).length }}
                </v-chip>
              </v-btn>
            </div>
          </div>
        </div>

        <v-dialog v-model="syncSettingsDialog" max-width="500" persistent>
          <v-card>
            <v-card-title>Sync paths</v-card-title>
            <v-card-text class="pt-4 pb-0">
              <v-text-field
                style="width: 140px"
                v-model.number="item.sync_interval"
                min="0"
                :label="$t('Auto-sync interval')"
                persistent-hint
                :disabled="formSaving"
                suffix="minutes"
                outlined
                dense
              ></v-text-field>

              <SecretStorageSyncOptionsForm v-model="item.sync_paths" />
            </v-card-text>
            <v-card-actions>
              <v-spacer />
              <v-btn text color="blue darken-1" @click="syncSettingsDialog = false">
                {{ $t('close') }}
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>

        <div>
          <v-subheader class="px-0">
            {{ $t('extraVariables') }}
            <v-tooltip v-if="needHelp" bottom color="black" open-delay="300" max-width="400">
              <template v-slot:activator="{ on, attrs }">
                <v-icon class="ml-1" v-bind="attrs" v-on="on">mdi-help-box </v-icon>
              </template>
              <div>
                <div><code>--extra-vars</code> for Ansible</div>
                <div><code>-var</code> for Terraform/OpenTofu</div>
              </div>
            </v-tooltip>

            <v-spacer />
            <v-btn icon @click="addSecret('var')" data-testid="varGroup-addSecretVar">
              <v-icon> mdi-plus </v-icon>
            </v-btn>
          </v-subheader>

          <v-alert
            color="warning"
            text
            v-if="secrets.filter((s) => !s.remove && s.type === 'var').length > 0"
          >
            Secrets passed this way may appear in plain text in Ansible logs.
          </v-alert>

          <v-data-table
            :items="secrets.filter((s) => !s.remove && s.type === 'var')"
            :items-per-page="-1"
            class="elevation-1 FieldTable"
            hide-default-footer
            :no-data-text="$t('noValues')"
            style="background: #8585850f"
          >
            <template v-slot:item="props">
              <tr>
                <td class="pa-1">
                  <v-text-field
                    solo-inverted
                    flat
                    hide-details
                    v-model="props.item.name"
                    class="v-text-field--solo--no-min-height"
                    :placeholder="$t('name')"
                  ></v-text-field>
                </td>

                <td class="pa-1">
                  <v-text-field
                    solo-inverted
                    flat
                    hide-details
                    v-model="props.item.value"
                    placeholder="*******"
                    class="v-text-field--solo--no-min-height"
                  ></v-text-field>
                </td>

                <td style="width: 38px">
                  <v-icon small class="pa-1" @click="removeSecret(props.item)"> mdi-delete </v-icon>
                </td>
              </tr>
            </template>
          </v-data-table>
        </div>

        <div>
          <v-subheader class="px-0 mt-4">
            {{ $t('environmentVariables') }}

            <v-spacer />

            <v-btn icon @click="addSecret('env')" data-testid="varGroup-addSecretEnv">
              <v-icon> mdi-plus </v-icon>
            </v-btn>
          </v-subheader>

          <v-data-table
            :items="secrets.filter((s) => !s.remove && s.type === 'env')"
            :items-per-page="-1"
            class="elevation-1 FieldTable"
            hide-default-footer
            :no-data-text="$t('noValues')"
            style="background: #8585850f"
          >
            <template v-slot:item="props">
              <tr>
                <td class="pa-1">
                  <v-text-field
                    solo-inverted
                    flat
                    hide-details
                    v-model="props.item.name"
                    class="v-text-field--solo--no-min-height"
                    :placeholder="$t('name')"
                  ></v-text-field>
                </td>

                <td class="pa-1">
                  <v-text-field
                    solo-inverted
                    flat
                    hide-details
                    v-model="props.item.value"
                    placeholder="*******"
                    class="v-text-field--solo--no-min-height"
                  ></v-text-field>
                </td>

                <td style="width: 38px">
                  <v-icon small class="pa-1" @click="removeSecret(props.item)"> mdi-delete </v-icon>
                </td>
              </tr>
            </template>
          </v-data-table>
        </div>
      </v-tab-item>
    </v-tabs-items>
  </v-form>
</template>
<style lang="scss">
.EnvironmentEditor {
  .CodeMirror {
    height: 160px !important;
  }
}

// Compact the RichEditor "expand" fab so it fits inside a variables table row.
.EnvVarExpandBtn {
  .v-btn--fab.v-size--small {
    height: 30px;
    width: 30px;
  }

  .v-btn__content .v-icon {
    font-size: 18px;
  }
}
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';

import { codemirror } from 'vue-codemirror';
import { load as loadYaml, dump as dumpYaml } from 'js-yaml';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
import 'codemirror/mode/yaml/yaml.js';
import 'codemirror/addon/display/placeholder.js';
import { getErrorMessage } from '@/lib/error';
import RichEditor from '@/components/RichEditor.vue';
import SecretStorageSyncOptionsForm from '@/components/SecretStorageSyncOptionsForm.vue';

export default {
  mixins: [ItemFormBase],

  props: {
    needHelp: Boolean,
    supportStorages: Boolean,
  },

  components: {
    RichEditor,
    codemirror,
    SecretStorageSyncOptionsForm,
  },

  computed: {
    secretStorage() {
      if (this.item && this.item.secret_storage_id && this.secretStorages) {
        return this.secretStorages.find((s) => s.id === this.item.secret_storage_id);
      }
      return null;
    },
  },

  watch: {
    // Handles Table/JSON/YAML toggling. The mode being left determines which
    // field is authoritative (extraVars for table, json for JSON, yaml for YAML);
    // it's parsed into a plain object which is then rendered into the mode being
    // entered.
    extraVarsEditMode(val, oldVal) {
      let source;
      switch (oldVal) {
        case 'json': {
          try {
            source = JSON.parse(this.json);
            this.formError = null;
          } catch (err) {
            this.formError = getErrorMessage(err);
            if (val === 'table') {
              this.extraVars = null;
            }
            return;
          }
          break;
        }
        case 'yaml': {
          try {
            source = loadYaml(this.yaml) || {};
            this.formError = null;
          } catch (err) {
            this.formError = getErrorMessage(err);
            if (val === 'table') {
              this.extraVars = null;
            }
            return;
          }
          break;
        }
        default: {
          // Coming from the table (or initial load): extraVars is authoritative.
          // Serialize leniently: a row whose list/dict value is not valid JSON yet
          // keeps its raw text (as a string) instead of throwing. This prevents the
          // toggle from blanking the target editor or dropping rows while the user
          // is still typing. Strict validation happens on save (see beforeSave).
          if (this.extraVars == null) {
            return;
          }
          source = this.extraVarsToObjectLenient(this.extraVars);
        }
      }

      switch (val) {
        case 'json':
          this.json = JSON.stringify(source, null, 2);
          break;
        case 'yaml': {
          const dumped = dumpYaml(source);
          this.yaml = dumped === '{}\n' ? '' : dumped;
          break;
        }
        case 'table': {
          // If the source still matches what the current table represents, the
          // user only switched tabs without editing it — keep the existing rows so
          // their chosen types (e.g. Dict) and in-progress values are preserved
          // instead of being re-inferred (and possibly downgraded to String).
          if (
            this.extraVars != null
            && JSON.stringify(source)
              === JSON.stringify(this.extraVarsToObjectLenient(this.extraVars))
          ) {
            return;
          }

          this.extraVars = this.objectToExtraVars(source);
          break;
        }
        default:
          throw new Error(`Invalid extra variables edit mode: ${val}`);
      }
    },
  },

  data() {
    return {
      // PREDEFINED_ENV_VARS,
      images: [
        'dind-runner:v2.0.0',
        'dind-runner:v2.0.2',
        'dind-runner:v2.0.3',
        'dind-runner:v2.0.4',
        'dind-runner:v2.0.5',
        'dind-runner:v2.0.6',
        'dind-runner:v2.0.7',
        'dind-runner:v2.0.8',
        'dind-runner:v2.0.9',
        'dind-runner:v2.0.10',
        'nodejs-runner:v2.0.0',
        'nodejs-runner:v2.0.3',
        'nodejs-runner:v2.0.4',
        'nodejs-runner:v2.0.5',
        'nodejs-runner:v2.0.6',
        'nodejs-runner:v2.0.7',
        'nodejs-runner:v2.0.8',
        'nodejs-runner:v2.0.9',
        'nodejs-runner:v2.0.10',
      ],

      json: '{}',
      yaml: '',
      extraVars: [],
      env: [],
      secrets: [],

      tab: 'variables',

      cmOptions: {
        tabSize: 2,
        mode: 'application/json',
        lineNumbers: true,
        line: true,
        lint: true,
        indentWithTabs: false,
      },

      cmYamlOptions: {
        tabSize: 2,
        mode: 'text/x-yaml',
        lineNumbers: true,
        line: true,
        indentWithTabs: false,
      },

      extraVarsEditMode: 'json',

      extraVarTypes: [
        { text: 'String', value: 'string' },
        { text: 'Number', value: 'number' },
        { text: 'List', value: 'list' },
        { text: 'Dict', value: 'dict' },
      ],

      secretStorages: null,

      syncSettingsDialog: false,
      syncing: false,
    };
  },

  methods: {
    getNewItem() {
      return {
        sync_enabled: false,
        sync_interval: 0,
        sync_paths: [],
      };
    },
    getIcon(type) {
      switch (type) {
        case 'aws_sm':
          return '$vuetify.icons.aws_sm';
        case 'vault':
          return '$vuetify.icons.hashicorp_vault';
        case 'openbao':
          return '$vuetify.icons.openbao';
        case 'dvls':
          return '$vuetify.icons.dvls';
        case 'azure_kv':
          return '$vuetify.icons.azure_kv';
        default:
          return '';
      }
    },

    addExtraVar(name = '', value = '', type = 'string') {
      this.extraVars.push({ name, value, type });
    },

    removeExtraVar(val) {
      const i = this.extraVars.indexOf(val);
      if (i > -1) {
        this.extraVars.splice(i, 1);
      }
    },

    extraVarValuePlaceholder(type) {
      switch (type) {
        case 'number':
          return '42';
        case 'list':
          return '["a", "b"]';
        case 'dict':
          return '{"key": "value"}';
        default:
          return this.$t('Value');
      }
    },

    // inferVarType maps a parsed JSON value to one of the editor's variable types
    // so that a value stored as a list/dict is shown with the right type in the
    // table editor. Numbers get the number type; other scalars are edited as strings.
    inferVarType(value) {
      if (Array.isArray(value)) {
        return 'list';
      }
      if (value !== null && typeof value === 'object') {
        return 'dict';
      }
      if (typeof value === 'number') {
        return 'number';
      }
      return 'string';
    },

    // rowToVarValue converts a table row back to its typed JSON value. It throws a
    // descriptive error (caught by save()/beforeSave) when the input is invalid.
    rowToVarValue(row) {
      switch (row.type) {
        case 'number': {
          const parsed = Number(row.value);
          if (row.value === '' || Number.isNaN(parsed)) {
            throw new Error(`Variable "${row.name}" must be a number, e.g. 42`);
          }
          return parsed;
        }
        case 'list': {
          let parsed;
          try {
            parsed = JSON.parse(row.value);
          } catch (e) {
            throw new Error(`Variable "${row.name}" is not a valid list, e.g. ["a", "b"]`);
          }
          if (!Array.isArray(parsed)) {
            throw new Error(`Variable "${row.name}" must be a list, e.g. ["a", "b"]`);
          }
          return parsed;
        }
        case 'dict': {
          let parsed;
          try {
            parsed = JSON.parse(row.value);
          } catch (e) {
            throw new Error(`Variable "${row.name}" is not a valid dict, e.g. {"key": "value"}`);
          }
          if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
            throw new Error(`Variable "${row.name}" must be a dict, e.g. {"key": "value"}`);
          }
          return parsed;
        }
        default:
          // Scalars are passed through as-is: an untouched number/boolean keeps its
          // original JSON type, while typed input stays a string.
          return row.value;
      }
    },

    extraVarsToObject(rows) {
      return (rows || []).reduce(
        (prev, curr) => ({
          ...prev,
          [curr.name]: this.rowToVarValue(curr),
        }),
        {},
      );
    },

    // extraVarsToObjectLenient is like extraVarsToObject but never throws: a row
    // whose typed value is still invalid keeps its raw text. Used when toggling to
    // the JSON view so editing modes never lose data mid-typing.
    extraVarsToObjectLenient(rows) {
      return (rows || []).reduce((prev, curr) => {
        let value;
        try {
          value = this.rowToVarValue(curr);
        } catch (e) {
          value = curr.value;
        }
        return { ...prev, [curr.name]: value };
      }, {});
    },

    objectToExtraVars(obj) {
      return Object.keys(obj).map((name) => {
        const value = obj[name];
        const type = this.inferVarType(value);
        return {
          name,
          type,
          value: type === 'string' ? value : JSON.stringify(value),
        };
      });
    },

    addEnvVar(name = '', value = '') {
      this.env.push({ name, value });
    },

    removeEnvVar(val) {
      const i = this.env.findIndex((v) => v.name === val.name);
      if (i > -1) {
        this.env.splice(i, 1);
      }
    },

    addSecret(type) {
      this.secrets.push({
        type,
        name: '',
        value: '',
        new: true,
      });
    },

    removeSecret(val) {
      const i = this.secrets.findIndex((v) => v.name === val.name);
      if (i > -1) {
        const s = this.secrets[i];
        this.secrets.splice(i, 1);

        if (!s.new) {
          this.secrets.push({
            ...s,
            remove: true,
          });
        }
      }
    },

    beforeSave() {
      switch (this.extraVarsEditMode) {
        case 'json':
          this.item.json = this.json;
          break;
        case 'yaml':
          try {
            this.item.json = JSON.stringify(loadYaml(this.yaml) || {});
          } catch (err) {
            throw new Error(`Extra variables: ${getErrorMessage(err)}`);
          }
          break;
        case 'table':
          if (this.extraVars == null) {
            this.item.json = this.json;
          } else {
            this.item.json = JSON.stringify(this.extraVarsToObject(this.extraVars));
          }
          break;
        default:
          throw new Error(`Invalid extra variables edit mode: ${this.extraVarsEditMode}`);
      }

      const env = (this.env || []).reduce(
        (prev, curr) => ({
          ...prev,
          [curr.name]: curr.value,
        }),
        {},
      );

      const secrets = (this.secrets || [])
        .map((s) => {
          let operation;
          if (s.new) {
            operation = 'create';
          } else if (s.remove) {
            operation = 'delete';
          } else {
            operation = 'update';
          }
          return {
            id: s.id,
            name: s.name,
            secret: s.value,
            type: s.type,
            operation,
          };
        })
        .filter((s) => s.operation != null);

      this.item.env = JSON.stringify(env);
      this.item.secrets = secrets;
    },

    async afterLoadData() {
      if (this.itemId === 'new') {
        [this.secretStorages] = await Promise.all([this.loadProjectResources('secret_storages')]);
      } else {
        this.secretStorages = [];

        if (this.item.secret_storage_id) {
          this.secretStorages.push(
            await this.loadProjectResource('secret_storages', this.item.secret_storage_id),
          );
        }
      }

      if (!this.item.sync_paths) {
        this.$set(this.item, 'sync_paths', []);
      }
      if (this.item.sync_enabled == null) {
        this.$set(this.item, 'sync_enabled', false);
      }
      if (this.item.sync_interval == null) {
        this.$set(this.item, 'sync_interval', 0);
      }

      this.json = JSON.stringify(JSON.parse(this.item?.json || '{}'), null, 2);

      const json = JSON.parse(this.item?.json || '{}');

      const env = JSON.parse(this.item?.env || '{}');

      const secrets = this.item?.secrets || [];

      this.extraVars = this.objectToExtraVars(json);
      this.extraVarsEditMode = 'table';

      this.env = Object.keys(env)
        // .filter((x) => {
        //   const index = PREDEFINED_ENV_VARS.findIndex((v) => v.name === x);
        //   return index === -1 || PREDEFINED_ENV_VARS[index].value !== env[x];
        // })
        .map((x) => ({
          name: x,
          value: env[x],
        }));

      this.secrets = secrets.map((x) => ({
        id: x.id,
        name: x.name,
        value: '',
        type: x.type,
      }));

      // Object.keys(env).forEach((x) => {
      //   const index = PREDEFINED_ENV_VARS.findIndex((v) => v.name === x);
      //   if (index !== -1 && PREDEFINED_ENV_VARS[index].value === env[x]) {
      //     this.predefinedEnvVars.push(index);
      //   }
      // });
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
