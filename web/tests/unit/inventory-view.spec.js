import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import InventoryView from '@/views/project/InventoryView.vue';
import EventBus from '@/event-bus';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountView() {
  return mount(InventoryView, {
    props: { projectId: 7 },
    global: {
      mocks: { $t: (k) => k, $route: { params: { inventoryId: '3' } } },
      stubs: { RouterLink: true, RouterView: true },
    },
  });
}

describe('InventoryView.vue', () => {
  let http;

  afterEach(() => {
    http.restore();
  });

  it('shows the number of inventory variables on the tab', async () => {
    http = mockAxios();
    http.respond(() => ({ id: 3, type: 'managed', variables: 'a: 1\nb: 2' }));
    const wrapper = mountView();
    await flush();

    expect(wrapper.find('[data-testid="inventoryView-variableCount"]').text()).to.equal('2');
  });

  it('follows the variables saved for this inventory', async () => {
    http = mockAxios();
    http.respond(() => ({ id: 3, type: 'managed', variables: 'a: 1' }));
    const wrapper = mountView();
    await flush();
    const chip = () => wrapper.find('[data-testid="inventoryView-variableCount"]');

    EventBus.$emit('i-inventory-variables', { projectId: 7, inventoryId: 4, variables: '' });
    await flush();
    expect(chip().text()).to.equal('1');

    EventBus.$emit('i-inventory-variables', { projectId: 7, inventoryId: 3, variables: 'a: 1\nb: 2' });
    await flush();
    expect(chip().text()).to.equal('2');

    EventBus.$emit('i-inventory-variables', { projectId: 7, inventoryId: 3, variables: '' });
    await flush();
    expect(chip().exists()).to.equal(false);
    expect(wrapper.find('[data-testid="inventoryView-page"]').attributes('data-state')).to.equal('ready');
  });

  it('shows no number for an inventory without variables', async () => {
    http = mockAxios();
    http.respond(() => ({ id: 3, type: 'managed', variables: '' }));
    const wrapper = mountView();
    await flush();

    expect(wrapper.find('[data-testid="inventoryView-variableCount"]').exists()).to.equal(false);
  });
});
