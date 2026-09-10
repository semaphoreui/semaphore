import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import ItemFormBase from '@/components/ItemFormBase';
import mockAxios, { httpError } from './helpers/axiosMock';

let formValid = true;

const FormStub = {
  render: (h) => h('form'),
  methods: {
    validate: () => formValid,
    resetValidation: () => {},
  },
};

const Host = {
  mixins: [ItemFormBase],
  components: { FormStub },
  render(h) {
    return h('div', [h('form-stub', { ref: 'form' })]);
  },
  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/keys`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/keys/${this.itemId}`;
    },
    getNewItem() {
      return { name: '', type: 'none' };
    },
  },
};

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountHost(propsData) {
  return mount(Host, { propsData: { projectId: 7, ...propsData } });
}

describe('ItemFormBase mixin', () => {
  let http;

  beforeEach(() => {
    formValid = true;
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  it('initialises a new item without requesting the backend', async () => {
    const wrapper = mountHost({ itemId: 'new' });
    await flush();

    expect(wrapper.vm.isNew).to.equal(true);
    expect(wrapper.vm.item).to.deep.equal({ name: '', type: 'none' });
    expect(http.requests).to.have.lengthOf(0);
  });

  it('loads an existing item from the single item URL', async () => {
    http.respond(() => ({ id: 3, name: 'deploy key' }));
    const wrapper = mountHost({ itemId: 3 });
    await flush();

    expect(http.requests).to.have.lengthOf(1);
    expect(http.requests[0].method).to.equal('get');
    expect(http.requests[0].url).to.equal('/api/project/7/keys/3');
    expect(wrapper.vm.item).to.deep.equal({ id: 3, name: 'deploy key' });
  });

  it('reports a load error through formError and the error event', async () => {
    http.respond(() => { throw httpError(404, { error: 'Key not found' }); });
    const wrapper = mountHost({ itemId: 3 });
    await flush();

    expect(wrapper.vm.item).to.equal(null);
    expect(wrapper.vm.formError).to.equal('Key not found');
    expect(wrapper.emitted().error[0][0]).to.deep.equal({ message: 'Key not found' });
  });

  it('creates a new item with POST and emits save', async () => {
    http.respond(() => ({ id: 10, name: 'k' }));
    const wrapper = mountHost({ itemId: 'new' });
    await flush();
    wrapper.vm.item.name = 'k';

    const result = await wrapper.vm.save({ extra: true });

    const req = http.requests[0];
    expect(req.method).to.equal('post');
    expect(req.url).to.equal('/api/project/7/keys');
    expect(req.data).to.deep.equal({
      name: 'k', type: 'none', project_id: 7, extra: true,
    });
    expect(result).to.deep.equal({ id: 10, name: 'k' });
    expect(wrapper.emitted().save[0][0]).to.deep.equal({
      item: { id: 10, name: 'k' },
      action: 'new',
    });
    expect(wrapper.vm.formSaving).to.equal(false);
  });

  it('updates an existing item with PUT', async () => {
    http.respond((config) => (config.method === 'get' ? { id: 3, name: 'old' } : ''));
    const wrapper = mountHost({ itemId: 3 });
    await flush();
    wrapper.vm.item.name = 'new name';

    const result = await wrapper.vm.save();

    const req = http.requests[1];
    expect(req.method).to.equal('put');
    expect(req.url).to.equal('/api/project/7/keys/3');
    expect(req.data).to.deep.equal({ id: 3, name: 'new name', project_id: 7 });
    // Empty response body: the local item is returned instead.
    expect(result).to.deep.equal({ id: 3, name: 'new name' });
    expect(wrapper.emitted().save[0][0].action).to.equal('edit');
  });

  it('does not send anything when the form is invalid', async () => {
    formValid = false;
    const wrapper = mountHost({ itemId: 'new' });
    await flush();

    const result = await wrapper.vm.save();

    expect(result).to.equal(null);
    expect(http.requests).to.have.lengthOf(0);
    expect(wrapper.emitted().error[0][0]).to.deep.equal({});
    expect(wrapper.emitted().save).to.equal(undefined);
  });

  it('reports a save error and resets formSaving', async () => {
    http.respond(() => { throw httpError(400, { error: 'Name is required' }); });
    const wrapper = mountHost({ itemId: 'new' });
    await flush();

    await wrapper.vm.save();

    expect(wrapper.vm.formError).to.equal('Name is required');
    expect(wrapper.vm.formSaving).to.equal(false);
    expect(wrapper.emitted().error[0][0]).to.deep.equal({ message: 'Name is required' });
    expect(wrapper.emitted().save).to.equal(undefined);
  });

  it('runs the beforeSave/afterSave hooks around the request', async () => {
    const calls = [];
    const HookedHost = {
      ...Host,
      methods: {
        ...Host.methods,
        beforeSave() { calls.push('before'); },
        afterSave(item) { calls.push(`after:${item.id}`); },
      },
    };
    http.respond(() => ({ id: 1 }));
    const wrapper = mount(HookedHost, { propsData: { projectId: 7, itemId: 'new' } });
    await flush();

    await wrapper.vm.save();

    expect(calls).to.deep.equal(['before', 'after:1']);
  });

  it('saves when the needSave prop becomes true', async () => {
    http.respond(() => ({ id: 1 }));
    const wrapper = mountHost({ itemId: 'new' });
    await flush();

    await wrapper.setProps({ needSave: true });
    await flush();

    expect(http.requests).to.have.lengthOf(1);
    expect(wrapper.emitted().save).to.have.lengthOf(1);
  });

  it('reloads data when the needReset prop becomes true', async () => {
    http.respond(() => ({ id: 3, name: 'fresh' }));
    const wrapper = mountHost({ itemId: 3 });
    await flush();
    wrapper.vm.item.name = 'dirty';
    wrapper.vm.formError = 'stale';

    await wrapper.setProps({ needReset: true });
    await flush();

    expect(http.requests).to.have.lengthOf(2);
    expect(wrapper.vm.item.name).to.equal('fresh');
    expect(wrapper.vm.formError).to.equal(null);
  });
});
