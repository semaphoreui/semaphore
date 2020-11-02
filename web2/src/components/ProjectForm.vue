<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="isNew || isLoaded"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      label="Playbook Alias"
      :rules="[v => !!v || 'Project name is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-checkbox
      v-model="item.alert"
      label="Allow alerts for this project"
    ></v-checkbox>

    <v-text-field
      v-model="item.alert_chat"
      label="Chat ID"
      :disabled="formSaving"
    ></v-text-field>
  </v-form>
</template>
<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    projectId: [Number, String],
  },

  data() {
    return {
      item: null,
      formValid: false,
      formError: null,
      formSaving: false,
    };
  },

  async created() {
    await this.loadData();
  },

  computed: {
    isLoaded() {
      return this.item != null;
    },
    isNew() {
      return this.projectId === 'new';
    },
    itemId() {
      return this.projectId;
    },
  },

  methods: {
    async reset() {
      this.$refs.form.resetValidation();
      await this.loadData();
    },

    async loadData() {
      if (this.isNew) {
        this.item = {};
      } else {
        this.item = (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}`,
          responseType: 'json',
        })).data;
      }
    },

    /**
     * Saves or creates project via API.
     * Method must be wrapped to try-catch block because it can throws exception.
     * @returns {Promise<null>} null if validation not passed or project data if project saved.
     */
    async save() {
      this.formError = null;

      if (!this.$refs.form.validate()) {
        return null;
      }

      this.formSaving = true;

      let item;
      try {
        item = (await axios({
          method: this.isNew ? 'post' : 'put',
          url: this.isNew
            ? '/api/projects'
            : `/api/project/${this.projectId}`,
          responseType: 'json',
          data: this.item,
        })).data;
      } catch (err) {
        this.formError = getErrorMessage(err);
      } finally {
        this.formSaving = false;
      }

      return item || this.item;
    },
  },
};
</script>
