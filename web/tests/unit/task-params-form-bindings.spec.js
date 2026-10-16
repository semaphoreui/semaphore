import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TaskParamsForm from '@/components/TaskParamsForm.vue';
import WorkflowVarAssign from '@/components/workflow/WorkflowVarAssign.vue';

const template = {
  id: 7,
  app: 'bash',
  survey_vars: [
    {
      name: 'image_tag', title: 'Image tag', type: '', required: true,
    },
    { name: 'replicas', title: 'Replicas', type: 'int' },
    {
      name: 'env', title: 'Env', type: 'enum', values: [{ name: 'prod', value: 'prod' }],
    },
    { name: 'notes', title: 'Notes', type: 'text' },
    { name: 'token', title: 'Token', type: 'secret' },
  ],
};

const workflowVars = [
  { name: 'tag', title: 'Release tag', type: '' },
  { name: 'count', title: 'count', type: 'int' },
];

function mountForm(props = {}) {
  const localVue = createLocalVue();
  return mount(TaskParamsForm, {
    localVue,
    vuetify: new Vuetify(),
    propsData: { value: { environment: '{"image_tag":"static"}' }, template, ...props },
    mocks: { $t: (k, params) => (params ? `${k}:${JSON.stringify(params)}` : k) },
    stubs: { WorkflowVarAssign: true, ArgsPicker: true },
  });
}

function fieldLabels(w) {
  return w.findAll('.v-label').wrappers.map((l) => l.text());
}

describe('TaskParamsForm.vue — workflow variable bindings', () => {
  it('shows no assign buttons without workflow variables (plain task forms)', () => {
    const w = mountForm();
    expect(w.findAllComponents(WorkflowVarAssign).length).to.equal(0);
    expect(w.find('.TaskParamsForm__bound').exists()).to.equal(false);
  });

  it('offers an assign button on every non-secret survey field, even with an empty list', () => {
    const w = mountForm({ workflowVars: [] });
    expect(w.findAllComponents(WorkflowVarAssign).length).to.equal(4);
    const w2 = mountForm({ workflowVars });
    expect(w2.findAllComponents(WorkflowVarAssign).length).to.equal(4);
    w2.findAllComponents(WorkflowVarAssign).wrappers.forEach((a) => {
      expect(a.props('vars')).to.deep.equal(workflowVars);
    });
  });

  it('emits bind(var, workflow_var) when a workflow variable is picked for a field', () => {
    const w = mountForm({ workflowVars });
    // Assign buttons follow the survey order: image_tag, replicas, env, notes.
    w.findAllComponents(WorkflowVarAssign).at(1).vm.$emit('select', 'count');
    expect(w.emitted('bind')).to.deep.equal([['replicas', 'count']]);
  });

  it('renders a bound field read-only with the workflow variable, no assign button, and an unbind cross', async () => {
    const w = mountForm({ workflowVars, surveyInputs: [{ var: 'image_tag', workflow_var: 'tag' }] });
    const bound = w.find('.TaskParamsForm__bound');
    expect(bound.exists()).to.equal(true);
    const input = bound.find('input');
    expect(input.attributes('readonly')).to.equal('readonly');
    expect(input.element.value).to.equal('Release tag (tag)');
    expect(bound.find('.v-label').text()).to.equal('Image tag *');
    expect(bound.find('.mdi-link-variant').exists()).to.equal(true);
    expect(w.findAllComponents(WorkflowVarAssign).length).to.equal(3, 'the bound field has no assign button');
    expect(fieldLabels(w)).to.include('Replicas');

    await bound.find('.TaskParamsForm__unbind').trigger('click');
    expect(w.emitted('unbind')).to.deep.equal([['image_tag']]);
  });

  it('flags a binding whose workflow variable no longer exists', () => {
    const w = mountForm({ workflowVars, surveyInputs: [{ var: 'replicas', workflow_var: 'gone' }] });
    const bound = w.find('.TaskParamsForm__bound');
    expect(bound.find('input').element.value).to.equal('gone');
    expect(bound.text()).to.contain('workflowNodeSurveyInputsUnknown:{"name":"gone"}');
    expect(bound.classes()).to.include('error--text');
  });

  it('never binds a secret field', () => {
    const w = mountForm({ workflowVars, surveyInputs: [{ var: 'token', workflow_var: 'tag' }] });
    expect(w.find('.TaskParamsForm__bound').exists()).to.equal(true, 'a stored binding is still shown');
    expect(w.vm.canBind({ name: 'token', type: 'secret' })).to.equal(false);
  });
});

describe('WorkflowVarAssign.vue', () => {
  function mountAssign(vars) {
    const localVue = createLocalVue();
    return mount(WorkflowVarAssign, {
      localVue,
      vuetify: new Vuetify(),
      propsData: { vars },
      mocks: { $t: (k) => k },
    });
  }

  it('lists the workflow variables with name and type and emits select', async () => {
    const w = mountAssign(workflowVars);
    const items = w.findAll('.WorkflowVarAssign__item');
    expect(items.length).to.equal(2);
    expect(items.at(0).text()).to.contain('Release tag');
    expect(items.at(0).text()).to.contain('tag · string');
    expect(items.at(1).text()).to.contain('count · int');
    await items.at(1).trigger('click');
    expect(w.emitted('select')).to.deep.equal([['count']]);
  });

  it('explains an empty list', () => {
    const w = mountAssign([]);
    expect(w.findAll('.WorkflowVarAssign__item').length).to.equal(0);
    expect(w.text()).to.contain('workflowNodeSurveyInputsNoWorkflowVars');
  });
});
