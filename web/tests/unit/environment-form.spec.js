import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import EnvironmentForm from '@/components/EnvironmentForm.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const keys = [
  { id: 1, name: 'shared token', type: 'string' },
  { id: 2, name: 'deploy key', type: 'ssh' },
  { id: 3, name: 'registry password', type: 'string' },
];

// One secret with a value of its own, one linked to a Key Store key.
const group = {
  id: 9,
  name: 'prod',
  project_id: 7,
  json: '{}',
  env: '{}',
  secrets: [{
    id: 21, type: 'var', name: 'token', key_id: 40, key_name: 'token', owned: true,
  }, {
    id: 22, type: 'env', name: 'REGISTRY_PASSWORD', key_id: 3, key_name: 'registry password', owned: false,
  }],
};

function respondTo(config) {
  if (config.method !== 'get') return {};
  if (config.url.endsWith('/keys')) return keys;
  if (config.url.endsWith('/secret_storages')) return [];
  if (/\/environment\/\d+$/.test(config.url)) return JSON.parse(JSON.stringify(group));
  return {};
}

function mountForm(props) {
  return mount(EnvironmentForm, {
    props: { projectId: 7, ...props },
    global: {
      mocks: { $t: (k) => k },
      stubs: { CodeMirror: true, RichEditor: true, SecretStorageSyncOptionsForm: true },
    },
  });
}

async function mountOnSecretsTab(props) {
  const wrapper = mountForm(props);
  await flush();
  wrapper.vm.tab = 'secrets';
  await flush();
  return wrapper;
}

describe('EnvironmentForm.vue secrets', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond(respondTo);
  });

  afterEach(() => {
    http.restore();
  });

  it('offers only the string keys of the Key Store', async () => {
    const wrapper = mountForm({ itemId: 'new' });
    await flush();

    expect(wrapper.vm.keys.map((k) => k.id)).to.deep.equal([1, 3]);
  });

  it('shows an owned secret as a value and a linked one as a key', async () => {
    const wrapper = await mountOnSecretsTab({ itemId: 9 });

    const varRows = wrapper.findAll('[data-testid="varGroup-secretVarRow"]');
    expect(varRows).to.have.length(1);
    expect(varRows[0].find('[data-testid="varGroup-secretVarValue"]').exists()).to.equal(true);
    expect(varRows[0].find('[data-testid="varGroup-secretVarKey"]').exists()).to.equal(false);
    expect(varRows[0].find('[data-testid="varGroup-secretVarLinked"]').exists()).to.equal(false);
    expect(varRows[0].find('[data-testid="varGroup-secretSource-var"]').exists())
      .to.equal(true, 'an owned secret can be switched to a key');

    const envRows = wrapper.findAll('[data-testid="varGroup-secretEnvRow"]');
    expect(envRows).to.have.length(1);
    expect(envRows[0].find('[data-testid="varGroup-secretEnvKey"]').exists()).to.equal(true);
    expect(envRows[0].find('[data-testid="varGroup-secretEnvValue"]').exists()).to.equal(false);
    expect(envRows[0].find('[data-testid="varGroup-secretEnvLinked"]').exists()).to.equal(true);
    expect(envRows[0].find('[data-testid="varGroup-secretSource-env"]').exists())
      .to.equal(false, 'a linked secret takes no value');

    expect(wrapper.find('[data-testid="varGroup-secretVarTable"]').attributes('data-state'))
      .to.equal('ready');
    expect(wrapper.find('[data-testid="varGroup-secretVarsWarning"]').exists()).to.equal(true);
  });

  it('switches a new secret between a value and a key', async () => {
    const wrapper = await mountOnSecretsTab({ itemId: 'new' });
    expect(wrapper.find('[data-testid="varGroup-secretEnvTable"]').attributes('data-state'))
      .to.equal('empty');

    await wrapper.find('[data-testid="varGroup-addSecretEnv"]').trigger('click');
    await flush();
    let row = wrapper.find('[data-testid="varGroup-secretEnvRow"]');
    expect(row.find('[data-testid="varGroup-secretEnvValue"]').exists()).to.equal(true);

    await row.find('[data-testid="varGroup-secretSource-env"]').trigger('click');
    await flush();
    row = wrapper.find('[data-testid="varGroup-secretEnvRow"]');
    expect(row.find('[data-testid="varGroup-secretEnvKey"]').exists()).to.equal(true);
    expect(row.find('[data-testid="varGroup-secretEnvValue"]').exists()).to.equal(false);
    expect(wrapper.vm.secrets[0].source).to.equal('key');
  });

  it('removes only the row that was clicked', async () => {
    const wrapper = await mountOnSecretsTab({ itemId: 'new' });
    wrapper.vm.addSecret('var');
    wrapper.vm.addSecret('var');
    wrapper.vm.secrets[1].value = 'kept';
    await flush();

    const rows = wrapper.findAll('[data-testid="varGroup-secretVarRow"]');
    expect(rows).to.have.length(2);
    await rows[0].find('[data-testid="varGroup-secretVarRemove"]').trigger('click');
    await flush();

    expect(wrapper.vm.secrets).to.have.length(1);
    expect(wrapper.vm.secrets[0].value).to.equal('kept');
  });

  it('keeps a link to a key outside the Key Store list visible by its name', async () => {
    const wrapper = await mountOnSecretsTab({ itemId: 9 });

    const items = wrapper.vm.keyItemsFor({ keyId: 77, keyName: 'key of another group' });

    expect(items[0]).to.deep.equal({ id: 77, name: 'key of another group' });
    expect(items).to.have.length(3);
  });

  it('sends a value for an owned secret and a key id for a linked one', async () => {
    const wrapper = await mountOnSecretsTab({ itemId: 9 });
    wrapper.vm.secrets[0].value = 'rotated';
    wrapper.vm.addSecret('env');
    Object.assign(wrapper.vm.secrets[2], { name: 'SHARED', source: 'key', keyId: 1 });
    wrapper.vm.addSecret('var');
    Object.assign(wrapper.vm.secrets[3], { name: 'fresh', value: 'v' });
    wrapper.vm.removeSecret(wrapper.vm.secrets[1]);

    wrapper.vm.beforeSave();

    expect(wrapper.vm.item.secrets).to.deep.equal([{
      id: 21, name: 'token', type: 'var', operation: 'update', secret: 'rotated',
    }, {
      id: undefined, name: 'SHARED', type: 'env', operation: 'create', key_id: 1,
    }, {
      id: undefined, name: 'fresh', type: 'var', operation: 'create', secret: 'v',
    }, {
      id: 22, name: 'REGISTRY_PASSWORD', type: 'env', operation: 'delete',
    }]);
  });

  it('refuses to save a key-backed secret with no key selected', async () => {
    const wrapper = await mountOnSecretsTab({ itemId: 'new' });
    wrapper.vm.addSecret('env');
    Object.assign(wrapper.vm.secrets[0], { name: 'SHARED', source: 'key' });

    expect(() => wrapper.vm.beforeSave()).to.throw('secretKeyRequired');
  });
});
