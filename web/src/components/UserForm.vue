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
      <v-tab key="2fa" v-if="canChangePassword || authMethods.totp || !isNew"> Security</v-tab>
    </v-tabs>

    <v-divider class="mb-6" style="margin-top: -1px" />

    <v-tabs-items v-model="tab" style="overflow: unset">
      <v-tab-item key="settings">
        <v-form ref="form" lazy-validation v-model="formValid" v-if="item != null">
          <v-alert :value="formError" color="error" class="pb-2">{{ formError }}</v-alert>

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

        <div :class="{ 'pt-10': canChangePassword }" v-if="authMethods.totp">
          <div class="title mb-2">Two-factor authentication</div>

          <v-switch
            class="mt-0"
            v-model="totpEnabled"
            label="Time-based one-time password"
          ></v-switch>

          <v-card
            class="pt-2 mt-1"
            style="background: var(--highlighted-card-bg-color)"
            v-if="totpQrUrl"
          >
            <div
              style="
                position: absolute;
                background: var(--highlighted-card-bg-color);
                width: 28px;
                height: 28px;
                transform: rotate(45deg);
                left: calc(50% - 14px);
                top: -14px;
                border-radius: 0;
              "
            ></div>

            <v-card-text>
              <img
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
            </v-card-text>
          </v-card>
        </div>

        <div
          v-if="!isNew"
          :class="{
            'pt-10': canChangePassword || authMethods.totp,
          }"
          :style="{ marginTop: (!authMethods.totp || totpEnabled) ? 0 : '-30px' }"
        >
          <template v-if="identities.length > 0">
            <div class="title mb-2">Linked accounts</div>

            <div style="margin-bottom: -8px;">
              <v-chip
                v-for="identity in identities"
                :key="identity.id"
                :color="identityProvider(identity).color"
                :title="identity.external_uid"
                class="mr-2 mb-2"
                close
                @click:close="unlinkIdentity(identity)"
              >
                <v-icon left v-if="identityProvider(identity).icon">
                  mdi-{{ identityProvider(identity).icon }}
                </v-icon>
                {{ identity.provider }}
              </v-chip>
            </div>
          </template>

          <div
            v-if="isSelf && (unlinkedLdapProviders.length > 0 || unlinkedOidcProviders.length > 0)"
            :class="{ 'pt-10': identities.length > 0 }"
          >
            <div class="title mb-2">Link account</div>

            <v-card class="mt-4" style="background: var(--highlighted-card-bg-color)">
              <v-card-text>
                <div>
                  <v-alert :value="!!linkError" color="error" dense text>{{ linkError }}</v-alert>

                  <v-btn-toggle
                    v-model="ldapLinkProvider"
                    mandatory
                    borderless
                    class="mb-7 d-flex"
                    :background-color="$vuetify.theme.dark ? '#212121' : 'grey lighten-3'"
                  >
                    <v-btn
                      v-for="provider in unlinkedLdapProviders"
                      :key="provider.id"
                      :value="provider.id"
                      small
                      class="flex-grow-1"
                    >
                      {{ provider.name }}
                    </v-btn>
                  </v-btn-toggle>

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
              </v-card-text>
            </v-card>

            <div :class="{ 'pt-5': unlinkedLdapProviders.length > 0 }">
              <v-btn
                v-for="provider in unlinkedOidcProviders"
                :key="provider.id"
                :color="provider.color || 'secondary'"
                dark
                class="mr-3 mb-3"
                rounded
                style="width: 100%"
                width="100%"
                @click="linkOidcIdentity(provider.id)"
              >
                <v-icon left dark v-if="provider.icon">mdi-{{ provider.icon }}</v-icon>
                Link {{ provider.name || provider.id }}
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
      return this.oidcProviders.filter(
        (provider) => !this.linkedProviders.has(`oidc:${provider.id}`),
      );
    },

    unlinkedLdapProviders() {
      return this.ldapProviders.filter(
        (provider) => !this.linkedProviders.has(`ldap:${provider.id}`),
      );
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

    identityProvider(identity) {
      const providers = identity.type === 'ldap' ? this.ldapProviders : this.oidcProviders;
      return providers.find((p) => p.id === identity.provider) || {};
    },

    linkOidcIdentity(providerId) {
      // Top-level form POST: the backend requires POST for link mode so that
      // SameSite=Lax blocks cross-site (CSRF) initiation of account linking.
      const form = document.createElement('form');
      form.method = 'POST';
      form.action = `${document.baseURI}api/auth/oidc/${providerId}/login?link=1`;
      document.body.appendChild(form);
      form.submit();
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
            provider:
              this.ldapLinkProvider
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
          this.linkError = (err.response && err.response.data && err.response.data.error)
            || 'Failed to link LDAP account.';
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
        this.linkError = (err.response && err.response.data && err.response.data.error)
          || 'Failed to unlink account.';
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
