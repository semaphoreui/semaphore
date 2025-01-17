<template>
  <div>
    <v-tabs v-model="tab">
      <v-tab key="settings">Settings</v-tab>
      <v-tab key="2fa">2FA</v-tab>
    </v-tabs>

    <v-divider class="mb-6" style="margin-top: -1px;" />

    <v-tabs-items v-model="tab">
      <v-tab-item key="settings">
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
            :disabled="formSaving"
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

          <v-row>
            <v-col>
              <v-checkbox
                v-model="item.admin"
                :label="$t('adminUser')"
                v-if="isAdmin"
              ></v-checkbox>
            </v-col>
            <v-col>
              <v-checkbox
                v-model="item.alert"
                :label="$t('sendAlerts')"
              ></v-checkbox>
            </v-col>
          </v-row>
        </v-form>
      </v-tab-item>
      <v-tab-item key="2fa" class="pb-4">
        <v-switch
          class="mt-0"
          v-model="totpEnabled"
          label="Time-based one-time password"
        ></v-switch>
        <img
          v-if="totpQrUrl"
          :src="totpQrUrl"
          style="
        width: 100%;
        aspect-ratio: 1;
        border-radius: 4px;
        display: block;
        margin: 0 auto 10px auto;
        border: 10px solid white;
        background-color: white;
      "
          alt="QR code"
        />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  props: {
    isAdmin: Boolean,
  },

  mixins: [ItemFormBase],

  data() {
    return {
      totpEnabled: false,
      totpQrUrl: null,

      tab: null,
    };
  },

  watch: {
    tab(value) {
      if (value === 0) {
        this.$emit('show-action-buttons');
      } else {
        this.$emit('hide-action-buttons');
      }
    },

    async totpEnabled(val) {
      if (val) {
        if (this.item.totp == null) {
          this.item.totp = (await axios({
            method: 'post',
            url: `/api/users/${this.itemId}/2fas/totp`,
            responseType: 'json',
          })).data;
          this.totpQrUrl = `/api/users/${this.itemId}/2fas/totp/${this.item.totp.id}/qr`;
        }
      } else if (this.item.totp != null) {
        await axios({
          method: 'delete',
          url: `/api/users/${this.itemId}/2fas/totp/${this.item.totp.id}`,
          responseType: 'json',
        });
        this.item.totp = null;
        this.totpQrUrl = null;
      }
    },
  },

  methods: {
    afterLoadData() {
      if (this.item.totp == null) {
        this.totpEnabled = false;
        this.totpQrUrl = null;
      } else {
        this.totpEnabled = true;
        this.totpQrUrl = `/api/users/${this.itemId}/2fas/totp/${this.item.totp.id}/qr`;
      }
    },

    getItemsUrl() {
      return '/api/users';
    },

    getSingleItemUrl() {
      return `/api/users/${this.itemId}`;
    },
  },
};
</script>
