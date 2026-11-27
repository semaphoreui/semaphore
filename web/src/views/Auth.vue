<template>
  <div
    class="auth"
    data-testid="auth-page"
    :data-state="loginWithPassword === null || signInProcess
      ? 'loading' : (signInError != null ? 'error' : 'ready')"
  >
    <v-dialog v-model="loginHelpDialog" max-width="600">
      <v-card data-testid="authHelpDialog">
        <v-card-title>
          {{ $t('howToFixSigninIssues') }}
          <v-spacer></v-spacer>
          <v-btn
            icon
            variant="text"
            @click="loginHelpDialog = false"
            data-testid="authHelpDialog-closeIcon"
          >
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text>
          <p class="text-body-large">
            {{ $t('firstlyYouNeedAccessToTheServerWhereSemaphoreRunni') }}
          </p>
          <p class="text-body-large">
            {{ $t('executeTheFollowingCommandOnTheServerToSeeExisting') }}
          </p>
          <v-alert
            density="compact"
            variant="tonal"
            color="info"
            style="font-family: monospace"
            data-testid="authHelpDialog-userListCmd"
          >
            {{ $t('semaphoreUserList') }}
          </v-alert>
          <p class="text-body-large">
            {{ $t('youCanChangePasswordOfExistingUser') }}
          </p>
          <v-alert
            density="compact"
            variant="tonal"
            color="info"
            style="font-family: monospace"
            data-testid="authHelpDialog-changePasswordCmd"
          >
            {{
              $t('semaphoreUserChangebyloginLoginUser123Password', {
                makePasswordExample: makePasswordExample(),
              })
            }}
          </v-alert>
          <p class="text-body-large">
            {{ $t('orCreateNewAdminUser') }}
          </p>
          <v-alert
            density="compact"
            variant="tonal"
            color="info"
            style="font-family: monospace"
            data-testid="authHelpDialog-addAdminCmd"
          >
            semaphore user add --admin --login user123 --name User123 --email user123@example.com
            --password {{ makePasswordExample() }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            color="blue-darken-1"
            variant="text"
            @click="loginHelpDialog = false"
            data-testid="authHelpDialog-close"
          >
            {{ $t('close2') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-container fluid class="pa-0 fill-height d-flex align-center justify-center">
      <v-card class="px-5 py-5" style="border-radius: 15px" data-testid="auth-card">
        <v-card-text>
          <v-form
            @submit.prevent
            ref="signInForm"
            v-model="signInFormValid"
            style="width: 350px; max-width: 100%;"
            data-testid="auth-form"
          >
            <v-img
              width="80"
              height="80"
              :transition="false"
              src="favicon.png"
              class="mx-auto mb-4"
              data-testid="auth-logo"
            />

            <h2
              v-if="screen === 'verification'"
              class="text-center pt-4 pb-6"
              data-testid="auth-title-verification"
            >
              Two-step verification
            </h2>

            <h2
              v-else-if="screen === 'recovery'"
              class="text-center pt-4 pb-6"
              data-testid="auth-title-recovery"
            >
              Account recovery
            </h2>

            <h2
              v-else
              class="text-center pt-4 pb-6"
              data-testid="auth-title-login"
            >Enter to your account</h2>

            <v-alert
              :model-value="signInError != null"
              color="error"
              style="margin-bottom: 20px"
              data-testid="auth-error"
              >{{ signInError }}
            </v-alert>

            <div v-if="screen === 'verification'" data-testid="auth-verification">
              <div
                v-if="verificationMethod === 'totp'"
                class="text-center mb-4"
                data-testid="auth-verificationHint-totp"
              >
                Open the two-step verification app on your mobile device to get your verification
                code.
              </div>

              <div
                v-else-if="isPortal && verificationMethod === 'email'"
                class="text-center mb-4"
                data-testid="auth-verificationHint-email"
              >
                Check your email for the verification code we just sent you.
              </div>

              <v-otp-input
                v-model="verificationCode"
                length="6"
                @finish="verify()"
                data-testid="auth-otp"
              ></v-otp-input>

              <v-divider class="my-6" />

              <div class="text-center">
                <a
                  @click="signOut()"
                  class="mr-6 text-primary cursor-pointer"
                  data-testid="auth-returnToLogin"
                >{{ $t('Return to login') }}</a>
                <a
                  v-if="
                    verificationMethod === 'totp' &&
                    authMethods.totp &&
                    authMethods.totp.allow_recovery
                  "
                  @click="screen = 'recovery'"
                  class="text-primary cursor-pointer"
                  data-testid="auth-useRecoveryCode"
                >
                  {{ $t('Use recovery code') }}
                </a>

                <v-btn
                  :width="200"
                  size="small"
                  :disabled="verificationEmailSending"
                  color="primary"
                  v-if="isPortal && verificationMethod === 'email'"
                  @click="resendEmailVerification()"
                  data-testid="auth-resendEmail"
                >
                  {{
                    verificationEmailSending ? $t('Email sending...') : $t('Resend code to email')
                  }}
                </v-btn>
              </div>
            </div>

            <div v-else-if="screen === 'recovery'" data-testid="auth-recovery">
              <div class="text-center mb-2">
                Use your recovery code to regain access to your account.
              </div>

              <v-text-field
                class="mt-6"
                variant="outlined"
                v-model="recoveryCode"
                @keyup.enter="signIn"
                :label="$t('Recovery code')"
                :rules="[(v) => !!v || $t('recoveryCode_required')]"
                required
                data-testid="auth-recoveryCode"
              />

              <div>
                <v-btn
                  style="width: 100%"
                  color="primary"
                  @click="recovery()"
                  data-testid="auth-recoverySend"
                > Send </v-btn>
              </div>

              <div class="text-center pt-6">
                <a
                  @click="screen = 'verification'"
                  class="text-primary cursor-pointer"
                  data-testid="auth-returnToVerification"
                >{{ $t('Return to verification') }}</a>
              </div>
            </div>

            <div v-else data-testid="auth-login">
              <v-btn-toggle
                v-if="loginTabs.length > 1"
                v-model="loginTab"
                mandatory
                variant="text"
                class="auth__tabs mb-7 d-flex"
                :class="$vuetify.theme.current.dark ? null : 'bg-grey-lighten-3'"
                :style="$vuetify.theme.current.dark ? { backgroundColor: '#212121' } : null"
                data-testid="auth-loginTabs"
              >
                <v-btn
                  v-for="tab in loginTabs"
                  :key="tab.id || 'local'"
                  size="small"
                  class="flex-grow-1"
                  :data-testid="`auth-loginTab-${tab.id || 'local'}`"
                >
                  {{ tab.name }}
                </v-btn>
              </v-btn-toggle>

              <div
                v-if="(activeLoginTab && activeLoginTab.ldap) || loginWithPassword"
                data-testid="auth-passwordLogin"
              >
                <v-text-field
                  v-model="username"
                  v-bind:label="$t('username')"
                  :rules="[(v) => !!v || $t('username_required')]"
                  required
                  variant="outlined"
                  :disabled="signInProcess"
                  id="auth-username"
                  density="compact"
                  data-testid="auth-username"
                ></v-text-field>

                <v-text-field
                  v-model="password"
                  :label="$t('password')"
                  :rules="[(v) => !!v || $t('password_required')]"
                  type="password"
                  required
                  variant="outlined"
                  density="compact"
                  :disabled="signInProcess"
                  @keyup.enter="signIn"
                  id="auth-password"
                  data-testid="auth-password"
                ></v-text-field>

                <v-btn
                  size="large"
                  color="primary"
                  @click="signIn"
                  :disabled="signInProcess"
                  block
                  rounded
                  data-testid="auth-signin"
                >
                  {{ $t('signIn') }}
                </v-btn>
              </div>

              <div v-else-if="isPortal" data-testid="auth-emailLogin">
                <v-text-field
                  v-model="email"
                  :label="$t('Email')"
                  :rules="[(v) => !!v || $t('email_required')]"
                  type="email"
                  required
                  :disabled="signInProcess"
                  @keyup.enter="signInWithEmail"
                  data-testid="auth-password"
                  variant="outlined"
                  class="mb-0"
                ></v-text-field>

                <v-btn
                  size="large"
                  color="primary"
                  @click="signInWithEmail"
                  :disabled="signInProcess"
                  block
                  rounded
                  data-testid="auth-signin-with-eamil"
                >
                  <v-icon start> mdi-email </v-icon>

                  {{ $t('Continue with Email') }}
                </v-btn>
              </div>

              <div
                class="auth__divider"
                data-testid="auth-divider"
                v-if="
                  (loginWithPassword || ldapProviders.length > 0 || isPortal) &&
                  oidcProviders.length > 0
                "
              >
                or
              </div>

              <v-btn
                size="large"
                v-for="provider in oidcProviders"
                :color="provider.color || 'secondary'"
                class="mt-3"
                @click="oidcSignIn(provider.id)"
                block
                :key="provider.id"
                rounded
                :data-testid="`auth-oidc-${provider.id}`"
              >
                <v-icon start v-if="provider.icon"> mdi-{{ provider.icon }} </v-icon>

                {{ provider.name }}
              </v-btn>

              <div class="text-center mt-6" v-if="loginWithPassword && false">
                <a
                  @click="loginHelpDialog = true"
                  class="text-primary cursor-pointer"
                  data-testid="auth-help"
                >{{ $t('dontHaveAccountOrCantSignIn') }}</a>
              </div>
            </div>
          </v-form>
        </v-card-text>
      </v-card>
    </v-container>
  </div>
</template>
<style lang="scss">
.auth__divider {
  margin-top: 15px;
  margin-bottom: 5px;

  display: flex;
  &:before,
  &:after {
    margin-top: 10px;
    width: 100%;
    content: '';
    border-top: 1px solid rgba(128, 128, 128, 0.51);
  }

  &:before {
    margin-right: 10px;
  }

  &:after {
    margin-left: 10px;
  }
}
// Sign-in / LDAP tabs as the borderless `v-btn-toggle` of Vuetify 2 with small buttons: 28 px
// high, text in the primary text colour (.87), buttons without their own background on the grey
// of the toggle, the inactive at .8 opacity, the active darkened by the .18 overlay of
// `.v-btn--active`.
.auth__tabs.v-btn-group {
  height: 28px;
  color: rgba(var(--v-theme-on-surface), var(--v-high-emphasis-opacity));

  > .v-btn {
    opacity: .8;
  }

  > .v-btn--active {
    opacity: 1;

    > .v-btn__overlay {
      opacity: .18;
    }
  }
}

.auth {
  // The large buttons (sign in, OIDC) in the type size of Vuetify 2 (`large` .875rem, as
  // `default`; Vuetify 4 grows it to 1rem). Only this screen has large buttons.
  .v-btn--size-large {
    --v-btn-size: .875rem;
  }

  // The OTP cells as in Vuetify 2: outlined fields 56 px high across the whole form width,
  // 8 px apart, a 4 px gutter above and 12 px below, digits in the field type size (1rem, .87).
  .v-otp-input {
    padding: 4px 0 12px;
  }

  .v-otp-input__content {
    max-width: none;
    height: 56px;
    padding: 0;
  }

  .v-otp-input__field {
    font-size: 1rem;
    opacity: var(--v-high-emphasis-opacity);
  }
}

.auth {
  height: 100dvh;
  background: #80808024;
}
.auth {
  background-image: url('../assets/background.svg');
  background-color: #005057;
  background-size: cover;
}
</style>
<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';
import EventBus from '@/event-bus';

export default {
  data() {
    return {
      signInFormValid: false,
      signInError: null,
      signInProcess: false,

      password: null,
      username: null,

      email: null,

      loginHelpDialog: null,

      oidcProviders: [],
      loginWithPassword: null,
      authMethods: {},

      ldapProviders: [],
      loginTab: 0,

      screen: null,

      verificationCode: null,
      verificationMethod: null,
      recoveryCode: null,
      verificationEmailSending: false,
    };
  },

  async created() {
    const { status, verificationMethod } = await this.getAuthenticationStatus();

    switch (status) {
      case 'authenticated':
        this.redirectAfterLogin();
        break;
      case 'unauthenticated':
        await this.loadLoginData();
        break;
      case 'unverified':
        this.screen = 'verification';
        this.verificationMethod = verificationMethod;
        await this.loadLoginData();
        break;
      default:
        throw new Error(`Unknown authentication status: ${status}`);
    }
  },

  computed: {
    isPortal() {
      return process.env.VUE_APP_BUILD_TYPE === 'pro_portal';
    },

    loginTabs() {
      const tabs = [];
      if (this.loginWithPassword || this.isPortal) {
        tabs.push({ id: null, name: this.$t('signIn') });
      }
      this.ldapProviders.forEach((p) => tabs.push({ id: p.id, name: p.name, ldap: true }));
      return tabs;
    },

    activeLoginTab() {
      return this.loginTabs[this.loginTab] || this.loginTabs[0];
    },
  },

  methods: {
    async resendEmailVerification() {
      if (this.verificationEmailSending) {
        return;
      }

      this.verificationEmailSending = true;
      try {
        await axios({
          method: 'post',
          url: '/api/auth/login/email/resend',
          responseType: 'json',
        });
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'Verification email sent successfully.',
        });
      } catch (e) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(e),
        });
      } finally {
        this.verificationEmailSending = false;
      }
    },

    async loadLoginData() {
      await axios({
        method: 'get',
        url: '/api/auth/login',
        responseType: 'json',
      }).then((resp) => {
        this.oidcProviders = resp.data.oidc_providers;
        this.loginWithPassword = resp.data.login_with_password;
        this.authMethods = resp.data.auth_methods || {};
        this.ldapProviders = resp.data.ldap_providers || [];
      });
    },

    async recovery() {
      this.signInProcess = true;

      try {
        await axios({
          method: 'post',
          url: '/api/auth/recovery',
          responseType: 'json',
          data: {
            recovery_code: this.recoveryCode,
          },
        });

        const { location } = document;
        document.location = location;
      } catch (e) {
        this.signInError = getErrorMessage(e);
      } finally {
        this.signInProcess = false;
      }
    },

    async signOut() {
      try {
        await axios({
          method: 'post',
          url: '/api/auth/logout',
          responseType: 'json',
        });

        const { location } = document;
        document.location = location;
      } catch (e) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(e),
        });
      }
    },

    makePasswordExample() {
      let pwd = '';
      const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
      const charactersLength = characters.length;
      for (let i = 0; i < 10; i += 1) {
        pwd += characters.charAt(Math.floor(Math.random() * charactersLength));
      }
      return pwd;
    },

    async getAuthenticationStatus() {
      try {
        await axios({
          method: 'get',
          url: '/api/user',
          responseType: 'json',
        });
      } catch (err) {
        if (err.response.status === 401) {
          switch (err.response.data.error) {
            case 'TOTP_REQUIRED':
              return {
                status: 'unverified',
                verificationMethod: 'totp',
              };
            case 'EMAIL_OTP_REQUIRED':
              return {
                status: 'unverified',
                verificationMethod: 'email',
              };
            default:
              return { status: 'unauthenticated' };
          }
        }
        throw err;
      }

      return { status: 'authenticated' };
    },

    async verify() {
      this.signInError = null;

      const { valid } = await this.$refs.signInForm.validate();
      if (!valid) {
        return;
      }

      this.signInProcess = true;

      try {
        await axios({
          method: 'post',
          url: '/api/auth/verify',
          responseType: 'json',
          data: {
            passcode: this.verificationCode,
          },
        });

        this.redirectAfterLogin();
      } catch (err) {
        this.signInError = getErrorMessage(err);
      } finally {
        this.signInProcess = false;
      }
    },

    async signInWithEmail() {
      this.signInError = null;

      const { valid } = await this.$refs.signInForm.validate();
      if (!valid) {
        return;
      }

      this.signInProcess = true;
      try {
        await axios({
          method: 'post',
          url: '/api/auth/login/email',
          responseType: 'json',
          data: {
            email: this.email,
          },
        });

        this.redirectAfterLogin();
      } catch (err) {
        if (err.response.status === 401) {
          this.signInError = this.$t('incorrectEmail');
        } else {
          this.signInError = getErrorMessage(err);
        }
      } finally {
        this.signInProcess = false;
      }
    },

    async signIn() {
      this.signInError = null;

      const { valid } = await this.$refs.signInForm.validate();
      if (!valid) {
        return;
      }

      this.signInProcess = true;
      try {
        const body = {
          auth: this.username,
          password: this.password,
        };
        if (this.activeLoginTab && this.activeLoginTab.ldap) {
          body.method = 'ldap';
          body.provider = this.activeLoginTab.id;
        } else if (this.loginTabs.length > 1) {
          // Internal tab explicitly requests password auth so a typo never
          // hits the LDAP directory (legacy single-form behavior).
          body.method = 'password';
        }

        await axios({
          method: 'post',
          url: '/api/auth/login',
          responseType: 'json',
          data: body,
        });

        this.redirectAfterLogin();
        // document.location = document.baseURI + window.location.search;
      } catch (err) {
        if (err.response.status === 401) {
          this.signInError = this.$t('incorrectUsrPwd');
        } else {
          this.signInError = getErrorMessage(err);
        }
      } finally {
        this.signInProcess = false;
      }
    },

    async oidcSignIn(provider) {
      const params = new URLSearchParams();
      const returnTo = this.$route.query.return;
      if (returnTo) {
        params.set('return', returnTo);
      } else if (this.$route.query.new_project === 'premium') {
        params.set('return', '/project/premium');
      }
      const qs = params.toString();
      const suffix = qs ? `?${qs}` : '';
      document.location = `${document.baseURI}api/auth/oidc/${provider}/login${suffix}`;
    },

    redirectAfterLogin() {
      const redirectTo = this.$route.query.return;
      let baseURI = document.baseURI;

      if (redirectTo) {
        if (baseURI.endsWith('/')) {
          baseURI = baseURI.substring(0, baseURI.length - 1);
        }

        document.location = baseURI + redirectTo;

        return;
      }

      document.location = document.baseURI + window.location.search;
    },
  },
};
</script>
