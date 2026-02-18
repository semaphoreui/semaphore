<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="isLoaded()"
    @submit.prevent="save()"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-alert
      color="blue"
      dark
      dismissible
      dense
      @input="item.commit_hash=null"
      v-model="hasCommit"
      class="overflow-hidden mt-2"
    >
      <div
        style="font-weight: bold;"
      >
        <v-icon small>mdi-source-fork</v-icon>
        {{ (item.commit_hash || '').substr(0, 10) }}
      </div>
      <div v-if="sourceTask && sourceTask.commit_message">
        {{ sourceTask.commit_message.substring(0, 50) }}
      </div>
    </v-alert>

    <v-autocomplete
      v-if="buildTasks != null && template.type === 'deploy'"
      v-model="item.build_task_id"
      :label="$t('buildVersion')"
      :items="buildTasks"
      item-value="id"
      :item-text="(itm) => getTaskMessage(itm)"
      :rules="[v => !!v || $t('build_version_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    />

    <v-skeleton-loader
      v-else-if="template.type === 'deploy'"
      type="card"
      height="54"
      style="margin-bottom: 16px; margin-top: 4px;"
    ></v-skeleton-loader>

    <v-text-field
      v-model="item.message"
      :label="$t('messageOptional')"
      :disabled="formSaving"
      outlined
      dense
    />

    <div v-for="(v) in template.survey_vars || []" :key="v.name">

      <v-text-field
        v-if="v.type === 'secret'"
        :label="v.title"
        :hint="v.description"
        v-model="editedSecretEnvironment[v.name]"
        :required="v.required"
        type="password"
        :rules="[
            val => !v.required || !!val || v.title + $t('isRequired'),
          ]"
        outlined
        dense
      />

      <v-select
        clearable
        v-else-if="v.type === 'enum'"
        :label="v.title + (v.required ? ' *' : '')"
        :hint="v.description"
        v-model="editedEnvironment[v.name]"
        :required="v.required"
        :rules="[
          val => !v.required || val != null || v.title + ' ' + $t('isRequired')
        ]"
        :items="v.values"
        item-text="name"
        item-value="value"
        outlined
        dense
      />

      <v-text-field
        v-else
        :label="v.title + (v.required ? ' *' : '')"
        :hint="v.description"
        v-model="editedEnvironment[v.name]"
        :required="v.required"
        :rules="[
          val => !v.required || !!val || v.title + ' ' + $t('isRequired'),
          val => !val || v.type !== 'int' || /^\d+$/.test(val) ||
          v.title + ' ' + $t('mustBeInteger'),
        ]"
        outlined
        dense
      />
    </div>

    <v-text-field
      v-model="git_branch"
      :label="fieldLabel('branch')"
      outlined
      dense
      required
      :disabled="formSaving"
      v-if="
        needField('allow_override_branch')
        && template.allow_override_branch_in_task"
    />

    <v-autocomplete
      v-model="inventory_id"
      :label="fieldLabel('inventory')"
      :items="inventory"
      item-value="id"
      item-text="name"
      outlined
      dense
      required
      :disabled="formSaving"
      v-if="inventory != null && needInventory"
    ></v-autocomplete>

    <v-skeleton-loader
      v-else-if="needInventory"
      type="card"
      height="46"
      style="margin-bottom: 16px; margin-top: 4px;"
    ></v-skeleton-loader>

    <TaskParamsAnsibleForm
      v-if="template.app === 'ansible'"
      v-model="item.params"
      :app="template.app"
      :template-params="template.task_params || {}"
    />

    <TaskParamsTerraformForm
      v-else-if="['terraform', 'tofu', 'terragrunt'].includes(template.app)"
      v-model="item.params"
      :app="template.app"
      :template-params="template.task_params || {}"
    />

    <ArgsPicker
      v-if="template.allow_override_args_in_task"
      :vars="args"
      title="CLI args"
      @change="setArgs"
    />

  </v-form>
</template>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import ArgsPicker from '@/components/ArgsPicker.vue';
import AppFieldsMixin from '@/components/AppFieldsMixin';
import TaskParamsAnsibleForm from '@/components/TaskParamsAnsibleForm.vue';
import TaskParamsTerraformForm from '@/components/TaskParamsTerraformForm.vue';

export default {
  mixins: [ItemFormBase, AppFieldsMixin],

  props: {
    template: Object,
    sourceTask: Object,
  },

  components: {
    TaskParamsAnsibleForm,
    TaskParamsTerraformForm,
    ArgsPicker,
  },

  data() {
    return {
      buildTasks: null,
      hasCommit: null,
      editedEnvironment: null,
      editedSecretEnvironment: null,
      cmOptions: {
        tabSize: 2,
        mode: 'application/json',
        lineNumbers: true,
        line: true,
        lint: true,
        indentWithTabs: false,
      },
      inventory: null,
    };
  },

  computed: {
    needInventory() {
      return this.needField('inventory') && this.template.task_params?.allow_override_inventory;
    },

    args() {
      let res = this.item.arguments;

      if (res == null) {
        res = this.template.arguments;
      }

      if (res == null) {
        res = '[]';
      }

      return JSON.parse(res);
    },

    app() {
      return this.template.app;
    },

    inventory_id: {
      get() {
        return (this.item || {}).inventory_id || this.template.inventory_id;
      },
      set(newValue) {
        this.item.inventory_id = newValue;
      },
    },

    git_branch: {
      get() {
        return (this.item || {}).git_branch || this.template.git_branch;
      },
      set(newValue) {
        this.item.git_branch = newValue;
      },
    },
  },

  watch: {
    needReset(val) {
      if (val) {
        if (this.item) {
          this.item.template_id = this.template.id;
        }
        this.buildTasks = null;
        this.inventory = null;
        // this.template = null;
      }
    },

    template(val) {
      if (this.item) {
        this.item.template_id = val?.id;
      }
    },

    sourceTask(val) {
      this.assignItem(val);
    },

    hasCommit(val) {
      if (val == null) {
        this.commit_hash = null;
      }
    },
  },

  created() {
    this.refreshItem();
  },

  methods: {

    setArgs(args) {
      this.item.arguments = JSON.stringify(args || []);
    },

    getTaskMessage(task) {
      let buildTask = task;

      while (buildTask.version == null && buildTask.build_task != null) {
        buildTask = buildTask.build_task;
      }

      if (!buildTask) {
        return '';
      }

      return buildTask.version + (buildTask.message ? ` — ${buildTask.message}` : '');
    },

    assignItem(val) {
      const v = val || {};

      if (this.item == null) {
        this.item = {};
      }

      Object.keys(v).forEach((field) => {
        this.item[field] = v[field];
      });

      this.editedEnvironment = JSON.parse(v.environment || '{}');
      this.editedSecretEnvironment = JSON.parse(v.secret || '{}');
      this.hasCommit = v.commit_hash != null;
    },

    isLoaded() {
      return this.item != null && this.template != null;
    },

    beforeSave() {
      this.item.environment = JSON.stringify(this.editedEnvironment);
      this.item.secret = JSON.stringify(this.editedSecretEnvironment);
    },

    refreshItem() {
      this.assignItem(this.sourceTask);

      this.item.template_id = this.template.id;

      if (!this.item.params) {
        this.item.params = {};
      }

      ['tags', 'limit', 'skip_tags'].forEach((param) => {
        if (!this.item.params[param]) {
          this.item.params[param] = (this.template.task_params || {})[param];
        }
      });
    },

    async afterLoadData() {
      this.refreshItem();

      [
        this.buildTasks,
        this.inventory,
      ] = await Promise.all([

        this.template.type === 'deploy' ? (await axios({
          keys: 'get',
          url: `/api/project/${this.projectId}/templates/${this.template.build_template_id}/tasks?status=success&limit=20`,
          responseType: 'json',
        })).data.filter((task) => task.status === 'success') : [],

        this.needInventory ? (await axios({
          keys: 'get',
          url: this.getInventoryUrl(),
          responseType: 'json',
        })).data : [],
      ]);

      if (this.item.build_task_id == null
        && this.buildTasks.length > 0
        && this.buildTasks.length > 0) {
        this.item.build_task_id = this.buildTasks[0].id;
      }

      ['tags', 'limit', 'skip_tags'].forEach((param) => {
        if (!this.item.params[param]) {
          this.item.params[param] = (this.template.task_params || {})[param];
        }
      });

      const defaultVars = (this.template.survey_vars || [])
        .filter((s) => s.default_value)
        .reduce((res, curr) => ({
          ...res,
          [curr.name]: curr.default_value,
        }), {});

      this.editedEnvironment = {
        ...defaultVars,
        ...this.editedEnvironment,
      };
    },

    getInventoryUrl() {
      let res = `/api/project/${this.projectId}/inventory?app=${this.app}`;
      switch (this.app) {
        case 'terraform':
        case 'tofu':
          res += `&template_id=${this.template.id}`;
          break;
        default:
          break;
      }
      return res;
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/tasks`;
    },
  },
};
</script>
