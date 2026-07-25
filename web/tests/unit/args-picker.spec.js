import { expect } from 'chai';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import ArgsPicker from '@/components/ArgsPicker.vue';

const localVue = createLocalVue();
localVue.use(Vuetify);

describe('ArgsPicker.vue', () => {
  let vuetify;

  beforeEach(() => {
    vuetify = new Vuetify();
  });

  it('saveVar adds the arg and closes the edit dialog', async () => {
    const wrapper = shallowMount(ArgsPicker, {
      localVue,
      vuetify,
      propsData: { vars: [] },
      mocks: { $t: (k) => k },
      stubs: {
        'v-dialog': {
          template: '<div><slot /></div>',
        },
      },
    });

    wrapper.vm.editVar(null);
    await wrapper.vm.$nextTick();

    wrapper.vm.editedVar.name = '--check';
    wrapper.vm.$refs.form = { validate: () => true };
    wrapper.vm.saveVar();
    await wrapper.vm.$nextTick();

    expect(wrapper.emitted().change[0][0]).to.deep.equal(['--check']);
    expect(wrapper.vm.editDialog).to.equal(false);
  });
});
