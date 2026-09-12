import { expect } from 'chai';
import { shallowMount } from '@vue/test-utils';
import AppFieldsMixin from '@/components/AppFieldsMixin';

const Host = {
  mixins: [AppFieldsMixin],
  props: { app: String },
  render: (h) => h('div'),
};

function mountFor(app) {
  return shallowMount(Host, {
    propsData: { app },
    mocks: { $t: (k) => `t:${k}` },
  }).vm;
}

describe('AppFieldsMixin', () => {
  describe('needField', () => {
    const tests = [
      { app: 'ansible', field: 'inventory', expected: true },
      { app: 'ansible', field: 'vault', expected: true },
      { app: 'ansible', field: 'auto_approve', expected: false },
      { app: 'terraform', field: 'auto_approve', expected: true },
      { app: 'terraform', field: 'vault', expected: false },
      { app: 'tofu', field: 'override_backend', expected: true },
      { app: 'terragrunt', field: 'inventory', expected: true },
      { app: 'bash', field: 'inventory', expected: false },
      { app: 'bash', field: 'playbook', expected: true },
      { app: 'bash', field: 'repository', expected: true },
      { app: 'unknown-app', field: 'vault', expected: false },
      { app: '', field: 'limit', expected: true },
    ];

    tests.forEach(({ app, field, expected }) => {
      it(`${app || '(empty)'} / ${field} -> ${expected}`, () => {
        expect(mountFor(app).needField(field)).to.equal(expected);
      });
    });
  });

  describe('isFieldRequired', () => {
    const tests = [
      { app: 'ansible', field: 'playbook', expected: true },
      { app: 'ansible', field: 'environment', expected: false },
      { app: 'terraform', field: 'playbook', expected: false },
      { app: 'terraform', field: 'repository', expected: true },
      { app: 'bash', field: 'inventory', expected: false },
      { app: 'bash', field: 'environment', expected: false },
    ];

    tests.forEach(({ app, field, expected }) => {
      it(`${app} / ${field} -> ${expected}`, () => {
        expect(mountFor(app).isFieldRequired(field)).to.equal(expected);
      });
    });
  });

  describe('fieldLabel', () => {
    it('translates the configured label', () => {
      expect(mountFor('ansible').fieldLabel('vault')).to.equal('t:vaultPassword2');
    });

    it('uses the app-specific label override', () => {
      expect(mountFor('terraform').fieldLabel('playbook')).to.equal('t:Subdirectory path (Optional)');
    });

    it('falls back to the field name for unknown fields', () => {
      expect(mountFor('ansible').fieldLabel('no_such_field')).to.equal('t:no_such_field');
    });
  });
});
