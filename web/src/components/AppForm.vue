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

  watch: {
    itemId(val) {
      this.id = val;
    },
  },

  created() {
    this.id = this.itemId;
  },

  computed: {
    isNew() {
      return false;
    },
  },

  data() {
    return {
      id: '',
    };
  },

  methods: {
    getItemsUrl() {
      return '/api/apps';
    },

    getSingleItemUrl() {
      return `/api/apps/${this.id}`;
    },
  },
};
</script>
