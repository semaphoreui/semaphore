<template>
  <v-container
    fluid
    fill-height
    align-center
    justify-center
    class="pa-0"
  >
    <v-card class="pa-6" style="max-width: 520px; border-radius: 12px;">
      <v-card-title>
        Accept Invitation
        <v-spacer />
      </v-card-title>
      <v-card-text>
        <div v-if="state === 'processing'" class="text-center py-6">
          <v-progress-circular indeterminate color="primary" />
          <div class="mt-4">Accepting invitation...</div>
        </div>

        <v-alert v-else-if="state === 'success'" type="success" text>
          Invitation accepted. You now have access to the project.
        </v-alert>

        <v-alert v-else type="error" text>
          {{ errorMessage }}
        </v-alert>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn v-if="state === 'success'" color="primary" @click="goToApp">
          Go to app
        </v-btn>
        <v-btn v-else-if="state === 'error'" text color="primary" @click="retry" :disabled="!token">
          Try again
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  name: 'AcceptInvite',
  data() {
    return {
      state: 'processing',
      token: null,
      errorMessage: null,
    };
  },
  async created() {
    this.token = this.$route.query.token;
    if (!this.token) {
      this.state = 'error';
      this.errorMessage = 'Missing invitation token.';
      return;
    }

    await this.process();
  },
  methods: {

    async process() {
      this.state = 'processing';
      this.errorMessage = null;

      try {
        await axios({
          method: 'post',
          url: '/api/invites/accept',
          responseType: 'json',
          data: { token: this.token },
        });
        this.state = 'success';
      } catch (err) {
        this.state = 'error';
        this.errorMessage = getErrorMessage(err);
      }
    },

    retry() {
      this.process();
    },

    goToApp() {
      this.$router.push('/tasks');
    },
  },
};
</script>
