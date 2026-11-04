import './setup';
import { expect } from 'chai';
import WorkflowRun from '@/views/project/WorkflowRun.vue';
import mockAxios from './helpers/axiosMock';

describe('WorkflowRun.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  it('keeps the caller permissions of the workflow on the run page', async () => {
    http.respond((config) => ({
      '/api/project/7/workflows/3/runs/9': {
        run: { id: 9, status: 'approval' }, workflow_name: 'release', nodes: [], edges: [],
      },
      '/api/project/7/templates': [],
      '/api/project/7/workflows/3': { id: 3, name: 'release', permissions: 80 },
    })[config.url]);
    const page = {
      projectId: 7, workflowId: 3, runId: 9, templates: [], workflow: null, details: null,
    };

    await WorkflowRun.methods.loadData.call(page);

    expect(page.workflow.permissions).to.equal(80);
    expect(WorkflowRun.computed.item.call(page)).to.equal(page.workflow);
  });
});
