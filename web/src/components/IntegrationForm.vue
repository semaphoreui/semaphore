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

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],
  data() {
    return {
      templates: [],
      authMethods: [{
        id: '',
        title: 'None',
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
        keys: 'get',
        url: `/api/project/${this.projectId}/keys`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
