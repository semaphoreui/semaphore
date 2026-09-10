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
        'v-form': {
          template: '<form><slot /></form>',
          methods: {
            validate: () => true,
            resetValidation: () => {},
          },
        },
      },
    });

    wrapper.vm.editVar(null);
    await wrapper.vm.$nextTick();

    wrapper.vm.editedVar.name = '--check';

    // `@keydown.enter` on a component is a component event, not a DOM event,
    // so emit it through the stubbed v-text-field instead of `trigger()`.
    wrapper.find('v-text-field-stub').vm.$emit('keydown', {
      type: 'keydown',
      key: 'Enter',
      keyCode: 13,
      preventDefault: () => {},
    });
    await wrapper.vm.$nextTick();

    expect(wrapper.emitted().change[0][0]).to.deep.equal(['--check']);
    expect(wrapper.vm.editDialog).to.equal(false);
  });
});
