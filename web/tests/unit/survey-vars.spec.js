import './setup';
import { expect } from 'chai';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import SurveyVars from '@/components/SurveyVars.vue';

let formValid = true;

function mountVars(vars = []) {
  const localVue = createLocalVue();
  return shallowMount(SurveyVars, {
    localVue,
    vuetify: new Vuetify(),
    propsData: { vars },
    mocks: { $t: (k) => k },
    stubs: {
      'v-dialog': { template: '<div><slot /></div>' },
      'v-form': {
        template: '<form><slot /></form>',
        methods: {
          validate: () => formValid,
          resetValidation: () => {},
        },
      },
    },
  });
}

describe('SurveyVars.vue', () => {
  beforeEach(() => {
    formValid = true;
  });

  describe('normalizeDefaultValue', () => {
    const tests = [
      {
        name: 'select keeps arrays', type: 'select', value: ['a', 'b'], expected: ['a', 'b'],
      },
      {
        name: 'select wraps a scalar', type: 'select', value: 'a', expected: ['a'],
      },
      {
        name: 'select turns null into an empty list', type: 'select', value: null, expected: [],
      },
      {
        name: 'select turns empty string into an empty list', type: 'select', value: '', expected: [],
      },
      {
        name: 'string takes the first array element', type: 'string', value: ['x', 'y'], expected: 'x',
      },
      {
        name: 'string turns an empty array into an empty string', type: 'string', value: [], expected: '',
      },
      {
        name: 'string keeps a scalar', type: 'string', value: 'v', expected: 'v',
      },
      {
        name: 'int keeps a number', type: 'int', value: 5, expected: 5,
      },
      {
        name: 'null becomes an empty string', type: 'string', value: null, expected: '',
      },
      {
        name: 'undefined becomes an empty string', type: 'string', value: undefined, expected: '',
      },
    ];

    tests.forEach(({
      name, type, value, expected,
    }) => {
      it(name, () => {
        expect(mountVars().vm.normalizeDefaultValue(type, value)).to.deep.equal(expected);
      });
    });
  });

  describe('selectedItemLabel', () => {
    it('returns the name of the matching option by value or name', () => {
      const wrapper = mountVars();
      wrapper.vm.editedValues = [{ name: 'Production', value: 'prod' }];
      expect(wrapper.vm.selectedItemLabel('prod')).to.equal('Production');
      expect(wrapper.vm.selectedItemLabel('Production')).to.equal('Production');
    });

    it('uses the name of an object item', () => {
      expect(mountVars().vm.selectedItemLabel({ name: 'N', value: 'v' })).to.equal('N');
    });

    it('falls back to the raw value and to an empty string', () => {
      const wrapper = mountVars();
      expect(wrapper.vm.selectedItemLabel('unknown')).to.equal('unknown');
      expect(wrapper.vm.selectedItemLabel(null)).to.equal('');
    });
  });

  describe('saveVar', () => {
    it('adds a new string variable and emits change', async () => {
      const wrapper = mountVars();
      wrapper.vm.editVar(null);
      await wrapper.vm.$nextTick();
      Object.assign(wrapper.vm.editedVar, { name: 'host', title: 'Host', type: 'string' });

      wrapper.vm.saveVar();

      expect(wrapper.vm.editDialog).to.equal(false);
      expect(wrapper.vm.formError).to.equal(null);
      const emitted = wrapper.emitted().change[0][0];
      expect(emitted).to.have.lengthOf(1);
      expect(emitted[0]).to.include({ name: 'host', title: 'Host', type: 'string' });
      expect(emitted[0].values).to.deep.equal([]);
    });

    it('replaces the edited variable in place', async () => {
      const wrapper = mountVars([
        { name: 'a', title: 'A', type: 'string' },
        { name: 'b', title: 'B', type: 'string' },
      ]);
      wrapper.vm.editVar(1);
      await wrapper.vm.$nextTick();
      wrapper.vm.editedVar.title = 'B2';

      wrapper.vm.saveVar();

      expect(wrapper.emitted().change[0][0].map((v) => v.title)).to.deep.equal(['A', 'B2']);
    });

    it('rejects an enum without values', async () => {
      const wrapper = mountVars();
      wrapper.vm.editVar(null);
      await wrapper.vm.$nextTick();
      Object.assign(wrapper.vm.editedVar, { name: 'e', type: 'enum' });
      await wrapper.vm.$nextTick();

      wrapper.vm.saveVar();

      expect(wrapper.vm.formError).to.equal('Enumeration must have values.');
      expect(wrapper.emitted().change).to.equal(undefined);
    });

    it('rejects duplicate and empty value names', async () => {
      const wrapper = mountVars();
      wrapper.vm.editVar(null);
      await wrapper.vm.$nextTick();
      Object.assign(wrapper.vm.editedVar, { name: 's', type: 'select' });
      await wrapper.vm.$nextTick();

      wrapper.vm.editedValues.push({ name: 'x', value: '1' }, { name: 'x', value: '2' });
      wrapper.vm.saveVar();
      expect(wrapper.vm.formError).to.equal('Select must have unique names.');

      wrapper.vm.editedValues.splice(0, 2, { name: '', value: '1' });
      wrapper.vm.saveVar();
      expect(wrapper.vm.formError).to.equal('Value name cannot be empty.');
    });

    it('normalises the default value of a select', async () => {
      const wrapper = mountVars();
      wrapper.vm.editVar(null);
      await wrapper.vm.$nextTick();
      Object.assign(wrapper.vm.editedVar, { name: 's', type: 'select' });
      await wrapper.vm.$nextTick();
      wrapper.vm.editedValues.push({ name: 'One', value: '1' });
      wrapper.vm.editedVar.default_value = '1';

      wrapper.vm.saveVar();

      expect(wrapper.emitted().change[0][0][0].default_value).to.deep.equal(['1']);
    });

    it('does nothing when the form is invalid', async () => {
      formValid = false;
      const wrapper = mountVars();
      wrapper.vm.editVar(null);
      await wrapper.vm.$nextTick();

      wrapper.vm.saveVar();

      expect(wrapper.vm.editDialog).to.equal(true);
      expect(wrapper.emitted().change).to.equal(undefined);
    });
  });

  it('deleteVar removes the variable and emits change', () => {
    const wrapper = mountVars([{ name: 'a' }, { name: 'b' }]);
    wrapper.vm.deleteVar(0);
    expect(wrapper.emitted().change[0][0].map((v) => v.name)).to.deep.equal(['b']);
  });
});
