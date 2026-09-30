import './setup';
import { expect } from 'chai';
import EditTemplatePermissionForm from '@/components/EditTemplatePermissionForm.vue';

describe('EditTemplatePermissionForm', () => {
  it('builds a role ID payload with the selected permission mask', () => {
    const context = {
      item: {
        role_id: 12,
        permissions: 0,
      },
      permissions: {
        1: true,
        2: false,
        4: true,
        8: false,
      },
    };

    EditTemplatePermissionForm.methods.beforeSave.call(context);
    const payload = EditTemplatePermissionForm.methods.getRequestOptions.call(context).data;

    expect(payload).to.deep.equal({
      role_id: 12,
      permissions: 5,
    });
  });
});
