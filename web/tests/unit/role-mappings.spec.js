import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import i18n from '@/plugins/i18';
import RoleMappings from '@/views/RoleMappings.vue';
import RoleMappingForm from '@/components/RoleMappingForm.vue';
import mockAxios from './helpers/axiosMock';

const MAPPING = {
  id: 3,
  provider: 'okta',
  group: 'DevOps',
  project_id: 1,
  project_name: 'Demo',
  role: 'deployer',
  role_name: 'Deployer',
  provider_configured: true,
};

const ROUTES = {
  '/api/auth/login': { oidc_providers: [{ id: 'okta', name: 'Okta' }] },
  '/api/role_mappings': [MAPPING, {
    ...MAPPING, id: 4, provider: 'gone', provider_configured: false,
  }],
  '/api/projects': [{ id: 1, name: 'Demo' }],
  '/api/project/1/roles/all': [{
    slug: 'deployer', name: 'Deployer', builtin: false, project_id: null,
  }],
};

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

describe('RoleMappings.vue', () => {
  let http;
  const wrappers = [];

  beforeEach(() => {
    http = mockAxios();
    http.respond((config) => ROUTES[config.url] || null);
  });

  afterEach(() => {
    http.restore();
    wrappers.splice(0).forEach((w) => w.destroy());
  });

  async function mountWith(component, propsData) {
    const wrapper = mount(component, {
      localVue: createLocalVue(),
      vuetify: new Vuetify(),
      i18n,
      propsData,
      stubs: { EditDialog: { template: '<div />' }, YesNoDialog: true },
    });
    wrappers.push(wrapper);
    await flush();
    return wrapper;
  }

  it('marks a mapping of an unconfigured provider', async () => {
    const wrapper = await mountWith(RoleMappings, {
      isAdmin: true, features: { custom_roles_management: true },
    });

    const rows = wrapper.findAll('tbody tr');
    expect(rows.at(0).text()).to.contain('Okta');
    expect(rows.at(0).text()).not.to.contain('Provider not configured');
    expect(rows.at(1).text()).to.contain('gone');
    expect(rows.at(1).text()).to.contain('Provider not configured');
  });

  it('deletes a mapping', async () => {
    const wrapper = await mountWith(RoleMappings, {
      isAdmin: true, features: { custom_roles_management: true },
    });

    await wrapper.vm.deleteItem(3);

    const deletes = http.requests.filter((r) => r.method === 'delete');
    expect(deletes.map((r) => r.url)).to.deep.equal(['/api/role_mappings/3']);
  });

  it('posts the form', async () => {
    const wrapper = await mountWith(RoleMappingForm, {
      itemId: 'new', providers: [{ id: 'okta', name: 'Okta' }],
    });
    wrapper.vm.item.provider = 'okta';
    wrapper.vm.item.group = 'DevOps';
    wrapper.vm.item.project_id = 1;
    await flush();
    wrapper.vm.item.role = 'deployer';
    await flush();

    await wrapper.vm.save();

    const post = http.requests.find((r) => r.method === 'post');
    expect(post.url).to.equal('/api/role_mappings');
    expect(post.data).to.deep.equal({
      provider: 'okta', group: 'DevOps', project_id: 1, role: 'deployer',
    });
  });
});
