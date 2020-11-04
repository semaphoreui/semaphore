<template>
  <div class="auth">
    <v-dialog
      v-model="forgotPasswordDialog"
      max-width="290"
      persistent
    >
      <v-card>
        <v-card-title class="headline">Forgot password?</v-card-title>

        <v-alert
          :value="forgotPasswordSubmitted"
          color="success"
        >
          Check your inbox.
        </v-alert>

        <v-alert
          :value="forgotPasswordError"
          color="error"
        >
          {{ forgotPasswordError }}
        </v-alert>

        <v-card-text>
          <v-form
            ref="forgotPasswordForm"
            lazy-validation
            v-model="forgotPasswordFormValid"
          >
            <v-text-field
              v-model="email"
              label="Email"
              :rules="emailRules"
              required
              :disabled="forgotPasswordSubmitting"
            ></v-text-field>
          </v-form>
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>

          <v-btn
            color="green darken-1"
            text
            :disabled="forgotPasswordSubmitting"
            @click="forgotPasswordDialog = false"
          >
            Cancel
          </v-btn>

          <v-btn
            color="green darken-1"
            text
            :disabled="forgotPasswordSubmitting"
            @click="submitForgotPassword()"
          >
            Reset Password
          </v-btn>
        </v-card-actions>

      </v-card>
    </v-dialog>

    <v-container
      fluid
      fill-height
      align-center
      justify-center
      class="pa-0"
    >
      <v-form
        ref="signInForm"
        lazy-validation
        v-model="signInFormValid"
        style="width: 300px; height: 300px;"
      >
        <h3 class="text-center mb-8">SEMAPHORE</h3>

        <v-alert
          :value="signInError"
          color="error"
          style="margin-bottom: 20px;"
        >{{ signInError }}</v-alert>

        <v-text-field
          v-model="username"
          label="Username"
          :rules="usernameRules"
          autofocus
          required
          :disabled="signInProcess"
        ></v-text-field>

        <v-text-field
          v-model="password"
          label="Password"
          :rules="[v => !!v || 'Password is required']"
          type="password"
          required
          :disabled="signInProcess"
          @keyup.enter.native="signIn"
          style="margin-bottom: 20px;"
        ></v-text-field>
        <v-btn
          color="primary"
          @click="signIn"
          :disabled="signInProcess"
          block
        >
          Sign In
        </v-btn>
      </v-form>
    </v-container>
  </div>
</template>
<style lang="scss">
.auth {
  height: 100vh;
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
      password: '',
      username: '',

      forgotPasswordFormValid: false,
      forgotPasswordError: false,
      forgotPasswordSubmitted: false,
      forgotPasswordSubmitting: false,
      forgotPasswordDialog: false,
      email: '',

      newPassword: '',
      newPassword2: '',

      emailRules: [
        (v) => !!v || 'Email is required',
      ],
      passwordRules: [
        (v) => !!v || 'Password is required',
        (v) => v.length >= 6 || 'Password too short. Min 6 characters',
      ],
      usernameRules: [
        (v) => !!v || 'Username is required',
      ],
    };
  },

  async created() {
    if (this.isAuthenticated()) {
      EventBus.$emit('i-session-create');
    }
  },

  methods: {
    isAuthenticated() {
      return document.cookie.includes('semaphore=');
    },

    async submitForgotPassword() {
      this.forgotPasswordSubmitted = false;
      this.forgotPasswordError = null;

      if (!this.$refs.forgotPasswordForm.validate()) {
        return;
      }

      this.forgotPasswordSubmitting = true;
      try {
        await axios({
          method: 'post',
          url: '/v1/session/forgot-password',
          data: {
            email: this.email,
          },
        });
        this.forgotPasswordSubmitted = true;
      } catch (err) {
        this.forgotPasswordError = err.response.data.error;
      } finally {
        this.forgotPasswordSubmitting = false;
      }
    },

    async signIn() {
      this.signInError = null;

      if (!this.$refs.signInForm.validate()) {
        return;
      }

      this.signInProcess = true;
      try {
        await axios({
          method: 'post',
          url: '/api/auth/login',
          responseType: 'json',
          data: {
            auth: this.username,
            password: this.password,
          },
        });

        EventBus.$emit('i-session-create');
      } catch (err) {
        this.signInError = getErrorMessage(err);
      } finally {
        this.signInProcess = false;
      }
    },

    forgotPassword() {
      this.forgotPasswordError = null;
      this.forgotPasswordSubmitted = false;
      this.email = '';
      this.$refs.forgotPasswordForm.resetValidation();
      this.forgotPasswordDialog = true;
    },
  },
};
</script>
