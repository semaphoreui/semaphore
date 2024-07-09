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
        v-model="id"
        :label="$t('ID')"
        :rules="[v => !!v || $t('id_required')]"
        required
        :disabled="formSaving"
    ></v-text-field>

    <v-text-field
        v-model="item.icon"
        :label="$t('Icon')"
        :rules="[v => !!v || $t('icon_required')]"
        required
        :disabled="formSaving"
    ></v-text-field>

    <v-text-field
        v-model="item.title"
        :label="$t('name')"
        :rules="[v => !!v || $t('name_required')]"
        required
        :disabled="formSaving"
    ></v-text-field>

    <v-text-field
        v-model="item.path"
        :label="$t('Path')"
        :rules="[v => !!v || $t('path_required')]"
        required
        :disabled="formSaving"
    ></v-text-field>

    <v-checkbox
        v-model="item.active"
        :label="$t('Active')"
    ></v-checkbox>

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],

  computed: {
    isNew() {
      return this.itemId === '';
    },
  },

  data() {
    return {
      id: null,
    };
  },

  watch: {
    itemId() {
      this.id = this.itemId;
    },
  },

  methods: {
    beforeLoadData() {
      if (!this.isNew) {
        this.id = this.itemId;
      }
    },

    afterReset() {
      this.id = null;
    },

    getItemsUrl() {
      return `/api/apps/${this.id}`;
    },

    getSingleItemUrl() {
      return `/api/apps/${this.id}`;
    },
  },
};
</script>
