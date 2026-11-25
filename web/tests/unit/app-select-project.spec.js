import './setup';
import { expect } from 'chai';
import App from '@/App.vue';
import mockAxios from './helpers/axiosMock';

// What App.vue gives selectProject: the project shown now and the router.
function app(extra = {}) {
  const pushed = [];
  return {
    pushed,
    ctx: {
      projectId: 1,
      userRole: null,
      $route: { path: '/project/1/workflows', query: {} },
      $router: { push: async (location) => pushed.push(location) },
      ...extra,
    },
  };
}

describe('App.vue selectProject', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond(() => ({ role: 'owner', permissions: 0 }));
  });

  afterEach(() => {
    http.restore();
    window.history.replaceState(null, '', '/');
    localStorage.clear();
  });

  it('closes the run panel when it switches to another project', async () => {
    window.history.replaceState(null, '', '/project/1/workflows?w=4/10&t=12');
    const { ctx, pushed } = app();

    await App.methods.selectProject.call(ctx, 2);

    expect(pushed).to.have.length(1);
    expect(pushed[0].path).to.equal('/project/2');
    expect(pushed[0].query.w).to.equal(undefined);
    expect(pushed[0].query.t).to.equal('12');
  });

  it('stays where it is on the project it already shows', async () => {
    window.history.replaceState(null, '', '/project/1/workflows?w=4/10');
    const { ctx, pushed } = app();

    await App.methods.selectProject.call(ctx, 1);

    expect(pushed).to.deep.equal([]);
  });
});
