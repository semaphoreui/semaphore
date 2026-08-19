<template>
  <v-form ref="form" lazy-validation v-model="formValid">
    <v-alert :value="formError" color="error" class="pb-2">{{ formError }}</v-alert>

    <v-text-field
      v-if="isSelf"
      v-model="item.current_password"
      :label="$t('currentPassword')"
      :class="{ 'masked-secret-input': !showCurrentPassword }"
      :append-icon="showCurrentPassword ? 'mdi-eye' : 'mdi-eye-off'"
      @click:append="showCurrentPassword = !showCurrentPassword"
      :rules="[(v) => !!v || $t('password_required')]"
      required
      dense
      outlined
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.password"
      :label="$t('password2')"
      :class="{ 'masked-secret-input': !showPassword }"
      :append-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
      @click:append="showPassword = !showPassword"
      :rules="[(v) => !!v || $t('password_required')]"
      required
      dense
      outlined
      :disabled="formSaving"
    ></v-text-field>
  </v-form>
</template>
<script>
import axios from 'axios';
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      showPassword: false,
      showCurrentPassword: false,
      isSelf: false,
    };
  },

  methods: {
    async loadData() {
      this.item = {};
      // Require the current password only when changing your own (CWE-620).
      const currentUser = (await axios.get('/api/user')).data;
      this.isSelf = currentUser.id === Number(this.itemId);
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
