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
      v-model="item.playbook"
      label="Playbook Override"
      :disabled="formSaving"
    ></v-text-field>

    <v-textarea
      v-model="item.environment"
      label="Environment Override (*MUST* be valid JSON)"
      :disabled="formSaving"
      rows="4"
    ></v-textarea>

    <v-textarea
      v-model="item.arguments"
      label="Extra CLI Arguments"
      :disabled="formSaving"
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
