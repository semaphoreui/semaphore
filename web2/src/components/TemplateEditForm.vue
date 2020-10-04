<template>
  <v-form
    ref="itemForm"
    lazy-validation
    v-model="itemFormValid"
    class="pa-4"
  >
    <v-text-field
      v-model="item.alias"
      label="Playbook Alias"
      required
      :disabled="itemFormSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.name"
      label="Playbook Name"
      required
      :disabled="itemFormSaving"
    ></v-text-field>

    <v-select
      v-model="item.ssh_key_id"
      label="SSH Key"
      :items="sshKeys"
      :rules="[v => !!v || 'Item is required']"
      required
    ></v-select>
  </v-form>
</template>
<script>
import axios from 'axios';

export default {
  props: {
    template: Object,
  },
  data() {
    return {
      sshKeys: [],
      itemFormValid: false,
      itemFormError: null,
      itemFormSaving: false,
    };
  },

  computed: {
    isNewItem() {
      return this.template == null;
    },
    item() {
      return this.template ? this.template : {};
    },
  },

  methods: {
    async saveItem() {
      if (!this.$refs.itemForm.validate()) {
        return null;
      }

      this.itemFormSaving = true;
      try {
        await axios({
          method: this.isNewItem ? 'post' : 'put',
          url: this.isNewItem
            ? `/api/project/${this.projectId}/templates`
            : `/api/project/${this.projectId}/templates/${this.item.id}`,
          responseType: 'json',
          data: this.item,
        });
      } finally {
        this.itemFormSaving = true;
      }

      return this.item;
    },
  },
};
</script>
