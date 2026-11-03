import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import ScheduleForm from '@/components/ScheduleForm.vue';
import mockAxios from './helpers/axiosMock';

const localVue = createLocalVue();
const workflows = [{
  id: 5,
  name: 'Deploy',
  survey_vars: [
    { name: 'region', type: 'string', default_value: 'eu' },
    { name: 'zones', type: 'select', default_value: ['a'] },
    { name: 'token', type: 'secret' },
  ],
}];

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountForm(features, itemId = 'new') {
  return mount(ScheduleForm, {
    localVue,
    vuetify: new Vuetify(),
    propsData: {
      projectId: 7, itemId, features, timezone: 'UTC', type: '',
    },
    mocks: { $t: (key) => key },
    stubs: { TaskParamsForm: true, SurveyFields: true },
  });
}

describe('ScheduleForm workflow targets', () => {
  let http;
  let wrapper;
  beforeEach(() => {
    http = mockAxios();
    http.respond((config) => {
      if (config.url.endsWith('/templates')) return [];
      if (config.url.endsWith('/workflows')) return workflows;
      if (config.url.endsWith('/schedules/9')) {
        return {
          id: 9, workflow_id: 5, task_params: { environment: JSON.stringify({ region: 'us' }) }, cron_format: '@daily',
        };
      }
      return {};
    });
  });
  afterEach(() => {
    if (wrapper) wrapper.destroy();
    http.restore();
  });

  it('does not fetch workflows when the feature is disabled', async () => {
    wrapper = mountForm({});
    await flush();
    expect(http.requests.some((r) => r.url.endsWith('/workflows'))).to.equal(false);
  });

  it('initializes survey defaults and sends only the workflow target', async () => {
    wrapper = mountForm({ workflows: true });
    await flush();
    wrapper.vm.target = 'workflow';
    wrapper.vm.changeTarget();
    wrapper.vm.item.workflow_id = 5;
    wrapper.vm.resetWorkflowParams();
    expect(wrapper.vm.workflowEnvironment).to.deep.equal({ region: 'eu', zones: ['a'] });
    await wrapper.vm.beforeSave();
    expect(wrapper.vm.item.template_id).to.equal(null);
    expect(JSON.parse(wrapper.vm.item.task_params.environment)).to.deep.equal({ region: 'eu', zones: ['a'] });
    expect(wrapper.vm.item.workflow_id).to.equal(5);
  });

  it('preserves saved workflow answers when editing', async () => {
    wrapper = mountForm({ workflows: true }, 9);
    await flush();
    expect(wrapper.vm.target).to.equal('workflow');
    expect(wrapper.vm.workflowEnvironment).to.deep.equal({ region: 'us' });
    await wrapper.vm.beforeSave();
    expect(wrapper.vm.workflowEnvironment).to.deep.equal({ region: 'us' });
  });

  it('clears workflow answers when switching to a template', async () => {
    wrapper = mountForm({ workflows: true }, 9);
    await flush();
    wrapper.vm.target = 'template';
    wrapper.vm.changeTarget();
    wrapper.vm.item.template_id = 11;
    await wrapper.vm.beforeSave();
    expect(wrapper.vm.item.workflow_id).to.equal(null);
    expect(wrapper.vm.workflowEnvironment).to.deep.equal({});
    expect(wrapper.vm.item.template_id).to.equal(11);
  });
});
