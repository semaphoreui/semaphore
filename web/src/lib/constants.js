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
  ansible: {
    icon: 'mdi-ansible',
    color: 'black',
    darkColor: 'white',
  },
  terraform: {
    icon: 'mdi-terraform',
    color: '#7b42bc',
    darkColor: '#7b42bc',
  },
  tofu: {
    icon: '$vuetify.icons.tofu',
    color: 'black',
    darkColor: 'white',
  },
  pulumi: {
    icon: '$vuetify.icons.pulumi',
    color: 'black',
    darkColor: 'white',
  },
  bash: {
    icon: 'mdi-bash',
    color: 'black',
    darkColor: 'white',
  },
  python: {
    icon: 'mdi-language-python',
  },
  powershell: {
    icon: 'mdi-powershell',
  },
};

export const APP_TITLE = {
  ansible: 'Ansible Playbook',
  terraform: 'Terraform Code',
  tofu: 'OpenTofu Code',
  bash: 'Bash Script',
  pulumi: 'Pulumi Code',
  python: 'Python Script',
  powershell: 'PowerShell Script',
};

export const APP_INVENTORY_TITLE = {
  ansible: 'Ansible Inventory',
  terraform: 'Terraform Workspace',
  tofu: 'OpenTofu Workspace',
};

export const DEFAULT_APPS = Object.keys(APP_ICONS);
