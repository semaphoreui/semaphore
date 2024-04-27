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

export const MATCHER_TYPE_TITLES = {
  '': 'Matcher',
  body: 'Body',
  header: 'Header',
};

export const MATCHER_TYPE_ICONS = {
  '': 'Matcher',
  body: 'mdi-page-layout-body',
  header: 'mdi-web',
};

export const EXTRACT_VALUE_TYPE_TITLES = {
  '': 'ExtractValue',
  body: 'Body',
  header: 'Header',
};

export const EXTRACT_VALUE_TYPE_ICONS = {
  '': 'ExtractValue',
  body: 'mdi-page-layout-body',
  header: 'mdi-web',
};

export const EXTRACT_VALUE_BODY_DATA_TYPE_TITLES = {
  '': 'BodyDataType',
  json: 'JSON',
  str: 'String',
};

export const EXTRACT_VALUE_BODY_DATA_TYPE_ICONS = {
  '': 'BodyDataType',
  json: 'mdi-code-json',
  str: 'mdi-text',
};

export const APP_ICONS = {
  '': {},
};

export const APP_TITLE = {
  '': '',
};
