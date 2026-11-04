import './setup';
import { expect } from 'chai';
import { shallowMount } from '@vue/test-utils';
import i18n from '@/plugins/i18';
import YesNoDialog from '@/components/YesNoDialog.vue';

function mountDialog(props = {}) {
  return shallowMount(YesNoDialog, {
    props,
    global: {
      plugins: [i18n],
      // the stubs render their default slot, as the stubs of @vue/test-utils 1 did
      renderStubDefaultSlot: true,
      stubs: {
        'v-dialog': {
          template: '<div><slot /></div>',
        },
      },
    },
  });
}

describe('YesNoDialog.vue', () => {
  it('renders title, text and default buttons', () => {
    const wrapper = mountDialog({ title: 'Delete item?', text: 'This cannot be undone.' });

    expect(wrapper.text()).to.include('Delete item?');
    expect(wrapper.text()).to.include('This cannot be undone.');
    expect(wrapper.text()).to.include('Cancel');
    expect(wrapper.text()).to.include('Yes');
  });

  it('renders custom button titles', () => {
    const wrapper = mountDialog({ yesButtonTitle: 'Remove', noButtonTitle: 'Keep' });

    expect(wrapper.text()).to.include('Remove');
    expect(wrapper.text()).to.include('Keep');
  });

  it('emits yes and closes the dialog', async () => {
    const wrapper = mountDialog();
    await wrapper.setProps({ modelValue: true });
    expect(wrapper.vm.dialog).to.equal(true);

    wrapper.vm.yes();
    await wrapper.vm.$nextTick();

    expect(wrapper.emitted().yes).to.have.lengthOf(1);
    expect(wrapper.vm.dialog).to.equal(false);
  });

  it('emits no and closes the dialog', async () => {
    const wrapper = mountDialog();
    await wrapper.setProps({ modelValue: true });

    wrapper.vm.no();
    await wrapper.vm.$nextTick();

    expect(wrapper.emitted().no).to.have.lengthOf(1);
    expect(wrapper.vm.dialog).to.equal(false);
  });
});
