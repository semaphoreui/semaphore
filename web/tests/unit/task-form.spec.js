import './setup';
import { expect } from 'chai';
import TaskForm from '@/components/TaskForm.vue';
import mockAxios, { httpError } from './helpers/axiosMock';

describe('TaskForm.vue build tasks', () => {
  let http;
  const ctx = { projectId: 7, template: { type: 'deploy', build_template_id: 3 } };

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  it('lists the successful builds', async () => {
    http.respond(() => [{ id: 1, status: 'success' }, { id: 2, status: 'error' }]);

    const builds = await TaskForm.methods.loadBuildTasks.call(ctx);

    expect(builds.map((b) => b.id)).to.deep.equal([1]);
    expect(http.requests[0].url).to.equal('/api/project/7/templates/3/tasks?status=success&limit=20');
  });

  it('has nothing to pick from when the build template is hidden', async () => {
    http.respond(() => { throw httpError(404, {}); });

    const builds = await TaskForm.methods.loadBuildTasks.call(ctx);

    expect(builds).to.deep.equal([]);
  });
});
