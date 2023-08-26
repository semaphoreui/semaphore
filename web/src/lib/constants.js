export const TEMPLATE_TYPE_ICONS = {
  '': 'mdi-cog',
  build: 'mdi-wrench',
  deploy: 'mdi-arrow-up-bold-box',
};

export const TEMPLATE_TYPE_TITLES = {
  '': 'Task',
  build: 'Build',
  deploy: 'Deploy',
};

export const TEMPLATE_TYPE_ACTION_TITLES = {
  '': 'Run',
  build: 'Build',
  deploy: 'Deploy',
};

export const USER_PERMISSIONS = {
  runProjectTasks: 1,
  updateProject: 2,
  manageProjectResources: 4,
  manageProjectUsers: 8,
};

export const USER_ROLES = [{
  slug: 'owner',
  title: 'Owner',
}, {
  slug: 'manager',
  title: 'Manager',
}, {
  slug: 'task_runner',
  title: 'Task Runner',
}, {
  slug: 'guest',
  title: 'Guest',
}];
