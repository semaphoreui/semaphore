import './setup';
import { expect } from 'chai';
import { h } from 'vue';
import { shallowMount, mount } from '@vue/test-utils';
import ItemListPageBase from '@/components/ItemListPageBase';
import ObjectRefsView from '@/components/ObjectRefsView.vue';
import { USER_PERMISSIONS } from '@/lib/constants';
import mockAxios from './helpers/axiosMock';

const Host = {
  mixins: [ItemListPageBase],
  render: () => h('div'),
  methods: {
    getHeaders() {
      return [{ title: 'Name', key: 'name' }];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/repositories`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/repositories/${this.itemId}`;
    },
    getEventName() {
      return 'i-repository';
    },
  },
};

const sourceRefs = {
  templates: [],
  repositories: [],
  inventories: [],
  access_keys: [],
  schedules: [],
  host_configs: [],
  inventory_sources: [{ id: 3, name: 'Static servers' }],
};

describe('refs of a repository used by an inventory source', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond((config) => (config.url.endsWith('/refs') ? sourceRefs : []));
  });

  afterEach(() => {
    http.restore();
  });

  it('opens the refs dialog instead of the delete confirmation', async () => {
    const wrapper = shallowMount(Host, {
      props: { projectId: 7, userPermissions: USER_PERMISSIONS.manageProjectResources },
    });

    await wrapper.vm.askDeleteItem(5);

    expect(wrapper.vm.itemRefsDialog).to.equal(true);
    expect(wrapper.vm.deleteItemDialog).to.not.equal(true);
  });

  it('renders the source name', () => {
    const wrapper = mount(ObjectRefsView, {
      props: { objectRefs: sourceRefs, projectId: 7, objectTitle: 'repository' },
      global: { mocks: { $t: (k) => k }, stubs: { 'router-link': { template: '<a><slot /></a>' } } },
    });

    expect(wrapper.text()).to.contain('Inventory sources');
    expect(wrapper.text()).to.contain('Static servers');
  });
});
