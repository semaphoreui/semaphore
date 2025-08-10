<template>
  <v-container
    fluid
    fill-height
    align-center
    justify-center
    class="pa-0"
  >
    <v-card
      class="pa-6"
      style="
        max-width: 520px;
        border-radius: 12px;
        background-color: var(--highlighted-card-bg-color);
      "
    >
      <v-card-title>
        Accept Invitation
        <v-spacer />
      </v-card-title>
      <v-card-text class="pb-0">
        <div v-if="state === 'processing'" class="text-center pt-6 pb-0">
          <v-progress-circular indeterminate color="primary" />
          <div class="mt-6">Accepting invitation...</div>
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

        <v-btn v-if="state === 'success'" color="primary" @click="goToProject">
          Go to project
        </v-btn>

        <div v-else-if="state === 'error'" >
          <v-btn text color="primary" @click="goToDashboard" :disabled="!token">
            Go to dashboard
          </v-btn>
          <v-btn text color="primary" @click="retry" :disabled="!token">
            Try again
          </v-btn>
        </div>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';
import delay from '@/lib/delay';

export default {
  name: 'AcceptInvite',
  data() {
    return {
      state: 'processing',
      token: null,
      errorMessage: null,
      projectId: null,
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

      await delay(2000);

      try {
        const res = (await axios({
          method: 'post',
          url: '/api/invites/accept',
          responseType: 'json',
          data: { token: this.token },
        })).data;
        this.projectId = res.project_id;
        this.state = 'success';
      } catch (err) {
        this.state = 'error';
        this.errorMessage = getErrorMessage(err);
      }
    },

    retry() {
      this.process();
    },

    goToProject() {
      // this.$router.push(`/project/${this.projectId}`);
      let baseURI = document.baseURI;
      if (baseURI.endsWith('/')) {
        baseURI = baseURI.slice(0, -1);
      }
      document.location = `${baseURI}/project/${this.projectId}`;
    },

    goToDashboard() {
      document.location = document.baseURI;
    },
  },
};
</script>
