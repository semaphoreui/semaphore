<template>
  <div
    data-testid="user-root"
    :data-state="formError ? 'error' : ((item == null || formSaving) ? 'loading' : 'ready')"
  >
    <EditDialog
      v-model="passwordDialog"
      test-id="changePasswordDialog"
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

    <v-tabs v-model="tab" data-testid="user-tabs">
      <v-tab key="settings" href="#settings" data-testid="user-tab-settings">Settings</v-tab>
      <v-tab
        key="2fa"
        href="#security"
        v-if="canChangePassword || authMethods.totp || !isNew"
        data-testid="user-tab-security"
      >
        Security
      </v-tab>
      <v-tab
        key="global-roles"
        href="#global-roles"
        v-if="canManageGlobalRoles && !isNew"
        data-testid="user-tab-globalRoles"
      >
        {{ $t('globalRolesTab') }}
      </v-tab>
    </v-tabs>

    <v-divider class="mb-6" style="margin-top: -1px" />

    <v-tabs-items v-model="tab" style="overflow: unset">
      <v-tab-item key="settings" value="settings" data-testid="user-panel-settings">
        <v-form
          ref="form"
          lazy-validation
          v-model="formValid"
          v-if="item != null"
          data-testid="user-form"
        >
          <v-alert
            :value="formError"
            color="error"
            class="pb-2"
            data-testid="user-error"
          >{{ formError }}</v-alert>

          <v-text-field
            v-model="item.name"
            data-testid="user-name"
            :label="$t('name')"
            :rules="[(v) => !!v || $t('name_required')]"
            required
            :disabled="formSaving"
            outlined
            dense
          ></v-text-field>

          <v-text-field
            v-model="item.username"
            data-testid="user-username"
            :label="$t('username')"
            :rules="[(v) => !!v || $t('user_name_required')]"
            required
            :disabled="formSaving"
            outlined
            dense
          ></v-text-field>

          <v-text-field
            v-model="item.email"
            data-testid="user-email"
            :label="$t('email')"
            :rules="[(v) => !!v || $t('email_required')]"
            required
            :disabled="(!isNew && item.external) || formSaving"
            outlined
            dense
          >
            <template v-slot:append>
              <v-chip
                outlined
                color="green"
                disabled
                small
                style="opacity: 1"
                data-testid="user-emailPrivate"
              >private</v-chip>
            </template>
          </v-text-field>

          <v-text-field
            v-if="isNew"
            v-model="item.password"
            data-testid="user-password"
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
                data-testid="user-alert"
                :label="$t('sendAlerts')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isAdmin">
              <v-checkbox
                dense
                hide-details
                v-model="item.admin"
                data-testid="user-admin"
                :label="$t('adminUser')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isPro">
              <v-checkbox
                :disabled="!isAdmin"
                dense
                hide-details
                v-model="item.pro"
                data-testid="user-pro"
                :label="$t('Pro user')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isAdmin">
              <v-checkbox
                :disabled="!isNew"
                dense
                hide-details
                v-model="item.external"
                data-testid="user-external"
                :label="$t('external')"
              ></v-checkbox>
            </v-col>
          </v-row>
        </v-form>
      </v-tab-item>

      <v-tab-item
        key="2fa"
        value="security"
        v-if="item != null && (canChangePassword || authMethods.totp || !isNew)"
        data-testid="user-panel-security"
      >
        <div v-if="canChangePassword">
          <div class="title mb-3" data-testid="user-passwordTitle">Password</div>
          <v-btn
            color="primary"
            @click="passwordDialog = true"
            data-testid="user-changePassword"
          >Change password</v-btn>
        </div>

        <div :class="{ 'pt-10': canChangePassword }" v-if="authMethods.totp">
          <div class="title mb-2" data-testid="user-totpTitle">Two-factor authentication</div>

          <v-switch
            class="mt-0"
            v-model="totpEnabled"
            label="Time-based one-time password"
            data-testid="user-totp"
          ></v-switch>

          <v-card
            class="pt-2 mt-1"
            style="background: var(--highlighted-card-bg-color)"
            v-if="totpQrUrl"
            data-testid="user-totpCard"
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
                data-testid="user-totpQr"
              />

              <div
                v-if="authMethods.totp.allow_recovery && item.totp && item.totp.recovery_code"
                class="mt-5 pb-3"
                data-testid="user-recovery"
              >
                <div class="subtitle-1 mb-2" data-testid="user-recoveryTitle">Recovery code</div>
                <div style="position: relative">
                  <code
                    style="font-size: 18px; background-color: #e03755"
                    data-testid="user-recoveryCode"
                  >
                    {{ item.totp.recovery_code }}
                  </code>

                  <CopyClipboardButton
                    style="position: absolute; right: -4px; top: -12px"
                    :text="item.totp.recovery_code"
                    large
                    color="white"
                    test-id="user-recoveryCodeCopy"
                  />
                </div>
              </div>
            </v-card-text>
          </v-card>
        </div>

        <div
          v-if="!isNew"
          data-testid="user-identities"
          :class="{
            'pt-10': canChangePassword || authMethods.totp,
          }"
          :style="{ marginTop: (!authMethods.totp || totpEnabled) ? 0 : '-30px' }"
        >
          <template v-if="identities.length > 0">
            <div class="title mb-2" data-testid="user-identitiesTitle">Linked accounts</div>

            <div style="margin-bottom: -8px;">
              <v-chip
                v-for="identity in identities"
                :key="identity.id"
                :color="identityProvider(identity).color"
                :title="identity.external_uid"
                class="mr-2 mb-2"
                close
                @click:close="unlinkIdentity(identity)"
                :data-testid="`user-identity-${identity.type}-${identity.provider}`"
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
            data-testid="user-link"
          >
            <div class="title mb-2" data-testid="user-linkTitle">Link account</div>

            <v-card
              class="mt-4"
              style="background: var(--highlighted-card-bg-color)"
              data-testid="user-linkCard"
            >
              <v-card-text>
                <div>
                  <v-alert
                    :value="!!linkError"
                    color="error"
                    dense
                    text
                    data-testid="user-linkError"
                  >{{ linkError }}</v-alert>

                  <v-btn-toggle
                    v-model="ldapLinkProvider"
                    mandatory
                    borderless
                    class="mb-7 d-flex"
                    :background-color="$vuetify.theme.dark ? '#212121' : 'grey lighten-3'"
                    data-testid="user-ldapProvider"
                  >
                    <v-btn
                      v-for="provider in unlinkedLdapProviders"
                      :key="provider.id"
                      :value="provider.id"
                      small
                      class="flex-grow-1"
                      :data-testid="`user-ldapProvider-${provider.id}`"
                    >
                      {{ provider.name }}
                    </v-btn>
                  </v-btn-toggle>

                  <v-text-field
                    v-model="ldapUsername"
                    label="LDAP username"
                    data-testid="user-ldapUsername"
                    outlined
                    dense
                  ></v-text-field>

                  <v-text-field
                    v-model="ldapPassword"
                    label="LDAP password"
                    data-testid="user-ldapPassword"
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
                    data-testid="user-ldapLink"
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
                :data-testid="`user-oidcLink-${provider.id}`"
              >
                <v-icon left dark v-if="provider.icon">mdi-{{ provider.icon }}</v-icon>
                Link {{ provider.name || provider.id }}
              </v-btn>
            </div>
          </div>
        </div>
      </v-tab-item>

      <v-tab-item
        key="global-roles"
        value="global-roles"
        v-if="item != null && canManageGlobalRoles && !isNew"
        data-testid="user-panel-globalRoles"
      >
        <GlobalRoleAssignments :user-id="itemId" />
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
import GlobalRoleAssignments from '@/components/GlobalRoleAssignments.vue';

export default {
  components: {
    CopyClipboardButton, ChangePasswordForm, EditDialog, GlobalRoleAssignments,
  },
  props: {
    isAdmin: Boolean,
    isSelf: Boolean,
    // The signed-in user may assign global roles; the tab is hidden otherwise.
    canManageGlobalRoles: Boolean,
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
      if (value === 'settings') {
        this.$emit('show-action-buttons');
      } else {
        this.$emit('hide-action-buttons');
      }
    },

    // The dialog keeps the form open after a user is created: load the user and go to the roles.
    async itemId(value, previous) {
      if (previous !== 'new' || value === 'new') {
        return;
      }
      await this.loadData();
      if (this.canManageGlobalRoles) {
        this.tab = 'global-roles';
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
