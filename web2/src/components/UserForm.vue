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
      label="Name"
      :rules="[v => !!v || 'Name is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.username"
      label="Username"
      :rules="[v => !!v || 'Username is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.email"
      label="Email"
      :rules="[v => !!v || 'Email is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-checkbox
      v-model="item.admin"
      label="Admin user"
    ></v-checkbox>

    <v-checkbox
      v-model="item.alert"
      label="Send alerts"
    ></v-checkbox>
  </v-form>
</template>
<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    userId: [Number, String],
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
      return this.userId === 'new';
    },
    itemId() {
      return this.userId;
    },
  },

  methods: {
    async reset() {
      this.item = null;
      this.$refs.form.resetValidation();
      await this.loadData();
    },

    async loadData() {
      if (this.isNew) {
        this.item = {};
      } else {
        this.item = (await axios({
          method: 'get',
          url: `/api/users/${this.userId}`,
          responseType: 'json',
        })).data;
      }
    },

    /**
     * Saves or creates user via API.
     * @returns {Promise<null>} null if validation didn't pass or user data if user saved.
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
            ? '/api/users'
            : `/api/users/${this.userId}`,
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
