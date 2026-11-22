import './setup';
import { expect } from 'chai';
import { shallowMount } from '@vue/test-utils';
import TaskLogView from '@/components/TaskLogView.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountView(item, props = {}) {
  return shallowMount(TaskLogView, {
    props: { item, projectId: 7, ...props },
    global: { mocks: { $t: (k) => k } },
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
    wrapper.unmount();
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
    wrapper.unmount();
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

describe('TaskLogView.vue agent log', () => {
  let http;
  let wrapper;

  const event = (output, time) => ({ time, output: JSON.stringify(output) });

  afterEach(() => {
    wrapper.unmount();
    http.restore();
  });

  it('folds the events of the stored log into rows', async () => {
    http = mockAxios();
    http.respond((config) => (config.url.endsWith('/output') ? [
      { time: 't1', output: 'Running claude with 1 MCP server(s)' },
      event({
        '@agent': 'tool', id: 'c1', name: 'Bash', input: { command: 'ls' },
      }, 't2'),
      event({ '@agent': 'result', id: 'c1', text: 'a.txt' }, 't3'),
      event({ '@agent': 'text', text: 'One file.' }, 't4'),
    ] : {}));
    wrapper = mountView({ id: 42, status: 'success' });
    await flush();

    const rows = wrapper.vm.output;
    expect(rows).to.have.lengthOf(3);
    expect(rows[0]).to.include({ output: 'Running claude with 1 MCP server(s)' });
    expect(rows[0].agent).to.equal(undefined);
    expect(rows[1].agent).to.include({ kind: 'tool', state: 'done', result: 'a.txt' });
    expect(rows[2].agent).to.include({ kind: 'text', text: 'One file.' });
  });

  it('adds the result that comes later to the row of its call, and opens a call', async () => {
    http = mockAxios();
    http.respond((config) => (config.url.endsWith('/output') ? [
      event({ '@agent': 'tool', id: 'c1', name: 'Bash' }, 't1'),
    ] : {}));
    wrapper = mountView({ id: 42, status: 'running' });
    await flush();

    expect(wrapper.vm.output[0].agent.state).to.equal('running');

    wrapper.vm.addRecords([{ id: 'x', ...event({ '@agent': 'result', id: 'c1', text: 'ok' }, 't2') }]);
    expect(wrapper.vm.output).to.have.lengthOf(1);
    expect(wrapper.vm.output[0].agent).to.include({ state: 'done', result: 'ok', open: false });

    wrapper.vm.toggleRecord(0);
    expect(wrapper.vm.output[0].agent.open).to.equal(true);
    wrapper.vm.toggleRecord(0);
    expect(wrapper.vm.output[0].agent.open).to.equal(false);
    wrapper.vm.toggleRecord(5);
  });

  it('starts the folding anew for another task', async () => {
    http = mockAxios();
    http.respond((config) => {
      if (config.url.endsWith('/tasks/42/output')) {
        return [event({ '@agent': 'tool', id: 'c1', name: 'Bash' }, 't1')];
      }
      if (config.url.endsWith('/tasks/43/output')) {
        return [event({ '@agent': 'result', id: 'c1', text: 'of another task' }, 't1')];
      }
      return {};
    });
    wrapper = mountView({ id: 42, status: 'success' });
    await flush();

    await wrapper.setProps({ item: { id: 43, status: 'success' } });
    await flush();

    expect(wrapper.vm.output).to.have.lengthOf(1);
    expect(wrapper.vm.output[0].agent).to.include({ kind: 'result', text: 'of another task' });
  });
});
