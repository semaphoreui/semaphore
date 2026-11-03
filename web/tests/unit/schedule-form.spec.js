import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import ScheduleForm from '@/components/ScheduleForm.vue';
import EventBus from '@/event-bus';
import mockAxios from './helpers/axiosMock';

const localVue = createLocalVue();

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const templates = [{
  id: 11, name: 'Deploy', app: 'bash', survey_vars: [],
}];
const workflows = [{ id: 5, name: 'Build and deploy' }];

function respondTo(url, item) {
  if (url.endsWith('/templates')) return templates;
  if (url.endsWith('/workflows')) return workflows;
  if (/\/schedules\/\d+$/.test(url)) return item;
  if (url.endsWith('/schedules/validate')) return {};
  return {};
}

function mountForm(propsData) {
  return mount(ScheduleForm, {
    localVue,
    vuetify: new Vuetify(),
    propsData: {
      projectId: 7, timezone: 'UTC', type: '', ...propsData,
    },
    mocks: { $t: (k) => k },
    stubs: { TaskParamsForm: true },
  });
}

describe('ScheduleForm.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  it('shows the target toggle only when the workflows feature is on', async () => {
    http.respond((c) => respondTo(c.url));

    const without = mountForm({ itemId: 'new', features: {} });
    await flush();
    expect(without.find('.ScheduleForm__target').exists()).to.equal(false);
    expect(http.requests.map((r) => r.url)).to.not.include('/api/project/7/workflows');

    const withWorkflows = mountForm({ itemId: 'new', features: { workflows: true } });
    await flush();
    expect(withWorkflows.find('.ScheduleForm__target').exists()).to.equal(true);
    expect(http.requests.map((r) => r.url)).to.include('/api/project/7/workflows');
  });

  it('derives the target from the loaded item', async () => {
    http.respond((c) => respondTo(c.url, {
      id: 3, name: 'nightly', workflow_id: 5, cron_format: '0 0 * * *', task_params: null,
    }));
    const wrapper = mountForm({ itemId: 3, features: { workflows: true } });
    await flush();

    expect(wrapper.vm.target).to.equal('workflow');
    expect(wrapper.vm.item.task_params).to.equal(null);
  });

  it('keeps a template item on the template target and initialises task params', async () => {
    http.respond((c) => respondTo(c.url, {
      id: 3, name: 'nightly', template_id: 11, cron_format: '0 0 * * *',
    }));
    const wrapper = mountForm({ itemId: 3, features: { workflows: true } });
    await flush();

    expect(wrapper.vm.target).to.equal('template');
    expect(wrapper.vm.item.workflow_id).to.equal(null);
    expect(wrapper.vm.item.task_params).to.deep.equal({});
  });

  it('clears the other target and task params when switching to a workflow', async () => {
    http.respond((c) => respondTo(c.url));
    const wrapper = mountForm({ itemId: 'new', features: { workflows: true } });
    await flush();

    wrapper.vm.item.template_id = 11;
    wrapper.vm.item.task_params = { params: { debug: true } };
    wrapper.vm.setTarget('workflow');

    expect(wrapper.vm.item.template_id).to.equal(null);
    expect(wrapper.vm.item.task_params).to.equal(null);

    wrapper.vm.item.workflow_id = 5;
    wrapper.vm.setTarget('template');

    expect(wrapper.vm.item.workflow_id).to.equal(null);
    expect(wrapper.vm.item.task_params).to.deep.equal({});
  });

  it('sends exactly one target to the API', async () => {
    http.respond((c) => {
      if (c.method === 'post' && c.url.endsWith('/schedules')) return { id: 9 };
      return respondTo(c.url);
    });
    const wrapper = mountForm({
      itemId: 'new',
      features: { workflows: true, workflow_schedules_management: true },
    });
    await flush();

    wrapper.vm.item.name = 'nightly';
    wrapper.vm.item.cron_format = '0 0 * * *';
    wrapper.vm.setTarget('workflow');
    wrapper.vm.item.workflow_id = 5;
    await flush(); // let the inputs pick up the values before v-form validates
    await wrapper.vm.save();

    expect(wrapper.vm.formError).to.equal(null);
    const post = http.requests.find((r) => r.method === 'post' && r.url === '/api/project/7/schedules');
    expect(post, 'POST request').to.not.equal(undefined);
    expect(post.data.workflow_id).to.equal(5);
    expect(post.data.template_id).to.equal(null);
    expect(post.data.task_params).to.equal(null);
    expect(post.data.cron_format).to.equal('0 0 * * *');
  });

  it('disables the workflow select with an upgrade chip without the Pro feature', async () => {
    http.respond((c) => respondTo(c.url));
    const wrapper = mountForm({ itemId: 'new', features: { workflows: true } });
    await flush();

    wrapper.vm.setTarget('workflow');
    await flush();

    expect(wrapper.find('.ScheduleForm__target').exists()).to.equal(true, 'the toggle stays');
    expect(wrapper.find('.ScheduleForm__workflow').classes()).to.include('v-input--is-disabled');
    const chip = wrapper.find('.ScheduleForm__upgrade');
    expect(chip.exists()).to.equal(true);
    expect(chip.text()).to.equal('Upgrade to PRO');

    const events = [];
    const listener = (e) => events.push(e);
    EventBus.$on('i-subscription', listener);
    try {
      await chip.trigger('click');
    } finally {
      EventBus.$off('i-subscription', listener);
    }
    expect(events).to.deep.equal([{ feature: 'workflow_schedules_management' }]);
  });

  it('shows the workflow select without the chip when the Pro feature is on', async () => {
    http.respond((c) => respondTo(c.url, {
      id: 3, name: 'nightly', workflow_id: 5, cron_format: '0 0 * * *',
    }));
    const wrapper = mountForm({
      itemId: 3,
      features: { workflows: true, workflow_schedules_management: true },
    });
    await flush();

    expect(wrapper.find('.ScheduleForm__workflow').classes()).to.not.include('v-input--is-disabled');
    expect(wrapper.find('.ScheduleForm__upgrade').exists()).to.equal(false);
  });

  it('shows a read-only notice for a workflow schedule when the feature is off', async () => {
    http.respond((c) => respondTo(c.url, {
      id: 3, name: 'nightly', workflow_id: 5, cron_format: '0 0 * * *',
    }));
    const wrapper = mountForm({ itemId: 3, features: {} });
    await flush();

    expect(wrapper.vm.target).to.equal('workflow');
    expect(wrapper.find('.ScheduleForm__workflowUnavailable').exists()).to.equal(true);
    expect(wrapper.find('.ScheduleForm__target').exists()).to.equal(false);
  });
});
