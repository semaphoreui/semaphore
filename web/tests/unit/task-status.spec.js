import { expect } from 'chai';
import { shallowMount } from '@vue/test-utils';
import TaskStatus from '@/components/TaskStatus.vue';

function vm(status) {
  return shallowMount(TaskStatus, { propsData: { status } }).vm;
}

describe('TaskStatus.vue', () => {
  const tests = [
    {
      status: 'waiting', icon: 'mdi-alarm', title: 'Waiting', color: '',
    },
    {
      status: 'starting', icon: 'mdi-play-circle', title: 'Starting...', color: 'warning',
    },
    {
      status: 'running', icon: '', title: 'Running', color: 'primary',
    },
    {
      status: 'success', icon: 'mdi-check-circle', title: 'Success', color: 'success',
    },
    {
      status: 'error', icon: 'mdi-information', title: 'Failed', color: 'error',
    },
    {
      status: 'stopping', icon: 'mdi-stop-circle', title: 'Stopping...', color: '',
    },
    {
      status: 'stopped', icon: 'mdi-stop-circle', title: 'Stopped', color: '',
    },
    {
      status: 'confirmed', icon: 'mdi-check-circle', title: 'Confirmed', color: 'warning',
    },
    {
      status: 'waiting_confirmation', icon: 'mdi-pause-circle', title: 'Waiting confirmation', color: 'warning',
    },
  ];

  tests.forEach(({
    status, icon, title, color,
  }) => {
    it(`maps "${status}"`, () => {
      const v = vm(status);
      expect(v.getStatusIcon(status)).to.equal(icon);
      expect(v.humanizeStatus(status)).to.equal(title);
      expect(v.getStatusColor(status)).to.equal(color);
    });
  });

  it('throws for an unknown status', () => {
    const v = vm('success');
    expect(() => v.getStatusIcon('bogus')).to.throw('Unknown task status bogus');
    expect(() => v.humanizeStatus('bogus')).to.throw('Unknown task status bogus');
    expect(() => v.getStatusColor('bogus')).to.throw('Unknown task status bogus');
  });
});
