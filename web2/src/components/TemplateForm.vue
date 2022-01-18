<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="isLoaded"
  >
    <v-dialog
      v-model="helpDialog"
      hide-overlay
      width="300"
    >
      <v-alert
        border="top"
        colored-border
        type="info"
        elevation="2"
        class="mb-0 pb-0"
      >
        <div v-if="helpKey === 'build_version'">
          <p>
            Defines start version of your
            <a target="_black" href="https://en.wikipedia.org/wiki/Software_build">artifact</a>.
            Each run increments the artifact version.
          </p>
          <p>
            For more information about building, see the
            <a href="https://docs.ansible-semaphore.com/user-guide/task-templates#build"
               target="_blank"
            >Task Template reference</a>.
          </p>
        </div>
        <div v-else-if="helpKey === 'build'">
          <p>
            Defines what
            <a target="_black" href="https://en.wikipedia.org/wiki/Software_build">artifact</a>
            should be deployed when the task run.
          </p>
          <p>
            For more information about deploying, see the
            <a href="https://docs.ansible-semaphore.com/user-guide/task-templates#build"
               target="_blank"
            >Task Template reference</a>.
          </p>
        </div>
        <div v-if="helpKey === 'cron'">
          <p>Defines autorun schedule.</p>
          <p>
            For more information about cron, see the
            <a href="https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format"
               target="_blank"
            >Cron expression format reference</a>.
          </p>
        </div>
      </v-alert>
    </v-dialog>

    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-row>
      <v-col cols="12" md="6" class="pb-0">
        <v-card class="mb-6">
          <v-tabs
            fixed-tabs
            v-model="itemTypeIndex"
          >
            <v-tab
              style="padding: 0"
              v-for="(key) in Object.keys(TEMPLATE_TYPE_ICONS)"
              :key="key"
            >
              <v-icon small class="mr-2">{{ TEMPLATE_TYPE_ICONS[key] }}</v-icon>
              {{ TEMPLATE_TYPE_TITLES[key] }}
            </v-tab>
          </v-tabs>

          <div class="ml-4 mr-4 mt-6" v-if="item.type">
            <v-text-field
              v-if="item.type === 'build'"
              v-model="item.start_version"
              label="Start Version"
              :rules="[v => !!v || 'Start Version is required']"
              required
              :disabled="formSaving"
              placeholder="Example: 0.0.0"
              append-outer-icon="mdi-help-circle"
              @click:append-outer="showHelpDialog('build_version')"
            ></v-text-field>

            <v-select
              v-if="item.type === 'deploy'"
              v-model="item.build_template_id"
              label="Build Template"
              :items="buildTemplates"
              item-value="id"
              item-text="alias"
              :rules="[v => !!v || 'Build Template is required']"
              required
              :disabled="formSaving"
              append-outer-icon="mdi-help-circle"
              @click:append-outer="showHelpDialog('build')"
            ></v-select>

            <v-checkbox
              v-if="item.type === 'deploy'"
              class="mt-0"
              label="Autorun"
              v-model="item.autorun"
            />
          </div>

        </v-card>

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
          v-model="item.vault_key_id"
          label="Vault Password"
          clearable
          :items="loginPasswordKeys"
          item-value="id"
          item-text="name"
          :disabled="formSaving"
        ></v-select>
      </v-col>

      <v-col cols="12" md="6" class="pb-0">
        <v-textarea
          outlined
          v-model="item.description"
          label="Description"
          :disabled="formSaving"
          rows="5"
        ></v-textarea>

        <SurveyVars :json="item.survey_vars" @change="setSurveyVars"/>

        <!--
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
        -->
        <v-select
          v-model="item.view_id"
          label="View"
          clearable
          :items="views"
          item-value="id"
          item-text="title"
          :disabled="formSaving"
        ></v-select>

        <v-text-field
          class="mt-6"
          v-model="cronFormat"
          label="Cron"
          :disabled="formSaving"
          placeholder="Example: * 1 * * * *"
          v-if="schedules == null || schedules.length <= 1"
          append-outer-icon="mdi-help-circle"
          @click:append-outer="showHelpDialog('cron')"
        ></v-text-field>

        <v-select
          v-model="cronRepositoryId"
          label="Cron Condition Repository"
          placeholder="Cron checks new commit before run"
          :items="repositories"
          item-value="id"
          item-text="name"
          :disabled="formSaving"
        ></v-select>

      </v-col>
    </v-row>
  </v-form>
</template>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

// import { codemirror } from 'vue-codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
// import 'codemirror/addon/lint/json-lint.js';
import 'codemirror/addon/display/placeholder.js';
import { TEMPLATE_TYPE_ICONS, TEMPLATE_TYPE_TITLES } from '../lib/constants';
import SurveyVars from './SurveyVars';

export default {
  mixins: [ItemFormBase],

  components: {
    SurveyVars,
    // codemirror,
  },

  props: {
    sourceItemId: String,
  },

  data() {
    return {
      itemTypeIndex: 0,
      TEMPLATE_TYPE_ICONS,
      TEMPLATE_TYPE_TITLES,
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
      views: null,
      schedules: null,
      cronFormat: null,
      cronRepositoryId: null,

      helpDialog: null,
      helpKey: null,
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

    itemTypeIndex(val) {
      this.item.type = Object.keys(TEMPLATE_TYPE_ICONS)[val];
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
        && this.schedules != null
        && this.views != null;
    },

    loginPasswordKeys() {
      if (this.keys == null) {
        return null;
      }
      return this.keys.filter((key) => key.type === 'login_password');
    },
  },

  methods: {
    setSurveyVars(v) {
      this.item.survey_vars = v;
    },

    showHelpDialog(key) {
      this.helpKey = key;
      this.helpDialog = true;
    },

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

      this.buildTemplates = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/templates?type=build`,
        responseType: 'json',
      })).data.filter((template) => template.type === 'build');

      this.schedules = this.isNew ? [] : (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/templates/${this.itemId}/schedules`,
        responseType: 'json',
      })).data;

      this.views = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/views`,
        responseType: 'json',
      })).data;

      if (this.schedules.length === 1) {
        this.cronFormat = this.schedules[0].cron_format;
      }

      this.itemTypeIndex = Object.keys(TEMPLATE_TYPE_ICONS).indexOf(this.item.type);
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/templates`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/templates/${this.itemId}`;
    },

    async beforeSave() {
      if (this.cronFormat == null || this.cronFormat === '') {
        return;
      }

      await axios({
        method: 'post',
        url: `/api/project/${this.projectId}/schedules/validate`,
        responseType: 'json',
        data: {
          cron_format: this.cronFormat,
        },
      });
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
