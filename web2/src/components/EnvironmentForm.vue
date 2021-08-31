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
      label="Environment Name"
      :rules="[v => !!v || 'Name is required']"
      required
      :disabled="formSaving"
      class="mb-4"
    ></v-text-field>

    <v-textarea
      v-model="item.json"
      label="Environment (This has to be a JSON object)"
      :disabled="formSaving"
      solo
    ></v-textarea>

    <v-alert
        dense
        type="info"
    >
      Must be valid JSON. You may use the key <code>ENV</code> to pass environment variables
      to ansible-playbook.
      Example:
      <pre style="font-size: 14px;">{
  "var_available_in_playbook_1": 1245,
  "var_available_in_playbook_2": "test",
  "ENV": {
    "VAR1": "Read by lookup('env', 'VAR1')"
  }
}</pre>
    </v-alert>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/environment`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/environment/${this.itemId}`;
    },
  },
};
</script>
