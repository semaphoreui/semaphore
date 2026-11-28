import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import IntegrationAuthForm from '@/components/IntegrationAuthForm.vue';
import EndpointForm from '@/components/EndpointForm.vue';
import IntegrationGroupForm from '@/components/IntegrationGroupForm.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const keys = [
  { id: 1, name: 'login', type: 'login_password' },
  { id: 2, name: 'token', type: 'string' },
  { id: 3, name: 'ssh', type: 'ssh' },
];
const endpoints = [{ id: 4, name: 'Public hooks', auth_method: 'github' }];

function respondTo(url, item) {
  if (url.endsWith('/keys')) return keys;
  if (url.endsWith('/endpoints')) return endpoints;
  if (/\/(endpoints|integration_groups)\/\d+$/.test(url)) return item;
  return {};
}

/** Mount options: `$t` answers with the key; `stubs` are added to it. */
function mountOptions(props, stubs = {}) {
  return {
    props,
    global: {
      mocks: { $t: (k) => k },
      stubs,
    },
  };
}

describe('IntegrationAuthForm.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond((c) => respondTo(c.url));
  });

  afterEach(() => {
    http.restore();
  });

  it('offers only login/password keys for BasicAuth and string keys too for the rest', async () => {
    const wrapper = mount(IntegrationAuthForm, mountOptions(
      { projectId: 7, modelValue: { auth_method: 'token', auth_secret_id: null } },
      { HighlightedCard: true },
    ));
    await flush();

    expect(wrapper.vm.authSecretKeys.map((k) => k.id)).to.deep.equal([1, 2]);

    await wrapper.setProps({ modelValue: { auth_method: 'basic', auth_secret_id: null } });
    expect(wrapper.vm.authSecretKeys.map((k) => k.id)).to.deep.equal([1]);
  });

  it('drops a secret that no longer fits the method and emits the change', async () => {
    const wrapper = mount(IntegrationAuthForm, mountOptions(
      { projectId: 7, modelValue: { auth_method: 'token', auth_secret_id: 2 } },
      { HighlightedCard: true },
    ));
    await flush();
    expect(wrapper.emitted('update:modelValue')).to.equal(undefined);

    await wrapper.setProps({ modelValue: { auth_method: 'basic', auth_secret_id: 2 } });
    await flush();

    const emitted = wrapper.emitted('update:modelValue');
    expect(emitted, 'update:modelValue event').to.not.equal(undefined);
    expect(emitted[emitted.length - 1][0]).to.deep.equal({ auth_method: 'basic', auth_secret_id: null });
  });

  it('shows the header field only for header-based methods', async () => {
    const wrapper = mount(IntegrationAuthForm, mountOptions({ projectId: 7, modelValue: { auth_method: 'hmac' } }));
    await flush();
    expect(wrapper.find('.IntegrationAuthForm__header').exists()).to.equal(true);

    await wrapper.setProps({ modelValue: { auth_method: 'github' } });
    expect(wrapper.find('.IntegrationAuthForm__header').exists()).to.equal(false);
    expect(wrapper.find('.IntegrationAuthForm__secret').exists()).to.equal(true);

    await wrapper.setProps({ modelValue: { auth_method: '' } });
    expect(wrapper.find('.IntegrationAuthForm__secret').exists()).to.equal(false);
  });
});

describe('EndpointForm.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  it('creates an endpoint with its authorization', async () => {
    http.respond((c) => (c.method === 'post' ? { id: 9 } : respondTo(c.url)));
    const wrapper = mount(EndpointForm, mountOptions({ projectId: 7, itemId: 'new' }));
    await flush();

    wrapper.vm.item.name = 'hooks';
    wrapper.vm.item = {
      ...wrapper.vm.item, auth_method: 'token', auth_header: 'X-Token', auth_secret_id: 2,
    };
    await flush();
    await wrapper.vm.save();

    expect(wrapper.vm.formError).to.equal(null);
    const post = http.requests.find((r) => r.method === 'post');
    expect(post, 'POST request').to.not.equal(undefined);
    expect(post.url).to.equal('/api/project/7/endpoints');
    expect(post.data.name).to.equal('hooks');
    expect(post.data.auth_method).to.equal('token');
    expect(post.data.auth_header).to.equal('X-Token');
    expect(post.data.auth_secret_id).to.equal(2);
  });

  it('fills the fields the API omits so the inputs can bind', async () => {
    http.respond((c) => respondTo(c.url, { id: 4, name: 'Public hooks' }));
    const wrapper = mount(EndpointForm, mountOptions({ projectId: 7, itemId: 4 }));
    await flush();

    expect(wrapper.vm.item.auth_method).to.equal('');
    expect(wrapper.vm.item.auth_secret_id).to.equal(null);
    expect(wrapper.vm.item.description).to.equal('');
  });
});

describe('IntegrationGroupForm.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  it('loads the endpoints and hides the auth block until authorization is overridden', async () => {
    http.respond((c) => respondTo(c.url));
    const wrapper = mount(IntegrationGroupForm, mountOptions({ projectId: 7, itemId: 'new' }));
    await flush();

    expect(http.requests.map((r) => r.url)).to.include('/api/project/7/endpoints');
    expect(wrapper.vm.endpoints).to.deep.equal(endpoints);
    expect(wrapper.find('.IntegrationAuthForm').exists()).to.equal(false);

    wrapper.vm.item.auth_override = true;
    await flush();
    expect(wrapper.find('.IntegrationAuthForm').exists()).to.equal(true);
  });

  it('sends empty auth fields when the group inherits the endpoint authorization', async () => {
    http.respond((c) => {
      if (c.method === 'put') return {};
      return respondTo(c.url, {
        id: 3, name: 'hooks', endpoint_id: 4, auth_override: true, auth_method: 'token', auth_header: 'X', auth_secret_id: 2,
      });
    });
    const wrapper = mount(IntegrationGroupForm, mountOptions({ projectId: 7, itemId: 3 }));
    await flush();

    wrapper.vm.item.auth_override = false;
    await flush();
    await wrapper.vm.save();

    expect(wrapper.vm.formError).to.equal(null);
    const put = http.requests.find((r) => r.method === 'put');
    expect(put, 'PUT request').to.not.equal(undefined);
    expect(put.url).to.equal('/api/project/7/integration_groups/3');
    expect(put.data.auth_override).to.equal(false);
    expect(put.data.auth_method).to.equal('');
    expect(put.data.auth_header).to.equal('');
    expect(put.data.auth_secret_id).to.equal(null);
    expect(put.data.endpoint_id).to.equal(4);
  });

  it('refuses to save without an endpoint', async () => {
    http.respond((c) => respondTo(c.url));
    const wrapper = mount(IntegrationGroupForm, mountOptions({ projectId: 7, itemId: 'new' }));
    await flush();

    wrapper.vm.item.name = 'hooks';
    await flush();
    await wrapper.vm.save();

    expect(http.requests.find((r) => r.method === 'post')).to.equal(undefined);
  });
});
