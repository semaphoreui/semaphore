import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import Vuetify from 'vuetify';
import ScheduleOffsetField from '@/components/ScheduleOffsetField.vue';
import { MAX_OFFSET_DAYS } from '@/lib/cronPresets';

function mountField(value = 0) {
  return mount(ScheduleOffsetField, {
    propsData: { value },
    vuetify: new Vuetify(),
    mocks: { $t: (key) => key, $tc: (key) => key },
  });
}

describe('ScheduleOffsetField.vue', () => {
  const tests = [
    { name: 'a number', typed: '3', expected: 3 },
    { name: 'a negative number', typed: '-2', expected: -2 },
    { name: 'too many days', typed: '99', expected: MAX_OFFSET_DAYS },
    { name: 'too few days', typed: '-99', expected: -MAX_OFFSET_DAYS },
    { name: 'a decimal', typed: '2.7', expected: 2 },
  ];

  tests.forEach(({ name, typed, expected }) => {
    it(`applies ${name} when the field changes`, () => {
      const wrapper = mountField();
      wrapper.vm.commit(typed);

      expect(wrapper.emitted().input).to.deep.equal([[expected]]);
      expect(wrapper.vm.days).to.equal(expected);
    });
  });

  it('does not emit when a lone "-" leaves the offset unchanged', () => {
    const wrapper = mountField();
    wrapper.vm.commit('-');

    expect(wrapper.emitted().input).to.equal(undefined);
    expect(wrapper.vm.days).to.equal(0);
  });

  it('steps by one day and stops at the limit', async () => {
    const wrapper = mountField(MAX_OFFSET_DAYS - 1);
    const [earlier, later] = wrapper.findAll('button').wrappers;

    await later.trigger('click');
    expect(wrapper.emitted().input).to.deep.equal([[MAX_OFFSET_DAYS]]);

    await wrapper.setProps({ value: MAX_OFFSET_DAYS });
    await later.trigger('click');
    expect(wrapper.emitted().input).to.have.lengthOf(1);

    await earlier.trigger('click');
    expect(wrapper.emitted().input[1]).to.deep.equal([MAX_OFFSET_DAYS - 1]);
  });

  it('names the buttons for screen readers', () => {
    const labels = mountField().findAll('button').wrappers.map((button) => button.attributes('aria-label'));

    expect(labels).to.deep.equal(['scheduleOffsetEarlier', 'scheduleOffsetLater']);
  });
});
