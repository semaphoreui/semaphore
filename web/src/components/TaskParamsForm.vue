<template>
  <div>
    <v-text-field v-model="item.message" :label="$t('messageOptional')" outlined dense />

    <div v-for="v in template.survey_vars || []" :key="v.name">
      <v-text-field
        v-if="v.type === 'secret'"
        :label="v.title"
        :hint="v.description"
        v-model="editedSecretEnvironment[v.name]"
        :required="v.required"
        type="password"
        :rules="[(val) => !v.required || !!val || v.title + $t('isRequired')]"
        outlined
        dense
      />

      <v-select
        clearable
        v-else-if="v.type === 'enum' || v.type === 'select'"
        :label="v.title + (v.required ? ' *' : '')"
        :hint="v.description"
        v-model="editedEnvironment[v.name]"
        :required="v.required"
        :rules="[(val) => !v.required || val != null || v.title + ' ' + $t('isRequired')]"
        :items="v.values"
        item-text="name"
        item-value="value"
        :multiple="v.type === 'select'"
        :chips="v.type === 'select'"
        outlined
        dense
      >
        <template v-if="v.type === 'select'" v-slot:selection="{ item, index }">
          <v-chip
            small
            close
            @click:close="removeSelectedItem(v.name, index)"
          >
            {{ item.name }}
          </v-chip>
        </template>
      </v-select>

      <v-text-field
        v-else
        :label="v.title + (v.required ? ' *' : '')"
        :hint="v.description"
        v-model="editedEnvironment[v.name]"
        :required="v.required"
        :rules="[
          (val) => !v.required || !!val || v.title + ' ' + $t('isRequired'),
          (val) =>
            !val || v.type !== 'int' || /^\d+$/.test(val) || v.title + ' ' + $t('mustBeInteger'),
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
      v-if="needField('allow_override_branch') && template.allow_override_branch_in_task"
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
      v-if="inventory != null && needInventory"
    ></v-autocomplete>

    <v-skeleton-loader
      v-else-if="needInventory"
      type="card"
      height="46"
      style="margin-bottom: 16px; margin-top: 4px"
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
  </div>
</template>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import axios from 'axios';
import ArgsPicker from '@/components/ArgsPicker.vue';
import AppFieldsMixin from '@/components/AppFieldsMixin';
import TaskParamsAnsibleForm from '@/components/TaskParamsAnsibleForm.vue';
import TaskParamsTerraformForm from '@/components/TaskParamsTerraformForm.vue';

export default {
  mixins: [AppFieldsMixin],

  props: {
    value: Object,
    template: Object,
  },

  components: {
    TaskParamsAnsibleForm,
    TaskParamsTerraformForm,
    ArgsPicker,
  },

  data() {
    return {
      item: null,
      editedEnvironment: null,
      editedSecretEnvironment: null,
      inventory: null,
    };
  },

  computed: {
    needInventory() {
      return this.template.task_params?.allow_override_inventory;
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
    editedEnvironment: {
      handler(newVal, oldVal) {
        if (oldVal == null) {
          return;
        }
        this.item.environment = JSON.stringify(this.editedEnvironment);
        this.$emit('input', this.item);
      },
      deep: true,
      immediate: true,
    },

    editedSecretEnvironment: {
      handler(newVal, oldVal) {
        if (oldVal == null) {
          return;
        }
        this.item.secret = JSON.stringify(this.editedSecretEnvironment);
        this.$emit('input', this.item);
      },
      deep: true,
      immediate: true,
    },

    item: {
      handler(newVal, oldVal) {
        if (oldVal == null) {
          return;
        }
        this.$emit('input', newVal);
      },
      deep: true,
      immediate: true,
    },

    needReset(val) {
      if (val) {
        // if (this.item) {
        //   this.item.template_id = this.template.id;
        // }
        this.buildTasks = null;
        this.inventory = null;
        // this.template = null;
      }
    },
  },

  async created() {
    await this.afterLoadData();
  },

  methods: {
    setArgs(args) {
      this.item.arguments = JSON.stringify(args || []);
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

      // Normalize select type variables to ensure they are arrays
      if (this.template && this.template.survey_vars) {
        this.template.survey_vars.forEach((surveyVar) => {
          if (surveyVar.type === 'select' && this.editedEnvironment[surveyVar.name] !== undefined) {
            const currentValue = this.editedEnvironment[surveyVar.name];
            if (!Array.isArray(currentValue)) {
              this.editedEnvironment[surveyVar.name] = currentValue == null || currentValue === '' ? [] : [currentValue];
            }
          }
        });
      }
    },

    isLoaded() {
      return this.item != null && this.template != null;
    },

    refreshItem() {
      this.assignItem(this.value);

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

      [this.buildTasks, this.inventory] = await Promise.all([
        this.template.type === 'deploy'
          ? (
            await axios({
              keys: 'get',
              url: `/api/project/${this.projectId}/templates/${this.template.build_template_id}/tasks?status=success&limit=20`,
              responseType: 'json',
            })
          ).data.filter((task) => task.status === 'success')
          : [],

        this.needInventory
          ? (
            await axios({
              keys: 'get',
              url: this.getInventoryUrl(),
              responseType: 'json',
            })
          ).data
          : [],
      ]);

      ['tags', 'limit', 'skip_tags'].forEach((param) => {
        if (!this.item.params[param]) {
          this.item.params[param] = (this.template.task_params || {})[param];
        }
      });

      const defaultVars = (this.template.survey_vars || [])
        .filter((s) => s.default_value)
        .reduce(
          (res, curr) => ({
            ...res,
            [curr.name]: curr.default_value,
          }),
          {},
        );

      this.editedEnvironment = { ...defaultVars, ...this.editedEnvironment };

      // Ensure select type variables without values are initialized as empty arrays
      (this.template.survey_vars || []).forEach((surveyVar) => {
        if (surveyVar.type === 'select' && this.editedEnvironment[surveyVar.name] === undefined) {
          this.editedEnvironment[surveyVar.name] = [];
        }
      });
    },

    getInventoryUrl() {
      let res = `/api/project/${this.template.project_id}/inventory?app=${this.app}`;
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

    removeSelectedItem(varName, index) {
      const env = this.editedEnvironment;
      if (!Object.prototype.hasOwnProperty.call(env, varName) || !Array.isArray(env[varName])) {
        return;
      }

      if (index < env[varName].length) {
        env[varName].splice(index, 1);
      }
    },
  },
};
</script>
