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

    <v-subheader class="px-0">
      <v-icon class="mr-1">mdi-variable</v-icon> {{ $t('extraVariables') }}

      <v-tooltip bottom color="black" open-delay="300" max-width="400">
        <template v-slot:activator="{ on, attrs }">
          <v-icon
            class="ml-1"
            v-bind="attrs"
            v-on="on"
          >mdi-help-circle</v-icon>
        </template>
        <span>
          Variables passed via <code>--extra-vars</code> (Ansible) or
          <code>-var</code> (Terraform/OpenTofu).
        </span>
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
        no-data-text="No values"
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
      <div class="mt-2 mb-4" v-if="extraVars != null">
        <v-btn
          color="primary"
          @click="addExtraVar()"
        >{{ $t('New Extra Variable') }}</v-btn>
      </div>
      <v-alert color="error" v-else>Can't be displayed as table.</v-alert>
    </div>

    <div>
      <v-subheader class="px-0 mt-4">
        <v-icon class="mr-1">mdi-application-settings</v-icon>
        {{ $t('environmentVariables') }}
      </v-subheader>
      <v-data-table
        :items="env"
        :items-per-page="-1"
        class="elevation-1"
        hide-default-footer
        no-data-text="No values"
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
      <div class="mt-2 mb-4">
        <v-btn
          color="primary"
          @click="addEnvVar()"
        >{{ $t('New Environment Variable') }}</v-btn>
      </div>
    </div>

    <div>
      <v-subheader class="px-0 mt-4">
        <v-icon class="mr-1">mdi-lock</v-icon>{{ $t('Secrets') }}
      </v-subheader>

      <v-data-table
        :items="secrets.filter(s => !s.remove)"
        :items-per-page="-1"
        class="elevation-1"
        hide-default-footer
        no-data-text="No values"
      >
        <template v-slot:item="props">
          <tr>
            <td class="pa-1">
              <v-icon>
                {{ props.item.type === 'var' ? 'mdi-variable' : 'mdi-application-settings' }}
              </v-icon>
            </td>
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

      <div class="mt-2 mb-4">
        <v-menu
          offset-y
        >
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              v-bind="attrs"
              v-on="on"
              color="primary"
            >New Secret</v-btn>
          </template>
          <v-list>
            <v-list-item
              link
              @click="addSecret('var')"
            >
              <v-list-item-icon>
                <v-icon>mdi-variable</v-icon>
              </v-list-item-icon>
              <v-list-item-title>{{ $t('Secret Extra Variable') }}</v-list-item-title>
            </v-list-item>
            <v-list-item
              link
              @click="addSecret('env')"
            >
              <v-list-item-icon>
                <v-icon>mdi-application-settings</v-icon>
              </v-list-item-icon>
              <v-list-item-title>{{ $t('Secret Environment Variable') }}</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </div>

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
import { getErrorMessage } from '@/lib/error';
// import EventBus from '@/event-bus';
// import { getErrorMessage } from '@/lib/error';

// const PREDEFINED_ENV_VARS = [{
//   name: 'ANSIBLE_HOST_KEY_CHECKING',
//   value: 'False',
//   description: 'Avoid host key checking by the tools Ansible uses to connect to the host.',
// }];

export default {
  mixins: [ItemFormBase],
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

      cmOptions: {
        tabSize: 2,
        mode: 'application/json',
        lineNumbers: true,
        line: true,
        lint: true,
        indentWithTabs: false,
      },

      extraVarsEditMode: 'json',
      // predefinedEnvVars: [],
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

    // setExtraVar(name, value) {
    //   try {
    //     const obj = JSON.parse(this.json || '{}');
    //     if (value == null) {
    //       delete obj[name];
    //     } else {
    //       obj[name] = value;
    //     }
    //     this.json = JSON.stringify(obj, null, 2);
    //   } catch (err) {
    //     EventBus.$emit('i-snackbar', {
    //       color: 'error',
    //       text: getErrorMessage(err),
    //     });
    //   }
    // },

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

      // this.predefinedEnvVars.forEach((index) => {
      //   const predefinedVar = PREDEFINED_ENV_VARS[index];
      //   env[predefinedVar.name] = predefinedVar.value;
      // });

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
