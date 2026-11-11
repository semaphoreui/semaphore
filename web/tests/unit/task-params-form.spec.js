import { expect } from 'chai';
import { shallowMount } from '@vue/test-utils';
import TaskParamsForm from '@/components/TaskParamsForm.vue';

describe('TaskParamsForm.vue', () => {
  it('emits update:modelValue when app params (debug/diff/etc) change', async () => {
    const wrapper = shallowMount(TaskParamsForm, {
      // v-model of a Vue 3 component: modelValue / update:modelValue (Vue 2: value / input)
      props: {
        modelValue: { params: { limit: ['web'] } },
        template: { app: 'ansible', survey_vars: [] },
      },
      global: { mocks: { $t: (k) => k } },
    });
    // Let the async created()/afterLoadData() chain fully settle first —
    // it emits `update:modelValue` itself, which must not count as the change event.
    await new Promise((resolve) => { setTimeout(resolve, 10); });
    const before = (wrapper.emitted()['update:modelValue'] || []).length;

    // What TaskParamsAnsibleForm's v-model does on a checkbox toggle.
    wrapper.vm.item.params = { ...wrapper.vm.item.params, debug: true, diff: true };
    await wrapper.vm.$nextTick();

    const emitted = wrapper.emitted()['update:modelValue'] || [];
    // emitted() holds live references, so check a NEW event fired after the
    // change — the payload object is always up to date by aliasing.
    expect(emitted.length, 'update:modelValue must be emitted on params change').to.be.greaterThan(before);
    const last = emitted[emitted.length - 1][0];
    expect(last.params.debug).to.equal(true);
    expect(last.params.diff).to.equal(true);
  });
});

describe('TaskParamsForm.vue bool survey variables', () => {
  const surveyVars = [
    {
      name: 'on', title: 'On', type: 'bool', default_value: 'true',
    },
    { name: 'off', title: 'Off', type: 'bool' },
    {
      name: 'saved', title: 'Saved', type: 'bool', default_value: 'true',
    },
    {
      name: 'legacy', title: 'Legacy', type: 'bool', required: true,
    },
    {
      name: 'host', title: 'Host', type: '', default_value: 'localhost',
    },
  ];

  async function seeded(environment) {
    const wrapper = shallowMount(TaskParamsForm, {
      props: {
        modelValue: { environment: JSON.stringify(environment) },
        template: { app: 'bash', survey_vars: surveyVars },
      },
      global: { mocks: { $t: (k) => k } },
    });
    await new Promise((resolve) => { setTimeout(resolve, 10); });
    return wrapper.vm.editedEnvironment;
  }

  it('seeds every bool as a boolean: its default, false without one, a saved answer wins', async () => {
    // a schedule saved through the API may hold the string form
    const env = await seeded({ saved: false, legacy: 'true' });
    expect(env).to.deep.equal({
      on: true, off: false, saved: false, legacy: true, host: 'localhost',
    });
  });
});
