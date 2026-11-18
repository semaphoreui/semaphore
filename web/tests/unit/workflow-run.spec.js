import './setup';
import { expect } from 'chai';
import { vi } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import WorkflowRunView from '@/components/WorkflowRunView.vue';
import EventBus from '@/event-bus';
import mockAxios, { httpError } from './helpers/axiosMock';

// What the run page and the run panel give the view: the run under its workflow.
function view(extra = {}) {
  return {
    projectId: 7, workflowId: 3, runId: 9, templates: [], workflow: null, details: null, ...extra,
  };
}

const RUN_URL = '/api/project/7/workflows/3/runs/9';

describe('WorkflowRunView.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
    EventBus.$off('i-workflow-run');
  });

  it('keeps the caller permissions of the workflow', async () => {
    http.respond((config) => ({
      [RUN_URL]: {
        run: { id: 9, status: 'approval' }, workflow_name: 'release', nodes: [], edges: [],
      },
      '/api/project/7/templates': [],
      '/api/project/7/workflows/3': { id: 3, name: 'release', permissions: 80 },
    })[config.url]);
    const run = view();

    await WorkflowRunView.methods.loadData.call(run);

    expect(run.workflow.permissions).to.equal(80);
    expect(WorkflowRunView.computed.item.call(run)).to.equal(run.workflow);
  });

  it('tells the lists of runs when the status of the run changes', async () => {
    let status = 'running';
    http.respond((config) => ({
      [RUN_URL]: {
        run: { id: 9, status }, workflow_name: 'release', nodes: [], edges: [],
      },
      '/api/project/7/templates': [],
      '/api/project/7/workflows/3': { id: 3, name: 'release', permissions: 80 },
    })[config.url]);
    const seen = [];
    EventBus.$on('i-workflow-run', (e) => seen.push(e));
    const run = view();

    await WorkflowRunView.methods.loadData.call(run);
    seen.length = 0;
    await WorkflowRunView.methods.loadData.call(run);
    expect(seen).to.deep.equal([]);

    status = 'success';
    await WorkflowRunView.methods.loadData.call(run);

    expect(seen).to.deep.equal([{ runId: 9, workflowId: 3, status: 'success' }]);
  });

  // The list was read before the panel opened: the run may have finished since, and a finished
  // run is not polled, so the first load is the only chance to tell the list.
  it('tells the lists of runs the status of the first load', async () => {
    http.respond((config) => ({
      [RUN_URL]: {
        run: { id: 9, status: 'success' }, workflow_name: 'release', nodes: [], edges: [],
      },
      '/api/project/7/templates': [],
      '/api/project/7/workflows/3': { id: 3, name: 'release', permissions: 80 },
    })[config.url]);
    const seen = [];
    EventBus.$on('i-workflow-run', (e) => seen.push(e));
    const run = view();

    await WorkflowRunView.methods.loadData.call(run);

    expect(seen).to.deep.equal([{ runId: 9, workflowId: 3, status: 'success' }]);
  });

  it('tells the lists nothing when the first load fails', async () => {
    http.respond((config) => {
      if (config.url === RUN_URL) {
        throw httpError(404, '');
      }
      return [];
    });
    const seen = [];
    EventBus.$on('i-workflow-run', (e) => seen.push(e));
    const run = view({ $t: (k) => k });

    await WorkflowRunView.methods.loadData.call(run);

    expect(seen).to.deep.equal([]);
  });

  describe('mounted in the panel', () => {
    const wrappers = [];

    afterEach(() => {
      while (wrappers.length) {
        wrappers.pop().unmount();
      }
    });

    async function mountPanel() {
      const wrapper = mount(WorkflowRunView, {
        props: {
          projectId: 7, workflowId: 3, runId: 9, panel: true,
        },
        global: {
          mocks: { $t: (k) => k },
          stubs: { WorkflowGraph: true, RouterLink: { template: '<a><slot /></a>' } },
        },
      });
      wrappers.push(wrapper);
      await flushPromises();
      return wrapper;
    }

    it('says that the run is not there instead of showing it', async () => {
      http.respond((config) => {
        if (config.url === RUN_URL) {
          throw httpError(404, '');
        }
        return [];
      });

      const wrapper = await mountPanel();

      expect(wrapper.attributes('data-state')).to.equal('error');
      expect(wrapper.find('[data-testid="workflowRun-error"]').text()).to.equal('workflowRunNotFound');
      expect(wrapper.find('[data-testid="workflowRun-graph"]').exists()).to.equal(false);
    });

    it('has a close button and no drawer toggle', async () => {
      http.respond((config) => ({
        [RUN_URL]: {
          run: { id: 9, status: 'success' },
          workflow_name: 'release',
          nodes: [],
          edges: [],
        },
        '/api/project/7/templates': [],
        '/api/project/7/workflows/3': { id: 3, name: 'release', permissions: 80 },
      })[config.url]);

      const wrapper = await mountPanel();

      expect(wrapper.attributes('data-state')).to.equal('ready');
      expect(wrapper.find('[data-testid="workflowRun-drawerToggle"]').exists()).to.equal(false);
      expect(wrapper.find('[data-testid="workflowRun-workflowLink"]').text()).to.equal('release');

      await wrapper.find('[data-testid="workflowRun-close"]').trigger('click');

      expect(wrapper.emitted('close')).to.have.length(1);
    });
  });

  describe('polling', () => {
    const RUNNING = {
      run: { id: 9, status: 'running' }, workflow_name: 'release', nodes: [], edges: [],
    };

    // The requests for the run itself: the first load and every poll after it.
    function runRequests() {
      return http.requests.filter((r) => r.url === RUN_URL).length;
    }

    function mountView() {
      return mount(WorkflowRunView, {
        props: {
          projectId: 7, workflowId: 3, runId: 9, panel: true,
        },
        global: {
          mocks: { $t: (k) => k },
          stubs: { WorkflowGraph: true, RouterLink: { template: '<a><slot /></a>' } },
        },
      });
    }

    beforeEach(() => {
      vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] });
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('polls a run in progress until the view is closed', async () => {
      http.respond((config) => ({
        [RUN_URL]: RUNNING,
        '/api/project/7/templates': [],
        '/api/project/7/workflows/3': { id: 3, name: 'release', permissions: 80 },
      })[config.url]);

      const wrapper = mountView();
      await flushPromises();
      expect(runRequests()).to.equal(1);

      vi.advanceTimersByTime(5000);
      await flushPromises();
      expect(runRequests()).to.equal(2);

      wrapper.unmount();
      vi.advanceTimersByTime(15000);
      await flushPromises();

      expect(runRequests()).to.equal(2);
      expect(vi.getTimerCount()).to.equal(0);
    });

    // The user follows the link to another run and comes back, closes the panel or leaves the
    // page before the run has loaded: the view is gone by the time its first load ends, and
    // nothing is left to stop a timer started then.
    it('starts no polling when it is closed before the run has loaded', async () => {
      let release;
      const loaded = new Promise((resolve) => {
        release = resolve;
      });
      http.respond(async (config) => {
        if (config.url === RUN_URL) {
          await loaded;
          return RUNNING;
        }
        return config.url === '/api/project/7/templates'
          ? []
          : { id: 3, name: 'release', permissions: 80 };
      });

      const wrapper = mountView();
      await flushPromises();
      wrapper.unmount();
      release();
      await flushPromises();
      expect(runRequests()).to.equal(1);

      vi.advanceTimersByTime(15000);
      await flushPromises();

      expect(runRequests()).to.equal(1);
      expect(vi.getTimerCount()).to.equal(0);
    });
  });

  describe('re-run', () => {
    const NEW_RUN = { id: 11, status: 'running' };

    // The view as `rerun` sees it, on the page or in the panel, with the router it moves.
    function rerunView(panel, pushed) {
      const run = view({
        panel,
        rerunning: false,
        $t: (k) => k,
        $route: { path: '/project/7/workflows', query: { t: '5' } },
        $router: { push: async (to) => pushed.push(to) },
      });
      run.runLocation = WorkflowRunView.methods.runLocation.bind(run);
      return run;
    }

    beforeEach(() => {
      http.respond(() => NEW_RUN);
    });

    it('repeats the run in the mode asked for', async () => {
      const run = rerunView(false, []);

      await WorkflowRunView.methods.rerun.call(run, 'failed');

      expect(http.requests).to.have.length(1);
      expect(http.requests[0].url).to.equal(`${RUN_URL}/rerun`);
      expect(http.requests[0].data).to.deep.equal({ mode: 'failed' });
      expect(run.rerunning).to.equal(false);
    });

    it('opens the new run at its address when it is the run page', async () => {
      const pushed = [];

      await WorkflowRunView.methods.rerun.call(rerunView(false, pushed), 'all');

      expect(pushed).to.deep.equal(['/project/7/workflows/3/runs/11']);
    });

    it('swaps the run of the panel and keeps the page under it', async () => {
      const pushed = [];

      await WorkflowRunView.methods.rerun.call(rerunView(true, pushed), 'all');

      expect(pushed).to.deep.equal([
        { path: '/project/7/workflows', query: { t: '5', w: '3/11' } },
      ]);
    });
  });
});
