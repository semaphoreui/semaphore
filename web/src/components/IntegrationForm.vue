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
    >{{ formError }}
    </v-alert>

    <v-text-field
      v-model="item.name"
      label="Name"
      :rules="[v => !!v || 'Name is required']"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-autocomplete
      v-model="item.template_id"
      label="Task Template to run"
      clearable
      :items="templates"
      item-value="id"
      item-text="name"
      :disabled="formSaving"
      outlined
      dense
    ></v-autocomplete>

    <v-card
      v-if="item.template_id"
      style="background: rgba(133, 133, 133, 0.06)"
      class="mb-6 pt-3"
    >

      <div style="
        position: absolute;
        background: var(--highlighted-card-bg-color);
        width: 28px;
        height: 28px;
        transform: rotate(45deg);
        left: calc(50% - 14px);
        top: -14px;
        border-radius: 0;
      "></div>

      <v-card-text>
        <TaskParamsForm
          :template="templates.find(t => t.id === item.template_id)"
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

    <v-text-field
      v-if="['token', 'hmac'].includes(item.auth_method)"
      v-model="item.auth_header"
      label="Auth header"
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-select
      v-if="item.auth_method"
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

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import TaskParamsForm from '@/components/TaskParamsForm.vue';

export default {
  components: { TaskParamsForm },
  mixins: [ItemFormBase],
  data() {
    return {
      templates: [],
      authMethods: [{
        id: '',
        title: 'None',
      }, {
        id: 'github',
        title: 'GitHub Webhooks',
      }, {
        id: 'bitbucket',
        title: 'Bitbucket Webhooks',
      }, {
        id: 'token',
        title: 'Token',
      }, {
        id: 'hmac',
        title: 'HMAC',
      }, {
        id: 'basic',
        title: 'BasicAuth',
      }],
      keys: null,
    };
  },
  async created() {
    this.templates = (await axios({
      templates: 'get',
      url: `/api/project/${this.projectId}/templates`,
      responseType: 'json',
    })).data;
  },

  computed: {
    isLoaded() {
      return this.item && this.keys != null;
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
      };
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/integrations`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.itemId}`;
    },

    async afterLoadData() {
      this.keys = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/keys`,
        responseType: 'json',
      })).data;

      if (this.item.task_params == null) {
        this.item.task_params = {};
      }
    },

  },
};
</script>
