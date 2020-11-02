<template>
  <v-form
    ref="itemForm"
    lazy-validation
    v-model="itemFormValid"
    v-if="isNewItem || isLoaded"
  >
    <v-text-field
      v-model="item.name"
      label="Playbook Alias"
      :rules="[v => !!v || 'Project name is required']"
      required
      :disabled="itemFormSaving"
    ></v-text-field>

    <v-checkbox
      v-model="item.alert"
      label="Allow alerts for this project"
    ></v-checkbox>

    <v-text-field
      v-model="item.alert_chat"
      label="Chat ID"
      :disabled="itemFormSaving"
    ></v-text-field>
  </v-form>
</template>
<script>
import axios from 'axios';

export default {
  props: {
    projectId: [Number, String],
  },

  data() {
    return {
      item: null,

      itemFormValid: false,
      itemFormError: null,
      itemFormSaving: false,
    };
  },

  async created() {
    if (this.isNewItem) {
      this.item = {};
    } else {
      this.item = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}`,
        responseType: 'json',
      })).data;
    }
  },

  computed: {
    isLoaded() {
      return this.item != null;
    },
    isNewItem() {
      return this.projectId === 'new';
    },
    itemId() {
      return this.projectId;
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
            ? '/api/project'
            : `/api/project/${this.projectId}`,
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
