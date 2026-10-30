<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="items != null"
    class="InventorySelectForm"
    data-testid="workspaceSelect-form"
    :data-state="formError ? 'error' : 'ready'"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="workspaceSelect-error"
    >{{ formError }}
    </v-alert>

    <v-select
      :rules="[v => !!v || $t('inventory_required')]"
      :items="items"
      v-model="itemId"
      label="Workspace"
      item-value="id"
      item-title="name"
      :menu-props="underlinedMenu"
      data-testid="workspaceSelect-inventoryId"
    />
  </v-form>
</template>
<style>
/* An underlined field of Vuetify 2 is a 4 px top margin, 12 px for the label and the 32 px line;
   the 48 px field of Vuetify 4 holds the margin. Under the 16 px margin of the error alert that
   margin collapsed, so the field stood 4 px higher (as RestoreProjectForm.vue). */
.InventorySelectForm > .v-alert + .v-input {
  margin-top: -4px;
}
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

  emits: ['error', 'save'],

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
      // The list of an underlined select opens over the input, below the label (Vuetify 2: the
      // menu covers `.v-input__slot`, 16 px under the top of the field), not over the label;
      // at most 304 px high (Vuetify 2 `menuProps`; Vuetify 4: 310).
      underlinedMenu: {
        location: 'top', origin: 'top', offset: -16, maxHeight: 304,
      },
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
          url: `/api/project/${this.projectId}/inventory?app=${this.app}`,
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

      const { valid } = await this.$refs.form.validate();
      if (!valid) {
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
