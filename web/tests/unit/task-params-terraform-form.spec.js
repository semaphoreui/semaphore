import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import TaskParamsTerraformForm from '@/components/TaskParamsTerraformForm.vue';

function mountForm(app, templateParams = {}) {
  return mount(TaskParamsTerraformForm, {
    props: { app, modelValue: { destroy: true }, templateParams },
    global: { mocks: { $t: (key) => key } },
  });
}

describe('TaskParamsTerraformForm.vue — auto approval', () => {
  ['terraform', 'tofu', 'terragrunt'].forEach((app) => {
    it(`${app}: hides auto approval when the template does not allow it`, () => {
      [{}, { allow_auto_approve: false }, { auto_approve: true }].forEach((params) => {
        const wrapper = mountForm(app, params);
        expect(wrapper.find('[data-testid="taskParams-autoApprove"]').exists()).to.equal(false);
        expect(wrapper.find('[data-testid="taskParams-destroy"]').exists()).to.equal(true);
        wrapper.unmount();
      });
    });

    it(`${app}: emits auto approval together with destroy when allowed`, async () => {
      const wrapper = mountForm(app, { allow_auto_approve: true });
      await wrapper.get('[data-testid="taskParams-autoApprove"] input').setValue(true);

      const events = wrapper.emitted('update:modelValue');
      expect(events[events.length - 1][0]).to.include({ destroy: true, auto_approve: true });
      wrapper.unmount();
    });
  });
});
