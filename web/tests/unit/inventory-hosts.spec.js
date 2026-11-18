import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import axios from 'axios';
import InventoryHosts from '@/views/project/InventoryHosts.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const hosts = [
  {
    id: 1, name: 'web1', description: 'frontend', enabled: true, variables: 'a: 1',
  },
  {
    id: 2, name: 'web2', description: '', enabled: false, variables: '',
  },
];

function mountHosts(props) {
  return mount(InventoryHosts, {
    props: { projectId: 7, inventoryId: 3, ...props },
    global: {
      mocks: {
        $t: (k) => k,
        $route: { query: {} },
        $router: { replace: () => {} },
      },
      stubs: { EditDialog: true, YesNoDialog: true },
    },
  });
}

describe('InventoryHosts.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    const adapter = axios.defaults.adapter;
    axios.defaults.adapter = async (config) => ({
      ...(await adapter(config)),
      headers: { 'x-total-count': '137' },
    });
    http.respond((c) => (c.method === 'get' ? hosts : {}));
  });

  afterEach(() => {
    http.restore();
  });

  it('renders the hosts and the range from the total count', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(http.requests[0].url).to.equal('/api/project/7/inventory/3/hosts');
    expect(wrapper.findAll('[data-testid="inventoryHosts-name"]').map((e) => e.text()))
      .to.deep.equal(['web1', 'web2']);
    expect(wrapper.find('[data-testid="inventoryHosts-range"]').text()).to.contain('137');
    expect([
      wrapper.find('[data-testid="inventoryHosts-enabled"] input').attributes('aria-label'),
      wrapper.find('[data-testid="inventoryHosts-rowEdit"]').attributes('aria-label'),
      wrapper.find('[data-testid="inventoryHosts-rowDelete"]').attributes('aria-label'),
    ]).to.deep.equal(['inventory_hosts_enabled_label', 'inventory_hosts_edit', 'inventory_hosts_delete']);
  });

  it('sends only enabled false when the switch is turned off', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    await wrapper.find('[data-testid="inventoryHosts-enabled"] input').setValue(false);
    await flush();

    const put = http.requests.find((r) => r.method === 'put');
    expect(put.url).to.equal('/api/project/7/inventory/3/hosts/1/enabled');
    expect(put.data).to.deep.equal({ enabled: false });
  });

  it('gives a read-only user no add button and a disabled switch', async () => {
    const wrapper = mountHosts({ isAdmin: false, userPermissions: 0 });
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-add"]').exists()).to.equal(false);
    expect(wrapper.find('[data-testid="inventoryHosts-rowEdit"]').exists()).to.equal(false);
    expect(wrapper.find('[data-testid="inventoryHosts-enabled"] input').element.disabled)
      .to.equal(true);
  });

  it('steps one page back when the last host of the last page is deleted', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    wrapper.vm.offset = 50;
    http.respond((c) => {
      if (c.method !== 'get') return {};
      return c.params.offset === 50 ? [] : hosts;
    });

    await wrapper.vm.load();
    await flush();

    expect(wrapper.vm.offset).to.equal(0);
    expect(wrapper.findAll('[data-testid="inventoryHosts-name"]')).to.have.lengthOf(2);
  });

  it('disables the switch of a row while its request is pending', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    let release;
    http.respond((c) => (c.method === 'put'
      ? new Promise((resolve) => { release = resolve; })
      : hosts));

    await wrapper.find('[data-testid="inventoryHosts-enabled"] input').setValue(false);
    await flush();
    expect(wrapper.find('[data-testid="inventoryHosts-enabled"] input').element.disabled)
      .to.equal(true);

    release({});
    await flush();
    expect(wrapper.find('[data-testid="inventoryHosts-enabled"] input').element.disabled)
      .to.equal(false);
  });

  it('disables the switches while the rows reload', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    let release;
    http.respond(() => new Promise((resolve) => { release = resolve; }));

    wrapper.vm.load();
    await flush();
    expect(wrapper.findAll('[data-testid="inventoryHosts-enabled"] input').map((e) => e.element.disabled))
      .to.deep.equal([true, true]);

    release(hosts);
    await flush();
    expect(wrapper.find('[data-testid="inventoryHosts-enabled"] input').element.disabled)
      .to.equal(false);
  });
});
