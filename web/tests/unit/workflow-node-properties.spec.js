import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import WorkflowNodeProperties from '@/components/workflow/WorkflowNodeProperties.vue';
import TaskParamsForm from '@/components/TaskParamsForm.vue';

const templates = [
  {
    id: 7,
    name: 'Deploy',
    app: 'bash',
    survey_vars: [
      { name: 'image_tag', title: 'Image tag', type: '' },
      { name: 'replicas', title: 'Replicas', type: 'int' },
      { name: 'token', title: 'Token', type: 'secret' },
    ],
  },
];

const workflowVars = [
  { name: 'tag', title: 'Release tag', type: '' },
  { name: 'count', title: 'count', type: 'int' },
  { name: 'pwd', title: 'Password', type: 'secret' },
];

function mountPanel(node, surveyVars = workflowVars) {
  return mount(WorkflowNodeProperties, {
    // v-model of a Vue 3 component: modelValue / update:modelValue (Vue 2: value / input)
    props: {
      modelValue: node, templates, surveyVars, canManage: true,
    },
    global: {
      mocks: {
        $t: (k, params) => (params ? `${k}:${JSON.stringify(params)}` : k),
        $tc: (k, n) => `${k}:${n}`,
      },
      stubs: { TaskParamsForm: true },
    },
  });
}

function lastEmitted(w) {
  const events = w.emitted('update:modelValue');
  return events[events.length - 1][0];
}

function taskNode(extra = {}) {
  return {
    id: 1, kind: 'task', template_id: 7, task_params: {}, ...extra,
  };
}

describe('WorkflowNodeProperties.vue — workflow variable bindings', () => {
  it('hands the form the non-secret workflow variables and the node bindings', () => {
    const w = mountPanel(taskNode({ survey_inputs: [{ var: 'image_tag', workflow_var: 'tag' }] }));
    const form = w.findComponent(TaskParamsForm);
    expect(form.exists()).to.equal(true);
    expect(form.props('workflowVars')).to.deep.equal([
      { name: 'tag', title: 'Release tag', type: '' },
      { name: 'count', title: 'count', type: 'int' },
    ]);
    expect(form.props('surveyInputs')).to.deep.equal([{ var: 'image_tag', workflow_var: 'tag' }]);
  });

  it('hands the form an empty list when the workflow has no variables, still enabling the UI', () => {
    const w = mountPanel(taskNode(), []);
    expect(w.findComponent(TaskParamsForm).props('workflowVars')).to.deep.equal([]);
  });

  it('writes a binding into node.survey_inputs on bind and emits the node', () => {
    const w = mountPanel(taskNode());
    const form = w.findComponent(TaskParamsForm);
    form.vm.$emit('bind', 'image_tag', 'tag');
    expect(lastEmitted(w).survey_inputs).to.deep.equal([{ var: 'image_tag', workflow_var: 'tag' }]);

    form.vm.$emit('bind', 'replicas', 'count');
    expect(lastEmitted(w).survey_inputs).to.have.deep.members([
      { var: 'image_tag', workflow_var: 'tag' },
      { var: 'replicas', workflow_var: 'count' },
    ]);
  });

  it('replaces the binding of a variable instead of adding a second one', () => {
    const w = mountPanel(taskNode({ survey_inputs: [{ var: 'image_tag', workflow_var: 'tag' }] }));
    w.findComponent(TaskParamsForm).vm.$emit('bind', 'image_tag', 'count');
    expect(lastEmitted(w).survey_inputs).to.deep.equal([{ var: 'image_tag', workflow_var: 'count' }]);
  });

  it('removes the field when the last binding is unbound', () => {
    const w = mountPanel(taskNode({ survey_inputs: [{ var: 'image_tag', workflow_var: 'tag' }] }));
    w.findComponent(TaskParamsForm).vm.$emit('unbind', 'image_tag');
    expect(lastEmitted(w)).to.not.have.property('survey_inputs');
    expect(w.vm.boundVar('image_tag')).to.equal(null);
  });

  it('drops the bindings when the template changes or the node stops being a task', () => {
    const w = mountPanel(taskNode({ survey_inputs: [{ var: 'image_tag', workflow_var: 'tag' }] }));
    w.vm.onTemplateChanged();
    expect(lastEmitted(w)).to.not.have.property('survey_inputs');

    const w2 = mountPanel(taskNode({ id: 2, survey_inputs: [{ var: 'image_tag', workflow_var: 'tag' }] }));
    w2.vm.node.kind = 'approval';
    w2.vm.onKindChanged();
    expect(lastEmitted(w2)).to.not.have.property('survey_inputs');
    expect(lastEmitted(w2).template_id).to.equal(null);
  });

  it('renders no task form for approval, delay and note nodes', () => {
    ['approval', 'delay', 'note'].forEach((kind) => {
      const w = mountPanel({ id: 3, kind, delay_seconds: 60 });
      expect(w.findComponent(TaskParamsForm).exists()).to.equal(false);
    });
  });
});
