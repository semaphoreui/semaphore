import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import TemplateDetails from '@/views/project/template/TemplateDetails.vue';

describe('TemplateDetails.vue', () => {
  it('renders without the project resources', () => {
    // Mounted whole: a shallow mount of VTU 2 does not render the content of stubbed Vuetify parts.
    const wrapper = mount(TemplateDetails, {
      props: {
        template: {
          id: 7, project_id: 1, repository_id: 3, inventory_id: 4, environment_ids: [5], app: 'ansible', type: '',
        },
        repositories: [],
        inventory: [],
        environment: [],
      },
      global: { mocks: { $t: (k) => k }, stubs: ['TaskStats'] },
    });

    expect(wrapper.text()).to.include('—');
    wrapper.unmount();
  });
});
