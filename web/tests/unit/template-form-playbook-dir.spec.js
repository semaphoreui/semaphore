import { expect } from 'chai';
import { shallowMount } from '@vue/test-utils';
import AppFieldsMixin from '@/components/AppFieldsMixin';
import TemplateForm from '@/components/TemplateForm.vue';

// Exercise TemplateForm's directory handler without mounting the full form.
const Host = {
  mixins: [AppFieldsMixin],
  props: { app: String },
  data() {
    return { playbookDir: '', loaded: 0 };
  },
  methods: {
    loadPlaybooks() {
      this.loaded += 1;
    },
    onPlaybookSearch: TemplateForm.methods.onPlaybookSearch,
  },
  render: (h) => h('div'),
};

const mountFor = (app) => shallowMount(Host, {
  propsData: { app },
  mocks: { $t: (k) => `t:${k}` },
}).vm;

describe('TemplateForm playbook directory drill-down', () => {
  it('asks for the directory when the typed path ends with a separator', () => {
    const vm = mountFor('terraform');

    vm.onPlaybookSearch('roles/');

    expect(vm.playbookDir).to.equal('roles/');
    expect(vm.loaded).to.equal(1);
  });

  it('goes a level deeper', () => {
    const vm = mountFor('terraform');

    vm.onPlaybookSearch('roles/');
    vm.onPlaybookSearch('roles/common/');

    expect(vm.playbookDir).to.equal('roles/common/');
    expect(vm.loaded).to.equal(2);
  });

  it('keeps the listing while the rest of a name is typed', () => {
    const vm = mountFor('terraform');

    vm.onPlaybookSearch('roles/');
    vm.onPlaybookSearch('roles/com');
    vm.onPlaybookSearch('roles/comm');

    expect(vm.playbookDir).to.equal('roles/');
    expect(vm.loaded).to.equal(1);
  });

  // v-combobox emits the typed text and then null, every time. Acting on the
  // null cancelled the request just made and listed the root instead.
  it('ignores the null vuetify emits after each change', () => {
    const vm = mountFor('terraform');

    vm.onPlaybookSearch('roles/');
    vm.onPlaybookSearch(null);

    expect(vm.playbookDir).to.equal('roles/');
    expect(vm.loaded).to.equal(1);
  });

  it('returns to the root when the path no longer has a separator', () => {
    const vm = mountFor('terraform');

    vm.onPlaybookSearch('roles/');
    vm.onPlaybookSearch('rol');

    expect(vm.playbookDir).to.equal('');
    expect(vm.loaded).to.equal(2);
  });

  ['tofu', 'terragrunt', 'pulumi'].forEach((app) => {
    it(`applies to ${app}`, () => {
      const vm = mountFor(app);

      vm.onPlaybookSearch('envs/');

      expect(vm.playbookDir).to.equal('envs/');
      expect(vm.loaded).to.equal(1);
    });
  });

  ['ansible', 'bash', 'python', ''].forEach((app) => {
    it(`does nothing for ${app || '(empty)'}, whose field names a file`, () => {
      const vm = mountFor(app);

      vm.onPlaybookSearch('scripts/');

      expect(vm.playbookDir).to.equal('');
      expect(vm.loaded).to.equal(0);
    });
  });
});
