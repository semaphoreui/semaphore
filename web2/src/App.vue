<template>
  <v-app v-if="state === 'success'"  class="app">
    <NewProjectDialog
      :project-id="projectId"
      v-model="newProjectDialog"
      @saved="onProjectSaved"
    />

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
      <v-menu bottom max-width="235">
        <template v-slot:activator="{ on, attrs }">
          <v-list class="pa-0">
            <v-list-item
              key="project"
              class="app__project-selector"
              v-bind="attrs"
              v-on="on"
            >
              <v-list-item-icon>
                <v-avatar :color="getProjectColor(project)" size="24">
                  <span class="white--text">{{ getProjectInitials(project) }}</span>
                </v-avatar>
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
          </v-list>
        </template>
        <v-list>
          <v-list-item
            v-for="(item, i) in projects"
            :key="i"
            :to="`/project/${item.id}`"
            @click="setLastProjectId(item.id)"
          >
            <v-list-item-icon>
              <v-avatar :color="getProjectColor(item)" size="24">
                <span class="white--text">{{ getProjectInitials(item) }}</span>
              </v-avatar>
            </v-list-item-icon>
            <v-list-item-content>{{ item.name }}</v-list-item-content>
          </v-list-item>

          <v-list-item @click="newProjectDialog = true">
            <v-list-item-icon>
              <v-icon>mdi-plus</v-icon>
            </v-list-item-icon>

            <v-list-item-content>
              New project...
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-menu>

      <v-list class="pt-0">
        <v-list-item key="dashboard" :to="`/project/${projectId}/history`">
          <v-list-item-icon>
            <v-icon>mdi-view-dashboard</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Dashboard</v-list-item-title>
          </v-list-item-content>
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

        <v-list-item key="environment" :to="`/project/${projectId}/environment`">
          <v-list-item-icon>
            <v-icon>mdi-code-braces</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Environment</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item key="keys" :to="`/project/${projectId}/keys`">
          <v-list-item-icon>
            <v-icon>mdi-key-change</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Key Store</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item key="repositories" :to="`/project/${projectId}/repositories`">
          <v-list-item-icon>
            <v-icon>mdi-git</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Playbook Repositories</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item key="team" :to="`/project/${projectId}/team`">
          <v-list-item-icon>
            <v-icon>mdi-account-multiple</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Team</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

      </v-list>

      <template v-slot:append>
        <v-menu top max-width="235" nudge-top="12">
          <template v-slot:activator="{ on, attrs }">
            <v-list class="pa-0">
              <v-list-item
                key="project"
                v-bind="attrs"
                v-on="on"
              >
                <v-list-item-icon>
                  <v-icon>mdi-account</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{ user.name }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </template>

          <v-list>
            <v-list-item key="users" to="/users">
              <v-list-item-icon>
                <v-icon>mdi-account-multiple</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                Users
              </v-list-item-content>
            </v-list-item>

            <v-list-item key="edit" to="/user">
              <v-list-item-icon>
                <v-icon>mdi-pencil</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                Edit Account
              </v-list-item-content>
            </v-list-item>

            <v-list-item key="sign_out" @click="signOut()">
              <v-list-item-icon>
                <v-icon>mdi-exit-to-app</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                Sign Out
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
      </template>
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
}

.theme--light.v-data-table > .v-data-table__wrapper > table > thead > tr:last-child > th {
  text-transform: uppercase;
}

.v-data-table > .v-data-table__wrapper > table > tbody > tr {
  background: transparent !important;
  & > td:first-child {
    font-weight: bold !important;
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
import NewProjectDialog from '@/components/NewProjectDialog.vue';
import EventBus from './event-bus';

const PROJECT_COLORS = [
  'red',
  'blue',
  'orange',
  'green',
];

export default {
  name: 'App',
  components: {
    NewProjectDialog,
  },
  data() {
    return {
      drawer: null,

      user: null,

      state: 'loading',

      snackbar: false,
      snackbarText: '',
      snackbarColor: '',

      projects: null,

      newProjectDialog: null,
    };
  },

  computed: {
    projectId() {
      return parseInt(this.$route.params.projectId, 10) || null;
    },
    project() {
      return this.projects.find((x) => x.id === this.projectId);
    },
  },

  async created() {
    if (!this.isAuthenticated()) {
      this.state = 'success';

      if (this.$route.path !== '/auth/login') {
        await this.$router.push({ path: '/auth/login' });
      }

      return;
    }

    try {
      await this.loadUserInfo();
      await this.loadProjects();

      if (!this.projectId) {
        const projectId = parseInt(localStorage.getItem('projectId'), 10)
          || this.projects[0].id;
        await this.$router.push({ path: `/project/${projectId}` });
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

    async onProjectSaved(e) {
      if (e.action === 'new') {
        await this.$router.push({ path: `/project/${e.item.id}` });
      }
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

    setLastProjectId(projectId) {
      localStorage.setItem('projectId', projectId);
    },

    getProjectColor(projectData) {
      const projectIndex = this.projects.length
        - this.projects.findIndex((p) => p.id === projectData.id);
      return PROJECT_COLORS[projectIndex % PROJECT_COLORS.length];
    },

    getProjectInitials(projectData) {
      const parts = projectData.name.split(/\s/);
      if (parts.length >= 2) {
        return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
      }
      return parts[0].substr(0, 2).toUpperCase();
    },

    async signOut() {
      (await axios({
        method: 'post',
        url: '/api/auth/logout',
        responseType: 'json',
      }));

      await this.$router.push({ path: '/auth/login' });
    },
  },
};
</script>
