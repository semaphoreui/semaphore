<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
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
      v-model="item.type"
      label="Username"
      :rules="[v => !!v || 'Type is required']"
      required
      :disabled="formSaving"
    ></v-text-field>
  </v-form>
</template>
<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    itemId: [Number, String],
    projectId: [Number, String],
    needSave: Boolean,
    needReset: Boolean,
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
    isNew() {
      return this.itemId === 'new';
    },
  },

  watch: {
    needSave(val) {
      if (val) {
        this.save();
      }
    },
    needReset(val) {
      if (val) {
        this.reset();
      }
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
          url: `/api/project/${this.projectId}/inventory/${this.itemId}`,
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
        return;
      }

      this.formSaving = true;
      try {
        const item = (await axios({
          method: this.isNew ? 'post' : 'put',
          url: this.isNew
            ? `/api/project/${this.projectId}/inventory`
            : `/api/project/${this.projectId}/inventory/${this.item.id}`,
          responseType: 'json',
          data: this.item,
        })).data;

        this.$emit('save', {
          item,
          action: this.isNew ? 'new' : 'edit',
        });
      } catch (err) {
        this.formError = getErrorMessage(err);
      } finally {
        this.formSaving = false;
      }
    },
  },
};
</script>
