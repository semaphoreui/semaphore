import './setup';
import { expect } from 'chai';
import { mount, flushPromises } from '@vue/test-utils';
import WorkflowRunDialog from '@/components/WorkflowRunDialog.vue';
import SurveyFields from '@/components/SurveyFields.vue';

const workflow = {
  id: 11,
  name: 'Deploy',
  survey_vars: [
    {
      name: 'env', title: 'Environment', type: 'enum', values: [{ name: 'Prod', value: 'prod' }, { name: 'Staging', value: 'staging' }], default_value: 'staging',
    },
    {
      name: 'image_tag', title: 'Image tag', type: '', required: true,
    },
    {
      name: 'replicas', title: 'Replicas', type: 'int', default_value: '2',
    },
    {
      name: 'subnets', title: 'Subnets', type: 'select', values: [{ name: 'a', value: 'a' }], default_value: ['a'],
    },
    {
      name: 'regions', title: 'Regions', type: 'select', values: [{ name: 'eu', value: 'eu' }],
    },
  ],
};

function mountDialog(props = {}) {
  return mount(WorkflowRunDialog, {
    // v-model of a Vue 3 component: modelValue / update:modelValue (Vue 2: value / input)
    props: { modelValue: true, workflow, ...props },
    global: {
      mocks: { $t: (k, params) => (params ? `${k}:${JSON.stringify(params)}` : k) },
      stubs: { 'v-dialog': { template: '<div><slot /></div>' } },
    },
  });
}

describe('WorkflowRunDialog.vue', () => {
  it('opens with the survey defaults filled in, a select always a list', () => {
    const w = mountDialog();
    expect(w.vm.environment).to.deep.equal({
      env: 'staging', image_tag: '', replicas: '2', subnets: ['a'], regions: [],
    });
    expect(w.findComponent(SurveyFields).props('vars').map((v) => v.name))
      .to.deep.equal(['env', 'image_tag', 'replicas', 'subnets', 'regions']);
    expect(w.text()).to.contain('workflowRunDialogTitle:{"name":"Deploy"}');
  });

  it('does not emit run while a required answer is missing', async () => {
    const w = mountDialog();
    await w.find('[data-testid="workflow-run-dialog-run"]').trigger('click');
    // run() awaits the validation of the form (Vuetify 3: validate() is async)
    await flushPromises();
    expect(w.emitted('run')).to.equal(undefined);
    expect(w.text()).to.contain('Image tag isRequired');
  });

  it('emits run with the answers as an object', async () => {
    const w = mountDialog();
    w.vm.environment.image_tag = '1.2.3';
    w.vm.environment.env = 'prod';
    await w.vm.$nextTick();
    await w.find('[data-testid="workflow-run-dialog-run"]').trigger('click');
    await flushPromises();
    expect(w.emitted('run')).to.deep.equal([[{
      env: 'prod', image_tag: '1.2.3', replicas: '2', subnets: ['a'], regions: [],
    }]]);
  });

  it('reseeds the answers every time it opens and mirrors v-model', async () => {
    const w = mountDialog({ modelValue: false });
    expect(w.vm.environment).to.equal(null);
    await w.setProps({ modelValue: true });
    expect(w.vm.environment.replicas).to.equal('2');
    w.vm.environment.replicas = '9';
    await w.setProps({ modelValue: false });
    await w.setProps({ modelValue: true });
    expect(w.vm.environment.replicas).to.equal('2');

    w.vm.dialog = false;
    await w.vm.$nextTick();
    expect(w.emitted('update:modelValue')).to.deep.equal([[false]]);
  });
});
