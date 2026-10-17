import './setup';
import { expect } from 'chai';
import EditRoleForm from '@/components/EditRoleForm.vue';

const ctx = { $t: (key) => key };

describe('EditRoleForm', () => {
  describe('name watcher', () => {
    const onNameChange = EditRoleForm.watch['item.name'];

    it('derives the slug of a new role from the name', () => {
      const form = { isNew: true, slugTouched: false, item: { slug: '' } };
      onNameChange.call(form, 'CI Deployer');
      expect(form.item.slug).to.equal('ci-deployer');
      onNameChange.call(form, '  Ops / Prod  ');
      expect(form.item.slug).to.equal('ops-prod');
      onNameChange.call(form, 'Тест');
      expect(form.item.slug).to.equal('');
    });

    it('keeps the slug the user has edited', () => {
      const form = { isNew: true, slugTouched: true, item: { slug: 'mine' } };
      onNameChange.call(form, 'CI Deployer');
      expect(form.item.slug).to.equal('mine');
    });

    it('keeps the slug of an existing role', () => {
      const form = { isNew: false, slugTouched: false, item: { slug: 'legacy' } };
      onNameChange.call(form, 'CI Deployer');
      expect(form.item.slug).to.equal('legacy');
    });
  });

  describe('global permissions', () => {
    const { hasGlobalPermission, setGlobalPermission, getNewItem } = EditRoleForm.methods;
    const form = (item, projectId) => ({
      item, projectId, $set(target, key, value) { Object.assign(target, { [key]: value }); },
    });

    it('toggles bits of item.global_permissions', () => {
      const f = form({ global_permissions: 0 });
      setGlobalPermission.call(f, 2, true);
      setGlobalPermission.call(f, 8, true);
      expect(f.item.global_permissions).to.equal(10);
      expect(hasGlobalPermission.call(f, 2)).to.equal(true);
      expect(hasGlobalPermission.call(f, 4)).to.equal(false);
      setGlobalPermission.call(f, 2, false);
      expect(f.item.global_permissions).to.equal(8);
    });

    it('treats a missing mask as empty', () => {
      const f = form({});
      expect(hasGlobalPermission.call(f, 1)).to.equal(false);
      setGlobalPermission.call(f, 1, true);
      expect(f.item.global_permissions).to.equal(1);
    });

    it('starts a global role with an empty mask and a project role without one', () => {
      expect(getNewItem.call(form(null, null))).to.deep.equal({
        name: '', slug: '', permissions: 0, global_permissions: 0,
      });
      expect(getNewItem.call(form(null, 7))).to.deep.equal({ name: '', slug: '', permissions: 0 });
    });
  });

  it('derives the slug again after the form is reset', () => {
    const form = { slugTouched: true };
    EditRoleForm.methods.afterReset.call(form);
    expect(form.slugTouched).to.equal(false);
  });

  it('accepts letters of both cases, digits, _ and -', () => {
    expect(EditRoleForm.methods.validateSlug.call(ctx, 'CI_bot-2')).to.equal(true);
    expect(EditRoleForm.methods.validateSlug.call(ctx, 'with space')).to.equal('invalidSlugFormat');
    expect(EditRoleForm.methods.validateSlug.call(ctx, 'a'.repeat(101))).to.equal('invalidSlugFormat');
  });

  it('applies the slug rules only to new roles', () => {
    const legacy = { isNew: false, ...ctx };
    expect(EditRoleForm.computed.slugRules.call(legacy)).to.deep.equal([]);

    const created = {
      isNew: true,
      validateSlug: EditRoleForm.methods.validateSlug,
      ...ctx,
    };
    const rules = EditRoleForm.computed.slugRules.call(created);
    expect(rules).to.have.length(2);
    expect(rules[0]('')).to.equal('slug_required');
    expect(rules[1].call(created, 'my.role')).to.equal('invalidSlugFormat');
  });
});
