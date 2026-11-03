import './setup';
import { expect } from 'chai';
import { USER_PERMISSIONS } from '@/lib/constants';
import PermissionsCheck from '@/components/PermissionsCheck';

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
  let TemplateForm;

  before(() => {
    // eslint-disable-next-line global-require
    Templates = require('@/views/project/Templates.vue').default;
    // eslint-disable-next-line global-require
    TemplateView = require('@/views/project/TemplateView.vue').default;
    // eslint-disable-next-line global-require
    TemplateForm = require('@/components/TemplateForm.vue').default;
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
    // A grant on the template carries read, the project role does not.
    const ctx = {
      ...ctxWithoutRead(loaded),
      can: PermissionsCheck.methods.can,
      item: { id: 7, permissions: USER_PERMISSIONS.readProject | USER_PERMISSIONS.runProjectTasks },
      userPermissions: 0,
      isAdmin: false,
    };

    await TemplateView.methods.loadData.call(ctx);

    expect(loaded).to.deep.equal([]);
    expect(ctx.item).to.deep.equal({ id: 7 });
    expect(ctx.inventory).to.deep.equal([]);
    expect(ctx.environment).to.deep.equal([]);
    expect(ctx.repositories).to.deep.equal([]);
  });

  it('TemplateForm.vue loads only what a template grant may read', async () => {
    const loaded = [];
    const endpoints = [];
    const ctx = {
      canRead: false,
      isNew: false,
      itemId: 7,
      app: 'ansible',
      loadProjectResources: async (name) => {
        loaded.push(name);
        return [];
      },
      loadProjectEndpoint: async (url) => {
        endpoints.push(url);
        return [];
      },
    };

    await TemplateForm.methods.loadRelativeData.call(ctx);

    expect(loaded).to.deep.equal(['views', 'templates']);
    expect(endpoints).to.deep.equal([]);
    expect(ctx.repositories).to.deep.equal([]);
    expect(ctx.inventory).to.deep.equal([]);
    expect(ctx.environment).to.deep.equal([]);
    expect(ctx.runnerTags).to.deep.equal([]);
    expect(ctx.schedules).to.deep.equal([]);
  });

  it('Templates.vue offers a new template only with project read', () => {
    const ctx = (canRead) => ({
      USER_PERMISSIONS,
      can: (p) => p !== USER_PERMISSIONS.readProject || canRead,
    });

    expect(Templates.computed.canCreate.call(ctx(true))).to.equal(true);
    expect(Templates.computed.canCreate.call(ctx(false))).to.equal(false);
  });
});
