import './setup';
import { expect } from 'chai';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TemplateDetails from '@/views/project/template/TemplateDetails.vue';

describe('TemplateDetails.vue', () => {
  it('renders without the project resources', () => {
    const localVue = createLocalVue();
    const wrapper = shallowMount(TemplateDetails, {
      localVue,
      vuetify: new Vuetify(),
      propsData: {
        template: {
          id: 7, project_id: 1, repository_id: 3, inventory_id: 4, environment_ids: [5], app: 'ansible', type: '',
        },
        repositories: [],
        inventory: [],
        environment: [],
      },
      mocks: { $t: (k) => k },
      stubs: ['TaskStats'],
    });

    expect(wrapper.text()).to.include('—');
    wrapper.destroy();
  });
});
