import './setup';
import { expect } from 'chai';
import fs from 'node:fs';
import { reactive } from 'vue';
import { mount } from '@vue/test-utils';
import axios from 'axios';
import InventoryHosts from '@/views/project/InventoryHosts.vue';
import HostStatus from '@/components/HostStatus.vue';
import InventoryHostForm from '@/components/InventoryHostForm.vue';
import EventBus from '@/event-bus';
import { messages } from '@/lang';
import mockAxios, { httpError } from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const hosts = [
  {
    id: 1,
    name: 'web1',
    description: 'frontend',
    enabled: true,
    variable_count: 3,
    groups: [2, 3],
  },
  {
    id: 2, name: 'web2', description: '', enabled: false, variable_count: 0, groups: [],
  },
];

const groups = [
  { id: 1, name: 'db', children: [] },
  { id: 2, name: 'eu', children: [] },
  { id: 3, name: 'web', children: [] },
];

const sources = [
  { id: 1, name: 'Static servers', variables: '' },
  { id: 2, name: 'Edge devices', variables: '' },
];

const sourceGroups = [
  ...groups,
  {
    id: 4, name: 'imported', children: [], source_id: 1,
  },
  {
    id: 5, name: 'other-imported', children: [], source_id: 2,
  },
];

const importedHost = {
  id: 1,
  name: 'web1',
  description: 'frontend',
  enabled: true,
  variables: 'a: 1',
  variable_count: 1,
  groups: [2, 4],
  source_id: 1,
};

// The text shows the values put into a message, to tell the numbers and names apart
const t = (key, values) => [key, ...Object.values(values || {})].join(' ');

const wrappers = [];
let route;

function mountHosts(props) {
  const wrapper = mount(InventoryHosts, {
    props: { projectId: 7, inventoryId: 3, ...props },
    global: {
      mocks: {
        $t: t,
        $route: route,
        $router: { replace: ({ query }) => { route.query = query; } },
      },
      stubs: {
        EditDialog: true,
        YesNoDialog: true,
        RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' },
        TaskLink: { props: ['taskId', 'label'], template: '<a data-testid="taskLink">{{ taskId }} {{ label }}</a>' },
        VDialog: { props: ['modelValue'], template: '<div v-if="modelValue"><slot /></div>' },
      },
    },
  });
  wrappers.push(wrapper);
  return wrapper;
}

function mountForm(itemId, props) {
  const wrapper = mount(InventoryHostForm, {
    props: {
      projectId: 7, inventoryId: 3, itemId, groups, ...props,
    },
    global: {
      mocks: { $t: t },
      stubs: { CodeMirror: true },
    },
  });
  wrappers.push(wrapper);
  return wrapper;
}

async function selectBoth(wrapper) {
  const boxes = wrapper.findAll('[data-testid="inventoryHosts-select"] input');
  await boxes[0].trigger('click');
  await boxes[1].trigger('click');
}

describe('InventoryHosts.vue', () => {
  let http;

  beforeEach(() => {
    route = reactive({ query: {} });
    http = mockAxios();
    const adapter = axios.defaults.adapter;
    axios.defaults.adapter = async (config) => ({
      ...(await adapter(config)),
      headers: { 'x-total-count': '137' },
    });
    http.respond((c) => {
      if (c.method !== 'get') return {};
      if (c.url.endsWith('/sources')) return sources;
      return c.url.endsWith('/groups') ? groups : hosts;
    });
  });

  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
    http.restore();
  });

  const hostRequests = () => http.requests
    .filter((r) => r.method === 'get' && r.url.endsWith('/hosts'));

  it('renders the hosts and the range from the total count', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(http.requests[0].url).to.equal('/api/project/7/inventory/3/hosts');
    expect(wrapper.findAll('[data-testid="inventoryHosts-name"]').map((e) => e.text()))
      .to.deep.equal(['web1', 'web2']);
    expect(wrapper.find('[data-testid="inventoryHosts-range"]').text()).to.contain('137');
    expect(wrapper.find('[data-testid="inventoryHosts-enabled"] input').attributes('aria-label'))
      .to.equal('inventory_hosts_enabled_label web1');
    expect(wrapper.find('[data-testid="inventoryHosts-rowEdit"]').attributes('aria-label'))
      .to.equal('inventory_hosts_edit');
    expect(wrapper.find('[data-testid="inventoryHosts-rowDelete"]').attributes('aria-label'))
      .to.equal('inventory_hosts_delete');
    expect(wrapper.find('[data-testid="inventoryHosts-select"] input').attributes('aria-label'))
      .to.equal('inventory_hosts_select web1');
    expect(wrapper.find('thead input').attributes('aria-label'))
      .to.equal('inventory_hosts_select_all');
  });

  it('drops the groups and closes the dialogs when the inventory changes', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    wrapper.vm.editDialog = true;
    wrapper.vm.groupsDialog = true;

    await wrapper.setProps({ inventoryId: 4 });

    expect(wrapper.vm.groups).to.equal(null);
    expect(wrapper.vm.editDialog).to.equal(false);
    expect(wrapper.vm.groupsDialog).to.equal(false);
  });

  it('shows the direct groups of each host as chips', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    const rows = wrapper.findAll('tbody tr');

    expect(http.requests.some((r) => r.url === '/api/project/7/inventory/3/groups')).to.equal(true);
    expect(rows[0].findAll('[data-testid="inventoryHosts-groups"]').map((e) => e.text()))
      .to.deep.equal(['eu', 'web']);
    expect(rows[1].findAll('[data-testid="inventoryHosts-groups"]')).to.have.length(0);
  });

  it('filters the hosts by a group and keeps the filter in the address', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    const filter = wrapper.findComponent('[data-testid="inventoryHosts-groupFilter"]');

    expect(filter.props('items').map((item) => item.title)).to.deep.equal([
      'inventory_hosts_any_group', 'inventory_hosts_no_groups', 'db', 'eu', 'web',
    ]);
    filter.vm.$emit('update:modelValue', 2);
    await flush();

    expect(route.query.group).to.equal(2);
    expect(hostRequests().at(-1).params).to.include({ group: 2, offset: 0 });
  });

  it('shows the hosts of a server that knows no groups', async () => {
    http.respond((c) => (c.url.endsWith('/groups') ? groups : hosts.map(({ groups: g, ...h }) => h)));
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(wrapper.findAll('[data-testid="inventoryHosts-name"]').map((e) => e.text()))
      .to.deep.equal(['web1', 'web2']);
  });

  it('shows no chip for a group whose name is not known', async () => {
    http.respond((c) => (c.url.endsWith('/groups') ? [] : [{ ...hosts[1], groups: [1] }]));
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-groups"]').exists()).to.equal(false);
    expect(wrapper.findAll('.v-data-table__tr--clickable')).to.have.length(0);
  });

  it('drops a group filter that is not an id', async () => {
    route.query = { group: 'abc' };
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(route.query.group).to.equal(undefined);
    expect(hostRequests().at(-1).params.group).to.equal(undefined);
    expect(wrapper.findAll('[data-testid="inventoryHosts-name"]')).to.have.length(2);
  });

  it('drops a filter by a group that is gone', async () => {
    route.query = { group: '9' };
    http.respond((c) => {
      if (c.url.endsWith('/groups')) return groups;
      if (c.params.group) throw httpError(404, { error: 'not found' });
      return hosts;
    });
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(route.query.group).to.equal(undefined);
    expect(hostRequests().at(-1).params.group).to.equal(undefined);
    expect(wrapper.findAll('[data-testid="inventoryHosts-name"]')).to.have.length(2);
  });

  it('shows how many hosts are selected', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-selectionBar"]').exists()).to.equal(false);
    await selectBoth(wrapper);

    const bar = wrapper.find('[data-testid="inventoryHosts-selectionBar"]');
    expect(bar.text()).to.contain('inventory_hosts_selected 2');
    expect(bar.find('[role="status"]').text()).to.equal('inventory_hosts_selected 2');
    expect(wrapper.find('[data-testid="inventoryHosts-clearSelection"]').attributes('aria-label'))
      .to.equal('inventory_hosts_clear_selection');
  });

  it('disables both buttons of the dialog while its request runs', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    let release;
    http.respond((c) => {
      if (c.method === 'put') return new Promise((resolve) => { release = resolve; });
      return c.url.endsWith('/groups') ? groups : hosts;
    });

    await selectBoth(wrapper);
    await wrapper.find('[data-testid="inventoryHosts-addToGroups"]').trigger('click');
    wrapper.findComponent('[data-testid="inventoryHosts-groupsDialogGroups"]')
      .vm.$emit('update:modelValue', [1]);
    await flush();
    await wrapper.find('[data-testid="inventoryHosts-groupsDialogApply"]').trigger('click');
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-groupsDialogApply"]').element.disabled)
      .to.equal(true);
    expect(wrapper.find('[data-testid="inventoryHosts-groupsDialogCancel"]').element.disabled)
      .to.equal(true);
    release({});
    await flush();
  });

  it('shows the selection bar in place of the search row', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    const find = (id) => wrapper.find(`[data-testid="inventoryHosts-${id}"]`);
    const inert = (id) => find(id).element.closest('[inert]') !== null;

    await selectBoth(wrapper);
    expect(find('selectionBar').exists()).to.equal(true);
    expect(find('toolbar').element.style.visibility).to.equal('hidden');
    expect(inert('search')).to.equal(true);
    expect(inert('groupFilter')).to.equal(true);
    expect(inert('add')).to.equal(true);

    await find('clearSelection').trigger('click');
    expect(find('selectionBar').exists()).to.equal(false);
    expect(find('toolbar').element.style.visibility).to.equal('');
    expect(inert('search')).to.equal(false);
    expect(inert('groupFilter')).to.equal(false);
    expect(inert('add')).to.equal(false);
  });

  it('adds the selected hosts to groups in one request', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    let counted = 0;
    const onCounts = () => { counted += 1; };
    EventBus.$on('i-inventory-counts', onCounts);

    await selectBoth(wrapper);
    await wrapper.find('[data-testid="inventoryHosts-addToGroups"]').trigger('click');
    const apply = () => wrapper.find('[data-testid="inventoryHosts-groupsDialogApply"]');
    expect(apply().element.disabled).to.equal(true);
    wrapper.findComponent('[data-testid="inventoryHosts-groupsDialogGroups"]')
      .vm.$emit('update:modelValue', [1, 2]);
    await flush();
    await apply().trigger('click');
    await flush();
    EventBus.$off('i-inventory-counts', onCounts);

    const writes = http.requests.filter((r) => r.method !== 'get');
    expect(writes).to.have.length(1);
    expect(writes[0].method).to.equal('put');
    expect(writes[0].url).to.equal('/api/project/7/inventory/3/host_groups');
    expect(writes[0].data).to.deep.equal({ hosts: [1, 2], add: [1, 2], remove: [] });
    expect(wrapper.find('[data-testid="inventoryHosts-groupsDialog"]').exists()).to.equal(false);
    expect(wrapper.find('[data-testid="inventoryHosts-selectionBar"]').exists()).to.equal(false);
    expect(counted).to.equal(1);
  });

  it('offers to remove only the groups a selected host is directly in', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    await selectBoth(wrapper);
    await wrapper.find('[data-testid="inventoryHosts-removeFromGroups"]').trigger('click');

    const select = wrapper.findComponent('[data-testid="inventoryHosts-groupsDialogGroups"]');
    expect(select.props('items').map((g) => g.name)).to.deep.equal(['eu', 'web']);
  });

  it('keeps the dialog open with the error of a failed change and reloads', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    http.respond((c) => {
      if (c.method === 'put') throw httpError(400, { error: 'no such group' });
      return c.url.endsWith('/groups') ? groups.slice(1) : hosts;
    });

    await selectBoth(wrapper);
    await wrapper.find('[data-testid="inventoryHosts-addToGroups"]').trigger('click');
    wrapper.findComponent('[data-testid="inventoryHosts-groupsDialogGroups"]')
      .vm.$emit('update:modelValue', [1, 2]);
    await flush();
    await wrapper.find('[data-testid="inventoryHosts-groupsDialogApply"]').trigger('click');
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-groupsDialog"]').text())
      .to.contain('no such group');
    expect(wrapper.findComponent('[data-testid="inventoryHosts-groupsDialogGroups"]')
      .props('modelValue')).to.deep.equal([2]);
    expect(wrapper.find('[data-testid="inventoryHosts-selectionBar"]').text())
      .to.contain('inventory_hosts_selected 2');
  });

  it('clears the selection when the search changes', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    await selectBoth(wrapper);

    route.query = { search: 'web' };
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-selectionBar"]').exists()).to.equal(false);
  });

  it('forgets a selected host that is no longer in the list', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    await selectBoth(wrapper);

    http.respond((c) => (c.url.endsWith('/groups') ? groups : [hosts[0]]));
    await wrapper.vm.load();
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-selectionBar"]').text())
      .to.contain('inventory_hosts_selected 1');
  });

  it('shows the stored number of variables', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(wrapper.findAll('[data-testid="inventoryHosts-variables"]').map((e) => e.text()))
      .to.deep.equal(['3']);
    expect(wrapper.findAll('[data-testid="inventoryHosts-noVariables"]')).to.have.length(1);
  });

  it('does not open a row', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    await wrapper.find('[data-testid="inventoryHosts-name"]').trigger('click');
    await wrapper.find('[data-testid="inventoryHosts-variables"]').trigger('click');

    expect(wrapper.findAll('.v-data-table__tr--clickable')).to.have.length(0);
    expect(wrapper.find('[data-testid="inventoryHosts-table"] [aria-expanded]').exists())
      .to.equal(false);
    expect(wrapper.find('[data-testid="inventoryHosts-details"]').exists()).to.equal(false);
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
    expect(wrapper.find('[data-testid="inventoryHosts-select"]').exists()).to.equal(false);
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

  it('shows the source of an imported host under its name and none for a manual one', async () => {
    http.respond((c) => {
      if (c.url.endsWith('/sources')) return sources;
      return c.url.endsWith('/groups') ? groups : [{ ...hosts[0], source_id: 2 }, hosts[1]];
    });
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-source-1"]').text()).to.equal('Edge devices');
    expect(wrapper.find('[data-testid="inventoryHosts-source-2"]').exists()).to.equal(false);
  });

  it('names the source of an imported host even when the sources did not load', async () => {
    http.respond((c) => {
      if (c.url.endsWith('/sources')) throw httpError(500, 'boom');
      return c.url.endsWith('/groups') ? groups : [{ ...hosts[0], source_id: 2 }];
    });
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-source-1"]').text())
      .to.equal('inventory_source_unknown');
  });

  it('keeps the source filter reachable when the address names a source that is not there', async () => {
    http.respond((c) => {
      if (c.url.endsWith('/sources')) return [];
      return c.url.endsWith('/groups') ? groups : hosts;
    });
    route.query = { source: '5' };
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-sourceFilter"]').exists()).to.equal(true);
  });

  it('shows no unknown source before the sources answer', async () => {
    let answer;
    http.respond((c) => {
      if (c.url.endsWith('/sources')) return new Promise((resolve) => { answer = resolve; });
      return c.url.endsWith('/groups') ? groups : [importedHost];
    });
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-source-1"]').text()).to.equal('');
    answer(sources);
    await flush();
    expect(wrapper.find('[data-testid="inventoryHosts-source-1"]').text()).to.equal('Static servers');
  });

  it('filters the hosts by the manual ones and by a source', async () => {
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    const filter = wrapper.findComponent('[data-testid="inventoryHosts-sourceFilter"]');

    expect(filter.props('items').map((item) => item.value)).to.deep.equal(['', 'manual', 1, 2]);
    filter.vm.$emit('update:modelValue', 'manual');
    await flush();
    expect(route.query.source).to.equal('manual');
    expect(hostRequests().at(-1).params).to.include({ source: 'manual', offset: 0 });

    filter.vm.$emit('update:modelValue', 2);
    await flush();
    expect(hostRequests().at(-1).params).to.include({ source: 2 });

    filter.vm.$emit('update:modelValue', '');
    await flush();
    expect(hostRequests().at(-1).params.source).to.equal(undefined);
  });

  it('shows no source filter when the inventory has no sources', async () => {
    http.respond((c) => {
      if (c.url.endsWith('/sources')) return [];
      return c.url.endsWith('/groups') ? groups : hosts;
    });
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    expect(wrapper.find('[data-testid="inventoryHosts-sourceFilter"]').exists()).to.equal(false);
  });

  it('offers no delete for an imported host', async () => {
    http.respond((c) => {
      if (c.url.endsWith('/sources')) return sources;
      return c.url.endsWith('/groups') ? groups : [importedHost, hosts[1]];
    });
    const wrapper = mountHosts({ isAdmin: true });
    await flush();

    const rows = wrapper.findAll('[data-testid="inventoryHosts-rowActions"]');
    expect(rows[0].find('[data-testid="inventoryHosts-rowDelete"]').exists()).to.equal(false);
    expect(rows[0].find('[data-testid="inventoryHosts-rowEdit"]').exists()).to.equal(true);
    expect(rows[1].find('[data-testid="inventoryHosts-rowDelete"]').exists()).to.equal(true);
  });

  it('offers only manual groups for the selected hosts', async () => {
    http.respond((c) => {
      if (c.url.endsWith('/sources')) return sources;
      return c.url.endsWith('/groups') ? sourceGroups : [importedHost, hosts[1]];
    });
    const wrapper = mountHosts({ isAdmin: true });
    await flush();
    await selectBoth(wrapper);

    await wrapper.find('[data-testid="inventoryHosts-addToGroups"]').trigger('click');
    const select = () => wrapper.findComponent('[data-testid="inventoryHosts-groupsDialogGroups"]');
    expect(select().props('items').map((g) => g.name)).to.deep.equal(['db', 'eu', 'web']);

    await wrapper.find('[data-testid="inventoryHosts-groupsDialogCancel"]').trigger('click');
    await wrapper.find('[data-testid="inventoryHosts-removeFromGroups"]').trigger('click');
    expect(select().props('items').map((g) => g.name)).to.deep.equal(['eu']);
  });

  it('uses only translation keys that exist', () => {
    const files = ['src/views/project/InventoryHosts.vue', 'src/components/InventoryHostForm.vue'];
    const keys = files.flatMap((file) => [...fs.readFileSync(file, 'utf8')
      .matchAll(/\$t\(\s*'([^']+)'/g)].map((m) => m[1]));

    expect(keys).to.include('inventory_source_unknown');
    expect(keys.filter((key) => !(key in messages.en))).to.deep.equal([]);
  });

  describe('with the hosts feature', () => {
    const runHosts = [
      {
        id: 1,
        name: 'web1',
        description: '',
        enabled: true,
        variable_count: 0,
        groups: [],
        host_id: 12,
        last_status: 'failed',
        last_run: '2026-11-20T10:00:00Z',
        last_task_id: 9,
      },
      {
        id: 2,
        name: 'web2',
        description: '',
        enabled: true,
        variable_count: 0,
        groups: [],
        host_id: null,
        last_status: null,
        last_run: null,
        last_task_id: null,
      },
    ];

    beforeEach(() => {
      http.respond((c) => {
        if (c.method !== 'get') return {};
        return c.url.endsWith('/groups') ? groups : runHosts;
      });
    });

    it('hides the last run and the name link without the feature', async () => {
      const wrapper = mountHosts({ isAdmin: true });
      await flush();

      expect(wrapper.find('[data-testid="inventoryHosts-lastRun"]').exists()).to.equal(false);
      expect(wrapper.find('[data-testid="inventoryHosts-neverRun"]').exists()).to.equal(false);
      expect(wrapper.find('[data-testid="inventoryHosts-name"] a').exists()).to.equal(false);
    });

    it('shows the status and the task link of the last run', async () => {
      const wrapper = mountHosts({ isAdmin: true, systemInfo: { features: { hosts: true } } });
      await flush();

      const cell = wrapper.get('[data-testid="inventoryHosts-lastRun"]');
      expect(cell.findComponent(HostStatus).props('status')).to.equal('failed');
      expect(cell.get('[data-testid="taskLink"]').text()).to.match(/^9 /);
      expect(wrapper.findAll('[data-testid="inventoryHosts-neverRun"]')).to.have.length(1);
      expect(wrapper.get('[data-testid="inventoryHosts-neverRun"]').text())
        .to.equal('inventory_hosts_never_run');
    });

    it('links the name to the project host when it has one', async () => {
      const wrapper = mountHosts({ isAdmin: true, systemInfo: { features: { hosts: true } } });
      await flush();

      const names = wrapper.findAll('[data-testid="inventoryHosts-name"]');
      expect(names[0].get('a').attributes('href')).to.equal('/project/7/hosts/12');
      expect(names[1].find('a').exists()).to.equal(false);
    });
  });
});

describe('InventoryHostForm.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond((c) => (c.method === 'get' ? hosts[0] : {}));
  });

  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
    http.restore();
  });

  it('opens a host of a server that knows no groups', async () => {
    http.respond(() => ({ ...hosts[0], groups: undefined }));
    const wrapper = mountForm(1);
    await flush();

    expect(wrapper.find('[data-testid="inventoryHostForm-error"]').text()).to.equal('');
    expect(wrapper.find('[data-testid="inventoryHostForm-groups"]').exists()).to.equal(true);
  });

  it('does not save a host that has not loaded yet', async () => {
    const wrapper = mountForm(1);

    await wrapper.setProps({ needSave: true });

    expect(wrapper.emitted('error')).to.have.length(1);
    await flush();
    expect(http.requests.filter((r) => r.method !== 'get')).to.have.length(0);
  });

  it('saves the fields and the change of groups in one request', async () => {
    const wrapper = mountForm(1);
    await flush();

    wrapper.findComponent('[data-testid="inventoryHostForm-groups"]')
      .vm.$emit('update:modelValue', [3, 1]);
    await wrapper.setProps({ needSave: true });
    await flush();

    const writes = http.requests.filter((r) => r.method !== 'get');
    expect(writes).to.have.length(1);
    expect(writes[0].method).to.equal('put');
    expect(writes[0].url).to.equal('/api/project/7/inventory/3/hosts/1');
    expect(writes[0].data).to.include({ name: 'web1', description: 'frontend' });
    expect(writes[0].data.add_groups).to.deep.equal([1]);
    expect(writes[0].data.remove_groups).to.deep.equal([2]);
    expect(writes[0].data).to.not.have.property('groups');
  });

  describe('an imported host', () => {
    beforeEach(() => {
      http.respond((c) => (c.method === 'get' ? { ...importedHost } : {}));
    });

    const mountImported = () => mountForm(1, {
      groups: sourceGroups,
      sources: { 1: 'Static servers', 2: 'Edge devices' },
    });

    it('says where it comes from and does not let name and variables change', async () => {
      const wrapper = mountImported();
      await flush();

      expect(wrapper.find('[data-testid="inventoryHostForm-imported"]').text())
        .to.equal('inventory_source_imported Static servers');
      expect(wrapper.find('[data-testid="inventoryHostForm-name"] input').attributes())
        .to.have.property('readonly');
      expect(wrapper.findComponent('[data-testid="inventoryHostForm-variables"]').props('readonly'))
        .to.equal(true);
    });

    it('marks the read-only name and variables with a lock and a dimmed editor', async () => {
      const wrapper = mountImported();
      await flush();

      expect(wrapper.find('[data-testid="inventoryHostForm-name"] .mdi-lock').exists()).to.equal(true);
      expect(wrapper.find('.InventoryHostForm__editor--readonly').exists()).to.equal(true);
    });

    it('marks nothing as locked for a manual host', async () => {
      http.respond((c) => (c.method === 'get' ? { ...hosts[0] } : {}));
      const wrapper = mountForm(1);
      await flush();

      expect(wrapper.find('[data-testid="inventoryHostForm-name"] .mdi-lock').exists()).to.equal(false);
      expect(wrapper.find('.InventoryHostForm__editor--readonly').exists()).to.equal(false);
    });

    it('names the lock of a group of the source', async () => {
      const wrapper = mountImported();
      await flush();

      const chip = wrapper.find('[aria-label="inventory_source_group_locked imported"]');
      expect(chip.exists()).to.equal(true);
    });

    it('sends the stored name and variables and only manual groups', async () => {
      const wrapper = mountImported();
      await flush();

      const groupsField = wrapper.findComponent('[data-testid="inventoryHostForm-groups"]');
      expect(groupsField.props('items').map((g) => g.name)).to.deep.equal(['db', 'eu', 'web', 'imported']);
      groupsField.vm.$emit('update:modelValue', [4, 3]);
      await wrapper.find('[data-testid="inventoryHostForm-description"] input').setValue('new note');
      await wrapper.setProps({ needSave: true });
      await flush();

      const [put] = http.requests.filter((r) => r.method !== 'get');
      expect(put.data).to.include({ name: 'web1', variables: 'a: 1', description: 'new note' });
      expect(put.data.add_groups).to.deep.equal([3]);
      expect(put.data.remove_groups).to.deep.equal([2]);
    });

    it('does not remove a group of the source', async () => {
      const wrapper = mountImported();
      await flush();

      wrapper.findComponent('[data-testid="inventoryHostForm-groups"]')
        .vm.$emit('update:modelValue', [2]);
      await wrapper.setProps({ needSave: true });
      await flush();

      const [put] = http.requests.filter((r) => r.method !== 'get');
      expect(put.data.remove_groups).to.deep.equal([]);
    });
  });
});
