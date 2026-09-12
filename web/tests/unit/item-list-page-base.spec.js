import './setup';
import { expect } from 'chai';
import { shallowMount } from '@vue/test-utils';
import ItemListPageBase from '@/components/ItemListPageBase';
import EventBus from '@/event-bus';
import { USER_PERMISSIONS } from '@/lib/constants';
import mockAxios, { httpError } from './helpers/axiosMock';

const Host = {
  mixins: [ItemListPageBase],
  render: (h) => h('div'),
  methods: {
    getHeaders() {
      return [
        { text: 'Name', value: 'name' },
        { text: 'Actions', value: 'actions' },
      ];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/keys`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/keys/${this.itemId}`;
    },
    getEventName() {
      return 'i-key';
    },
  },
};

const noRefs = {
  templates: [], repositories: [], inventories: [], access_keys: [], schedules: [],
};

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountHost(propsData = {}) {
  return shallowMount(Host, {
    propsData: {
      projectId: 7,
      userPermissions: USER_PERMISSIONS.manageProjectResources,
      ...propsData,
    },
  });
}

describe('ItemListPageBase mixin', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond(() => []);
  });

  afterEach(() => {
    http.restore();
  });

  it('hides the actions column without manageProjectResources permission', () => {
    const wrapper = mountHost({ userPermissions: USER_PERMISSIONS.runProjectTasks });
    expect(wrapper.vm.headers.map((h) => h.value)).to.deep.equal(['name']);
  });

  it('keeps the actions column with manageProjectResources permission', () => {
    const wrapper = mountHost();
    expect(wrapper.vm.headers.map((h) => h.value)).to.deep.equal(['name', 'actions']);
  });

  it('keeps the actions column for admins', () => {
    const wrapper = mountHost({ userPermissions: 0, isAdmin: true });
    expect(wrapper.vm.headers.map((h) => h.value)).to.deep.equal(['name', 'actions']);
  });

  it('loads items on creation', async () => {
    http.respond(() => [{ id: 1, name: 'a' }]);
    const wrapper = mountHost();
    await flush();

    expect(http.requests[0].method).to.equal('get');
    expect(http.requests[0].url).to.equal('/api/project/7/keys');
    expect(wrapper.vm.items).to.deep.equal([{ id: 1, name: 'a' }]);
  });

  it('opens the edit dialog for the selected item', () => {
    const wrapper = mountHost();
    wrapper.vm.editItem(5);
    expect(wrapper.vm.itemId).to.equal(5);
    expect(wrapper.vm.editDialog).to.equal(true);
  });

  it('asks for confirmation when the item is not referenced', async () => {
    http.respond((config) => (config.url.endsWith('/refs') ? noRefs : []));
    const wrapper = mountHost();
    await flush();

    await wrapper.vm.askDeleteItem(5);

    expect(http.requests[1].url).to.equal('/api/project/7/keys/5/refs');
    expect(wrapper.vm.deleteItemDialog).to.equal(true);
    expect(wrapper.vm.itemRefsDialog).to.equal(null);
  });

  it('shows references instead of the confirmation when the item is in use', async () => {
    const refs = { ...noRefs, templates: [{ id: 1, name: 'T' }] };
    http.respond((config) => (config.url.endsWith('/refs') ? refs : []));
    const wrapper = mountHost();
    await flush();

    await wrapper.vm.askDeleteItem(5);

    expect(wrapper.vm.itemRefsDialog).to.equal(true);
    expect(wrapper.vm.itemRefs).to.deep.equal(refs);
    expect(wrapper.vm.deleteItemDialog).to.equal(null);
  });

  it('falls back to the confirmation when references cannot be loaded', async () => {
    http.respond((config) => {
      if (config.url.endsWith('/refs')) {
        throw httpError(500, {});
      }
      return [];
    });
    const wrapper = mountHost();
    await flush();

    await wrapper.vm.askDeleteItem(5);

    expect(wrapper.vm.deleteItemDialog).to.equal(true);
  });

  it('deletes the item, emits the event and reloads the list', async () => {
    let items = [{ id: 5, name: 'to delete' }, { id: 6, name: 'keep' }];
    http.respond((config) => {
      if (config.method === 'delete') {
        items = items.filter((x) => x.id !== 5);
        return '';
      }
      return items;
    });
    const wrapper = mountHost();
    await flush();

    const events = [];
    const handler = (e) => events.push(e);
    EventBus.$on('i-key', handler);
    try {
      await wrapper.vm.deleteItem(5);
    } finally {
      EventBus.$off('i-key', handler);
    }

    const del = http.requests.find((r) => r.method === 'delete');
    expect(del.url).to.equal('/api/project/7/keys/5');
    expect(events).to.deep.equal([{ action: 'delete', item: { id: 5, name: 'to delete' } }]);
    expect(wrapper.vm.items).to.deep.equal([{ id: 6, name: 'keep' }]);
  });

  it('shows a snackbar when deletion fails', async () => {
    http.respond((config) => {
      if (config.method === 'delete') {
        throw httpError(400, { error: 'Cannot delete' });
      }
      return [{ id: 5 }];
    });
    const wrapper = mountHost();
    await flush();

    const snackbars = [];
    const handler = (e) => snackbars.push(e);
    EventBus.$on('i-snackbar', handler);
    try {
      await wrapper.vm.deleteItem(5);
    } finally {
      EventBus.$off('i-snackbar', handler);
    }

    expect(snackbars).to.deep.equal([{ color: 'error', text: 'Cannot delete' }]);
  });
});
