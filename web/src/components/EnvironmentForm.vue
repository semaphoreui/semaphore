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
      {{ $t('extraVariables') }}

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
        <v-btn value="json" small class="mr-0" style="border-radius: 4px;" disabled>
          JSON
        </v-btn>
      </v-btn-toggle>
    </v-subheader>

    <codemirror
      :style="{ border: '1px solid lightgray' }"
      v-model="json"
      :options="cmOptions"
      :placeholder="$t('enterExtraVariablesJson')"
    />

    <div>
      <v-subheader class="px-0 mt-4">
        {{ $t('environmentVariables') }}

        <v-tooltip bottom color="black" open-delay="300">
          <template v-slot:activator="{ on, attrs }">
            <v-icon
              class="ml-1"
              v-bind="attrs"
              v-on="on"
              color="lightgray"
            >mdi-help-circle</v-icon>
          </template>
          <span>Variables passed as process environment variables.</span>
        </v-tooltip>
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
              ></v-text-field>
            </td>
            <td class="pa-1">
              <v-text-field
                solo-inverted
                flat
                hide-details
                v-model="props.item.value"
                class="v-text-field--solo--no-min-height"
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
      <div class="mt-2 mb-4 mx-1">
        <v-btn
          color="primary"
          @click="addEnvVar()"
        >New Environment Variable</v-btn>
      </div>
    </div>

    <div>
      <v-subheader class="px-0 mt-4">
        {{ $t('Secrets') }}

        <v-tooltip bottom color="black" open-delay="300" max-width="400">
          <template v-slot:activator="{ on, attrs }">
            <v-icon
              class="ml-1"
              v-bind="attrs"
              v-on="on"
              color="lightgray"
            >mdi-help-circle</v-icon>
          </template>
          <span>
            Secrets are stored in the database in encrypted form.
            Secrets passed via <code>--extra-vars</code> (Ansible) or
            <code>-var</code> (Terraform/OpenTofu).
          </span>
        </v-tooltip>
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
              <v-text-field
                solo-inverted
                flat
                hide-details
                v-model="props.item.name"
                class="v-text-field--solo--no-min-height"
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

      <div class="mt-2 mb-4 mx-1">
        <v-btn
          color="primary"
          @click="addSecret()"
        >New Secret</v-btn>
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
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

const PREDEFINED_ENV_VARS = [{
  name: 'ANSIBLE_HOST_KEY_CHECKING',
  value: 'False',
  description: 'Avoid host key checking by the tools Ansible uses to connect to the host.',
}];

export default {
  mixins: [ItemFormBase],
  components: {
    codemirror,
  },

  created() {
  },

  data() {
    return {
      PREDEFINED_ENV_VARS,
      images: [
        'dind-runner:latest',
      ],
      advancedOptions: false,

      json: '{}',
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
      predefinedEnvVars: [],
    };
  },

  methods: {
    addEnvVar(name = '', value = '') {
      this.env.push({ name, value });
    },

    removeEnvVar(val) {
      const i = this.env.findIndex((v) => v.name === val.name);
      if (i > -1) {
        this.env.splice(i, 1);
      }
    },

    addSecret(name = '', value = '') {
      this.secrets.push({ name, value, new: true });
    },

    removeSecret(val) {
      const i = this.secrets.findIndex((v) => v.name === val.name);
      if (i > -1) {
        const s = this.secrets[i];
        this.secrets.splice(i, 1);

        if (!this.secrets[i].new) {
          this.secrets.push({
            ...s,
            remove: true,
          });
        }
      }
    },

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

      const env = (this.env || []).reduce((prev, curr) => ({
        ...prev,
        [curr.name]: curr.value,
      }), {});

      this.predefinedEnvVars.forEach((index) => {
        const predefinedVar = PREDEFINED_ENV_VARS[index];
        env[predefinedVar.name] = predefinedVar.value;
      });

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
          operation,
        };
      }).filter((s) => s.operation != null);

      this.item.env = JSON.stringify(env);
      this.item.secrets = secrets;
    },

    afterLoadData() {
      this.json = this.item?.json || '{}';

      const env = JSON.parse(this.item?.env || '{}');

      const secrets = this.item?.secrets || [];

      this.env = Object.keys(env)
        .filter((x) => {
          const index = PREDEFINED_ENV_VARS.findIndex((v) => v.name === x);
          return index === -1 || PREDEFINED_ENV_VARS[index].value !== env[x];
        })
        .map((x) => ({
          name: x,
          value: env[x],
        }));

      this.secrets = secrets.map((x) => ({
        id: x.id,
        name: x.name,
        value: '',
      }));

      Object.keys(env).forEach((x) => {
        const index = PREDEFINED_ENV_VARS.findIndex((v) => v.name === x);
        if (index !== -1 && PREDEFINED_ENV_VARS[index].value === env[x]) {
          this.predefinedEnvVars.push(index);
        }
      });
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
