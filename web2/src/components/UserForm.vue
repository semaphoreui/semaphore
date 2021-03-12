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
      label="Name"
      :rules="[v => !!v || 'Name is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.username"
      label="Username"
      :rules="[v => !!v || 'Username is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.email"
      label="Email"
      :rules="[v => !!v || 'Email is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.password"
      label="Password"
      type="password"
      :required="isNew"
      :rules="isNew ? [v => !!v || 'Password is required'] : []"
      :disabled="formSaving"
    ></v-text-field>

    <v-checkbox
      v-model="item.admin"
      label="Admin user"
    ></v-checkbox>

    <v-checkbox
      v-model="item.alert"
      label="Send alerts"
    ></v-checkbox>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  methods: {
    getItemsUrl() {
      return '/api/users';
    },

    getSingleItemUrl() {
      return `/api/users/${this.itemId}`;
    },
  },
};
</script>
