<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="isLoaded()"
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

    <v-select
      v-if="template.type === 'deploy'"
      v-model="item.build_task_id"
      :label="$t('buildVersion')"
      :items="buildTasks"
      item-value="id"
      :item-text="(itm) => getTaskMessage(itm)"
      :rules="[v => !!v || $t('build_version_required')]"
      required
      :disabled="formSaving"
    />

    <v-text-field
      v-model="item.message"
      :label="$t('messageOptional')"
      :disabled="formSaving"
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
      />
    </div>

    <div class="pt-3"></div>

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

    <v-select
      v-model="inventory_id"
      :label="fieldLabel('inventory')"
      :items="inventory"
      item-value="id"
      item-text="name"
      outlined
      dense
      required
      :disabled="formSaving"
      v-if="needField('inventory') && (template.task_params || {}).allow_override_inventory"
    ></v-select>

    <ArgsPicker
      v-if="needField('limit') && (template.task_params || {}).allow_override_limit"
      :vars="item.params.limit"
      @change="setLimit"
      :title="$t('limit')"
      :arg-title="$t('limit')"
      :add-arg-title="$t('addLimit')"
    />

    <ArgsPicker
      v-if="needField('tags') && (template.task_params || {}).allow_override_tags"
      :vars="item.params.tags"
      @change="setTags"
      :title="$t('tags')"
      :arg-title="$t('tags')"
      :add-arg-title="$t('addTag')"
    />

    <ArgsPicker
      v-if="needField('skip_tags') && (template.task_params || {}).allow_override_limit"
      :vars="item.params.skip_tags"
      @change="setSkipTags"
      :title="$t('skipTags')"
      :arg-title="$t('tag')"
      :add-arg-title="$t('addSkippedTag')"
    />

    <TaskParamsForm
      v-if="template.app === 'ansible'"
      v-model="item.params"
      :app="template.app"
      :template-params="template.task_params || {}"
    />

    <TaskParamsForm
      v-else
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
import TaskParamsForm from '@/components/TaskParamsForm.vue';
import ArgsPicker from '@/components/ArgsPicker.vue';
import AppFieldsMixin from '@/components/AppFieldsMixin';

export default {
  mixins: [ItemFormBase, AppFieldsMixin],

  props: {
    templateId: Number,
    sourceTask: Object,
  },

  components: {
    ArgsPicker,
    TaskParamsForm,
  },

  data() {
    return {
      template: null,
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
    args() {
      return JSON.parse(this.item.arguments || '[]');
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
          this.item.template_id = this.templateId;
        }
        this.inventory = null;
        this.template = null;
      }
    },

    templateId(val) {
      if (this.item) {
        this.item.template_id = val;
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

  methods: {

    setSkipTags(tags) {
      this.item.params.skip_tags = tags;
    },

    setTags(tags) {
      this.item.params.tags = tags;
    },

    setLimit(limit) {
      this.item.params.limit = limit;
    },

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
      return this.item != null
        && this.template != null
        && this.buildTasks != null
        && this.inventory != null;
    },

    beforeSave() {
      this.item.environment = JSON.stringify(this.editedEnvironment);
      this.item.secret = JSON.stringify(this.editedSecretEnvironment);
    },

    async afterLoadData() {
      this.assignItem(this.sourceTask);

      this.item.template_id = this.templateId;

      if (!this.item.params) {
        this.item.params = {};
      }

      this.template = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/templates/${this.templateId}`,
        responseType: 'json',
      })).data;

      this.buildTasks = this.template.type === 'deploy' ? (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/templates/${this.template.build_template_id}/tasks?status=success`,
        responseType: 'json',
      })).data.filter((task) => task.status === 'success') : [];

      this.inventory = (await axios({
        keys: 'get',
        url: this.getInventoryUrl(),
        responseType: 'json',
      })).data;

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
    },

    getInventoryUrl() {
      let res = `/api/project/${this.projectId}/inventory?app=${this.app}`;
      switch (this.app) {
        case 'terraform':
        case 'tofu':
          res += `&template_id=${this.templateId}`;
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
