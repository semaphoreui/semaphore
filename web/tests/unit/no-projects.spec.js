import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import i18n from '@/plugins/i18';
import New from '@/views/project/New.vue';
import RestoreProject from '@/views/project/RestoreProject.vue';
import ProjectForm from '@/components/ProjectForm.vue';
import mockAxios from './helpers/axiosMock';

const USER = { username: 'alice', email: 'alice@example.com', can_create_project: false };

describe('No projects screen', () => {
  let http;
  const wrappers = [];

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
    wrappers.splice(0).forEach((w) => w.destroy());
  });

  function mountWith(component, user) {
    const wrapper = mount(component, {
      localVue: createLocalVue(),
      vuetify: new Vuetify(),
      i18n,
      propsData: { user },
      stubs: { ProjectForm: true, RestoreProjectForm: true },
    });
    wrappers.push(wrapper);
    return wrapper;
  }

  it('New.vue shows the screen instead of the form without the right', () => {
    const wrapper = mountWith(New, USER);

    const text = wrapper.text();
    expect(text).to.contain('No projects yet');
    expect(text).to.contain('Ask your Semaphore administrator');
    expect(text).to.contain('alice');
    expect(text).to.contain('alice@example.com');
    expect(wrapper.findComponent(ProjectForm).exists()).to.equal(false);
  });

  it('leaves out the email brackets when the user has no email', () => {
    const wrapper = mountWith(New, { ...USER, email: '' });

    expect(wrapper.text()).to.contain('Signed in as alice');
    expect(wrapper.text()).not.to.contain('(');
  });

  it('New.vue keeps the form for a user who may create projects', () => {
    const wrapper = mountWith(New, { ...USER, can_create_project: true });

    expect(wrapper.findComponent(ProjectForm).exists()).to.equal(true);
    expect(wrapper.text()).not.to.contain('No projects yet');
  });

  it('New.vue keeps the form while the user is not loaded', () => {
    const wrapper = mountWith(New, null);

    expect(wrapper.findComponent(ProjectForm).exists()).to.equal(true);
  });

  it('RestoreProject.vue shows the screen without the right', () => {
    const wrapper = mountWith(RestoreProject, USER);

    expect(wrapper.text()).to.contain('No projects yet');
    expect(wrapper.find('restoreprojectform-stub').exists()).to.equal(false);
  });
});
