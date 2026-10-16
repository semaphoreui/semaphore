import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TaskParamsForm from '@/components/TaskParamsForm.vue';
import SurveyFields from '@/components/SurveyFields.vue';

const template = {
  id: 7,
  app: 'bash',
  survey_vars: [
    {
      name: 'image_tag', title: 'Image tag', type: '', required: true,
    },
    {
      name: 'replicas', title: 'Replicas', type: 'int', default_value: '1',
    },
  ],
};

const workflowVars = [{ name: 'tag', title: 'Release tag', type: '' }];

function mountForm(props = {}) {
  const localVue = createLocalVue();
  return mount(TaskParamsForm, {
    localVue,
    vuetify: new Vuetify(),
    propsData: { value: { environment: '{"image_tag":"static"}' }, template, ...props },
    mocks: { $t: (k, params) => (params ? `${k}:${JSON.stringify(params)}` : k) },
    stubs: { SurveyFields: true, ArgsPicker: true },
  });
}

const settle = () => new Promise((resolve) => { setTimeout(resolve, 10); });

describe('TaskParamsForm.vue — survey fields and bindings', () => {
  it('renders the survey through SurveyFields with the edited environment and defaults', async () => {
    const w = mountForm();
    await settle();
    const fields = w.findComponent(SurveyFields);
    expect(fields.exists()).to.equal(true);
    expect(fields.props('vars')).to.deep.equal(template.survey_vars);
    expect(fields.props('value')).to.deep.equal({ image_tag: 'static', replicas: '1' });
    expect(fields.props('workflowVars')).to.equal(null, 'no binding UI for a plain task form');
  });

  it('passes the workflow variables and bindings through and re-emits bind/unbind', async () => {
    const surveyInputs = [{ var: 'image_tag', workflow_var: 'tag' }];
    const w = mountForm({ workflowVars, surveyInputs });
    await settle();
    const fields = w.findComponent(SurveyFields);
    expect(fields.props('workflowVars')).to.deep.equal(workflowVars);
    expect(fields.props('surveyInputs')).to.deep.equal(surveyInputs);

    fields.vm.$emit('bind', 'replicas', 'tag');
    expect(w.emitted('bind')).to.deep.equal([['replicas', 'tag']]);
    fields.vm.$emit('unbind', 'image_tag');
    expect(w.emitted('unbind')).to.deep.equal([['image_tag']]);
  });

  it('still emits input with the environment when a survey value changes', async () => {
    const w = mountForm();
    await settle();
    const before = (w.emitted('input') || []).length;
    w.vm.editedEnvironment.image_tag = '1.2.3';
    await w.vm.$nextTick();
    const emitted = w.emitted('input');
    expect(emitted.length).to.be.greaterThan(before);
    expect(JSON.parse(emitted[emitted.length - 1][0].environment).image_tag).to.equal('1.2.3');
  });
});
