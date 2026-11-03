import './setup';
import { expect } from 'chai';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TaskLogView from '@/components/TaskLogView.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountView(item, props = {}) {
  const localVue = createLocalVue();
  return shallowMount(TaskLogView, {
    localVue,
    vuetify: new Vuetify(),
    propsData: { item, projectId: 7, ...props },
    mocks: { $t: (k) => k },
  });
}

describe('TaskLogView.vue task origin', () => {
  let http;
  let wrapper;

  beforeEach(() => {
    http = mockAxios();
    http.respond((config) => (config.url.endsWith('/output') ? [] : { id: 5 }));
  });

  afterEach(() => {
    wrapper.destroy();
    http.restore();
  });

  it('loads the schedule and the integration with read_project', async () => {
    wrapper = mountView({
      id: 42, status: 'success', schedule_id: 5, integration_id: 6,
    }, { canRead: true });
    await flush();

    const urls = http.requests.map((r) => r.url);
    expect(urls).to.include('/api/project/7/schedules/5');
    expect(urls).to.include('/api/project/7/integrations/6');
    expect(wrapper.vm.schedule.status).to.equal('ready');
  });

  it('shows the origin ids only without read_project', async () => {
    wrapper = mountView({
      id: 42, status: 'success', schedule_id: 5, integration_id: 6,
    });
    await flush();

    const urls = http.requests.map((r) => r.url);
    expect(urls).to.deep.equal(['/api/project/7/tasks/42/output']);
    expect(wrapper.vm.schedule).to.equal(null);
    expect(wrapper.vm.integration).to.equal(null);
  });
});

describe('TaskLogView.vue websocket handling', () => {
  let http;
  let wrapper;

  beforeEach(async () => {
    http = mockAxios();
    http.respond((config) => (config.url.endsWith('/output') ? [] : {}));
    wrapper = mountView({ id: 42, status: 'running' });
    await flush();
  });

  afterEach(() => {
    wrapper.destroy();
    http.restore();
  });

  it('ignores messages of other tasks and projects', () => {
    wrapper.vm.onWebsocketDataReceived({
      project_id: 7, task_id: 43, type: 'log', output: 'x', time: 't',
    });
    wrapper.vm.onWebsocketDataReceived({
      project_id: 8, task_id: 42, type: 'log', output: 'x', time: 't',
    });
    expect(wrapper.vm.outputBuffer).to.deep.equal([]);
  });

  it('buffers log records with a derived id', () => {
    wrapper.vm.onWebsocketDataReceived({
      project_id: 7, task_id: 42, type: 'log', output: 'hello', time: '10:00',
    });
    expect(wrapper.vm.outputBuffer).to.have.lengthOf(1);
    expect(wrapper.vm.outputBuffer[0]).to.include({ output: 'hello', id: '10:00hello' });
    expect(wrapper.vm.output).to.deep.equal([]);
  });

  it('merges update messages into the item without copying the message type', () => {
    wrapper.vm.onWebsocketDataReceived({
      project_id: 7, task_id: 42, type: 'update', status: 'success', end: '2026-01-01',
    });
    expect(wrapper.vm.item.status).to.equal('success');
    expect(wrapper.vm.item.end).to.equal('2026-01-01');
    expect(wrapper.vm.item.type).to.equal(undefined);
  });

  it('ignores unknown message types', () => {
    wrapper.vm.onWebsocketDataReceived({ project_id: 7, task_id: 42, type: 'weird' });
    expect(wrapper.vm.outputBuffer).to.deep.equal([]);
    expect(wrapper.vm.item.status).to.equal('running');
  });
});
