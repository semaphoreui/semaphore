import './setup';
import { expect } from 'chai';
import ObjectPerms from '@/views/project/ObjectPerms.vue';

describe('ObjectPerms.vue', () => {
  it('builds the permissions url for the object it is given', () => {
    const forTemplate = { projectId: 1, template: { id: 5 } };
    expect(ObjectPerms.computed.scope.call(forTemplate)).to.equal('template');
    expect(ObjectPerms.methods.getItemsUrl.call({
      ...forTemplate, scope: 'template', object: { id: 5 },
    })).to.equal('/api/project/1/templates/5/perms');

    const forWorkflow = { projectId: 1, workflow: { id: 7 } };
    expect(ObjectPerms.computed.scope.call(forWorkflow)).to.equal('workflow');
    expect(ObjectPerms.methods.getSingleItemUrl.call({
      ...forWorkflow, scope: 'workflow', object: { id: 7 }, itemId: 3, getItemsUrl: ObjectPerms.methods.getItemsUrl,
    })).to.equal('/api/project/1/workflows/7/perms/3');
  });
});
