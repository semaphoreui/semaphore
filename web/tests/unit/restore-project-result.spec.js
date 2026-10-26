import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import i18n from '@/plugins/i18';
import RestoreProjectResult from '@/components/RestoreProjectResult.vue';
import mockAxios, { httpError } from './helpers/axiosMock';

const ROUTES = {
  '/api/project/7/roles/all': [
    {
      slug: 'ops', name: 'Ops', builtin: false, project_id: null,
    },
    {
      slug: 'manager', name: 'Manager', builtin: true, project_id: null,
    },
    {
      slug: 'deployer', name: 'Deployer', builtin: false, project_id: null,
    },
  ],
};

const PROJECT = {
  id: 7,
  name: 'copy',
  unassigned_roles: [
    { slug: 'ops', reason: 'not_in_file', templates: [{ id: 3, name: 'Deploy', permissions: 1 }] },
    {
      slug: 'deployer',
      reason: 'slug_taken',
      name: 'Deployer',
      permissions: 5,
      templates: [{ id: 3, name: 'Deploy', permissions: 2 }],
    },
  ],
};

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

describe('RestoreProjectResult.vue', () => {
  let http;
  const wrappers = [];

  beforeEach(() => {
    http = mockAxios();
    http.respond((config) => ROUTES[config.url] || {});
  });

  afterEach(() => {
    http.restore();
    wrappers.splice(0).forEach((w) => w.destroy());
  });

  async function mountWith(project, emptyKeys = 0) {
    const wrapper = mount(RestoreProjectResult, {
      localVue: createLocalVue(),
      vuetify: new Vuetify(),
      i18n,
      propsData: { project, emptyKeys },
    });
    wrappers.push(wrapper);
    await flush();
    return wrapper;
  }

  it('renders a row per role with its templates', async () => {
    const wrapper = await mountWith(PROJECT);

    expect(wrapper.text()).to.include('Project copy successfully restored');
    expect(wrapper.text()).to.include('ops');
    expect(wrapper.text()).to.include('deployer');
    expect(wrapper.text()).to.include('Deploy');
  });

  it('preselects the role with the same slug', async () => {
    const wrapper = await mountWith(PROJECT);

    expect(wrapper.vm.choices.ops).to.equal('ops');
    expect(wrapper.vm.choices.deployer).to.equal('deployer');
  });

  it('creates the role and then grants its templates', async () => {
    const wrapper = await mountWith(PROJECT);
    wrapper.vm.choices.deployer = wrapper.vm.CREATE;
    wrapper.vm.newSlugs.deployer = 'deployer2';

    await wrapper.vm.assign();
    await flush();

    const posts = http.requests.filter((r) => r.method === 'post');
    expect(posts.map((r) => [r.url, r.data])).to.deep.equal([
      ['/api/project/7/templates/3/perms', { role_slug: 'ops', permissions: 1 }],
      ['/api/project/7/roles', { name: 'Deployer', slug: 'deployer2', permissions: 5 }],
      ['/api/project/7/templates/3/perms', { role_slug: 'deployer2', permissions: 2 }],
    ]);
    expect(wrapper.vm.rows).to.have.lengthOf(0);
  });

  it('keeps Assign off while the new slug is invalid', async () => {
    const taken = PROJECT.unassigned_roles[1];
    const wrapper = await mountWith({ ...PROJECT, unassigned_roles: [taken] });
    wrapper.vm.choices.deployer = wrapper.vm.CREATE;
    wrapper.vm.newSlugs.deployer = '';
    await wrapper.vm.$nextTick();

    expect(wrapper.vm.hasChoices).to.equal(false);
  });

  it('retries only what failed', async () => {
    const project = {
      id: 7,
      name: 'copy',
      unassigned_roles: [{
        slug: 'deployer',
        reason: 'slug_taken',
        name: 'Deployer',
        permissions: 5,
        templates: [
          { id: 3, name: 'Deploy', permissions: 2 },
          { id: 4, name: 'Stage', permissions: 1 },
        ],
      }],
    };
    let stageFails = true;
    http.respond((config) => {
      if (config.url === '/api/project/7/templates/4/perms' && stageFails) {
        stageFails = false;
        throw httpError(400, { error: 'Role does not exist in this project' });
      }
      return ROUTES[config.url] || {};
    });
    const wrapper = await mountWith(project);
    wrapper.vm.choices.deployer = wrapper.vm.CREATE;
    wrapper.vm.newSlugs.deployer = 'deployer2';

    await wrapper.vm.assign();
    await flush();
    expect(wrapper.vm.rows).to.have.lengthOf(1);
    expect(wrapper.text()).to.include('Stage: Role does not exist in this project');
    await wrapper.vm.assign();
    await flush();

    const posts = http.requests.filter((r) => r.method === 'post').map((r) => r.url);
    expect(posts).to.deep.equal([
      '/api/project/7/roles',
      '/api/project/7/templates/3/perms',
      '/api/project/7/templates/4/perms',
      '/api/project/7/templates/4/perms',
    ]);
    expect(wrapper.vm.rows).to.have.lengthOf(0);
  });

  it('locks the role once one of its templates is granted', async () => {
    http.respond((config) => {
      if (config.url === '/api/project/7/templates/4/perms') {
        throw httpError(400, { error: 'Role already has permissions on this template' });
      }
      return ROUTES[config.url] || {};
    });
    const wrapper = await mountWith({
      id: 7,
      name: 'copy',
      unassigned_roles: [{
        slug: 'ops',
        reason: 'not_in_file',
        templates: [{ id: 3, name: 'Deploy', permissions: 1 }, { id: 4, name: 'Stage', permissions: 1 }],
      }],
    });

    await wrapper.vm.assign();
    await flush();

    expect(wrapper.find('.v-select').classes()).to.include('v-input--is-disabled');
  });

  it('shows a role whose slug is an object property name', async () => {
    const wrapper = await mountWith({
      id: 7,
      name: 'copy',
      unassigned_roles: [
        { slug: 'constructor', reason: 'not_in_file', templates: [{ id: 3, name: 'Deploy', permissions: 1 }] },
      ],
    });

    expect(wrapper.vm.rows.map((r) => r.slug)).to.deep.equal(['constructor']);
    expect(wrapper.text()).not.to.include('function');
    wrapper.vm.choices.constructor = 'manager';
    await wrapper.vm.assign();
    await flush();
    expect(wrapper.vm.rows).to.have.lengthOf(0);
  });

  it('assigns to a role whose slug is __create__', async () => {
    http.respond((config) => {
      if (config.url === '/api/project/7/roles/all') {
        return [{
          slug: '__create__', name: 'Odd', builtin: false, project_id: null,
        }];
      }
      return {};
    });
    const taken = PROJECT.unassigned_roles[1];
    const wrapper = await mountWith({ ...PROJECT, unassigned_roles: [taken] });
    wrapper.vm.choices.deployer = '__create__';

    await wrapper.vm.assign();
    await flush();

    const posts = http.requests.filter((r) => r.method === 'post');
    expect(posts.map((r) => [r.url, r.data])).to.deep.equal([
      ['/api/project/7/templates/3/perms', { role_slug: '__create__', permissions: 2 }],
    ]);
  });

  it('hides the keys alert when no key is empty', async () => {
    const wrapper = await mountWith({ id: 7, name: 'copy', unassigned_roles: [] }, 0);

    expect(wrapper.text()).not.to.include('empty keys added');
  });

  it('hides the list when nothing is unassigned', async () => {
    const wrapper = await mountWith({ id: 7, name: 'copy', unassigned_roles: [] });

    expect(wrapper.text()).not.to.include('were not restored');
    expect(wrapper.find('button').exists()).to.equal(false);
  });
});
