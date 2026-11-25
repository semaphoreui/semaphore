import './setup';
import { expect } from 'chai';
import TaskStats from '@/components/TaskStats.vue';

function ctx(canRead, calls) {
  return {
    canRead,
    users: [{ text: 'All users', value: null }],
    refreshData: async () => {},
    loadUsers: async () => {
      calls.push('users');
      return [{ value: 3, text: 'Ann' }];
    },
  };
}

describe('TaskStats.vue', () => {
  it('does not request project users without read_project', async () => {
    const calls = [];
    const c = ctx(false, calls);

    await TaskStats.created.call(c);

    expect(calls).to.deep.equal([]);
    expect(c.users).to.deep.equal([{ text: 'All users', value: null }]);
  });

  it('requests project users with read_project', async () => {
    const calls = [];
    const c = ctx(true, calls);

    await TaskStats.created.call(c);

    expect(calls).to.deep.equal(['users']);
    expect(c.users).to.deep.equal([{ text: 'All users', value: null }, { value: 3, text: 'Ann' }]);
  });
});
