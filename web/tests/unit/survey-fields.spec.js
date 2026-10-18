import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import SurveyFields from '@/components/SurveyFields.vue';
import WorkflowVarAssign from '@/components/workflow/WorkflowVarAssign.vue';

const vars = [
  {
    name: 'image_tag', title: 'Image tag', type: '', required: true,
  },
  { name: 'replicas', title: 'Replicas', type: 'int' },
  {
    name: 'env', title: 'Env', type: 'enum', values: [{ name: 'prod', value: 'prod' }],
  },
  {
    name: 'subnets', title: 'Subnets', type: 'select', values: [{ name: 'a', value: 'a' }, { name: 'b', value: 'b' }],
  },
  { name: 'notes', title: 'Notes', type: 'text' },
  { name: 'token', title: 'Token', type: 'secret' },
];

const workflowVars = [
  { name: 'tag', title: 'Release tag', type: '' },
  { name: 'count', title: 'count', type: 'int' },
];

function mountFields(props = {}) {
  const localVue = createLocalVue();
  return mount(SurveyFields, {
    localVue,
    vuetify: new Vuetify(),
    propsData: {
      vars, value: { image_tag: 'static', subnets: ['a', 'b'] }, secretValue: {}, ...props,
    },
    mocks: { $t: (k, params) => (params ? `${k}:${JSON.stringify(params)}` : k) },
    stubs: { WorkflowVarAssign: true },
  });
}

function labels(w) {
  return w.findAll('.v-label').wrappers.map((l) => l.text());
}

describe('SurveyFields.vue', () => {
  describe('fields by type', () => {
    it('renders one field per variable, required ones starred, secrets masked', () => {
      const w = mountFields();
      expect(labels(w)).to.deep.equal(['Image tag *', 'Replicas', 'Env', 'Subnets', 'Notes', 'Token']);
      expect(w.findAll('.v-select').length).to.equal(2);
      expect(w.findAll('textarea').length).to.equal(1);
      expect(w.find('.masked-secret-input').exists()).to.equal(true);
    });

    it('edits the given objects in place', async () => {
      const value = { image_tag: 'static', subnets: [] };
      const secretValue = {};
      const w = mountFields({ value, secretValue });
      const inputs = w.findAll('input[type="text"], input:not([type])');
      await inputs.at(0).setValue('1.2.3');
      expect(value.image_tag).to.equal('1.2.3');
      await w.find('.masked-secret-input input').setValue('s3cr3t');
      expect(secretValue.token).to.equal('s3cr3t');
    });

    it('follows a replaced value object', async () => {
      const w = mountFields();
      const next = { image_tag: 'next', subnets: [] };
      await w.setProps({ value: next });
      expect(w.vm.env).to.equal(next);
      expect(w.find('input').element.value).to.equal('next');
    });

    it('removes a chip of a select variable from the shared object', () => {
      const value = { image_tag: 'x', subnets: ['a', 'b'] };
      const w = mountFields({ value });
      w.vm.removeSelectedItem('subnets', 0);
      expect(value.subnets).to.deep.equal(['b']);
      w.vm.removeSelectedItem('missing', 0);
      w.vm.removeSelectedItem('image_tag', 0);
      expect(value.image_tag).to.equal('x');
    });

    it('shows no assign buttons without workflow variables', () => {
      const w = mountFields();
      expect(w.findAllComponents(WorkflowVarAssign).length).to.equal(0);
      expect(w.find('.SurveyFields__bound').exists()).to.equal(false);
    });
  });

  describe('workflow variable bindings', () => {
    it('offers an assign button on every non-secret field, even with an empty list', () => {
      const empty = mountFields({ workflowVars: [] });
      expect(empty.findAllComponents(WorkflowVarAssign).length).to.equal(5);
      const w = mountFields({ workflowVars });
      expect(w.findAllComponents(WorkflowVarAssign).length).to.equal(5);
      w.findAllComponents(WorkflowVarAssign).wrappers.forEach((a) => {
        expect(a.props('vars')).to.deep.equal(workflowVars);
      });
    });

    it('emits bind(var, workflow_var) when a workflow variable is picked for a field', () => {
      const w = mountFields({ workflowVars });
      // Assign buttons follow the survey order: image_tag, replicas, env, subnets, notes.
      w.findAllComponents(WorkflowVarAssign).at(1).vm.$emit('select', 'count');
      expect(w.emitted('bind')).to.deep.equal([['replicas', 'count']]);
    });

    it('renders a bound field read-only with the workflow variable and an unbind cross', async () => {
      const w = mountFields({ workflowVars, surveyInputs: [{ var: 'image_tag', workflow_var: 'tag' }] });
      const bound = w.find('.SurveyFields__bound');
      expect(bound.exists()).to.equal(true);
      const input = bound.find('input');
      expect(input.attributes('readonly')).to.equal('readonly');
      expect(input.element.value).to.equal('Release tag (tag)');
      expect(bound.find('.v-label').text()).to.equal('Image tag *');
      expect(bound.find('.mdi-link-variant').exists()).to.equal(true);
      expect(bound.text()).to.contain('workflowNodeSurveyInputsBoundHint');
      expect(w.findAllComponents(WorkflowVarAssign).length)
        .to.equal(4, 'the bound field has no assign button');

      await bound.find('.SurveyFields__unbind').trigger('click');
      expect(w.emitted('unbind')).to.deep.equal([['image_tag']]);
    });

    it('flags a binding whose workflow variable no longer exists', () => {
      const w = mountFields({ workflowVars, surveyInputs: [{ var: 'replicas', workflow_var: 'gone' }] });
      const bound = w.find('.SurveyFields__bound');
      expect(bound.find('input').element.value).to.equal('gone');
      expect(bound.text()).to.contain('workflowNodeSurveyInputsUnknown:{"name":"gone"}');
      expect(bound.classes()).to.include('error--text');
    });

    it('never offers to bind a secret field', () => {
      const w = mountFields({ workflowVars });
      expect(w.vm.canBind({ name: 'token', type: 'secret' })).to.equal(false);
      expect(w.vm.canBind({ name: 'image_tag', type: '' })).to.equal(true);
    });
  });
});

describe('WorkflowVarAssign.vue', () => {
  function mountAssign(list) {
    const localVue = createLocalVue();
    return mount(WorkflowVarAssign, {
      localVue,
      vuetify: new Vuetify(),
      propsData: { vars: list },
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
