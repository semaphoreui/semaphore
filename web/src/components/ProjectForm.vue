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
      :label="$t('projectName')"
      :rules="[v => !!v || $t('project_name_required')]"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-checkbox
      v-model="item.alert"
      :label="$t('allowAlertsForThisProject')"
    ></v-checkbox>

    <v-text-field
      v-model="item.alert_chat"
      :label="$t('telegramChatIdOptional')"
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model.number="item.max_parallel_tasks"
      :label="$t('maxNumberOfParallelTasksOptional')"
      :disabled="formSaving"
      :rules="[
        v => (v == null || v === '' || Math.floor(v) === v) || $t('mustBeInteger'),
        v => (v == null || v === '' || v >= 0) || $t('mustBe0OrGreater'),
      ]"
      hint="Should be 0 or greater, 0 - unlimited."
      type="number"
      :step="1"
    ></v-text-field>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  methods: {
    getItemsUrl() {
      return '/api/projects';
    },
    getSingleItemUrl() {
      return `/api/project/${this.itemId}`;
    },
  },
};
</script>
