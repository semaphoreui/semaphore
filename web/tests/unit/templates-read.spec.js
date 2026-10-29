import './setup';
import { expect } from 'chai';
import { USER_PERMISSIONS } from '@/lib/constants';

function ctxWithoutRead(loaded) {
  return {
    USER_PERMISSIONS,
    can: (p) => p !== USER_PERMISSIONS.readProject,
    loadProjectResources: async (name) => {
      loaded.push(name);
      return [{ id: 1 }];
    },
    loadProjectResource: async () => ({ id: 7 }),
    itemId: 7,
  };
}

describe('template pages without read_project', () => {
  let Templates;
  let TemplateView;

  before(() => {
    // eslint-disable-next-line global-require
    Templates = require('@/views/project/Templates.vue').default;
    // eslint-disable-next-line global-require
    TemplateView = require('@/views/project/TemplateView.vue').default;
  });

  it('Templates.vue loads the list without project resources', async () => {
    const loaded = [];
    const ctx = ctxWithoutRead(loaded);

    await Templates.methods.loadData.call(ctx);

    expect(loaded).to.deep.equal([]);
    expect(ctx.inventory).to.deep.equal([]);
    expect(ctx.environment).to.deep.equal([]);
    expect(ctx.repositories).to.deep.equal([]);
  });

  it('TemplateView.vue loads the template without project resources', async () => {
    const loaded = [];
    const ctx = ctxWithoutRead(loaded);

    await TemplateView.methods.loadData.call(ctx);

    expect(loaded).to.deep.equal([]);
    expect(ctx.item).to.deep.equal({ id: 7 });
    expect(ctx.inventory).to.deep.equal([]);
    expect(ctx.environment).to.deep.equal([]);
    expect(ctx.repositories).to.deep.equal([]);
  });
});
