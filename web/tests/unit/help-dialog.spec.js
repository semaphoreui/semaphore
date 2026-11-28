import { expect } from 'chai';
import { mount, flushPromises } from '@vue/test-utils';
import HelpDialog from '@/components/HelpDialog.vue';
import HelpButton from '@/components/HelpButton.vue';
// Vitest has run the setup ahead of the spec already (`test.setupFiles`).
import createTestVuetify from './setup';

vi.mock('@/help', () => ({
  hasHelpDoc: (key) => key === 'demo-dialog',
  helpDocLocale: (key) => (key === 'demo-dialog' ? 'en' : null),
  loadHelpDoc: (key) => Promise.resolve({
    key,
    locale: 'en',
    title: 'New Demo',
    intro: [{ type: 'p', inline: [{ type: 'text', text: 'What the dialog is for.' }] }],
    sections: ['demo-name', 'demo-gone'].map((field) => ({
      field,
      title: field,
      blocks: [{ type: 'p', inline: [{ type: 'text', text: `About ${field}.` }] }],
    })),
  }),
}));

// A plain dialog as Tokens.vue has one: a card with a title, a field, a button.
const Page = {
  components: { HelpDialog, HelpButton },
  props: { helpDoc: String, maxWidth: { type: [Number, String], default: 480 } },
  data() {
    return { dialog: true };
  },
  template: `
    <HelpDialog v-model="dialog" :help-doc="helpDoc" :max-width="maxWidth" persistent v-slot="{ docs }">
      <v-card data-testid="demoDialog">
        <v-card-title>
          New Demo
          <HelpButton v-if="docs.available" :open="docs.open" data-testid="demoDialog-docs" @click="docs.toggle()" />
        </v-card-title>
        <v-card-text>
          <div data-testid="demo-name"><div class="v-field"><input class="demo-name"></div></div>
        </v-card-text>
      </v-card>
    </HelpDialog>`,
};

describe('HelpDialog.vue', () => {
  const wrappers = [];
  const byTestId = (id) => document.querySelector(`[data-testid="${id}"]`);
  const panel = () => byTestId('editDialog-docsPanel');
  const content = () => byTestId('demoDialog').parentElement;

  async function mountPage(props = {}, plugins = []) {
    const app = document.createElement('div');
    app.setAttribute('data-app', 'true');
    document.body.appendChild(app);
    const wrapper = mount(Page, {
      attachTo: app,
      props: { helpDoc: 'demo-dialog', ...props },
      global: { mocks: { $t: (k) => k }, plugins },
    });
    wrappers.push(wrapper);
    await flushPromises();
    await wrapper.vm.$nextTick();
    return wrapper;
  }

  async function settled(wrapper) {
    await flushPromises();
    await wrapper.vm.$nextTick();
  }

  async function openPanel(wrapper) {
    byTestId('demoDialog-docs').click();
    await settled(wrapper);
  }

  afterEach(() => {
    while (wrappers.length) {
      wrappers.pop().unmount();
    }
    document.body.innerHTML = '';
  });

  it('is a dialog with its card as the direct child of the content element', async () => {
    await mountPage();

    expect(content().classList.contains('v-overlay__content')).to.equal(true);
    expect(content().style.maxWidth).to.equal('480px');
  });

  it('offers no button without a document', async () => {
    await mountPage({ helpDoc: 'nope' });

    expect(byTestId('demoDialog')).to.not.equal(null);
    expect(byTestId('demoDialog-docs')).to.equal(null);
  });

  it('opens the panel next to the card, inside the content element', async () => {
    const wrapper = await mountPage();
    await openPanel(wrapper);

    expect(panel().parentElement).to.equal(content());
    expect(content().classList.contains('item-dialog--docs')).to.equal(true);
    expect(content().classList.contains('item-dialog--docs-beside')).to.equal(true);
  });

  it('narrows a wide dialog to what the panel leaves of the window', async () => {
    const wrapper = await mountPage({ maxWidth: 1000 });
    await openPanel(wrapper);

    expect(content().style.maxWidth).to.equal(`${window.innerWidth - 468}px`);
  });

  it('lets the panel lie over the dialog in a narrow window', async () => {
    const innerWidth = window.innerWidth;
    window.innerWidth = 860;
    const vuetify = createTestVuetify();
    window.innerWidth = innerWidth;

    const wrapper = await mountPage({}, [vuetify]);
    await openPanel(wrapper);

    expect(panel()).to.not.equal(null);
    expect(content().classList.contains('item-dialog--docs-beside')).to.equal(false);
    expect(content().style.maxWidth).to.equal('480px');
  });

  it('follows the focus in the card and marks the field', async () => {
    const wrapper = await mountPage();
    await openPanel(wrapper);

    document.querySelector('.demo-name').dispatchEvent(new FocusEvent('focusin', { bubbles: true }));
    await settled(wrapper);

    expect(byTestId('editDialog-docsSection-demo-name').getAttribute('aria-current')).to.equal('true');
    expect(document.querySelector('[data-help-docs-field]')).to.equal(document.querySelector('.v-field'));
  });

  it('keeps a section in focus active: the focus in the panel is not taken for the card', async () => {
    const wrapper = await mountPage();
    await openPanel(wrapper);

    byTestId('editDialog-docsSection-demo-name').dispatchEvent(new FocusEvent('focusin', { bubbles: true }));
    await settled(wrapper);

    expect(byTestId('editDialog-docsSection-demo-name').getAttribute('aria-current')).to.equal('true');
  });

  it('disables the section of a field the card does not show', async () => {
    const wrapper = await mountPage();
    await openPanel(wrapper);

    expect(byTestId('editDialog-docsSection-demo-gone').classList
      .contains('HelpPanel__section--disabled')).to.equal(true);
  });

  it('closes the panel when the dialog closes', async () => {
    const wrapper = await mountPage();
    await openPanel(wrapper);

    wrapper.vm.dialog = false;
    await settled(wrapper);
    wrapper.vm.dialog = true;
    await settled(wrapper);

    expect(panel()).to.equal(null);
    expect(document.querySelector('[data-help-docs-field]')).to.equal(null);
  });
});
