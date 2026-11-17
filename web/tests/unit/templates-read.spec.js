import './setup';
import { expect } from 'chai';
import { USER_PERMISSIONS } from '@/lib/constants';
import PermissionsCheck from '@/components/PermissionsCheck';
import Templates from '@/views/project/Templates.vue';
import TemplateView from '@/views/project/TemplateView.vue';
import TemplateForm from '@/components/TemplateForm.vue';
import TemplateTerraformState from '@/views/project/template/TemplateTerraformState.vue';
import TaskLogDialog from '@/components/TaskLogDialog.vue';

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

  it('TemplateForm.vue copies a template without its schedule when read is missing', async () => {
    const endpoints = [];
    const ctx = {
      canRead: false,
      sourceItemId: 3,
      item: {},
      schedules: [],
      loadProjectResource: async () => ({ id: 3, arguments: '[]' }),
      loadProjectEndpoint: async (url) => {
        endpoints.push(url);
        return [{ cron_format: '* * * * *', repository_id: 1 }];
      },
      loadRelativeData: async () => {},
      needField: () => false,
      $set: (o, k, v) => Object.assign(o, { [k]: v }),
    };

    await TemplateForm.methods.afterLoadData.call(ctx);

    expect(endpoints).to.deep.equal([]);
    expect(ctx.cronVisible).to.equal(undefined);
    expect(ctx.item.id).to.equal(null);
  });

  it('TemplateForm.vue leaves the vaults out without read', () => {
    const needField = () => true;

    expect(TemplateForm.computed.hasVaults.call({ canRead: true, needField })).to.equal(true);
    expect(TemplateForm.computed.hasVaults.call({ canRead: false, needField })).to.equal(false);
  });

  it('TemplateView.vue shows the state tab only with read', () => {
    const ctx = (canRead) => ({
      USER_PERMISSIONS,
      item: { app: 'tofu' },
      can: (p) => p !== USER_PERMISSIONS.readProject || canRead,
    });

    expect(TemplateView.computed.hasStateTab.call(ctx(true))).to.equal(true);
    expect(TemplateView.computed.hasStateTab.call(ctx(false))).to.equal(false);
  });

  it('TemplateTerraformState.vue loads nothing without read', async () => {
    let loaded = false;
    const ctx = {
      canRead: false,
      template: { inventory_id: 1 },
      inventories: [],
      loadInventories: async () => { loaded = true; },
      loadAliases: async () => { loaded = true; },
      loadStates: async () => { loaded = true; },
    };

    await TemplateTerraformState.created.call(ctx);
    await TemplateTerraformState.watch.inventoryId.call(ctx);

    expect(loaded).to.equal(false);
  });

  it('TaskLogDialog.vue opens a task whose template is hidden', async () => {
    const ctx = {
      itemId: 1,
      item: null,
      template: null,
      templateHidden: false,
      loadProjectResource: async (name) => {
        if (name === 'templates') {
          throw new Error('Request failed with status code 404');
        }
        return { id: 1, template_id: 2 };
      },
    };

    await TaskLogDialog.methods.loadData.call(ctx);

    expect(ctx.item).to.deep.equal({ id: 1, template_id: 2 });
    expect(ctx.template).to.equal(null);
    expect(ctx.templateHidden).to.equal(true);
  });

  it('TemplateView.vue reports a template it may not load', async () => {
    const ctx = {
      ...ctxWithoutRead([]),
      item: null,
      notFound: false,
      loadProjectResource: async () => {
        const err = new Error('Request failed with status code 404');
        err.response = { status: 404 };
        throw err;
      },
    };

    await TemplateView.methods.loadData.call(ctx);

    expect(ctx.notFound).to.equal(true);
    expect(ctx.item).to.equal(null);
  });
});
