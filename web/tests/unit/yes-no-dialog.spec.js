import './setup';
import { expect } from 'chai';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import i18n from '@/plugins/i18';
import YesNoDialog from '@/components/YesNoDialog.vue';

const localVue = createLocalVue();

function mountDialog(propsData = {}) {
  return shallowMount(YesNoDialog, {
    localVue,
    vuetify: new Vuetify(),
    i18n,
    propsData,
    stubs: {
      'v-dialog': {
        template: '<div><slot /></div>',
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
    await wrapper.setProps({ value: true });
    expect(wrapper.vm.dialog).to.equal(true);

    wrapper.vm.yes();
    await wrapper.vm.$nextTick();

    expect(wrapper.emitted().yes).to.have.lengthOf(1);
    expect(wrapper.vm.dialog).to.equal(false);
  });

  it('emits no and closes the dialog', async () => {
    const wrapper = mountDialog();
    await wrapper.setProps({ value: true });

    wrapper.vm.no();
    await wrapper.vm.$nextTick();

    expect(wrapper.emitted().no).to.have.lengthOf(1);
    expect(wrapper.vm.dialog).to.equal(false);
  });
});
