import { expect } from 'chai';
import { mount, flushPromises } from '@vue/test-utils';
import HelpPage from '@/components/HelpPage.vue';
import HelpButton from '@/components/HelpButton.vue';
// Vitest has run the setup ahead of the spec already (`test.setupFiles`).
import createTestVuetify from './setup';

// The help documents the page can open: the registry is replaced, so the specs do not depend on
// the documents the application ships.
vi.mock('@/help', () => ({
  hasHelpDoc: (key) => ['demo-page', 'other-page'].includes(key),
  helpDocLocale: (key) => (['demo-page', 'other-page'].includes(key) ? 'en' : null),
  loadHelpDoc: (key) => Promise.resolve({
    key,
    locale: 'en',
    title: `Help of ${key}`,
    intro: [{ type: 'p', inline: [{ type: 'text', text: 'What the page is for.' }] }],
    // The last element is one the page of the specs does not show.
    sections: ['demo-name', 'demo-check', 'demo-delete', 'demo-gone'].map((field) => ({
      field,
      title: field,
      blocks: [{ type: 'p', inline: [{ type: 'text', text: `About ${field}.` }] }],
    })),
  }),
}));

// A page as the Settings page is: a toolbar with the help button, a field of a form (the
// `data-testid` on the root of the input, around its frame) and a button.
const Page = {
  components: { HelpPage, HelpButton },
  props: { helpDoc: String },
  template: `
    <HelpPage :help-doc="helpDoc" data-testid="demo-page" v-slot="{ docs }">
      <HelpButton
        v-if="docs.available"
        page
        :open="docs.open"
        data-testid="demo-docs"
        @click="docs.toggle()"
      />
      <div data-testid="demo-name"><div class="v-field"><input class="demo-name"></div></div>
      <div data-testid="demo-check">
        <div class="v-input__control"><input type="checkbox" class="demo-check"></div>
        <div class="v-input__details"></div>
      </div>
      <button type="button" data-testid="demo-delete" class="demo-delete">Delete</button>
      <button type="button" class="demo-other">Other</button>
    </HelpPage>`,
};

const MARK = 'data-help-docs-field';

describe('HelpPage.vue', () => {
  const wrappers = [];

  const byTestId = (id) => document.querySelector(`[data-testid="${id}"]`);
  const panel = () => byTestId('editDialog-docsPanel');
  const section = (field) => byTestId(`editDialog-docsSection-${field}`);
  const marked = () => [...document.querySelectorAll(`[${MARK}]`)];

  function mountPage(helpDoc = 'demo-page', plugins = []) {
    const el = document.createElement('div');
    document.body.appendChild(el);
    const wrapper = mount(Page, {
      attachTo: el,
      props: { helpDoc },
      global: { mocks: { $t: (k) => k }, plugins },
    });
    wrappers.push(wrapper);
    return wrapper;
  }

  async function settled(wrapper) {
    await flushPromises();
    await wrapper.vm.$nextTick();
  }

  async function openPanel(wrapper) {
    byTestId('demo-docs').click();
    await settled(wrapper);
  }

  async function focusIn(wrapper, el) {
    el.dispatchEvent(new FocusEvent('focusin', { bubbles: true }));
    await settled(wrapper);
  }

  async function pointer(wrapper, el, type) {
    el.dispatchEvent(new MouseEvent(type));
    await settled(wrapper);
  }

  afterEach(() => {
    while (wrappers.length) {
      wrappers.pop().unmount();
    }
    document.body.innerHTML = '';
  });

  it('keeps the attributes of the page on its root', () => {
    const wrapper = mountPage();

    expect(wrapper.attributes('data-testid')).to.equal('demo-page');
    expect(wrapper.classes()).to.include('HelpPage');
  });

  it('offers no help button for a page without a document', () => {
    mountPage('no-such-page');

    expect(byTestId('demo-docs')).to.equal(null);
    expect(byTestId('demo-name'), 'the page itself is drawn').to.not.equal(null);
  });

  it('labels the button as the help of a page', () => {
    mountPage();

    expect(byTestId('demo-docs').getAttribute('aria-label')).to.equal('helpPanelOpenPage');
    expect(byTestId('demo-docs').getAttribute('aria-pressed')).to.equal('false');
  });

  it('opens the panel with the button and closes it with the button or the panel', async () => {
    const wrapper = mountPage();
    expect(panel()).to.equal(null);

    await openPanel(wrapper);

    expect(byTestId('editDialog-docsTitle').textContent).to.equal('Help of demo-page');
    expect(byTestId('demo-docs').getAttribute('aria-pressed')).to.equal('true');

    await openPanel(wrapper);
    expect(panel()).to.equal(null);

    await openPanel(wrapper);
    byTestId('editDialog-docsClose').click();
    await settled(wrapper);
    expect(panel()).to.equal(null);
  });

  it('draws the panel next to the content of the page, not inside it', async () => {
    const wrapper = mountPage();
    await openPanel(wrapper);

    const content = wrapper.find('.HelpPage__content').element;
    expect(panel().parentElement).to.equal(wrapper.element);
    expect(content.contains(panel())).to.equal(false);
    expect(content.contains(byTestId('demo-name'))).to.equal(true);
  });

  describe('layout', () => {
    // The display is measured when the Vuetify instance is created.
    function vuetifyOfWidth(width) {
      const innerWidth = window.innerWidth;
      window.innerWidth = width;
      const vuetify = createTestVuetify();
      window.innerWidth = innerWidth;
      return vuetify;
    }

    it('leaves the panel its place in a wide window', async () => {
      const wrapper = mountPage('demo-page', [vuetifyOfWidth(1100)]);
      expect(wrapper.classes()).to.not.include('HelpPage--docs-beside');

      await openPanel(wrapper);
      expect(wrapper.classes()).to.include('HelpPage--docs-beside');

      await openPanel(wrapper);
      expect(wrapper.classes()).to.not.include('HelpPage--docs-beside');
    });

    it('lets the panel lie over the page in a narrow window', async () => {
      const wrapper = mountPage('demo-page', [vuetifyOfWidth(1099)]);

      await openPanel(wrapper);

      expect(panel()).to.not.equal(null);
      expect(wrapper.classes()).to.not.include('HelpPage--docs-beside');
    });
  });

  describe('elements of the page', () => {
    it('follows the focus: a field, then a button, then an element without a section', async () => {
      const wrapper = mountPage();
      await openPanel(wrapper);

      await focusIn(wrapper, document.querySelector('.demo-name'));
      expect(section('demo-name').getAttribute('aria-current')).to.equal('true');
      expect(marked(), 'the frame of the input').to.deep.equal([document.querySelector('.v-field')]);

      await focusIn(wrapper, byTestId('demo-delete'));
      expect(section('demo-delete').getAttribute('aria-current')).to.equal('true');
      expect(section('demo-name').getAttribute('aria-current')).to.equal(null);
      expect(marked(), 'the button itself').to.deep.equal([byTestId('demo-delete')]);

      await focusIn(wrapper, document.querySelector('.demo-other'));
      expect(document.querySelectorAll('[aria-current="true"]')).to.have.length(0);
      expect(marked()).to.deep.equal([]);
    });

    it('marks the control of a checkbox, not its root with the messages under it', async () => {
      const wrapper = mountPage();
      await openPanel(wrapper);

      await focusIn(wrapper, document.querySelector('.demo-check'));

      expect(section('demo-check').getAttribute('aria-current')).to.equal('true');
      expect(marked()).to.deep.equal([document.querySelector('.v-input__control')]);
    });

    it('does nothing while the panel is closed', async () => {
      const wrapper = mountPage();

      await focusIn(wrapper, document.querySelector('.demo-name'));

      expect(marked()).to.deep.equal([]);
    });

    it('marks the element of the section under the pointer and of the section in focus', async () => {
      const wrapper = mountPage();
      await openPanel(wrapper);

      await pointer(wrapper, section('demo-delete'), 'mouseenter');
      expect(marked()).to.deep.equal([byTestId('demo-delete')]);

      await pointer(wrapper, section('demo-delete'), 'mouseleave');
      expect(marked()).to.deep.equal([]);

      await focusIn(wrapper, section('demo-name'));
      expect(marked()).to.deep.equal([document.querySelector('.v-field')]);
    });

    it('keeps the section in focus active while the focus is in the panel', async () => {
      const wrapper = mountPage();
      await openPanel(wrapper);

      // The focus event of the panel must not be taken for a focus in the page.
      await focusIn(wrapper, section('demo-delete'));

      expect(section('demo-delete').getAttribute('aria-current')).to.equal('true');
      expect(marked()).to.deep.equal([byTestId('demo-delete')]);
    });

    it('disables the section of an element the page does not show', async () => {
      const wrapper = mountPage();
      await openPanel(wrapper);

      expect(section('demo-gone').classList.contains('HelpPanel__section--disabled')).to.equal(true);
      expect(section('demo-name').classList.contains('HelpPanel__section--disabled')).to.equal(false);

      const button = document.createElement('button');
      button.setAttribute('data-testid', 'demo-gone');
      wrapper.find('.HelpPage__content').element.appendChild(button);
      await settled(wrapper);

      expect(section('demo-gone').classList.contains('HelpPanel__section--disabled')).to.equal(false);
    });

    it('leaves no mark when the panel closes or the page goes', async () => {
      const wrapper = mountPage();
      await openPanel(wrapper);
      await focusIn(wrapper, section('demo-delete'));
      expect(marked()).to.have.length(1);

      byTestId('editDialog-docsClose').click();
      await settled(wrapper);
      expect(marked()).to.deep.equal([]);

      await openPanel(wrapper);
      await pointer(wrapper, section('demo-delete'), 'mouseenter');
      const button = byTestId('demo-delete');
      expect(button.hasAttribute(MARK)).to.equal(true);

      wrappers.pop().unmount();
      expect(button.hasAttribute(MARK)).to.equal(false);
    });

    it('closes the panel when the page gets another document', async () => {
      const wrapper = mountPage();
      await openPanel(wrapper);

      await wrapper.setProps({ helpDoc: 'other-page' });
      expect(panel()).to.equal(null);

      await openPanel(wrapper);
      expect(byTestId('editDialog-docsTitle').textContent).to.equal('Help of other-page');
    });
  });
});
