import { expect } from 'chai';
import { shallowMount, flushPromises } from '@vue/test-utils';
import ArgsPicker from '@/components/ArgsPicker.vue';

describe('ArgsPicker.vue', () => {
  it('saveVar adds the arg and closes the edit dialog', async () => {
    const wrapper = shallowMount(ArgsPicker, {
      props: { vars: [] },
      global: {
        // the stubs render their default slot, as the stubs of @vue/test-utils 1 did
        renderStubDefaultSlot: true,
        mocks: { $t: (k) => k },
        stubs: {
          'v-dialog': {
            template: '<div><slot /></div>',
          },
          // Vuetify 3: validate() resolves to { valid }.
          'v-form': {
            template: '<form><slot /></form>',
            methods: {
              validate: async () => ({ valid: true }),
              resetValidation: () => {},
            },
          },
        },
      },
    });

    wrapper.vm.editVar(null);
    await wrapper.vm.$nextTick();

    wrapper.vm.editedVar.name = '--check';

    // `@keydown.enter` on a component is a component event, not a DOM event,
    // so emit it through the stubbed v-text-field instead of `trigger()`.
    wrapper.findComponent({ name: 'VTextField' }).vm.$emit('keydown', {
      type: 'keydown',
      key: 'Enter',
      keyCode: 13,
      preventDefault: () => {},
    });
    // saveVar awaits the form validation
    await flushPromises();

    expect(wrapper.emitted().change[0][0]).to.deep.equal(['--check']);
    expect(wrapper.vm.editDialog).to.equal(false);
  });
});
