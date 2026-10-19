<template>
  <v-form ref="form" lazy-validation v-model="formValid" v-if="isLoaded">
    <v-alert :value="formError" color="error" class="pb-2">{{ formError }} </v-alert>

    <v-text-field
      v-model="item.name"
      label="Name"
      :rules="[(v) => !!v || 'Name is required']"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

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
        style="position: absolute; top: -10px; right: 15px"
        @click="upgradeToPro('workflow_integrations_management')"
      >
        Upgrade to PRO
      </v-chip>
    </div>

    <v-autocomplete
      v-else
      v-model="item.template_id"
      label="Task Template to run"
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

    <v-select
      v-model="item.auth_method"
      label="Auth method"
      :items="authMethods"
      item-value="id"
      item-text="title"
      :disabled="formSaving"
      outlined
      dense
    ></v-select>

    <HighlightedCard v-if="item.auth_method">
      <template>
        <v-text-field
            v-if="['token', 'hmac', 'hmac-sha512'].includes(item.auth_method)"
            v-model="item.auth_header"
            label="Auth header"
            :disabled="formSaving"
            outlined
            dense
        ></v-text-field>

        <v-select
            v-model="item.auth_secret_id"
            :label="$t('vaultPassword2')"
            clearable
            :items="loginPasswordKeys"
            item-value="id"
            item-text="name"
            :disabled="formSaving"
            outlined
            dense
        ></v-select>
      </template>
    </HighlightedCard>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import TaskParamsForm from '@/components/TaskParamsForm.vue';
import HighlightedCard from '@/components/HighlightedCard.vue';

export default {
  components: { HighlightedCard, TaskParamsForm },
  mixins: [ItemFormBase],
  props: {
    /**
     * Server feature flags (`systemInfo.features`): `workflows` shows the workflow target,
     * `workflow_integrations_management` (Pro) lets the user pick a workflow.
     */
    features: Object,
  },
  data() {
    return {
      templates: [],
      workflows: [],
      /** What the integration runs: 'template' or 'workflow'. Derived from the item on load. */
      target: 'template',
      authMethods: [
        {
          id: '',
          title: 'None',
        },
        {
          id: 'github',
          title: 'GitHub Webhooks',
        },
        {
          id: 'bitbucket',
          title: 'Bitbucket Webhooks',
        },
        {
          id: 'token',
          title: 'Token',
        },
        {
          id: 'hmac',
          title: 'HMAC',
        },
        {
          id: 'hmac-sha512',
          title: 'HMAC (SHA-512)',
        },
        {
          id: 'basic',
          title: 'BasicAuth',
        },
      ],
      keys: null,
    };
  },
  async created() {
    this.templates = (
      await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates`,
        responseType: 'json',
      })
    ).data;

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
      return this.item && this.keys != null;
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

    loginPasswordKeys() {
      if (this.keys == null) {
        return null;
      }
      return this.keys.filter((key) => key.type === 'login_password');
    },

    template() {
      return this.templates.find((t) => t.id === this.item.template_id);
    },
  },

  methods: {
    getNewItem() {
      return {
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
      this.keys = (
        await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/keys`,
          responseType: 'json',
        })
      ).data;

      // The API omits workflow_id when it is null; add the key reactively so the
      // workflow autocomplete can bind to it.
      if (this.item.workflow_id === undefined) {
        this.$set(this.item, 'workflow_id', null);
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
