import './setup';
import fs from 'node:fs';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import InventoryGroups from '@/views/project/InventoryGroups.vue';
import InventoryGroupForm from '@/components/InventoryGroupForm.vue';
import EventBus from '@/event-bus';
import { messages } from '@/lang';
import EditDialog from '@/components/EditDialog.vue';
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

const sources = [
  { id: 1, name: 'Static servers' },
  { id: 2, name: 'Edge devices' },
];

const withSources = [
  ...groups.slice(0, 2),
  {
    id: 3,
    name: 'imported-a',
    description: '',
    variable_count: 0,
    children: [],
    host_count: 1,
    source_id: 1,
  },
  {
    id: 4,
    name: 'imported-b',
    description: '',
    variable_count: 0,
    children: [],
    host_count: 1,
    source_id: 2,
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

function mountForm(itemId, props) {
  const wrapper = mount(InventoryGroupForm, {
    props: {
      projectId: 7, inventoryId: 3, itemId, groups, ...props,
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
    http.respond((c) => {
      if (c.method !== 'get') return {};
      return c.url.endsWith('/sources') ? [] : groups;
    });
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
    expect(http.requests.filter((r) => r.url.endsWith('/groups')).at(-1).url)
      .to.equal('/api/project/7/inventory/4/groups');
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

  describe('with imported groups', () => {
    beforeEach(() => {
      http.respond((c) => {
        if (c.method !== 'get') return {};
        return c.url.endsWith('/sources') ? sources : withSources;
      });
    });

    const names = (wrapper) => wrapper.findAll('[data-testid="inventoryGroups-name"]')
      .map((e) => e.text());

    it('shows the source under the name of an imported group and none for a manual one', async () => {
      const wrapper = mountGroups({ isAdmin: true });
      await flush();

      expect(wrapper.find('[data-testid="inventoryGroups-source-3"]').text()).to.equal('Static servers');
      expect(wrapper.find('[data-testid="inventoryGroups-source-1"]').exists()).to.equal(false);
    });

    it('filters the groups by the manual ones and by a source', async () => {
      const wrapper = mountGroups({ isAdmin: true });
      await flush();
      const filter = () => wrapper.findComponent('[data-testid="inventoryGroups-sourceFilter"]');

      expect(filter().props('items').map((item) => item.value)).to.deep.equal(['', 'manual', 1, 2]);
      filter().vm.$emit('update:modelValue', 'manual');
      await flush();
      expect(names(wrapper)).to.deep.equal(['web', 'web-eu']);

      filter().vm.$emit('update:modelValue', 2);
      await flush();
      expect(names(wrapper)).to.deep.equal(['imported-b']);
    });

    it('offers no delete for an imported group', async () => {
      const wrapper = mountGroups({ isAdmin: true });
      await flush();

      const rows = wrapper.findAll('[data-testid="inventoryGroups-rowActions"]');
      expect(rows[0].find('[data-testid="inventoryGroups-rowDelete"]').exists()).to.equal(true);
      expect(rows[2].find('[data-testid="inventoryGroups-rowDelete"]').exists()).to.equal(false);
      expect(rows[2].find('[data-testid="inventoryGroups-rowEdit"]').exists()).to.equal(true);
    });

    it('gives the dialog of an imported group no save button', async () => {
      const wrapper = mountGroups({ isAdmin: true });
      await flush();
      const dialog = () => wrapper.findComponent(EditDialog);

      await wrapper.findAll('[data-testid="inventoryGroups-rowEdit"]')[2].trigger('click');
      expect(dialog().props('saveButtonText')).to.equal(undefined);

      await wrapper.findAll('[data-testid="inventoryGroups-rowEdit"]')[0].trigger('click');
      expect(dialog().props('saveButtonText')).to.equal('save');
    });

    it('renders no save button without a save text', async () => {
      const wrapper = mount(EditDialog, {
        props: { modelValue: false, title: 'x', saveButtonText: undefined },
        global: {
          mocks: { $t: (k) => k },
          stubs: { VDialog: { props: ['modelValue'], template: '<div v-if="modelValue"><slot /></div>' } },
        },
      });
      wrappers.push(wrapper);
      await wrapper.setProps({ modelValue: true });
      await flush();

      expect(wrapper.find('[data-testid="editDialog-cancel"]').exists()).to.equal(true);
      expect(wrapper.find('[data-testid="editDialog-save"]').exists()).to.equal(false);
    });
  });

  it('uses only translation keys that exist', () => {
    const files = ['src/views/project/InventoryGroups.vue', 'src/components/InventoryGroupForm.vue'];
    const keys = files.flatMap((file) => [...fs.readFileSync(file, 'utf8')
      .matchAll(/\$t\(\s*'([^']+)'/g)].map((m) => m[1]));

    expect(keys).to.include('inventory_source_unknown');
    expect(keys.filter((key) => !(key in messages.en))).to.deep.equal([]);
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

  it('marks an imported group as locked and a manual one not', async () => {
    http.respond(() => ({ ...withSources[2], variables: 'a: 1' }));
    const imported = mountForm(3);
    await flush();
    http.respond(() => ({ ...groups[0], variables: 'a: 1' }));
    const manual = mountForm(1);
    await flush();

    ['name', 'description'].forEach((field) => {
      const lock = `[data-testid="inventoryGroupForm-${field}"] .mdi-lock`;
      expect(imported.find(lock).exists()).to.equal(true);
      expect(manual.find(lock).exists()).to.equal(false);
    });
    expect(imported.find('.InventoryGroupForm__editor--readonly').exists()).to.equal(true);
    expect(manual.find('.InventoryGroupForm__editor--readonly').exists()).to.equal(false);
  });

  it('shows an imported group read-only with the source it comes from', async () => {
    http.respond(() => ({ ...withSources[2], variables: 'a: 1' }));
    const wrapper = mountForm(3, { sources: { 1: 'Static servers' } });
    await flush();

    expect(wrapper.find('[data-testid="inventoryGroupForm-imported"]').text())
      .to.equal('inventory_source_imported_group');
    ['name', 'description'].forEach((field) => {
      expect(wrapper.find(`[data-testid="inventoryGroupForm-${field}"] input`).attributes())
        .to.have.property('readonly');
    });
    expect(wrapper.findComponent('[data-testid="inventoryGroupForm-children"]').props('readonly'))
      .to.equal(true);
    expect(wrapper.findComponent('[data-testid="inventoryGroupForm-variables"]').props('readonly'))
      .to.equal(true);
  });
});
