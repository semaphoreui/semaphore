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

    <v-textarea
      outlined
      class="mt-4"
      v-model="item.environment"
      label="Environment Override"
      placeholder='Example: {"version": 10, "author": "John"}'
      :disabled="formSaving"
      rows="4"
    ></v-textarea>

    <v-textarea
      outlined
      v-model="item.arguments"
      label="Extra CLI Arguments"
      :disabled="formSaving"
      placeholder='Example: ["-i", "@myinventory.sh", "--private-key=/there/id_rsa", "-vvvv"]'
      rows="4"
    ></v-textarea>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  props: {
    templateId: Number,
  },
  watch: {
    needReset(val) {
      if (val) {
        this.item.template_id = this.templateId;
      }
    },

    templateId(val) {
      this.item.template_id = val;
    },
  },
  created() {
    this.item.template_id = this.templateId;
  },
  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/tasks`;
    },
  },
};
</script>
