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
      v-model="item.username"
      :label="$t('username')"
      :rules="[v => !!v || $t('user_name_required')]"
      required
      :disabled="item.external || formSaving"
    ></v-text-field>

    <v-text-field
      v-model="item.email"
      :label="$t('email')"
      :rules="[v => !!v || $t('email_required')]"
      required
      :disabled="item.external || formSaving"
    >

      <template v-slot:append>
        <v-chip outlined color="green" disabled small style="opacity: 1">private</v-chip>
      </template>
    </v-text-field>

    <v-text-field
      v-model="item.password"
      :label="$t('password')"
      type="password"
      :required="isNew"
      :rules="isNew ? [v => !!v || $t('password_required')] : []"
      :disabled="item.external || formSaving"
    ></v-text-field>

    <v-checkbox
      v-model="item.admin"
      :label="$t('adminUser')"
      v-if="isAdmin"
    ></v-checkbox>

    <v-checkbox
      v-model="item.alert"
      :label="$t('sendAlerts')"
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
      return '/api/users';
    },

    getSingleItemUrl() {
      return `/api/users/${this.itemId}`;
    },
  },
};
</script>
