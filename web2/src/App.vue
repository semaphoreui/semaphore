<template>
  <v-app v-if="state === 'success'" class="app">
    <EditDialog
      v-model="passwordDialog"
      save-button-text="Save"
      title="Change password"
      v-if="user"
      event-name="i-user"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <ChangePasswordForm
          :project-id="projectId"
          :item-id="user.id"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <EditDialog
      v-model="userDialog"
      save-button-text="Save"
      title="Edit User"
      v-if="user"
      event-name="i-user"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <UserForm
          :project-id="projectId"
          :item-id="user.id"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <EditDialog
      v-model="taskLogDialog"
      save-button-text="Delete"
      :max-width="1000"
      :hide-buttons="true"
      @close="onTaskLogDialogClosed()"
    >
      <template v-slot:title={}>
        <router-link
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/templates/${template ? template.id : null}`"
          @click="taskLogDialog = false"
        >{{ template ? template.alias : null }}</router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">Task #{{ task ? task.id : null }}</span>
        <v-spacer></v-spacer>
        <v-btn
          icon
        >
          <v-icon @click="taskLogDialog = false; onTaskLogDialogClosed()">mdi-close</v-icon>
        </v-btn>
      </template>
      <template v-slot:form="{}">
        <TaskLogView :project-id="projectId" :item-id="task ? task.id : null" />
      </template>
    </EditDialog>

    <EditDialog
      v-model="newProjectDialog"
      save-button-text="Create"
      title="New Project"
      event-name="i-project"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <ProjectForm
          item-id="new"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

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
      color="#005057"
      fixed
      width="260"
      v-model="drawer"
      mobile-breakpoint="960"
      v-if="$route.path.startsWith('/project/')"
    >
      <v-menu bottom max-width="235" v-if="project">
        <template v-slot:activator="{ on, attrs }">
          <v-list class="pa-0">
            <v-list-item
              key="project"
              class="app__project-selector"
              v-bind="attrs"
              v-on="on"
            >
              <v-list-item-icon>
                <v-avatar
                  :color="getProjectColor(project)"
                  size="24"
                  style="font-size: 13px; font-weight: bold;"
                >
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
            @click="selectProject(item.id)"
          >
            <v-list-item-icon>
              <v-avatar
                :color="getProjectColor(item)"
                size="24"
                style="font-size: 13px; font-weight: bold;"
              >
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

      <v-list class="pt-0" v-if="!project">
        <v-list-item key="new_project" :to="`/project/new`">
          <v-list-item-icon>
            <v-icon>mdi-plus</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>New Project</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>

      <v-list class="pt-0" v-if="project">
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

            <v-list-item key="edit" @click="userDialog = true">
              <v-list-item-icon>
                <v-icon>mdi-pencil</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                Edit Account
              </v-list-item-content>
            </v-list-item>

            <v-list-item key="password" @click="passwordDialog = true">
              <v-list-item-icon>
                <v-icon>mdi-lock</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                Change Password
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
      <router-view :projectId="projectId" :userId="user ? user.id : null"></router-view>
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

.breadcrumbs {

}

.breadcrumbs__item {
}

.breadcrumbs__item--link {
  text-decoration-line: none;
  &:hover {
    text-decoration-line: underline;
  }
}

.breadcrumbs__separator {
  padding: 0 10px;
}

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
  white-space: nowrap;
}

.v-data-table > .v-data-table__wrapper > table > tbody > tr {
  background: transparent !important;
  & > td {
    white-space: nowrap;
  }
  & > td:first-child {
    //font-weight: bold !important;
    a {
      text-decoration-line: none;
      &:hover {
        text-decoration-line: underline;
      }
    }
  }
}

.v-data-table > .v-data-table__wrapper > table > tbody > tr > th,
.v-data-table > .v-data-table__wrapper > table > thead > tr > th,
.v-data-table > .v-data-table__wrapper > table > tfoot > tr > th,
.v-data-table > .v-data-table__wrapper > table > tbody > tr > td {
  font-size: 1rem !important;
}

.v-data-footer {
  font-size: 1rem !important;
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
import { getErrorMessage } from '@/lib/error';
import EditDialog from '@/components/EditDialog.vue';
import TaskLogView from '@/components/TaskLogView.vue';
import ProjectForm from '@/components/ProjectForm.vue';
import UserForm from '@/components/UserForm.vue';
import ChangePasswordForm from '@/components/ChangePasswordForm.vue';
import EventBus from '@/event-bus';
import socket from '@/socket';

const PROJECT_COLORS = [
  'red',
  'blue',
  'orange',
  'green',
];

export default {
  name: 'App',
  components: {
    ChangePasswordForm,
    UserForm,
    EditDialog,
    TaskLogView,
    ProjectForm,
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
      userDialog: null,
      passwordDialog: null,

      taskLogDialog: null,
      task: null,
      template: null,
    };
  },

  watch: {
    async projects(val) {
      if (val.length === 0
        && this.$route.path.startsWith('/project/')
        && this.$route.path !== '/project/new') {
        await this.$router.push({ path: '/project/new' });
      }
    },

    async $route(val) {
      if (val.query.t == null) {
        this.taskLogDialog = false;
      } else {
        const taskId = parseInt(this.$route.query.t || '', 10);
        if (taskId) {
          EventBus.$emit('i-show-task', { taskId });
        }
      }
    },
  },

  computed: {
    projectId() {
      return parseInt(this.$route.params.projectId, 10) || null;
    },

    project() {
      return this.projects.find((x) => x.id === this.projectId);
    },

    isAuthenticated() {
      return document.cookie.includes('semaphore=');
    },
  },

  async created() {
    if (!this.isAuthenticated) {
      if (this.$route.path !== '/auth/login') {
        await this.$router.push({ path: '/auth/login' });
      }
      this.state = 'success';
      return;
    }

    try {
      await this.reloadData();
      this.state = 'success';
    } catch (err) {
      EventBus.$emit('i-snackbar', {
        color: 'error',
        text: getErrorMessage(err),
      });
    }
  },

  mounted() {
    EventBus.$on('i-snackbar', (e) => {
      this.snackbar = true;
      this.snackbarColor = e.color;
      this.snackbarText = e.text;
    });

    EventBus.$on('i-account-change', async () => {
      await this.loadUserInfo();
    });

    EventBus.$on('i-show-drawer', async () => {
      this.drawer = true;
    });

    EventBus.$on('i-show-task', async (e) => {
      if (parseInt(this.$route.query.t || '', 10) !== e.taskId) {
        const query = { ...this.$route.query, t: e.taskId };
        await this.$router.replace({ query });
      }

      this.task = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/tasks/${e.taskId}`,
        responseType: 'json',
      })).data;

      this.template = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates/${this.task.template_id}`,
        responseType: 'json',
      })).data;

      this.taskLogDialog = true;
    });

    EventBus.$on('i-open-last-project', async () => {
      await this.trySelectMostSuitableProject();
    });

    EventBus.$on('i-user', async (e) => {
      let text;

      switch (e.action) {
        case 'new':
          text = `User ${e.item.name} created`;
          break;
        case 'edit':
          text = `User ${e.item.name} saved`;
          break;
        case 'delete':
          text = `User ${e.item.name} deleted`;
          break;
        default:
          throw new Error('Unknown project action');
      }

      EventBus.$emit('i-snackbar', {
        color: 'success',
        text,
      });

      if (this.user && e.item.id === this.user.id) {
        await this.loadUserInfo();
      }
    });

    EventBus.$on('i-project', async (e) => {
      let text;

      const project = this.projects.find((p) => p.id === e.item.id) || e.item;
      const projectName = project.name || `#${project.id}`;

      switch (e.action) {
        case 'new':
          text = `Project ${projectName} created`;
          break;
        case 'edit':
          text = `Project ${projectName} saved`;
          break;
        case 'delete':
          text = `Project ${projectName} deleted`;
          break;
        default:
          throw new Error('Unknown project action');
      }

      EventBus.$emit('i-snackbar', {
        color: 'success',
        text,
      });

      await this.loadProjects();

      switch (e.action) {
        case 'new':
          await this.selectProject(e.item.id);
          break;
        case 'delete':
          if (this.projectId === e.item.id && this.projects.length > 0) {
            await this.selectProject(this.projects[0].id);
          }
          break;
        default:
          break;
      }
    });
  },

  methods: {
    async onTaskLogDialogClosed() {
      const query = { ...this.$route.query, t: undefined };
      await this.$router.replace({ query });
    },

    async reloadData() {
      if (!socket.isRunning()) {
        socket.start();
      }

      await this.loadUserInfo();
      await this.loadProjects();

      if (this.$route.path === '/'
        || this.$route.path === '/project'
        || (this.$route.path.startsWith('/project/'))) {
        // try to find project and switch to it
        await this.trySelectMostSuitableProject();
      }

      if (this.$route.query.t) {
        const taskId = parseInt(this.$route.query.t || '', 10);
        if (taskId) {
          EventBus.$emit('i-show-task', { taskId });
        }
      }
    },

    async trySelectMostSuitableProject() {
      if (this.projects.length === 0) {
        if (this.$route.path !== '/project/new') {
          await this.$router.push({ path: '/project/new' });
        }
        return;
      }

      let projectId;

      if (this.projectId) {
        projectId = this.projectId;
      }

      if ((projectId == null || !this.projects.some((p) => p.id === projectId))
        && localStorage.getItem('projectId')) {
        projectId = parseInt(localStorage.getItem('projectId'), 10);
      }

      if (projectId == null || !this.projects.some((p) => p.id === projectId)) {
        projectId = this.projects[0].id;
      }

      if (projectId != null) {
        await this.selectProject(projectId);
      }
    },

    async selectProject(projectId) {
      localStorage.setItem('projectId', projectId);
      if (this.projectId === projectId) {
        return;
      }
      await this.$router.push({ path: `/project/${projectId}` });
    },

    async loadProjects() {
      this.projects = (await axios({
        method: 'get',
        url: '/api/projects',
        responseType: 'json',
      })).data;
    },

    async loadUserInfo() {
      if (!this.isAuthenticated) {
        return;
      }
      this.user = (await axios({
        method: 'get',
        url: '/api/user/',
        responseType: 'json',
      })).data;
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
      this.snackbar = false;
      this.snackbarColor = '';
      this.snackbarText = '';

      socket.stop();

      (await axios({
        method: 'post',
        url: '/api/auth/logout',
        responseType: 'json',
      }));

      if (this.$route.path !== '/auth/login') {
        await this.$router.push({ path: '/auth/login' });
      }
    },
  },
};
</script>
