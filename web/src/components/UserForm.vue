<template>
  <div>
    <EditDialog
      v-model="passwordDialog"
      save-button-text="Save"
      :title="$t('changePassword')"
      v-if="item"
      event-name="i-user"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <ChangePasswordForm
          :project-id="projectId"
          :item-id="item.id"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <v-tabs v-model="tab">
      <v-tab key="settings">Settings</v-tab>
      <v-tab key="2fa" v-if="canChangePassword || authMethods.totp || !isNew"> Security </v-tab>
    </v-tabs>

    <v-divider class="mb-6" style="margin-top: -1px" />

    <v-tabs-items v-model="tab" style="overflow: unset">
      <v-tab-item key="settings">
        <v-form ref="form" lazy-validation v-model="formValid" v-if="item != null">
          <v-alert :value="formError" color="error" class="pb-2">{{ formError }} </v-alert>

          <v-text-field
            v-model="item.name"
            :label="$t('name')"
            :rules="[(v) => !!v || $t('name_required')]"
            required
            :disabled="formSaving"
            outlined
            dense
          ></v-text-field>

          <v-text-field
            v-model="item.username"
            :label="$t('username')"
            :rules="[(v) => !!v || $t('user_name_required')]"
            required
            :disabled="formSaving"
            outlined
            dense
          ></v-text-field>

          <v-text-field
            v-model="item.email"
            :label="$t('email')"
            :rules="[(v) => !!v || $t('email_required')]"
            required
            :disabled="(!isNew && item.external) || formSaving"
            outlined
            dense
          >
            <template v-slot:append>
              <v-chip outlined color="green" disabled small style="opacity: 1">private</v-chip>
            </template>
          </v-text-field>

          <v-text-field
            v-if="isNew"
            v-model="item.password"
            :label="$t('password')"
            class="masked-secret-input"
            :required="isNew && !item.external"
            :rules="isNew && !item.external ? [(v) => !!v || $t('password_required')] : []"
            :disabled="item.external || formSaving"
            outlined
            dense
          ></v-text-field>

          <v-row class="pb-5 pt-2">
            <v-col cols="6">
              <v-checkbox
                dense
                hide-details
                v-model="item.alert"
                :label="$t('sendAlerts')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isAdmin">
              <v-checkbox
                dense
                hide-details
                v-model="item.admin"
                :label="$t('adminUser')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isPro">
              <v-checkbox
                :disabled="!isAdmin"
                dense
                hide-details
                v-model="item.pro"
                :label="$t('Pro user')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isAdmin">
              <v-checkbox
                :disabled="!isNew"
                dense
                hide-details
                v-model="item.external"
                :label="$t('external')"
              ></v-checkbox>
            </v-col>
          </v-row>
        </v-form>
      </v-tab-item>

      <v-tab-item
        key="2fa"
        v-if="item != null && (canChangePassword || authMethods.totp || !isNew)"
      >
        <div v-if="canChangePassword">
          <div class="title mb-3">Password</div>
          <v-btn color="primary" @click="passwordDialog = true">Change password</v-btn>
        </div>

        <div :class="{ 'pt-10': !isNew }" v-if="authMethods.totp">
          <div class="title mb-2">Two-factor authentication</div>

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

          <div
            v-if="authMethods.totp.allow_recovery && item.totp && item.totp.recovery_code"
            class="mt-5 pb-3"
          >
            <div class="subtitle-1 mb-2">Recovery code</div>
            <div style="position: relative">
              <code style="font-size: 18px; background-color: #e03755">
                {{ item.totp.recovery_code }}
              </code>

              <CopyClipboardButton
                style="position: absolute; right: -4px; top: -12px"
                :text="item.totp.recovery_code"
                large
                color="white"
              />
            </div>
          </div>
        </div>

        <div class="" v-if="!isNew">
          <div class="title mb-2">Linked accounts</div>

          <v-alert :value="!!linkError" color="error" dense text>{{ linkError }}</v-alert>

          <v-simple-table v-if="identities.length > 0" class="mb-3">
            <tbody>
              <tr v-for="identity in identities" :key="identity.id">
                <td style="width: 30%">{{ providerName(identity) }}</td>
                <td
                  style="
                    max-width: 200px;
                    overflow: hidden;
                    text-overflow: ellipsis;
                    white-space: nowrap;
                  "
                  :title="identity.external_uid"
                >
                  {{ identity.external_uid }}
                </td>
                <td style="text-align: right">
                  <v-btn icon @click="unlinkIdentity(identity)">
                    <v-icon>mdi-link-off</v-icon>
                  </v-btn>
                </td>
              </tr>
            </tbody>
          </v-simple-table>

          <div v-else class="text--secondary mb-3">No linked accounts.</div>

          <div v-if="isSelf">
            <v-btn
              v-for="provider in unlinkedOidcProviders"
              :key="provider.id"
              :color="provider.color || 'secondary'"
              dark
              class="mr-3 mb-3"
              @click="linkOidcIdentity(provider.id)"
            >
              <v-icon left dark v-if="provider.icon">mdi-{{ provider.icon }}</v-icon>
              Link {{ provider.name || provider.id }}
            </v-btn>

            <div v-if="unlinkedLdapProviders.length > 0" class="mt-3">
              <div class="subtitle-1 mb-2">Link LDAP account</div>

              <v-select
                v-if="unlinkedLdapProviders.length > 1"
                v-model="ldapLinkProvider"
                :items="unlinkedLdapProviders"
                item-text="name"
                item-value="id"
                label="Provider"
                outlined
                dense
              ></v-select>

              <v-text-field
                v-model="ldapUsername"
                label="LDAP username"
                outlined
                dense
              ></v-text-field>

              <v-text-field
                v-model="ldapPassword"
                label="LDAP password"
                type="password"
                autocomplete="new-password"
                outlined
                dense
              ></v-text-field>

              <v-btn
                color="primary"
                :disabled="!ldapUsername || !ldapPassword || linkingLdap"
                :loading="linkingLdap"
                @click="linkLdapIdentity()"
              >
                Link
              </v-btn>
            </div>
          </div>
        </div>
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import EditDialog from '@/components/EditDialog.vue';
import ChangePasswordForm from '@/components/ChangePasswordForm.vue';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';

export default {
  components: { CopyClipboardButton, ChangePasswordForm, EditDialog },
  props: {
    isAdmin: Boolean,
    isSelf: Boolean,
    authMethods: Object,
    LoginWithPassword: Boolean,
  },

  mixins: [ItemFormBase],

  data() {
    return {
      passwordDialog: null,
      totpEnabled: false,
      totpQrUrl: null,

      identities: [],
      oidcProviders: [],
      ldapProviders: [],
      ldapLinkProvider: null,
      ldapUsername: '',
      ldapPassword: '',
      linkingLdap: false,
      linkError: null,

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
          this.item.totp = (
            await axios({
              method: 'post',
              url: `/api/users/${this.itemId}/2fas/totp`,
              responseType: 'json',
            })
          ).data;

          // let baseURI = document.baseURI;
          // if (baseURI.endsWith('/')) {
          //   baseURI = baseURI.substring(0, baseURI.length - 1);
          // }

          this.totpQrUrl = `${document.baseURI}api/users/${this.itemId}/2fas/totp/${this.item.totp.id}/qr`;
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

  computed: {
    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },

    canChangePassword() {
      return !this.isNew && !this.item.external && this.LoginWithPassword;
    },

    linkedProviders() {
      return new Set(this.identities.map((identity) => `${identity.type}:${identity.provider}`));
    },

    unlinkedOidcProviders() {
      return this.oidcProviders.filter((provider) => !this.linkedProviders.has(`oidc:${provider.id}`));
    },

    unlinkedLdapProviders() {
      return this.ldapProviders.filter((provider) => !this.linkedProviders.has(`ldap:${provider.id}`));
    },
  },

  methods: {
    afterLoadData() {
      if (this.item.totp == null) {
        this.totpEnabled = false;
        this.totpQrUrl = null;
      } else {
        this.totpEnabled = true;
        this.totpQrUrl = `${document.baseURI}api/users/${this.itemId}/2fas/totp/${this.item.totp.id}/qr`;
      }

      if (!this.isNew) {
        this.loadIdentities();
        this.loadAuthMetadata();
      }
    },

    async loadIdentities() {
      this.identities = (
        await axios({
          method: 'get',
          url: `/api/users/${this.itemId}/identities`,
          responseType: 'json',
        })
      ).data || [];
    },

    async loadAuthMetadata() {
      const data = (
        await axios({
          method: 'get',
          url: '/api/auth/login',
          responseType: 'json',
        })
      ).data;
      this.oidcProviders = data.oidc_providers || [];
      this.ldapProviders = data.ldap_providers || [];
    },

    providerName(identity) {
      const providers = identity.type === 'ldap' ? this.ldapProviders : this.oidcProviders;
      const provider = providers.find((p) => p.id === identity.provider);
      const name = (provider && provider.name) || identity.provider;
      return `${name} (${identity.type})`;
    },

    linkOidcIdentity(providerId) {
      document.location = `${document.baseURI}api/auth/oidc/${providerId}/login?link=1`;
    },

    async linkLdapIdentity() {
      this.linkError = null;
      this.linkingLdap = true;
      try {
        await axios({
          method: 'post',
          url: '/api/user/identities/ldap',
          responseType: 'json',
          data: {
            username: this.ldapUsername,
            password: this.ldapPassword,
            provider: this.ldapLinkProvider
              || (this.unlinkedLdapProviders[0] && this.unlinkedLdapProviders[0].id)
              || undefined,
          },
        });
        this.ldapUsername = '';
        this.ldapPassword = '';
        this.ldapLinkProvider = null;
        await this.loadIdentities();
      } catch (err) {
        if (err.response && err.response.status === 401) {
          this.linkError = 'Invalid LDAP credentials.';
        } else {
          this.linkError = (err.response
            && err.response.data
            && err.response.data.error) || 'Failed to link LDAP account.';
        }
      } finally {
        this.linkingLdap = false;
      }
    },

    async unlinkIdentity(identity) {
      this.linkError = null;
      try {
        await axios({
          method: 'delete',
          url: `/api/users/${this.itemId}/identities/${identity.type}/${identity.provider}`,
          responseType: 'json',
        });
        await this.loadIdentities();
      } catch (err) {
        this.linkError = (err.response
          && err.response.data
          && err.response.data.error) || 'Failed to unlink account.';
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
