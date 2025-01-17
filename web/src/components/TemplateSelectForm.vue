<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="items != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-select
      :rules="[v => !!v || $t('template_required')]"
      :items="items"
      v-model="itemId"
      label="Template"
      item-value="id"
      item-text="name"
    />
  </v-form>
</template>
<style>
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {

  props: {
    app: String,
    projectId: Number,
    needSave: Boolean,
    needReset: Boolean,
  },

  components: {},

  watch: {
    async needSave(val) {
      if (val) {
        await this.save();
      }
    },
    async needReset(val) {
      if (val) {
        await this.reset();
      }
    },
  },

  data() {
    return {
      itemId: null,
      items: null,
      formValid: false,
      formError: null,
    };
  },

  async created() {
    await this.loadData();
  },

  methods: {
    async loadData() {
      this.formError = null;
      try {
        this.items = (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/templates?app=${this.app}`,
          responseType: 'json',
        })).data;
      } catch (err) {
        this.formError = getErrorMessage(err);
        this.$emit('error', {
          message: this.formError,
        });
      }
    },

    async reset() {
      this.itemId = null;
      this.formError = null;
      if (this.$refs.form) {
        this.$refs.form.resetValidation();
      }
      await this.loadData();
    },

    async save() {
      this.formError = null;

      if (!this.$refs.form.validate()) {
        this.$emit('error', {});
        return null;
      }

      this.$emit('save', {
        itemId: this.itemId,
      });

      return this.itemId;
    },

  },
};
</script>
