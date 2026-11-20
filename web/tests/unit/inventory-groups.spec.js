import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import InventoryGroups from '@/views/project/InventoryGroups.vue';
import InventoryGroupForm from '@/components/InventoryGroupForm.vue';
import EventBus from '@/event-bus';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const groups = [
  {
    id: 1,
    name: 'web',
    description: 'all frontends',
    variable_count: 2,
    children: [2, 3],
    host_count: 12,
  },
  {
    id: 2, name: 'web-eu', description: '', variable_count: 0, children: [4], host_count: 7,
  },
  {
    id: 3, name: 'web-us', description: '', variable_count: 0, children: [], host_count: 5,
  },
  {
    id: 4, name: 'web-eu-1', description: '', variable_count: 0, children: [], host_count: 3,
  },
  {
    id: 5, name: 'db', description: '', variable_count: 0, children: [], host_count: 4,
  },
];

const wrappers = [];

function mountGroups(props) {
  const wrapper = mount(InventoryGroups, {
    props: { projectId: 7, inventoryId: 3, ...props },
    global: {
      mocks: { $t: (k) => k },
      stubs: { EditDialog: true, YesNoDialog: true },
    },
  });
  wrappers.push(wrapper);
  return wrapper;
}

function mountForm(itemId) {
  const wrapper = mount(InventoryGroupForm, {
    props: {
      projectId: 7, inventoryId: 3, itemId, groups,
    },
    global: {
      mocks: { $t: (k) => k },
      stubs: { CodeMirror: true },
    },
  });
  wrappers.push(wrapper);
  return wrapper;
}

describe('InventoryGroups.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond((c) => (c.method === 'get' ? groups : {}));
  });

  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
    http.restore();
  });

  it('shows the hosts, the child groups and the variables of each group', async () => {
    const wrapper = mountGroups({ isAdmin: true });
    await flush();
    const rows = wrapper.findAll('tbody tr');
    const hosts = wrapper.findAllComponents('[data-testid="inventoryGroups-hosts"]');

    expect(http.requests[0].url).to.equal('/api/project/7/inventory/3/groups');
    expect(wrapper.findAll('[data-testid="inventoryGroups-name"]').map((e) => e.text()))
      .to.deep.equal(['web', 'web-eu', 'web-us', 'web-eu-1', 'db']);
    expect(hosts.map((e) => e.text())).to.deep.equal(['12', '7', '5', '3', '4']);
    expect(hosts[0].props('to')).to.equal('/project/7/inventory/3/hosts?group=1');
    expect(rows[0].findAll('[data-testid="inventoryGroups-child"]').map((e) => e.text()))
      .to.deep.equal(['web-eu', 'web-us']);
    expect(wrapper.findAll('[data-testid="inventoryGroups-variables"]').map((e) => e.text()))
      .to.deep.equal(['2']);
    expect(hosts[0].attributes('aria-label')).to.equal('inventory_groups_hosts_label');
    expect(wrapper.find('[data-testid="inventoryGroups-rowEdit"]').attributes('aria-label'))
      .to.equal('inventory_groups_edit');
    expect(wrapper.find('[data-testid="inventoryGroups-rowDelete"]').attributes('aria-label'))
      .to.equal('inventory_groups_delete');
  });

  it('does not open a row', async () => {
    const wrapper = mountGroups({ isAdmin: true });
    await flush();

    await wrapper.find('[data-testid="inventoryGroups-name"]').trigger('click');
    await wrapper.find('[data-testid="inventoryGroups-variables"]').trigger('click');

    expect(wrapper.findAll('.v-data-table__tr--clickable')).to.have.length(0);
    expect(wrapper.find('[data-testid="inventoryGroups-table"] [aria-expanded]').exists())
      .to.equal(false);
    expect(wrapper.find('[data-testid="inventoryGroups-details"]').exists()).to.equal(false);
    expect(wrapper.findAll('[data-testid="inventoryGroups-noVariables"]')).to.have.length(4);
  });

  it('finds groups by name', async () => {
    const wrapper = mountGroups({ isAdmin: true });
    await flush();

    await wrapper.find('[data-testid="inventoryGroups-search"] input').setValue('EU');

    expect(wrapper.findAll('[data-testid="inventoryGroups-name"]').map((e) => e.text()))
      .to.deep.equal(['web-eu', 'web-eu-1']);
  });

  it('gives a read-only user no add, edit or delete buttons', async () => {
    const wrapper = mountGroups({ isAdmin: false, userPermissions: 0 });
    await flush();

    expect(wrapper.find('[data-testid="inventoryGroups-add"]').exists()).to.equal(false);
    expect(wrapper.find('[data-testid="inventoryGroups-rowEdit"]').exists()).to.equal(false);
    expect(wrapper.find('[data-testid="inventoryGroups-rowDelete"]').exists()).to.equal(false);
  });

  it('forgets the groups and closes the dialogs when the inventory changes', async () => {
    const wrapper = mountGroups({ isAdmin: true });
    await flush();
    await wrapper.find('[data-testid="inventoryGroups-rowDelete"]').trigger('click');
    await wrapper.find('[data-testid="inventoryGroups-rowEdit"]').trigger('click');
    http.respond(() => new Promise(() => {}));

    await wrapper.setProps({ inventoryId: 4 });

    expect(wrapper.findAll('[data-testid="inventoryGroups-name"]')).to.have.length(0);
    expect(wrapper.findComponent({ name: 'YesNoDialog' }).props('modelValue')).to.equal(false);
    expect(wrapper.findComponent({ name: 'EditDialog' }).props('modelValue')).to.equal(false);
    expect(http.requests.at(-1).url).to.equal('/api/project/7/inventory/4/groups');
  });

  it('asks to count again after a group is deleted', async () => {
    const wrapper = mountGroups({ isAdmin: true });
    await flush();
    let counted = 0;
    const onCounts = () => { counted += 1; };
    EventBus.$on('i-inventory-counts', onCounts);

    await wrapper.find('[data-testid="inventoryGroups-rowDelete"]').trigger('click');
    wrapper.findComponent({ name: 'YesNoDialog' }).vm.$emit('yes');
    await flush();
    EventBus.$off('i-inventory-counts', onCounts);

    expect(http.requests.find((r) => r.method === 'delete').url)
      .to.equal('/api/project/7/inventory/3/groups/1');
    expect(counted).to.equal(1);
  });
});

describe('InventoryGroupForm.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond((c) => (c.method === 'get' ? groups.find((g) => c.url.endsWith(`/${g.id}`)) : {}));
  });

  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
    http.restore();
  });

  [
    { name: 'web-eu', hint: true },
    { name: 'web_eu', hint: false },
    { name: 'вэб', hint: false },
  ].forEach(({ name, hint }) => {
    it(`${hint ? 'warns' : 'does not warn'} about the name ${name}`, async () => {
      const wrapper = mountForm('new');
      await flush();

      await wrapper.find('[data-testid="inventoryGroupForm-name"] input').setValue(name);

      expect(wrapper.find('[data-testid="inventoryGroupForm-name"]').text()
        .includes('inventory_groups_name_hint')).to.equal(hint);
    });
  });

  it('offers neither the group nor its ancestors as child groups', async () => {
    const wrapper = mountForm(4);
    await flush();

    const children = wrapper.findComponent('[data-testid="inventoryGroupForm-children"]');
    expect(children.props('items').map((g) => g.name)).to.deep.equal(['web-us', 'db']);
  });

  it('does not save a group that has not loaded yet', async () => {
    const wrapper = mountForm(1);

    await wrapper.setProps({ needSave: true });

    expect(wrapper.emitted('error')).to.have.length(1);
    await flush();
    expect(http.requests.filter((r) => r.method !== 'get')).to.have.length(0);
  });

  it('saves the fields and the change of child groups in one request', async () => {
    const wrapper = mountForm(1);
    await flush();

    wrapper.findComponent('[data-testid="inventoryGroupForm-children"]')
      .vm.$emit('update:modelValue', [3, 5]);
    await wrapper.setProps({ needSave: true });
    await flush();

    const writes = http.requests.filter((r) => r.method !== 'get');
    expect(writes).to.have.length(1);
    expect(writes[0].method).to.equal('put');
    expect(writes[0].url).to.equal('/api/project/7/inventory/3/groups/1');
    expect(writes[0].data).to.include({ name: 'web', description: 'all frontends' });
    expect(writes[0].data.add_children).to.deep.equal([5]);
    expect(writes[0].data.remove_children).to.deep.equal([2]);
    expect(writes[0].data).to.not.have.property('children');
  });
});
