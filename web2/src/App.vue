<template>
  <v-app v-if="state === 'success'"  class="app">
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
      dark
      fixed
      width="260"
      v-model="drawer"
      mobile-breakpoint="960"
      v-if="$route.path.startsWith('/project/')"
    >

      <v-list class="pt-0">
        <v-list-item key="project" class="app__project-selector">
          <v-list-item-icon>
            <v-icon color="#015157">mdi-checkbox-blank-circle</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title class="app__project-selector-title">
              {{ project.name }}
            </v-list-item-title>
          </v-list-item-content>

          <v-list-item-icon>
            <v-icon>mdi-chevron-down</v-icon>
          </v-list-item-icon>
        </v-list-item>

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
.app__project-selector {
  height: 64px;
  .v-list-item__icon {
    margin-top: 20px !important;
  }
}

.app__project-selector-title {
  font-size: 1.25rem !important;
  font-weight: bold;
}

.v-application--is-ltr .v-list-item__action:first-child,
.v-application--is-ltr .v-list-item__icon:first-child {
  margin-right: 16px !important;
}

.v-toolbar__content {
  height: 64px !important;
}

.v-data-table-header {
  //background: #f7f7f7 !important;
}

.theme--light.v-data-table > .v-data-table__wrapper > table > thead > tr:last-child > th {
  //border-bottom: 0 !important;
  text-transform: uppercase;
}

.v-data-table > .v-data-table__wrapper > table > tbody > tr {
  background: transparent !important;
  & > td:first-child {
    font-weight: bold !important;
    //font-family: monospace, monospace !important;
  }
}

.v-data-table > .v-data-table__wrapper > table > tbody > tr > th,
.v-data-table > .v-data-table__wrapper > table > thead > tr > th,
.v-data-table > .v-data-table__wrapper > table > tfoot > tr > th,
.v-data-table > .v-data-table__wrapper > table > tbody > tr > td {
  font-size: 16px !important;
}

.v-toolbar__title {
  font-weight: bold !important;
}

.v-app-bar__nav-icon {
  margin-left: 0 !important;
}

.v-toolbar__title:not(:first-child) {
  margin-left: 10px !important;
}

@media (min-width: 960px) {
  .v-app-bar__nav-icon {
    display: none !important;
  }

  .v-toolbar__title:not(:first-child) {
    padding-left: 0 !important;
    margin-left: 0 !important;
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
      drawer: null,

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

    EventBus.$on('i-show-drawer', async () => {
      this.drawer = true;
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
