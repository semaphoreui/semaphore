<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.password"
      :label="$t('password2')"
      :type="showPassword ? 'text' : 'password'"
      :append-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
      @click:append="showPassword = !showPassword"
      :rules="[v => !!v || $t('password_required')]"
      required
      :disabled="formSaving"
    ></v-text-field>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      showPassword: false,
    };
  },

  methods: {
    async loadData() {
      this.item = {};
    },

    getItemsUrl() {
      return null;
    },

    getSingleItemUrl() {
      return null;
    },

    getRequestOptions() {
      return {
        method: 'post',
        url: `/api/users/${this.itemId}/password`,
      };
    },
  },
};
</script>
