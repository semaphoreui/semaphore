import './setup';
import { expect } from 'chai';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import AppsMixin from '@/components/AppsMixin';

const backendApps = [
  { id: 'ansible', active: true },
  { id: 'terraform', active: false },
  {
    id: 'custom', active: true, title: 'My App', icon: 'rocket', color: 'purple',
  },
];

function mountHost(apps = backendApps) {
  const localVue = createLocalVue();
  const Host = {
    mixins: [AppsMixin],
    methods: {
      loadAppsDataFromBackend: async () => apps,
    },
    render: (h) => h('div'),
  };
  return shallowMount(Host, { localVue, vuetify: new Vuetify() });
}

async function mountLoaded(apps) {
  const wrapper = mountHost(apps);
  await wrapper.vm.$nextTick();
  await wrapper.vm.$nextTick();
  expect(wrapper.vm.isAppsLoaded).to.equal(true);
  return wrapper;
}

describe('AppsMixin', () => {
  it('is not loaded until the backend responds', () => {
    const wrapper = mountHost();
    expect(wrapper.vm.isAppsLoaded).to.equal(false);
  });

  it('collects active app ids and indexes apps by id', async () => {
    const wrapper = await mountLoaded();
    expect(wrapper.vm.appsMixin.activeAppIds).to.deep.equal(['ansible', 'custom']);
    expect(wrapper.vm.appsMixin.apps.custom.title).to.equal('My App');
  });

  describe('getAppTitle', () => {
    it('prefers the title from the backend', async () => {
      const wrapper = await mountLoaded();
      expect(wrapper.vm.getAppTitle('custom')).to.equal('My App');
      expect(wrapper.vm.getAppTitle('custom', true)).to.equal('My App');
    });

    it('uses built-in long and short titles', async () => {
      const wrapper = await mountLoaded();
      expect(wrapper.vm.getAppTitle('terraform')).to.equal('Terraform Code');
      expect(wrapper.vm.getAppTitle('terraform', true)).to.equal('Terraform');
    });

    it('returns an empty string for unknown apps', async () => {
      const wrapper = await mountLoaded();
      expect(wrapper.vm.getAppTitle('nope')).to.equal('');
    });
  });

  describe('getAppIcon', () => {
    it('prefixes backend icons with mdi-', async () => {
      const wrapper = await mountLoaded();
      expect(wrapper.vm.getAppIcon('custom')).to.equal('mdi-rocket');
    });

    it('uses built-in icons', async () => {
      const wrapper = await mountLoaded();
      expect(wrapper.vm.getAppIcon('ansible')).to.equal('mdi-ansible');
    });

    it('returns a help icon for unknown apps', async () => {
      const wrapper = await mountLoaded();
      expect(wrapper.vm.getAppIcon('nope')).to.equal('mdi-help');
    });
  });

  describe('getAppColor', () => {
    it('prefers the backend color', async () => {
      const wrapper = await mountLoaded();
      expect(wrapper.vm.getAppColor('custom')).to.equal('purple');
    });

    it('depends on the theme for built-in apps', async () => {
      const wrapper = await mountLoaded();
      wrapper.vm.$vuetify.theme.dark = false;
      expect(wrapper.vm.getAppColor('ansible')).to.equal('black');
      wrapper.vm.$vuetify.theme.dark = true;
      expect(wrapper.vm.getAppColor('ansible')).to.equal('white');
    });

    it('returns gray for unknown apps', async () => {
      const wrapper = await mountLoaded();
      expect(wrapper.vm.getAppColor('nope')).to.equal('gray');
    });
  });
});
