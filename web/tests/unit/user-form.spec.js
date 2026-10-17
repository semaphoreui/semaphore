import './setup';
import { expect } from 'chai';
import UserForm from '@/components/UserForm.vue';

describe('UserForm.vue', () => {
  const { itemId, tab } = UserForm.watch;

  it('loads the created user and opens the roles tab', async () => {
    const form = {
      canManageGlobalRoles: true, tab: 'settings', loaded: 0, async loadData() { this.loaded += 1; },
    };
    await itemId.call(form, 12, 'new');
    expect(form.loaded).to.equal(1);
    expect(form.tab).to.equal('global-roles');
  });

  it('stays on the settings when roles cannot be assigned', async () => {
    const form = {
      canManageGlobalRoles: false, tab: 'settings', loaded: 0, async loadData() { this.loaded += 1; },
    };
    await itemId.call(form, 12, 'new');
    expect(form.loaded).to.equal(1);
    expect(form.tab).to.equal('settings');
  });

  it('ignores other id changes', async () => {
    const form = {
      canManageGlobalRoles: true, tab: 'settings', loaded: 0, async loadData() { this.loaded += 1; },
    };
    await itemId.call(form, 13, 12);
    await itemId.call(form, 'new', 13);
    expect(form.loaded).to.equal(0);
  });

  it('shows the dialog buttons on the settings tab only', () => {
    const emitted = [];
    const form = { $emit: (name) => emitted.push(name) };
    tab.call(form, 'settings');
    tab.call(form, 'global-roles');
    expect(emitted).to.deep.equal(['show-action-buttons', 'hide-action-buttons']);
  });
});
