import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import TaskParamsAgentForm from '@/components/TaskParamsAgentForm.vue';
import TaskForm from '@/components/TaskForm.vue';
import { AGENT_PROMPT_EXPRESSION } from '@/lib/constants';

function mountForm(props) {
  return mount(TaskParamsAgentForm, {
    props,
    global: { mocks: { $t: (k) => k } },
  });
}

function lastEmitted(wrapper) {
  const events = wrapper.emitted()['update:modelValue'] || [];
  return events.length ? events[events.length - 1][0] : undefined;
}

describe('TaskParamsAgentForm.vue', () => {
  const allowed = { prompt: 'fix the build', allow_override_prompt: true };

  it('shows nothing when the template does not let a task replace the prompt', () => {
    const wrapper = mountForm({ modelValue: {}, templateParams: { prompt: 'fix the build' } });

    expect(wrapper.find('textarea').exists()).to.equal(false);
  });

  it('starts with the prompt of the template', () => {
    const wrapper = mountForm({ modelValue: {}, templateParams: allowed });

    expect(wrapper.find('textarea').element.value).to.equal('fix the build');
  });

  it('starts with the prompt of the task it repeats', () => {
    const wrapper = mountForm({ modelValue: { prompt: 'fix the tests' }, templateParams: allowed });

    expect(wrapper.find('textarea').element.value).to.equal('fix the tests');
  });

  it('tells that the expressions are not replaced in a prompt given here', () => {
    const wrapper = mount(TaskParamsAgentForm, {
      props: { modelValue: {}, templateParams: allowed },
      global: { mocks: { $t: (k, params) => `${k} ${params?.expression ?? ''}` } },
    });

    expect(wrapper.text()).to.contain(`agentTaskPromptHint ${AGENT_PROMPT_EXPRESSION}`);
    expect(AGENT_PROMPT_EXPRESSION).to.contain('{ … }');
  });

  it('sends a prompt of its own and keeps the other params', async () => {
    const wrapper = mountForm({ modelValue: { debug: true }, templateParams: allowed });

    await wrapper.find('textarea').setValue('fix the tests');

    expect(lastEmitted(wrapper)).to.deep.equal({ debug: true, prompt: 'fix the tests' });
  });

  it('sends no prompt when it is the template one or empty', async () => {
    const wrapper = mountForm({ modelValue: { prompt: 'old', debug: true }, templateParams: allowed });

    await wrapper.find('textarea').setValue('fix the build');
    expect(lastEmitted(wrapper)).to.deep.equal({ debug: true });

    await wrapper.find('textarea').setValue('   ');
    expect(lastEmitted(wrapper)).to.deep.equal({ debug: true });
    // The emptied field stays empty instead of jumping back to the template prompt.
    expect(wrapper.find('textarea').element.value).to.equal('   ');
  });
});

describe('TaskForm.vue agent prompt', () => {
  const save = (template, params) => {
    const form = {
      isAgent: TaskForm.computed.isAgent.call({ template }),
      template,
      item: { params },
      editedEnvironment: {},
      editedSecretEnvironment: {},
    };
    TaskForm.methods.beforeSave.call(form);
    return form.item.params;
  };

  it('drops a repeated prompt the template no longer lets a task replace', () => {
    expect(save({ app: 'claude', task_params: { prompt: 'p' } }, { prompt: 'old' })).to.deep.equal({});
  });

  it('keeps the prompt when the template allows it', () => {
    const template = { app: 'codex', task_params: { allow_override_prompt: true } };
    expect(save(template, { prompt: 'new' })).to.deep.equal({ prompt: 'new' });
  });

  it('leaves the params of other apps alone', () => {
    expect(save({ app: 'ansible', task_params: {} }, { prompt: 'x' })).to.deep.equal({ prompt: 'x' });
  });
});
