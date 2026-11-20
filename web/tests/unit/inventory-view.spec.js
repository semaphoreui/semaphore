import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import axios from 'axios';
import InventoryView from '@/views/project/InventoryView.vue';
import EventBus from '@/event-bus';
import mockAxios, { httpError } from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const wrappers = [];

function mountView() {
  const wrapper = mount(InventoryView, {
    props: { projectId: 7 },
    global: {
      mocks: { $t: (k) => k, $route: { params: { inventoryId: '3' } } },
      stubs: { RouterLink: true, RouterView: true },
    },
  });
  wrappers.push(wrapper);
  return wrapper;
}

describe('InventoryView.vue', () => {
  let http;

  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
    http.restore();
  });

  it('shows the number of inventory variables on the tab', async () => {
    http = mockAxios();
    http.respond(() => ({ id: 3, type: 'managed', variable_count: 2 }));
    const wrapper = mountView();
    await flush();

    expect(wrapper.find('[data-testid="inventoryView-variableCount"]').text()).to.equal('2');
  });

  it('reads the number again when the variables of this inventory are saved', async () => {
    http = mockAxios();
    http.respond(() => ({ id: 3, type: 'managed', variable_count: 1 }));
    const wrapper = mountView();
    await flush();
    const chip = () => wrapper.find('[data-testid="inventoryView-variableCount"]');

    http.respond(() => ({ id: 3, type: 'managed', variable_count: 2 }));
    const loaded = http.requests.length;
    EventBus.$emit('i-inventory-variables', { projectId: 7, inventoryId: 4 });
    await flush();
    expect(chip().text()).to.equal('1');
    expect(http.requests).to.have.length(loaded);

    EventBus.$emit('i-inventory-variables', { projectId: 7, inventoryId: 3 });
    await flush();
    expect(chip().text()).to.equal('2');

    http.respond(() => ({ id: 3, type: 'managed', variable_count: 0 }));
    EventBus.$emit('i-inventory-variables', { projectId: 7, inventoryId: 3 });
    await flush();
    expect(chip().exists()).to.equal(false);
    expect(wrapper.find('[data-testid="inventoryView-page"]').attributes('data-state')).to.equal('ready');
  });

  it('keeps the number of the last save when an older answer comes late', async () => {
    http = mockAxios();
    http.respond(() => ({ id: 3, type: 'managed', variable_count: 0 }));
    const wrapper = mountView();
    await flush();

    let answerFirst;
    http.respond(() => new Promise((resolve) => { answerFirst = resolve; }));
    EventBus.$emit('i-inventory-variables', { projectId: 7, inventoryId: 3 });
    await flush();
    http.respond(() => ({ id: 3, type: 'managed', variable_count: 2 }));
    EventBus.$emit('i-inventory-variables', { projectId: 7, inventoryId: 3 });
    await flush();
    answerFirst({ id: 3, type: 'managed', variable_count: 1 });
    await flush();

    expect(wrapper.find('[data-testid="inventoryView-variableCount"]').text()).to.equal('2');
  });

  it('shows no number for an inventory without variables', async () => {
    http = mockAxios();
    http.respond(() => ({ id: 3, type: 'managed', variable_count: 0 }));
    const wrapper = mountView();
    await flush();

    expect(wrapper.find('[data-testid="inventoryView-variableCount"]').exists()).to.equal(false);
  });

  describe('the numbers of hosts and groups', () => {
    let hostTotal;
    let groupList;

    beforeEach(() => {
      hostTotal = '137';
      groupList = [{ id: 1 }, { id: 2 }, { id: 3 }, { id: 4 }];
      http = mockAxios();
      const adapter = axios.defaults.adapter;
      axios.defaults.adapter = async (config) => {
        const response = await adapter(config);
        return config.url.endsWith('/hosts')
          ? { ...response, headers: { 'x-total-count': hostTotal } }
          : response;
      };
      http.respond((c) => {
        if (c.url.endsWith('/hosts')) return [];
        if (c.url.endsWith('/groups')) return groupList;
        return { id: 3, type: 'managed', variable_count: 3 };
      });
    });

    const text = (wrapper, id) => wrapper.find(`[data-testid="inventoryView-${id}"]`).text();

    it('shows the numbers of hosts, groups and variables on the tabs', async () => {
      const wrapper = mountView();
      await flush();

      const hostsRequest = http.requests.find((r) => r.url.endsWith('/hosts'));
      expect(hostsRequest.params).to.deep.equal({ count: 1 });
      expect(text(wrapper, 'hostCount')).to.equal('137');
      expect(text(wrapper, 'groupCount')).to.equal('4');
      expect(text(wrapper, 'variableCount')).to.equal('3');
      expect(wrapper.find('[data-testid="inventoryView-tab-groups"]').exists()).to.equal(true);
    });

    it('counts again when a tab asks for it', async () => {
      const wrapper = mountView();
      await flush();

      hostTotal = '0';
      groupList = [{ id: 1 }];
      EventBus.$emit('i-inventory-counts');
      await flush();

      expect(wrapper.find('[data-testid="inventoryView-hostCount"]').exists()).to.equal(false);
      expect(text(wrapper, 'groupCount')).to.equal('1');

      hostTotal = undefined;
      EventBus.$emit('i-inventory-counts');
      await flush();

      expect(text(wrapper, 'hostCount')).to.equal('?');
    });

    it('shows a question mark for a number that could not be loaded', async () => {
      http.respond((c) => {
        if (c.url.endsWith('/hosts')) throw httpError(500, { error: 'boom' });
        if (c.url.endsWith('/groups')) return groupList;
        return { id: 3, type: 'managed', variables: '' };
      });
      const wrapper = mountView();
      await flush();

      expect(text(wrapper, 'hostCount')).to.equal('?');
      expect(text(wrapper, 'groupCount')).to.equal('4');
    });

    it('keeps the later numbers when an earlier answer comes last', async () => {
      const wrapper = mountView();
      await flush();
      const releases = [];
      http.respond((c) => new Promise((resolve) => {
        const data = c.url.endsWith('/groups') ? groupList : [];
        releases.push(() => resolve(data));
      }));

      groupList = [{ id: 1 }, { id: 2 }];
      EventBus.$emit('i-inventory-counts');
      await flush();
      groupList = [{ id: 1 }];
      EventBus.$emit('i-inventory-counts');
      await flush();
      releases[2]();
      releases[3]();
      await flush();
      releases[0]();
      releases[1]();
      await flush();

      expect(text(wrapper, 'groupCount')).to.equal('1');
    });
  });
});
