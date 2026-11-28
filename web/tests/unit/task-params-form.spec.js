import { expect } from 'chai';
import { mount, shallowMount } from '@vue/test-utils';
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

// A schedule or an integration keeps one TaskParamsForm while its template is changed
// (ScheduleForm, IntegrationForm); `item.task_params` is what they save.
describe('TaskParamsForm.vue agent prompt', () => {
  const agent = (id, taskParams) => ({
    id, app: 'claude', survey_vars: [], task_params: taskParams,
  });
  const open = agent(1, { prompt: 'fix the build', allow_override_prompt: true });
  const alsoOpen = agent(2, { prompt: 'write the docs', allow_override_prompt: true });
  const closed = agent(3, { prompt: 'review the code' });

  // Mounted in full: the prompt field is a real textarea.
  async function mountForm(template, params) {
    const wrapper = mount(TaskParamsForm, {
      props: { modelValue: { params }, template },
      global: { mocks: { $t: (k) => k } },
    });
    await new Promise((resolve) => { setTimeout(resolve, 10); });
    return wrapper;
  }

  const prompt = (wrapper) => wrapper.find('[data-testid="taskParams-agentPrompt"] textarea');

  // The params as they are sent: the form keeps `tags`, `limit` and `skip_tags` as undefined keys.
  const sent = (params) => JSON.parse(JSON.stringify(params));

  function lastEmitted(wrapper) {
    const events = wrapper.emitted()['update:modelValue'] || [];
    return events.length ? events[events.length - 1][0] : undefined;
  }

  it('drops the prompt when the new template does not let a task replace it', async () => {
    const wrapper = await mountForm(open, { prompt: 'fix the tests', debug: true });
    expect(prompt(wrapper).element.value).to.equal('fix the tests');

    await wrapper.setProps({ template: closed });

    expect(prompt(wrapper).exists()).to.equal(false);
    expect(sent(wrapper.vm.item.params)).to.deep.equal({ debug: true });
    // The owner saves what it was told, not what it gave.
    expect(sent(lastEmitted(wrapper).params)).to.deep.equal({ debug: true });
  });

  it('starts over with the prompt of the new template', async () => {
    const wrapper = await mountForm(open, { prompt: 'fix the tests' });

    await wrapper.setProps({ template: alsoOpen });

    expect(prompt(wrapper).element.value).to.equal('write the docs');
    expect(wrapper.vm.item.params).to.not.have.property('prompt');
    expect(lastEmitted(wrapper).params).to.not.have.property('prompt');
  });

  it('shows the prompt of the new template when the task had none of its own', async () => {
    const wrapper = await mountForm(open, {});
    expect(prompt(wrapper).element.value).to.equal('fix the build');

    await wrapper.setProps({ template: alsoOpen });

    expect(prompt(wrapper).element.value).to.equal('write the docs');
  });

  it('keeps a prompt typed after the change', async () => {
    const wrapper = await mountForm(open, { prompt: 'fix the tests' });
    await wrapper.setProps({ template: alsoOpen });

    await prompt(wrapper).setValue('write the changelog');

    expect(sent(lastEmitted(wrapper).params)).to.deep.equal({ prompt: 'write the changelog' });
  });

  it('drops a saved prompt the template no longer lets a task replace', async () => {
    const params = { prompt: 'fix the tests', debug: true };
    const wrapper = await mountForm(closed, params);

    expect(prompt(wrapper).exists()).to.equal(false);
    expect(sent(wrapper.vm.item.params)).to.deep.equal({ debug: true });
    // The owner's params are the same object: what it saves has no prompt either.
    expect(sent(params)).to.deep.equal({ debug: true });
  });

  it('keeps a saved prompt the template allows, and the template when it is loaded again', async () => {
    const wrapper = await mountForm(open, { prompt: 'fix the tests' });

    // The same template, a new object: the list of templates was loaded again.
    await wrapper.setProps({ template: { ...open } });

    expect(prompt(wrapper).element.value).to.equal('fix the tests');
    expect(sent(wrapper.vm.item.params)).to.deep.equal({ prompt: 'fix the tests' });
  });

  it('leaves the params of other apps alone', async () => {
    const ansible = { id: 4, app: 'ansible', survey_vars: [] };
    const wrapper = await mountForm(ansible, { prompt: 'x' });

    expect(wrapper.vm.item.params.prompt).to.equal('x');
  });
});
