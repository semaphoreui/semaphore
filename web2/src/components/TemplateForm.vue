<template>
  <v-form
      ref="form"
      lazy-validation
      v-model="formValid"
      v-if="isLoaded"
  >
    <v-alert
        :value="formError"
        color="error"
        class="pb-2"
    >{{ formError }}</v-alert>

    <v-row>
      <v-col cols="12" md="6" class="pb-0">
        <v-text-field
            v-model="item.alias"
            label="Playbook Alias"
            :rules="[v => !!v || 'Playbook Alias is required']"
            required
            :disabled="formSaving"
        ></v-text-field>

        <v-text-field
            v-model="item.playbook"
            label="Playbook Filename"
            :rules="[v => !!v || 'Playbook Filename is required']"
            required
            :disabled="formSaving"
            placeholder="Example: site.yml"
        ></v-text-field>

        <v-select
            v-model="item.inventory_id"
            label="Inventory"
            :items="inventory"
            item-value="id"
            item-text="name"
            :rules="[v => !!v || 'Inventory is required']"
            required
            :disabled="formSaving"
        ></v-select>

        <v-select
            v-model="item.repository_id"
            label="Playbook Repository"
            :items="repositories"
            item-value="id"
            item-text="name"
            :rules="[v => !!v || 'Playbook Repository is required']"
            required
            :disabled="formSaving"
        ></v-select>

        <v-select
            v-model="item.environment_id"
            label="Environment"
            :items="environment"
            item-value="id"
            item-text="name"
            :rules="[v => !!v || 'Environment is required']"
            required
            :disabled="formSaving"
        ></v-select>

        <v-select
            v-model="item.vault_pass_id"
            label="Vault Password"
            clearable
            :items="loginPasswordKeys"
            item-value="id"
            item-text="name"
            :disabled="formSaving"
        ></v-select>

        <v-text-field
            v-model="cronFormat"
            label="Cron"
            :disabled="formSaving"
            placeholder="Example: * 1 * * * *"
            v-if="schedules.length <= 1"
        ></v-text-field>
      </v-col>

      <v-col cols="12" md="6" class="pb-0">
        <v-textarea
            outlined
            v-model="item.description"
            label="Description"
            :disabled="formSaving"
            rows="5"
        ></v-textarea>

        <codemirror
            :style="{ border: '1px solid lightgray' }"
            v-model="item.arguments"
            :options="cmOptions"
            :disabled="formSaving"
            placeholder='Enter extra CLI Arguments...
Example:
[
  "-i",
  "@myinventory.sh",
  "--private-key=/there/id_rsa",
  "-vvvv"
]'
        />
      </v-col>
    </v-row>
  </v-form>
</template>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

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

  props: {
    sourceItemId: String,
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
      item: null,
      keys: null,
      inventory: null,
      repositories: null,
      environment: null,
      schedules: null,
      cronFormat: null,
    };
  },

  watch: {
    needReset(val) {
      if (val) {
        if (this.item != null) {
          this.item.template_id = this.templateId;
        }
      }
    },

    sourceItemId(val) {
      this.item.template_id = val;
    },
  },

  computed: {
    isLoaded() {
      if (this.isNew && this.sourceItemId == null) {
        return true;
      }

      return this.keys != null
          && this.repositories != null
          && this.inventory != null
          && this.environment != null
          && this.item != null
          && this.schedules != null;
    },

    loginPasswordKeys() {
      if (this.keys == null) {
        return null;
      }
      return this.keys.filter((key) => key.type === 'login_password');
    },
  },

  methods: {
    async afterLoadData() {
      if (this.sourceItemId) {
        this.item = (await axios({
          keys: 'get',
          url: `/api/project/${this.projectId}/templates/${this.sourceItemId}`,
          responseType: 'json',
        })).data;
      }
      this.keys = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/keys`,
        responseType: 'json',
      })).data;
      this.repositories = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/repositories`,
        responseType: 'json',
      })).data;
      this.inventory = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/inventory`,
        responseType: 'json',
      })).data;
      this.environment = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/environment`,
        responseType: 'json',
      })).data;
      this.schedules = this.isNew ? [] : (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/templates/${this.itemId}/schedules`,
        responseType: 'json',
      })).data;
      if (this.schedules.length === 1) {
        this.cronFormat = this.schedules[0].cron_format;
      }
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/templates`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/templates/${this.itemId}`;
    },

    async afterSave(newItem) {
      if (newItem || this.schedules.length === 0) {
        if (this.cronFormat != null && this.cronFormat !== '') {
          // new schedule
          await axios({
            method: 'post',
            url: `/api/project/${this.projectId}/schedules`,
            responseType: 'json',
            data: {
              project_id: this.projectId,
              template_id: newItem ? newItem.id : this.itemId,
              cron_format: this.cronFormat,
            },
          });
        }
      } else if (this.schedules.length > 1) {
        // do nothing
      } else if (this.cronFormat == null || this.cronFormat === '') {
        // drop schedule
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/schedules/${this.schedules[0].id}`,
          responseType: 'json',
        });
      } else {
        // update schedule
        await axios({
          method: 'put',
          url: `/api/project/${this.projectId}/schedules/${this.schedules[0].id}`,
          responseType: 'json',
          data: {
            id: this.schedules[0].id,
            project_id: this.projectId,
            template_id: this.itemId,
            cron_format: this.cronFormat,
          },
        });
      }
    },
  },
};
</script>
