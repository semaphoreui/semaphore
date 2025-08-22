import Vue from 'vue';
import VueRouter from 'vue-router';
import RestoreProject from '@/views/project/RestoreProject.vue';
import Tasks from '@/views/Tasks.vue';
import TaskList from '@/components/TaskList.vue';
import TemplateDetails from '@/views/project/template/TemplateDetails.vue';
import TemplateTerraformState from '@/views/project/template/TemplateTerraformState.vue';
import Invites from '@/views/project/Invites.vue';
import Schedule from '../views/project/Schedule.vue';
import History from '../views/project/History.vue';
import Activity from '../views/project/Activity.vue';
import Settings from '../views/project/Settings.vue';
import Templates from '../views/project/Templates.vue';
import TemplateView from '../views/project/TemplateView.vue';
import Environment from '../views/project/Environment.vue';
import Inventory from '../views/project/Inventory.vue';
import Keys from '../views/project/Keys.vue';
import Repositories from '../views/project/Repositories.vue';
import Team from '../views/project/Team.vue';
import Users from '../views/Users.vue';
import Auth from '../views/Auth.vue';
import New from '../views/project/New.vue';
import Integrations from '../views/project/Integrations.vue';
import IntegrationExtractor from '../views/project/IntegrationExtractor.vue';
import Apps from '../views/Apps.vue';
import Runners from '../views/Runners.vue';
import Stats from '../views/project/Stats.vue';
import Tokens from '../views/Tokens.vue';
import AcceptInvite from '../views/AcceptInvite.vue';
import SecretStorage from '../views/project/SecretStorages.vue';

Vue.use(VueRouter);

const routes = [
  {
    path: '/project/new',
    component: New,
  },
  {
    path: '/project/restore',
    component: RestoreProject,
  },
  {
    path: '/project/:projectId',
    redirect: '/project/:projectId/history',
  },
  {
    path: '/project/:projectId/secret_storages',
    component: SecretStorage,
  },
  {
    path: '/project/:projectId/history',
    component: History,
  },
  {
    path: '/project/:projectId/stats',
    component: Stats,
  },
  {
    path: '/project/:projectId/activity',
    component: Activity,
  },
  {
    path: '/project/:projectId/runners',
    component: Runners,
  },
  {
    path: '/project/:projectId/schedule',
    component: Schedule,
  },
  {
    path: '/project/:projectId/settings',
    component: Settings,
  },
  {
    path: '/project/:projectId/templates',
    component: Templates,
  },
  {
    path: '/project/:projectId/views/:viewId/templates',
    component: Templates,
  },
  {
    path: '/project/:projectId/templates/:templateId',
    redirect: '/project/:projectId/templates/:templateId/tasks',
    component: TemplateView,
    children: [{
      path: 'tasks',
      component: TaskList,
    }, {
      path: 'details',
      component: TemplateDetails,
    }, {
      path: 'state',
      component: TemplateTerraformState,
    }],
  },
  {
    path: '/project/:projectId/views/:viewId/templates/:templateId',
    redirect: '/project/:projectId/views/:viewId/templates/:templateId/tasks',
    component: TemplateView,
    children: [{
      path: 'tasks',
      component: TaskList,
    }, {
      path: 'details',
      component: TemplateDetails,
    }, {
      path: 'state',
      component: TemplateTerraformState,
    }],
  },
  {
    path: '/project/:projectId/environment',
    component: Environment,
  },
  {
    path: '/project/:projectId/inventory',
    component: Inventory,
  },
  {
    path: '/project/:projectId/integrations',
    component: Integrations,
  },
  {
    path: '/project/:projectId/integrations/:integrationId',
    component: IntegrationExtractor,
  },
  {
    path: '/project/:projectId/repositories',
    component: Repositories,
  },
  {
    path: '/project/:projectId/keys',
    component: Keys,
  },
  {
    path: '/project/:projectId/team',
    component: Team,
  },
  {
    path: '/project/:projectId/invites',
    component: Invites,
  },
  {
    path: '/auth/login',
    component: Auth,
  },
  {
    path: '/users',
    component: Users,
  },
  {
    path: '/runners',
    component: Runners,
  },
  {
    path: '/tasks',
    component: Tasks,
  },
  {
    path: '/apps',
    component: Apps,
  },
  {
    path: '/tokens',
    component: Tokens,
  },
  {
    path: '/accept-invite',
    component: AcceptInvite,
  },
];

const router = new VueRouter({
  mode: 'history',
  routes,
});

export default router;
