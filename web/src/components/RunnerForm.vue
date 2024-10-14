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
      :label="$t('name')"
      :rules="[v => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.webhook"
      :label="$t('webhook')"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      type="number"
      v-model.number="item.max_parallel_tasks"
      :label="$t('maxNumberOfParallelTasksOptional')"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-checkbox
      v-model="item.active"
      :label="$t('enabled')"
    ></v-checkbox>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  props: {
    isAdmin: Boolean,
  },
  mixins: [ItemFormBase],
  methods: {
    getItemsUrl() {
      return '/api/runners';
    },

    getSingleItemUrl() {
      return `/api/runners/${this.itemId}`;
    },
  },
};
</script>
