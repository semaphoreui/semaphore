<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
  >
    <v-alert
      :value="formError"
      :color="(formError || '').includes('already activated') ? 'warning' : 'error'"
    >{{ formError }}
    </v-alert>

    <v-text-field
      v-model="item.key"
      label="Subscription Key"
      :rules="[v => !!v || $t('key_required')]"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-list v-if="item.plan">
      <v-list-item class="pa-0">
        <v-list-item-content>
          <v-list-item-title>Plan</v-list-item-title>
          <v-list-item-subtitle>{{ item.plan }}</v-list-item-subtitle>
        </v-list-item-content>
      </v-list-item>
      <v-list-item class="pa-0">
        <v-list-item-content>
          <v-list-item-title>Expires at</v-list-item-title>
          <v-list-item-subtitle>{{ item.expiresAt }}</v-list-item-subtitle>
        </v-list-item-content>
      </v-list-item>
      <v-list-item class="pa-0">
        <v-list-item-content>
          <v-list-item-title>Users</v-list-item-title>
          <v-list-item-subtitle>{{ item.users }}</v-list-item-subtitle>
        </v-list-item-content>
      </v-list-item>
    </v-list>
    <v-alert color="info" v-else>There is no active subscription.</v-alert>

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],

  data() {
    return {};
  },

  methods: {
    async loadData() {
      this.item = (await axios({
        method: 'get',
        url: '/api/subscription',
        responseType: 'json',
      })).data;
    },

    async afterSave() {
      await this.loadData();
    },

    getItemsUrl() {
      return '/api/subscription';
    },

    getSingleItemUrl() {
      return '/api/subscription';
    },

    getRequestOptions() {
      return {
        method: 'post',
        url: '/api/subscription',
      };
    },
  },
};
</script>
