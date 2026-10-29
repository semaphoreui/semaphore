import './setup';
import { expect } from 'chai';
import EditTemplatePermissionForm from '@/components/EditTemplatePermissionForm.vue';

describe('EditTemplatePermissionForm', () => {
  const pack = EditTemplatePermissionForm.watch.permissions.handler;

  it('lets a verb on the object include seeing it', () => {
    const form = { item: { permissions: 0 }, permissions: {}, scope: 'workflow' };
    pack.call(form, { 16: true });
    expect(form.item.permissions).to.equal(144);

    pack.call(form, { 128: true });
    expect(form.item.permissions).to.equal(128);

    pack.call(form, {});
    expect(form.item.permissions).to.equal(0);
  });

  it('keeps the see checkbox on while a verb is checked', () => {
    const verbChecked = EditTemplatePermissionForm.computed.verbChecked;
    expect(verbChecked.call({ permissions: { 1: true, 128: false } })).to.equal(true);
    expect(verbChecked.call({ permissions: { 128: true } })).to.equal(false);
  });
});
