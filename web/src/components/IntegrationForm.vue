<template>
  <v-form ref="form" lazy-validation v-model="formValid" v-if="isLoaded">
    <v-alert :value="formError" color="error" class="pb-2">{{ formError }} </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-autocomplete
      v-model="item.group_id"
      :label="$t('integrationGroup')"
      :items="groups"
      item-value="id"
      item-text="name"
      :rules="[(v) => !!v || $t('groupRequired')]"
      :disabled="formSaving"
      :no-data-text="$t('noIntegrationGroups')"
      outlined
      dense
      class="IntegrationForm__group"
    ></v-autocomplete>

    <div
      v-if="workflowsEnabled"
      class="d-flex justify-space-between align-center mb-2 IntegrationForm__target"
    >
      <b style="font-size: 13px; margin-left: 5px">{{ $t('integrationRunTarget') }}</b>
      <v-btn-toggle
        :value="target"
        @change="setTarget($event)"
        tile
        group
        mandatory
        :disabled="formSaving"
      >
        <v-btn value="template" small class="ma-0" style="border-radius: 4px">
          {{ $t('integrationTargetTemplate') }}
        </v-btn>
        <v-btn value="workflow" small class="ma-0" style="border-radius: 4px">
          {{ $t('integrationTargetWorkflow') }}
        </v-btn>
      </v-btn-toggle>
    </div>

    <v-alert
      v-if="target === 'workflow' && !workflowsEnabled"
      type="info"
      text
      dense
      class="IntegrationForm__workflowUnavailable"
    >
      {{ $t('integrationWorkflowTargetUnavailable', { id: item.workflow_id }) }}
    </v-alert>

    <div v-else-if="target === 'workflow'" style="position: relative">
      <v-autocomplete
        v-model="item.workflow_id"
        :label="$t('workflowToRun')"
        clearable
        :items="workflows"
        item-value="id"
        item-text="name"
        :rules="[(v) => !!v || $t('integrationTargetRequired')]"
        :disabled="formSaving || !workflowIntegrationsAvailable"
        outlined
        dense
        class="IntegrationForm__workflow"
      ></v-autocomplete>

      <v-chip
        v-if="!workflowIntegrationsAvailable"
        color="hsl(348deg, 86%, 61%)"
        text-color="white"
        small
        label
        class="IntegrationForm__upgrade"
        style="position: absolute; top: 8px; right: 8px"
        @click="upgradeToPro('workflow_integrations_management')"
      >
        Upgrade to PRO
      </v-chip>
    </div>

    <v-autocomplete
      v-else
      v-model="item.template_id"
      :label="$t('templateToRun')"
      clearable
      :items="templates"
      item-value="id"
      item-text="name"
      :rules="[(v) => !!v || $t('integrationTargetRequired')]"
      :disabled="formSaving"
      outlined
      dense
    ></v-autocomplete>

    <v-card
      v-if="target === 'template' && item.template_id"
      style="background: var(--highlighted-card-bg-color)"
      class="mb-6 pt-3"
    >
      <div
        style="
          position: absolute;
          background: var(--highlighted-card-bg-color);
          width: 28px;
          height: 28px;
          transform: rotate(45deg);
          left: calc(50% - 14px);
          top: -14px;
          border-radius: 0;
        "
      ></div>

      <v-card-text>
        <TaskParamsForm
          :template="templates.find((t) => t.id === item.template_id)"
          v-model="item.task_params"
        />
      </v-card-text>
    </v-card>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import TaskParamsForm from '@/components/TaskParamsForm.vue';

export default {
  components: { TaskParamsForm },
  mixins: [ItemFormBase],
  props: {
    /**
     * Server feature flags (`systemInfo.features`): `workflows` shows the workflow target,
     * `workflow_integrations_management` (Pro) lets the user pick a workflow.
     */
    features: Object,
    /** Group to preselect for a new integration (opened from a group's "New integration"). */
    groupId: Number,
  },
  data() {
    return {
      templates: [],
      workflows: [],
      groups: [],
      /** What the integration runs: 'template' or 'workflow'. Derived from the item on load. */
      target: 'template',
    };
  },
  async created() {
    const [templates, groups] = await Promise.all([
      axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates`,
        responseType: 'json',
      }),
      axios({
        method: 'get',
        url: `/api/project/${this.projectId}/integration_groups`,
        responseType: 'json',
      }),
    ]);
    this.templates = templates.data || [];
    this.groups = groups.data || [];

    if (this.workflowsEnabled) {
      this.workflows = (
        await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/workflows`,
          responseType: 'json',
        })
      ).data || [];
    }
  },

  computed: {
    isLoaded() {
      return this.item != null;
    },

    workflowsEnabled() {
      return !!(this.features && this.features.workflows);
    },

    /**
     * Whether this user may point an integration at a workflow. Without it the workflow
     * select is shown disabled with an "Upgrade to PRO" chip, and the server refuses the
     * target with 403 anyway.
     */
    workflowIntegrationsAvailable() {
      return !!(this.features && this.features.workflow_integrations_management);
    },

    template() {
      return this.templates.find((t) => t.id === this.item.template_id);
    },
  },

  methods: {
    getNewItem() {
      return {
        group_id: this.groupId || null,
        template_id: null,
        workflow_id: null,
      };
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/integrations`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.itemId}`;
    },

    /**
     * Switches the run target. Exactly one of template_id / workflow_id is kept;
     * task params only make sense for a template.
     */
    setTarget(target) {
      if (!target || target === this.target) {
        return;
      }
      this.target = target;
      if (target === 'workflow') {
        this.item.template_id = null;
        this.item.task_params = null;
      } else {
        this.item.workflow_id = null;
        if (this.item.task_params == null) {
          this.item.task_params = {};
        }
      }
    },

    async afterLoadData() {
      // The API omits workflow_id when it is null; add the key reactively so the
      // workflow autocomplete can bind to it.
      if (this.item.workflow_id === undefined) {
        this.$set(this.item, 'workflow_id', null);
      }
      if (this.item.group_id === undefined) {
        this.$set(this.item, 'group_id', this.groupId || null);
      }

      this.target = this.item.workflow_id ? 'workflow' : 'template';

      if (this.target === 'template' && this.item.task_params == null) {
        this.item.task_params = {};
      }
    },

    beforeSave() {
      if (this.target === 'workflow') {
        this.item.template_id = null;
        this.item.task_params = null;
      } else {
        this.item.workflow_id = null;
      }
    },
  },
};
</script>
