import { expect } from 'chai';
import { shallowMount } from '@vue/test-utils';
import YesNoDialog from '@/components/YesNoDialog.vue';

describe('YesNoDialog.vue', () => {
  it('renders props.msg when passed', () => {
    const wrapper = shallowMount(YesNoDialog);
    expect(wrapper.text()).to.include('Cancel  Yes');
  });
});
