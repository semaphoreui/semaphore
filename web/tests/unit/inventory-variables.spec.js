import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import InventoryVariables from '@/views/project/InventoryVariables.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountVariables() {
  return mount(InventoryVariables, {
    props: { projectId: 7, inventoryId: 3, isAdmin: true },
    global: {
      mocks: { $t: (k) => k },
      stubs: { CodeMirror: true, YesNoDialog: true },
    },
  });
}

describe('InventoryVariables.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond((c) => (c.method === 'get'
      ? {
        id: 3, name: 'prod', type: 'managed', ssh_key_id: 1, variables: 'a: 1',
      }
      : {}));
  });

  afterEach(() => {
    http.restore();
  });

  it('saves the variables over a fresh copy of the inventory', async () => {
    const wrapper = mountVariables();
    await flush();
    wrapper.vm.inventory.variables = 'a: 2';
    http.respond((c) => (c.method === 'get'
      ? {
        id: 3, name: 'prod', type: 'managed', ssh_key_id: 5, variables: 'a: 1',
      }
      : {}));

    await wrapper.find('[data-testid="inventoryVariables-save"]').trigger('click');
    await flush();

    const put = http.requests.find((r) => r.method === 'put');
    expect(put.url).to.equal('/api/project/7/inventory/3');
    expect(put.data.variables).to.equal('a: 2');
    expect(put.data.ssh_key_id).to.equal(5);
  });
});
