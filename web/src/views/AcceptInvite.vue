<template>
  <v-container
    fluid
    class="pa-0 fill-height d-flex align-center justify-center"
    data-testid="acceptInvite-page"
    :data-state="state === 'processing' ? 'loading' : (state === 'success' ? 'ready' : 'error')"
  >
    <v-card
      data-testid="acceptInvite-card"
      class="pa-6"
      style="
        max-width: 520px;
        border-radius: 12px;
        background-color: var(--highlighted-card-bg-color);
      "
    >
      <v-card-title data-testid="acceptInvite-title">
        Accept Invitation
        <v-spacer />
      </v-card-title>
      <v-card-text class="pb-0">
        <div
          v-if="state === 'processing'"
          class="text-center pt-6 pb-0"
          data-testid="acceptInvite-loading"
        >
          <v-progress-circular indeterminate color="primary" />
          <div class="mt-6">Accepting invitation...</div>
        </div>

        <v-alert
          v-else-if="state === 'success'"
          type="success"
          variant="tonal"
          data-testid="acceptInvite-success"
        >
          Invitation accepted. You now have access to the project.
        </v-alert>

        <v-alert v-else type="error" variant="tonal" data-testid="acceptInvite-error">
          {{ errorMessage }}
        </v-alert>
      </v-card-text>

      <!-- No gap: Vuetify 2 had none after the spacer (Vuetify 4: 8 px between all children), the
           card takes the width of its content. -->
      <v-card-actions class="ga-0">
        <v-spacer />

        <v-btn
          v-if="state === 'success'"
          color="primary"
          variant="elevated"
          @click="goToProject"
          data-testid="acceptInvite-goToProject"
        >
          Go to project
        </v-btn>

        <!-- Not slim: Vuetify 2 narrowed (0 8px) only the direct buttons of v-card-actions,
             Vuetify 4 does it for every button inside. -->
        <div v-else-if="state === 'error'" >
          <v-btn
            variant="text"
            color="primary"
            @click="goToDashboard"
            :disabled="!token"
            :slim="false"
            data-testid="acceptInvite-goToDashboard"
          >
            Go to dashboard
          </v-btn>
          <v-btn
            variant="text"
            color="primary"
            @click="retry"
            :disabled="!token"
            :slim="false"
            data-testid="acceptInvite-retry"
          >
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
  props: {
    token: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      state: 'processing',
      errorMessage: null,
      projectId: null,
    };
  },
  async created() {
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
