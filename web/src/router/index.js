import Vue from 'vue';
import VueRouter from 'vue-router';
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

Vue.use(VueRouter);

const routes = [
  {
    path: '/project/new',
    component: New,
  },
  {
    path: '/project/:projectId',
    redirect: '/project/:projectId/history',
  },
  {
    path: '/project/:projectId/history',
    component: History,
  },
  {
    path: '/project/:projectId/activity',
    component: Activity,
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
    component: TemplateView,
  },
  {
    path: '/project/:projectId/views/:viewId/templates/:templateId',
    component: TemplateView,
  },
  // {
  //   path: '/project/:projectId/views/:viewId/templates/:templateId/edit',
  //   component: TemplateEdit,
  // },
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
    path: '/project/:projectId/integration/:integrationId',
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
    path: '/apps',
    component: Apps,
  },
];

const router = new VueRouter({
  mode: 'history',
  routes,
});

export default router;
