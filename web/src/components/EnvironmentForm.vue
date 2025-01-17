<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
    class="pb-3"
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
      class="mb-2"
    ></v-text-field>

    <v-tabs grow v-model="tab" class="mb-7">
      <v-tab key="variables">Variables</v-tab>
      <v-tab key="secrets">Secrets</v-tab>
    </v-tabs>

    <v-tabs-items v-model="tab">
      <v-tab-item key="variables">

        <v-subheader class="px-0">
          {{ $t('extraVariables') }}

          <v-tooltip v-if="needHelp" bottom color="black" open-delay="300" max-width="400">
            <template v-slot:activator="{ on, attrs }">
              <v-icon
                class="ml-1"
                v-bind="attrs"
                v-on="on"
              >mdi-help-box</v-icon>
            </template>
            <div>
              <div><code>--extra-vars</code> for Ansible</div>
              <div><code>-var</code> for Terraform/OpenTofu</div>
            </div>
          </v-tooltip>

          <v-spacer />

          <v-btn-toggle
            v-model="extraVarsEditMode"
            tile
            group
          >
            <v-btn value="table" small class="mr-0" style="border-radius: 4px;">
              Table
            </v-btn>
            <v-btn value="json" small class="mr-0" style="border-radius: 4px;">
              JSON
            </v-btn>
          </v-btn-toggle>

          <v-btn icon @click="addExtraVar()">
            <v-icon>
              mdi-plus
            </v-icon>
          </v-btn>

        </v-subheader>

        <codemirror
          v-if="extraVarsEditMode === 'json'"
          :style="{ border: '1px solid lightgray' }"
          v-model="json"
          :options="cmOptions"
          :placeholder="$t('enterExtraVariablesJson')"
        />

        <div v-else-if="extraVarsEditMode === 'table'">
          <v-data-table
            v-if="extraVars != null"
            :items="extraVars"
            :items-per-page="-1"
            class="elevation-1"
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
                <td style="width: 38px;">
                  <v-icon
                    small
                    class="pa-1"
                    @click="removeExtraVar(props.item)"
                  >
                    mdi-delete
                  </v-icon>
                </td>
              </tr>
            </template>
          </v-data-table>

          <v-alert color="error" v-else>Can't be displayed as table.</v-alert>
        </div>

        <div>
          <v-subheader class="px-0 mt-4">
            {{ $t('environmentVariables') }}

            <v-spacer />

            <v-btn icon @click="addEnvVar()">
              <v-icon>
                mdi-plus
              </v-icon>
            </v-btn>
          </v-subheader>
          <v-data-table
            :items="env"
            :items-per-page="-1"
            class="elevation-1"
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
                <td style="width: 38px;">
                  <v-icon
                    small
                    class="pa-1"
                    @click="removeEnvVar(props.item)"
                  >
                    mdi-delete
                  </v-icon>
                </td>
              </tr>
            </template>
          </v-data-table>
        </div>
      </v-tab-item>

      <v-tab-item key="secrets">

        <div>
          <v-subheader class="px-0">
            {{ $t('extraVariables') }}

            <v-tooltip v-if="needHelp" bottom color="black" open-delay="300" max-width="400">
              <template v-slot:activator="{ on, attrs }">
                <v-icon
                  class="ml-1"
                  v-bind="attrs"
                  v-on="on"
                >mdi-help-box</v-icon>
              </template>
              <div>
                <div><code>--extra-vars</code> for Ansible</div>
                <div><code>-var</code> for Terraform/OpenTofu</div>
              </div>
            </v-tooltip>

            <v-spacer />
            <v-btn icon @click="addSecret('var')">
              <v-icon>
                mdi-plus
              </v-icon>
            </v-btn>
          </v-subheader>

          <v-data-table
            :items="secrets.filter(s => !s.remove && s.type === 'var')"
            :items-per-page="-1"
            class="elevation-1"
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

                <td style="width: 38px;">
                  <v-icon
                    small
                    class="pa-1"
                    @click="removeSecret(props.item)"
                  >
                    mdi-delete
                  </v-icon>
                </td>
              </tr>
            </template>
          </v-data-table>
        </div>

        <div>
          <v-subheader class="px-0 mt-4">
            {{ $t('environmentVariables') }}

            <v-spacer />

            <v-btn icon @click="addSecret('env')">
              <v-icon>
                mdi-plus
              </v-icon>
            </v-btn>
          </v-subheader>

          <v-data-table
            :items="secrets.filter(s => !s.remove && s.type === 'env')"
            :items-per-page="-1"
            class="elevation-1"
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

                <td style="width: 38px;">
                  <v-icon
                    small
                    class="pa-1"
                    @click="removeSecret(props.item)"
                  >
                    mdi-delete
                  </v-icon>
                </td>
              </tr>
            </template>
          </v-data-table>
        </div>

      </v-tab-item>
    </v-tabs-items>

  </v-form>
</template>

<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';

import { codemirror } from 'vue-codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
import 'codemirror/addon/display/placeholder.js';
import { getErrorMessage } from '@/lib/error';

export default {
  mixins: [ItemFormBase],

  props: {
    needHelp: Boolean,
  },

  components: {
    codemirror,
  },

  created() {
  },

  watch: {
    extraVarsEditMode(val) {
      let extraVars;

      switch (val) {
        case 'json':
          if (this.extraVars == null) {
            return;
          }

          this.json = JSON.stringify(this.extraVars.reduce((prev, curr) => ({
            ...prev,
            [curr.name]: curr.value,
          }), {}), null, 2);
          break;
        case 'table':
          try {
            extraVars = JSON.parse(this.json);
            this.formError = null;
          } catch (err) {
            this.formError = getErrorMessage(err);
            this.extraVars = null;
            return;
          }
          if (Object.keys(extraVars).some((x) => typeof extraVars[x] === 'object')) {
            this.extraVars = null;
          } else {
            this.extraVars = Object.keys(extraVars)
              .map((x) => ({
                name: x,
                value: extraVars[x],
              }));
          }
          break;
        default:
          throw new Error(`Invalid extra variables edit mode: ${val}`);
      }
    },
  },

  data() {
    return {
      // PREDEFINED_ENV_VARS,
      images: [
        'dind-runner:latest',
      ],
      advancedOptions: false,

      json: '{}',
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

      extraVarsEditMode: 'json',
    };
  },

  methods: {
    addExtraVar(name = '', value = '') {
      this.extraVars.push({ name, value });
    },

    removeExtraVar(val) {
      const i = this.extraVars.findIndex((v) => v.name === val.name);
      if (i > -1) {
        this.extraVars.splice(i, 1);
      }
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
        type, name: '', value: '', new: true,
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
        case 'table':
          if (this.extraVars == null) {
            this.item.json = this.json;
          } else {
            this.item.json = JSON.stringify(this.extraVars.reduce((prev, curr) => ({
              ...prev,
              [curr.name]: curr.value,
            }), {}));
          }
          break;
        default:
          throw new Error(`Invalid extra variables edit mode: ${this.extraVarsEditMode}`);
      }

      const env = (this.env || []).reduce((prev, curr) => ({
        ...prev,
        [curr.name]: curr.value,
      }), {});

      const secrets = (this.secrets || []).map((s) => {
        let operation;
        if (s.new) {
          operation = 'create';
        } else if (s.remove) {
          operation = 'delete';
        } else if (s.value !== '') {
          operation = 'update';
        }
        return {
          id: s.id,
          name: s.name,
          secret: s.value,
          type: s.type,
          operation,
        };
      }).filter((s) => s.operation != null);

      this.item.env = JSON.stringify(env);
      this.item.secrets = secrets;
    },

    afterLoadData() {
      this.json = JSON.stringify(JSON.parse(this.item?.json || '{}'), null, 2);

      const json = JSON.parse(this.item?.json || '{}');

      const env = JSON.parse(this.item?.env || '{}');

      const secrets = this.item?.secrets || [];

      if (Object.keys(json).some((x) => typeof json[x] === 'object')) {
        this.extraVars = null;
        this.extraVarsEditMode = 'json';
      } else {
        this.extraVars = Object.keys(json)
          .map((x) => ({
            name: x,
            value: json[x],
          }));
        this.extraVarsEditMode = 'table';
      }

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
