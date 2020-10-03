<template>
  <v-container
    fluid
    fill-height
    align-center
    justify-center
    class="pa-0"
  >
    <v-form
      ref="changePasswordForm"
      lazy-validation
      v-model="changePasswordFormValid"
      style="width: 300px;"
    >

      <v-alert
        :value="changePasswordError"
        color="error"
        style="margin-bottom: 20px;"
      >{{ changePasswordError }}</v-alert>

      <v-text-field
        v-model="newPassword"
        label="Password"
        :rules="passwordRules"
        type="password"
        counter
        required
      ></v-text-field>

      <v-text-field
        v-model="newPassword2"
        label="Repeat password"
        type="password"
        required
        counter
        style="margin-bottom: 20px;"
      ></v-text-field>

      <div class="text-xs-right">
        <v-btn
          color="primary"
          @click="changePassword"
          style="margin-right: 0;"
          align-end
        >
          Change password
        </v-btn>
      </div>
    </v-form>
  </v-container>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  data() {
    return {
      changePasswordFormValid: false,
      changePasswordError: null,
      changePasswordInProgress: false,
      newPassword: '',
      newPassword2: '',

      passwordRules: [
        (v) => !!v || 'Password is required',
        (v) => v.length >= 6 || 'Password too short. Min 6 characters',
      ],
    };
  },

  methods: {
    async changePassword() {
      this.changePasswordError = null;

      if (!this.$refs.changePasswordForm.validate()) {
        return;
      }

      if (this.newPassword !== this.newPassword2) {
        this.changePasswordError = 'Passwords not equal';
        return;
      }

      this.changePasswordInProgress = true;

      try {
        await axios({
          method: 'post',
          url: '/v1/session/change-password',
          data: {
            token: this.$route.query.token,
            password: this.newPassword,
          },
        });
        await this.$router.replace('/');
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'Password changed',
        });
      } catch (err) {
        this.changePasswordError = getErrorMessage(err);
      } finally {
        this.changePasswordInProgress = false;
      }
    },
  },
};
</script>
