import './setup';
import { expect } from 'chai';
import EditGlobalRoleForm from '@/components/EditGlobalRoleForm.vue';

const ctx = { $t: (key) => key };

describe('EditGlobalRoleForm', () => {
  const form = (item) => ({
    item, $set(target, key, value) { Object.assign(target, { [key]: value }); },
  });

  it('derives the slug of a new role from the name until the user edits it', () => {
    const onNameChange = EditGlobalRoleForm.watch['item.name'];
    const f = { isNew: true, slugTouched: false, item: { slug: '' } };
    onNameChange.call(f, 'Audit Reader');
    expect(f.item.slug).to.equal('audit-reader');
    f.slugTouched = true;
    onNameChange.call(f, 'Other');
    expect(f.item.slug).to.equal('audit-reader');
    const legacy = { isNew: false, slugTouched: false, item: { slug: 'legacy' } };
    onNameChange.call(legacy, 'Audit Reader');
    expect(legacy.item.slug).to.equal('legacy');
  });

  it('toggles bits of item.permissions', () => {
    const { hasPermission, setPermission } = EditGlobalRoleForm.methods;
    const f = form({ permissions: 0 });
    setPermission.call(f, 2, true);
    setPermission.call(f, 8, true);
    expect(f.item.permissions).to.equal(10);
    expect(hasPermission.call(f, 2)).to.equal(true);
    expect(hasPermission.call(f, 4)).to.equal(false);
    setPermission.call(f, 2, false);
    expect(f.item.permissions).to.equal(8);
    const empty = form({});
    expect(hasPermission.call(empty, 1)).to.equal(false);
  });

  it('starts a new role without permissions and talks to the global roles API', () => {
    const { getNewItem, getItemsUrl, getSingleItemUrl } = EditGlobalRoleForm.methods;
    expect(getNewItem.call({})).to.deep.equal({ name: '', slug: '', permissions: 0 });
    expect(getItemsUrl.call({})).to.equal('/api/global-roles');
    expect(getSingleItemUrl.call({ itemId: 4 })).to.equal('/api/global-roles/4');
  });

  it('applies the slug rules only to new roles', () => {
    expect(EditGlobalRoleForm.computed.slugRules.call({ isNew: false, ...ctx })).to.deep.equal([]);
    const created = { isNew: true, validateSlug: EditGlobalRoleForm.methods.validateSlug, ...ctx };
    const rules = EditGlobalRoleForm.computed.slugRules.call(created);
    expect(rules).to.have.length(2);
    expect(rules[0]('')).to.equal('slug_required');
    expect(rules[1].call(created, 'my.role')).to.equal('invalidSlugFormat');
    expect(rules[1].call(created, 'CI_bot-2')).to.equal(true);
  });
});
