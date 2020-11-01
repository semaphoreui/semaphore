import Vue from 'vue';
import VueRouter from 'vue-router';
import Dashboard from '../views/project/Dashboard.vue';
import Templates from '../views/project/Templates.vue';
import TemplateView from '../views/project/TemplateView.vue';
import TemplateEdit from '../views/project/TemplateEdit.vue';
import Environment from '../views/project/Environment.vue';
import Inventory from '../views/project/Inventory.vue';
import Keys from '../views/project/Keys.vue';
import Repositories from '../views/project/Repositories.vue';
import Team from '../views/project/Team.vue';
import Users from '../views/Users.vue';
import Auth from '../views/Auth.vue';
import ChangePassword from '../views/ChangePassword.vue';

Vue.use(VueRouter);

const routes = [
  {
    path: '/project/:projectId',
    redirect: '/project/:projectId/dashboard',
  },
  {
    path: '/project/:projectId/dashboard',
    component: Dashboard,
  },
  {
    path: '/project/:projectId/templates',
    component: Templates,
  },
  {
    path: '/project/:projectId/templates/:templateId',
    component: TemplateView,
  },
  {
    path: '/project/:projectId/templates/:templateId/edit',
    component: TemplateEdit,
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
    path: '/change-password',
    component: ChangePassword,
  },
];

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes,
});

export default router;
