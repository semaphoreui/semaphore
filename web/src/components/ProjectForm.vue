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
      :label="$t(projectNameTitle)"
      :rules="[v => !!v || $t('project_name_required')]"
      required
      :disabled="formSaving"
      data-testid="newProject-name"
      outlined
      dense
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
      outlined
      dense
    ></v-text-field>

    <v-text-field
      v-model="item.alert_chat"
      :label="$t('telegramChatIdOptional')"
      :disabled="formSaving"
      data-testid="newProject-tg"
      outlined
      dense
    ></v-text-field>

    <v-checkbox
      class="mt-0"
      v-model="item.alert"
      :label="$t('allowAlertsForThisProject')"
      data-testid="newProject-alert"
    ></v-checkbox>

    <v-switch
      v-if="itemId === 'new'"
      v-model="item.demo"
      label="Demo"
      style="position: absolute; left: 24px; bottom: 15px;"
      hide-details
    />

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  props: {
    projectNameTitle: {
      type: String,
      default: 'projectName',
    },
  },
  methods: {
    getItemsUrl() {
      return '/api/projects';
    },
    getSingleItemUrl() {
      return `/api/project/${this.itemId}`;
    },
    beforeSave() {
      if (this.item.max_parallel_tasks === '') {
        this.item.max_parallel_tasks = 0;
      }
    },
  },
};
</script>
