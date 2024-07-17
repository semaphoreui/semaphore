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
    ></v-text-field>

    <v-select
      v-model="item.template_id"
      label="Task Template to run"
      clearable
      :items="templates"
      item-value="id"
      item-text="name"
      :disabled="formSaving"
    ></v-select>

    <v-select
      v-model="item.auth_method"
      label="Auth method"
      :items="authMethods"
      item-value="id"
      item-text="title"
      :disabled="formSaving"
    ></v-select>

    <v-text-field
      v-if="['token', 'hmac'].includes(item.auth_method)"
      v-model="item.auth_header"
      label="Auth header"
      :disabled="formSaving"
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
    ></v-select>

    <TaskParamsForm
      v-if="item.template_id"
      v-model="item.task_params"
      :app="(template || {}).app"
    />
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
        id: 'token',
        title: 'Token',
      }, {
        id: 'hmac',
        title: 'HMAC',
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
      return this.keys != null;
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
        template_id: {},
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
