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
      <v-tab key="settings" value="settings" data-testid="user-tab-settings">Settings</v-tab>
      <v-tab
        key="2fa"
        value="security"
        v-if="canChangePassword || authMethods.totp || !isNew"
        data-testid="user-tab-security"
      >
        Security
      </v-tab>
      <v-tab
        key="global-roles"
        value="global-roles"
        v-if="canManageGlobalRoles && !isNew"
        data-testid="user-tab-globalRoles"
      >
        {{ $t('globalRolesTab') }}
      </v-tab>
    </v-tabs>

    <v-divider class="mb-6" style="margin-top: -1px" />

    <v-window v-model="tab" style="overflow: unset">
      <v-window-item value="settings" data-testid="user-panel-settings">
        <v-form
          ref="form"
          v-model="formValid"
          v-if="item != null"
          data-testid="user-form"
        >
          <v-alert
            :model-value="!!formError"
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
            variant="outlined"
            density="compact"
          ></v-text-field>

          <v-text-field
            v-model="item.username"
            data-testid="user-username"
            :label="$t('username')"
            :rules="[(v) => !!v || $t('user_name_required')]"
            required
            :disabled="formSaving"
            variant="outlined"
            density="compact"
          ></v-text-field>

          <v-text-field
            v-model="item.email"
            data-testid="user-email"
            :label="$t('email')"
            :rules="[(v) => !!v || $t('email_required')]"
            required
            :disabled="(!isNew && item.external) || formSaving"
            variant="outlined"
            density="compact"
          >
            <template v-slot:append-inner>
              <v-chip
                variant="outlined"
                color="green"
                disabled
                size="small"
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
            variant="outlined"
            density="compact"
          ></v-text-field>

          <v-row class="pb-5 pt-2 UserForm__checks">
            <v-col cols="6">
              <v-checkbox
                density="compact"
                hide-details
                v-model="item.alert"
                data-testid="user-alert"
                :label="$t('sendAlerts')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isAdmin">
              <v-checkbox
                density="compact"
                hide-details
                v-model="item.admin"
                data-testid="user-admin"
                :label="$t('adminUser')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isPro">
              <v-checkbox
                :disabled="!isAdmin"
                density="compact"
                hide-details
                v-model="item.pro"
                data-testid="user-pro"
                :label="$t('Pro user')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isAdmin">
              <v-checkbox
                :disabled="!isNew"
                density="compact"
                hide-details
                v-model="item.external"
                data-testid="user-external"
                :label="$t('external')"
              ></v-checkbox>
            </v-col>
          </v-row>
        </v-form>
      </v-window-item>

      <v-window-item
        value="security"
        v-if="item != null && (canChangePassword || authMethods.totp || !isNew)"
        data-testid="user-panel-security"
      >
        <div v-if="canChangePassword">
          <div class="text-headline-small mb-3" data-testid="user-passwordTitle">Password</div>
          <v-btn
            color="primary"
            @click="passwordDialog = true"
            data-testid="user-changePassword"
          >Change password</v-btn>
        </div>

        <div :class="{ 'pt-10': canChangePassword }" v-if="authMethods.totp">
          <div
            class="text-headline-small mb-2"
            data-testid="user-totpTitle"
          >Two-factor authentication</div>

          <v-switch
            class="mt-0"
            v-model="totpEnabled"
            label="Time-based one-time password"
            data-testid="user-totp"
          ></v-switch>

          <v-card
            class="pt-2 mt-1"
            style="background: var(--highlighted-card-bg-color); overflow: visible"
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
                <div
                  class="mb-2 UserForm__subtitle"
                  data-testid="user-recoveryTitle"
                >Recovery code</div>
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
            <div
              class="text-headline-small mb-2"
              data-testid="user-identitiesTitle"
            >Linked accounts</div>

            <div style="margin-bottom: -8px;">
              <!-- A closable chip hides itself on close; here the identities list decides. -->
              <v-chip
                v-for="identity in identities"
                :key="identity.id"
                :color="identityProvider(identity).color"
                :title="identity.external_uid"
                class="mr-2 mb-2 UserForm__identity"
                :model-value="true"
                @update:model-value="() => {}"
                closable
                @click:close="unlinkIdentity(identity)"
                :data-testid="`user-identity-${identity.type}-${identity.provider}`"
              >
                <v-icon start v-if="identityProvider(identity).icon">
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
            <div class="text-headline-small mb-2" data-testid="user-linkTitle">Link account</div>

            <v-card
              class="mt-4"
              style="background: var(--highlighted-card-bg-color)"
              data-testid="user-linkCard"
            >
              <v-card-text>
                <div>
                  <v-alert
                    :model-value="!!linkError"
                    color="error"
                    density="compact"
                    variant="tonal"
                    data-testid="user-linkError"
                  >{{ linkError }}</v-alert>

                  <v-btn-toggle
                    v-model="ldapLinkProvider"
                    mandatory="force"
                    class="mb-7 d-flex UserForm__ldapProviders"
                    :class="$vuetify.theme.current.dark ? '' : 'bg-grey-lighten-3'"
                    :style="$vuetify.theme.current.dark ? { backgroundColor: '#212121' } : null"
                    data-testid="user-ldapProvider"
                  >
                    <v-btn
                      v-for="provider in unlinkedLdapProviders"
                      :key="provider.id"
                      :value="provider.id"
                      size="small"
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
                    variant="outlined"
                    density="compact"
                  ></v-text-field>

                  <v-text-field
                    v-model="ldapPassword"
                    label="LDAP password"
                    data-testid="user-ldapPassword"
                    type="password"
                    autocomplete="new-password"
                    variant="outlined"
                    density="compact"
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
                class="mr-3 mb-3"
                rounded
                style="width: 100%"
                width="100%"
                @click="linkOidcIdentity(provider.id)"
                :data-testid="`user-oidcLink-${provider.id}`"
              >
                <v-icon start v-if="provider.icon">mdi-{{ provider.icon }}</v-icon>
                Link {{ provider.name || provider.id }}
              </v-btn>
            </div>
          </div>
        </div>
      </v-window-item>

      <v-window-item
        value="global-roles"
        v-if="item != null && canManageGlobalRoles && !isNew"
        data-testid="user-panel-globalRoles"
      >
        <GlobalRoleAssignments :user-id="itemId" />
      </v-window-item>
    </v-window>
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
    loginWithPassword: Boolean,
  },

  mixins: [ItemFormBase],

  emits: ['show-action-buttons', 'hide-action-buttons'],

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
      return !this.isNew && !this.item.external && this.loginWithPassword;
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
<style lang="scss" scoped>
// `subtitle-1` of Vuetify 2 (the codemod made it `text-body-large`, the body-1 of Vuetify 2,
// whose utility would win over this layer): 16 px on a 28 px line, letter-spacing .009375em.
.UserForm__subtitle {
  font-size: 1rem;
  font-weight: 400;
  line-height: 1.75rem;
  letter-spacing: .009375em;
}

// The dense check boxes of the settings (`v-checkbox dense`): a 20 px box icon in the 24 px
// control; a disabled one in the disabled text colour (.38), icon and label, under the .5 of a
// disabled input (App.vue), as in Vuetify 2 — not the extra .38 of the Vuetify 4 control.
.UserForm__checks :deep(.v-input--density-compact) {
  .v-selection-control__input > .v-icon {
    font-size: 20px;
  }

  &.v-input--disabled {
    .v-selection-control--disabled {
      opacity: 1;
    }

    .v-selection-control__input > .v-icon,
    .v-label {
      color: rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
    }
  }
}

// The provider icon of a linked account is a 24 px icon (`.v-chip .v-icon`), not scaled down to
// the chip text as in Vuetify 4.
.UserForm__identity .v-icon--start {
  font-size: 24px;
}

// The LDAP provider switch of Vuetify 2 (`v-btn-toggle borderless` with a background colour and
// small buttons): as high as its 28 px buttons, the buttons without a frame or a background of
// their own (the grey of the group shows through), the selected one darkened by the active
// overlay (`.v-btn--active::before`: .18).
.UserForm__ldapProviders {
  height: 28px;

  > .v-btn:not(.v-btn--variant-text):not(.v-btn--variant-plain) {
    border-width: 0;
    background-color: transparent;

    &.v-btn--active > :deep(.v-btn__overlay) {
      opacity: .18;
    }
  }
}
</style>
