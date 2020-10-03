<template>
  <v-app v-if="state === 'success'" style="background: white;">
    <v-snackbar
      v-model="snackbar"
      :color="snackbarColor"
      :timeout="3000"
      top
    >
      {{ snackbarText }}
      <v-btn
        text
        @click="snackbar = false"
      >
        Close
      </v-btn>
    </v-snackbar>

    <v-navigation-drawer
      app
      permanent
      v-if="$route.path.startsWith('/project/')"
    >
      <v-toolbar flat class="white">
        <v-toolbar-title>
          <v-icon large style="position: absolute; left: 10px; top: 13px;">
            mdi-play-circle-outline
          </v-icon>
          <router-link to="/" style="color: black; text-decoration: none; margin-left: 46px;">
            {{ project.name }}
          </router-link>
        </v-toolbar-title>
      </v-toolbar>
      <v-divider></v-divider>

      <div style="padding: 10px 15px;">
        <v-select
          solo-inverted
          flat
          hide-details
          :items="projects"
          item-value="id"
          item-text="name"
          v-model="projectId"
        ></v-select>
      </div>

      <v-list class="pt-0" rounded>
        <v-list-item key="templates" :to="`/project/${projectId}/templates`">
          <v-list-item-icon>
            <v-icon>mdi-check-all</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Task Templates</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item key="inventory" :to="`/project/${projectId}/inventory`">
          <v-list-item-icon>
            <v-icon>mdi-monitor-multiple</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Inventory</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-main>
      <router-view v-bind:projectId="projectId"></router-view>
    </v-main>

  </v-app>

  <v-app v-else-if="state === 'loading'">
    <v-main>
      <v-container
        fluid
        fill-height
        align-center
        justify-center
        class="pa-0"
      >
        <v-progress-circular
          :size="70"
          color="primary"
          indeterminate
        ></v-progress-circular>
      </v-container>
    </v-main>
  </v-app>

  <v-app v-else></v-app>
</template>
<style lang="scss">
.site-select {
  & > .v-input__control > .v-input__slot {
    border-radius: 0 !important;
  }
}
</style>

<script>
import axios from 'axios';
import EventBus from './event-bus';

export default {
  name: 'App',
  data() {
    return {
      user: null,

      state: 'loading',

      snackbar: false,
      snackbarText: '',
      snackbarColor: '',

      projects: null,

      projectId: null,
    };
  },

  computed: {
    project() {
      return this.projects.find((x) => x.id === this.projectId);
    },
  },

  watch: {
    async projectId(val) {
      if (val == null) {
        return;
      }
      const projectId = parseInt(this.$route.params.projectId, 10) || null;
      if (val === projectId) {
        return;
      }
      await this.$router.push({ path: `/project/${val}` });
    },
  },

  async created() {
    if (!this.isAuthenticated()) {
      await this.$router.push({ path: '/auth/login' });
      return;
    }

    try {
      await this.loadUserInfo();
      await this.loadProjects();

      if (!this.projectId) {
        if (this.projects.length > 0) {
          this.projectId = this.projects[0].id;
        } else {
          // TODO: create project page
          this.projectId = parseInt(this.$route.params.projectId, 10) || null;
        }
      }

      this.state = 'success';
    } catch (err) {
      EventBus.$emit('i-session-end');
    }
  },

  mounted() {
    EventBus.$on('i-snackbar', (e) => {
      this.snackbar = true;
      this.snackbarColor = e.color;
      this.snackbarText = e.text;
    });

    EventBus.$on('i-session-end', async () => {
      this.snackbar = false;
      this.snackbarColor = '';
      this.snackbarText = '';
      await this.$router.push({ path: '/auth/login' });
    });

    EventBus.$on('i-account-changed', async () => {
      await this.loadUserInfo();
    });

    EventBus.$on('i-site-changed', async () => {
      await this.loadProjects();
    });
  },

  methods: {
    isAuthenticated() {
      return document.cookie.includes('semaphore=');
    },

    async loadProjects() {
      this.projects = (await axios({
        method: 'get',
        url: '/api/projects',
        responseType: 'json',
      })).data;
    },

    async loadUserInfo() {
      if (!this.isAuthenticated()) {
        return;
      }
      this.user = (await axios({
        method: 'get',
        url: '/api/user',
        responseType: 'json',
      })).data;
    },
  },
};
</script>
