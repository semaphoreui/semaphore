import { expect } from 'chai';
import fs from 'node:fs';
import { vi } from 'vitest';
import { createI18n } from 'vue-i18n';
import { messages } from '@/lang';
import { mount } from '@vue/test-utils';
import InventorySources from '@/views/project/InventorySources.vue';
import InventorySourceForm from '@/components/InventorySourceForm.vue';
import EventBus from '@/event-bus';
import createTestVuetify from './setup';
import mockAxios, { httpError } from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const ago = (ms) => new Date(Date.now() - ms).toISOString();

const sources = [
  {
    id: 1,
    name: 'Static servers',
    type: 'git',
    repository_id: 5,
    params: { path: 'inventory/prod.yml' },
    status: 'success',
    last_sync: ago(12 * 60 * 1000),
    last_success: ago(12 * 60 * 1000),
    last_error: '',
    host_count: 1204,
    group_count: 18,
    last_result: {
      hosts_added: 10,
      hosts_changed: 3,
      hosts_removed: 1,
      groups_added: 2,
      groups_changed: 0,
      groups_removed: 0,
      adopted: 0,
      variables_changed: true,
      skipped: [
        { kind: 'host', name: 'fe80::1', reason: 'invalid name' },
        { kind: 'host', name: 'b', reason: 'owned by other' },
        { kind: 'group', name: 'c', reason: 'reserved' },
        { kind: 'group', name: 'd', reason: 'reserved' },
      ],
      skipped_count: 5,
    },
  },
  {
    id: 2,
    name: 'Edge devices',
    type: 'git',
    repository_id: 6,
    params: { path: '' },
    status: 'failed',
    last_sync: ago(2 * 60 * 60 * 1000),
    last_success: ago(3 * 24 * 60 * 60 * 1000),
    last_error: 'Repository clone failed: authentication required',
    host_count: 96,
    group_count: 4,
    last_result: null,
  },
  {
    id: 3,
    name: 'Lab',
    type: 'git',
    repository_id: 7,
    params: { path: 'inventory.yml' },
    status: 'syncing',
    last_sync: null,
    last_success: null,
    last_error: '',
    host_count: 0,
    group_count: 0,
    last_result: null,
  },
  {
    id: 4,
    name: 'Fresh',
    type: 'git',
    repository_id: 5,
    params: { path: 'a.ini' },
    status: '',
    last_sync: null,
    last_success: null,
    last_error: '',
    host_count: 0,
    group_count: 0,
    last_result: null,
  },
];

const repositories = [
  { id: 5, name: 'infra', git_branch: 'main' },
  { id: 6, name: 'edge-inventory', git_branch: 'main' },
  { id: 7, name: 'lab', git_branch: 'develop' },
];

const wrappers = [];

function mountSources(props, vuetify, t) {
  const wrapper = mount(InventorySources, {
    props: { projectId: 7, inventoryId: 3, ...props },
    global: {
      mocks: { $t: t || ((k, p) => (p ? `${k} ${JSON.stringify(p)}` : k)) },
      stubs: { EditDialog: true, YesNoDialog: true },
      plugins: vuetify ? [vuetify] : [],
    },
  });
  wrappers.push(wrapper);
  return wrapper;
}

const text = (wrapper, id) => wrapper.find(`[data-testid="${id}"]`).text();
const has = (wrapper, id) => wrapper.find(`[data-testid="${id}"]`).exists();

describe('InventorySources.vue', () => {
  let http;
  let list;

  beforeEach(() => {
    list = sources;
    http = mockAxios();
    http.respond((c) => {
      if (c.url.endsWith('/repositories')) return repositories;
      if (c.method === 'get') return list;
      return {};
    });
  });

  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
    http.restore();
    vi.useRealTimers();
  });

  it('shows the status, the outcome and the error of each source', async () => {
    const wrapper = mountSources({ isAdmin: true });
    await flush();

    expect(text(wrapper, 'inventorySources-card-1')).to.contain('Static servers');
    expect(text(wrapper, 'inventorySources-card-1')).to.contain('infra · main · inventory/prod.yml');
    expect(text(wrapper, 'inventorySources-status-1')).to.contain('inventory_sources_status_success');
    expect(text(wrapper, 'inventorySources-result-1')).to.contain('1204').and.to.contain('18');
    expect(text(wrapper, 'inventorySources-result-1')).to.contain('"added":12,"changed":3,"removed":1');
    expect(text(wrapper, 'inventorySources-result-1')).to.contain('inventory_sources_variables_changed');
    expect(text(wrapper, 'inventorySources-skipped-1')).to.contain('fe80::1 (invalid name), b (owned by other), c (reserved), …');
    expect(has(wrapper, 'inventorySources-error-1')).to.equal(false);

    expect(text(wrapper, 'inventorySources-status-2')).to.contain('inventory_sources_status_failed');
    expect(text(wrapper, 'inventorySources-card-2')).to.contain('edge-inventory · main · /');
    expect(text(wrapper, 'inventorySources-error-2')).to.contain('authentication required');
    expect(text(wrapper, 'inventorySources-result-2')).to.contain('inventory_sources_kept');
    expect(text(wrapper, 'inventorySources-result-2')).to.contain('96').and.to.contain('4');

    expect(text(wrapper, 'inventorySources-status-3')).to.contain('inventory_sources_status_syncing');
    expect(has(wrapper, 'inventorySources-result-3')).to.equal(false);
    expect(wrapper.find('[data-testid="inventorySources-sync-3"]').attributes('disabled')).to.not.equal(undefined);

    expect(has(wrapper, 'inventorySources-status-4')).to.equal(false);
    expect(has(wrapper, 'inventorySources-skipped-4')).to.equal(false);
  });

  it('starts a sync and reads the list again', async () => {
    const wrapper = mountSources({ isAdmin: true });
    await flush();

    await wrapper.find('[data-testid="inventorySources-sync-1"]').trigger('click');
    await flush();

    const post = http.requests.find((r) => r.method === 'post');
    expect(post.url).to.equal('/api/project/7/inventory/3/sources/1/sync');
    expect(http.requests.filter((r) => r.url.endsWith('/sources'))).to.have.length(2);
  });

  it('only reads the list again when the source is already syncing', async () => {
    http.respond((c) => {
      if (c.url.endsWith('/repositories')) return repositories;
      if (c.method === 'post') throw httpError(409, { error: 'source is already syncing' });
      return list;
    });
    const events = [];
    const onSnack = (e) => events.push(e);
    EventBus.$on('i-snackbar', onSnack);
    const wrapper = mountSources({ isAdmin: true });
    await flush();

    await wrapper.find('[data-testid="inventorySources-sync-1"]').trigger('click');
    await flush();
    EventBus.$off('i-snackbar', onSnack);

    expect(events).to.deep.equal([]);
    expect(http.requests.filter((r) => r.url.endsWith('/sources'))).to.have.length(2);
  });

  it('shows no controls to a viewer', async () => {
    const wrapper = mountSources({ isAdmin: false, userPermissions: 0 });
    await flush();

    expect(has(wrapper, 'inventorySources-card-1')).to.equal(true);
    expect(has(wrapper, 'inventorySources-add')).to.equal(false);
    expect(has(wrapper, 'inventorySources-sync-1')).to.equal(false);
    expect(has(wrapper, 'inventorySources-menu-1')).to.equal(false);
  });

  it('names the source and the numbers in the delete dialog', async () => {
    const wrapper = mountSources({ isAdmin: true });
    await flush();

    wrapper.vm.askDelete(sources[0]);
    await flush();
    const dialog = wrapper.findComponent({ name: 'YesNoDialog' });

    expect(dialog.props('modelValue')).to.equal(true);
    expect(dialog.props('text')).to.contain('Static servers').and.to.contain('1204');

    dialog.vm.$emit('yes');
    await flush();
    const del = http.requests.find((r) => r.method === 'delete');
    expect(del.url).to.equal('/api/project/7/inventory/3/sources/1');
  });

  it('puts the Sync button on its own line on a phone', async () => {
    window.innerWidth = 400;
    const wrapper = mountSources({ isAdmin: true }, createTestVuetify());
    await flush();
    window.innerWidth = 1024;

    expect(wrapper.find('[data-testid="inventorySources-card-1"] .InventorySources__phoneActions [data-testid="inventorySources-sync-1"]').exists()).to.equal(true);
  });

  it('offers a refresh when the first read fails', async () => {
    http.respond((c) => {
      if (c.url.endsWith('/repositories')) return repositories;
      throw httpError(500, { error: 'database is down' });
    });
    const wrapper = mountSources({ isAdmin: true });
    await flush();

    expect(has(wrapper, 'inventorySources-empty')).to.equal(false);
    http.respond((c) => (c.url.endsWith('/repositories') ? repositories : list));
    await wrapper.find('[data-testid="inventorySources-refresh"]').trigger('click');
    await flush();

    expect(has(wrapper, 'inventorySources-card-1')).to.equal(true);
  });

  it('shows the type and the restored hosts of a source that never synced', async () => {
    list = [{
      ...sources[3], type: 'servicenow', host_count: 7, group_count: 2,
    }];
    const wrapper = mountSources({ isAdmin: true });
    await flush();

    expect(text(wrapper, 'inventorySources-card-4')).to.contain('servicenow');
    expect(text(wrapper, 'inventorySources-result-4')).to.contain('inventory_sources_totals');
  });

  describe('polling', () => {
    // setTimeout is fake here, so the pending answers are waited for with setImmediate
    const settle = async () => {
      await new Promise((resolve) => { setImmediate(resolve); });
      await new Promise((resolve) => { setImmediate(resolve); });
    };
    const reads = () => http.requests.filter((r) => r.url.endsWith('/sources')).length;

    beforeEach(() => {
      vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    });

    it('reads the list every 2 seconds while a source syncs and then stops', async () => {
      const counts = [];
      const onCounts = () => counts.push(1);
      EventBus.$on('i-inventory-counts', onCounts);
      const wrapper = mountSources({ isAdmin: true });
      await settle();

      expect(reads()).to.equal(1);
      await vi.advanceTimersByTimeAsync(2000);
      await settle();
      expect(reads()).to.equal(2);

      list = sources.map((s) => (s.id === 3 ? { ...s, status: 'success', last_sync: ago(1000) } : s));
      await vi.advanceTimersByTimeAsync(2000);
      await settle();
      expect(reads()).to.equal(3);
      expect(counts).to.have.length(1);
      expect(text(wrapper, 'inventorySources-status-3')).to.contain('inventory_sources_status_success');

      await vi.advanceTimersByTimeAsync(10000);
      expect(reads()).to.equal(3);
      EventBus.$off('i-inventory-counts', onCounts);
    });

    it('refreshes the counts after a sync that ended between two reads', async () => {
      const counts = [];
      const onCounts = () => counts.push(1);
      EventBus.$on('i-inventory-counts', onCounts);
      mountSources({ isAdmin: true });
      await settle();

      list = sources.map((s) => (s.id === 1 ? { ...s, last_sync: ago(1000) } : s));
      await vi.advanceTimersByTimeAsync(2000);
      await settle();
      EventBus.$off('i-inventory-counts', onCounts);

      expect(counts).to.have.length(1);
    });

    it('starts the next read only after the previous one answered', async () => {
      const wrapper = mountSources({ isAdmin: true });
      await settle();
      let answer;
      http.respond(() => new Promise((resolve) => { answer = resolve; }));

      await vi.advanceTimersByTimeAsync(2000);
      await vi.advanceTimersByTimeAsync(10000);
      expect(reads()).to.equal(2);

      answer(sources.map((s) => (s.id === 3 ? { ...s, status: 'success', last_sync: ago(1000) } : s)));
      await settle();
      expect(text(wrapper, 'inventorySources-status-3')).to.contain('inventory_sources_status_success');
    });

    it('shows no error when a background read fails', async () => {
      const events = [];
      const onSnack = (e) => events.push(e);
      EventBus.$on('i-snackbar', onSnack);
      mountSources({ isAdmin: true });
      await settle();
      http.respond(() => { throw httpError(500, { error: 'database is down' }); });

      await vi.advanceTimersByTimeAsync(2000);
      await settle();
      EventBus.$off('i-snackbar', onSnack);

      expect(reads()).to.equal(2);
      expect(events).to.deep.equal([]);
    });

    it('stops when the page goes away', async () => {
      const wrapper = mountSources({ isAdmin: true });
      await settle();
      wrapper.unmount();
      const before = http.requests.length;

      await vi.advanceTimersByTimeAsync(10000);

      expect(http.requests.length).to.equal(before);
    });
  });
});

describe('Sources tab plurals', () => {
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en: messages.en } });
  const one = { ...sources[0], host_count: 1, group_count: 1 };
  const many = { ...sources[0], host_count: 2, group_count: 2 };
  let http;

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
    http.restore();
  });

  [[one, '1 host, 1 group', '1 host and 1 group'], [many, '2 hosts, 2 groups', '2 hosts and 2 groups']]
    .forEach(([source, line, confirm]) => {
      it(`writes ${line}`, async () => {
        http.respond((c) => (c.method === 'get' ? [source] : {}));
        const wrapper = mountSources({}, undefined, i18n.global.t);
        await flush();

        expect(text(wrapper, 'inventorySources-result-1')).to.contain(line);
        wrapper.vm.askDelete(source);
        await flush();
        expect(wrapper.findComponent({ name: 'YesNoDialog' }).props('text')).to.contain(confirm);
      });
    });
});

describe('Sources tab translations', () => {
  it('uses only translation keys that exist', () => {
    const files = [
      'src/views/project/InventorySources.vue',
      'src/components/InventorySourceForm.vue',
      'src/views/project/InventoryView.vue',
    ];
    const keys = files.flatMap((file) => [...fs.readFileSync(file, 'utf8')
      .matchAll(/\$t\(\s*'([^']+)'/g)].map((m) => m[1]));
    const dynamic = ['success', 'failed', 'syncing'].map((s) => `inventory_sources_status_${s}`);

    const missing = [...keys, ...dynamic].filter((key) => !(key in messages.en));

    expect(keys).to.include('inventory_sources_delete');
    expect(missing).to.deep.equal([]);
  });
});

describe('InventorySourceForm.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond((c) => {
      if (c.url.endsWith('/repositories')) return repositories;
      return {};
    });
  });

  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
    http.restore();
  });

  it('creates a git source with the chosen repository and path', async () => {
    const wrapper = mount(InventorySourceForm, {
      props: { projectId: 7, inventoryId: 3, itemId: 'new' },
      global: { mocks: { $t: (k) => k } },
    });
    wrappers.push(wrapper);
    await flush();

    const field = (id) => wrapper.findComponent(`[data-testid="inventorySourceForm-${id}"]`);
    field('name').vm.$emit('update:modelValue', 'Static');
    field('repository').vm.$emit('update:modelValue', 5);
    field('path').vm.$emit('update:modelValue', 'inventory/prod.yml');
    await flush();
    await wrapper.setProps({ needSave: true });
    await flush();

    const post = http.requests.find((r) => r.method === 'post');
    expect(post.url).to.equal('/api/project/7/inventory/3/sources');
    expect(post.data).to.deep.include({
      name: 'Static', type: 'git', repository_id: 5, params: { path: 'inventory/prod.yml' },
    });
  });

  it('does not save a source without a path', async () => {
    const wrapper = mount(InventorySourceForm, {
      props: { projectId: 7, inventoryId: 3, itemId: 'new' },
      global: { mocks: { $t: (k) => k } },
    });
    wrappers.push(wrapper);
    await flush();

    const field = (id) => wrapper.findComponent(`[data-testid="inventorySourceForm-${id}"]`);
    field('name').vm.$emit('update:modelValue', 'Static');
    field('repository').vm.$emit('update:modelValue', 5);
    await flush();
    await wrapper.setProps({ needSave: true });
    await flush();

    expect(http.requests.some((r) => r.method === 'post')).to.equal(false);
    expect(field('path').props('placeholder')).to.equal('inventory/prod.yml');
  });
});
