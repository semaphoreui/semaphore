import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import HelpPanel from '@/components/HelpPanel.vue';
import HelpContent from '@/components/HelpContent';

const text = (value) => ({ type: 'text', text: value });

const DOC = {
  key: 'demo',
  title: 'Edit Demo',
  intro: [{ type: 'p', inline: [text('What the dialog is for.')] }],
  sections: [{
    field: 'demoForm-name',
    title: 'Name',
    blocks: [{
      type: 'p',
      inline: [
        text('A '),
        { type: 'code', text: 'code' },
        text(', a '),
        { type: 'strong', text: 'bold' },
        text(' and a '),
        { type: 'link', text: 'link', href: 'https://example.com/a' },
        text('.'),
      ],
    }],
  }, {
    field: 'demoForm-type',
    title: 'Type',
    blocks: [
      { type: 'ul', items: [[text('one')], [text('two '), { type: 'code', text: 'x' }]] },
      { type: 'code', text: '[web]\n10.0.0.1' },
    ],
  }],
};

function mountPanel(doc = DOC) {
  return mount(HelpPanel, {
    props: { doc },
    global: { mocks: { $t: (k) => k } },
  });
}

describe('HelpPanel.vue', () => {
  it('shows the title, the intro and the sections in the order of the document', () => {
    const wrapper = mountPanel();

    expect(wrapper.attributes('data-testid')).to.equal('editDialog-docsPanel');
    expect(wrapper.find('[data-testid="editDialog-docsTitle"]').text()).to.equal('Edit Demo');
    expect(wrapper.find('[data-testid="editDialog-docsIntro"]').text())
      .to.equal('What the dialog is for.');

    const sections = wrapper.findAll('section');
    expect(sections.map((s) => s.attributes('data-testid'))).to.deep.equal([
      'editDialog-docsSection-demoForm-name',
      'editDialog-docsSection-demoForm-type',
    ]);
    expect(sections.map((s) => s.find('h3').text())).to.deep.equal(['Name', 'Type']);
  });

  it('labels the close button and emits close', async () => {
    const wrapper = mountPanel();
    const button = wrapper.find('[data-testid="editDialog-docsClose"]');

    expect(button.attributes('aria-label')).to.equal('helpPanelClose');
    await button.trigger('click');

    expect(wrapper.emitted('close')).to.have.length(1);
  });
});

describe('HelpPanel.vue sections', () => {
  function mountWith(props = {}) {
    const el = document.createElement('div');
    document.body.appendChild(el);
    return mount(HelpPanel, {
      attachTo: el,
      props: { doc: DOC, ...props },
      global: { mocks: { $t: (k) => k } },
    });
  }

  const sectionOf = (wrapper, field) => wrapper
    .find(`[data-testid="editDialog-docsSection-${field}"]`);

  it('makes every section focusable', () => {
    const wrapper = mountWith();

    expect(wrapper.findAll('section').map((s) => s.attributes('tabindex'))).to.deep.equal(['0', '0']);
    wrapper.unmount();
  });

  it('marks the section of the active field', async () => {
    const wrapper = mountWith({ activeField: 'demoForm-type' });

    expect(sectionOf(wrapper, 'demoForm-type').attributes('aria-current')).to.equal('true');
    expect(sectionOf(wrapper, 'demoForm-type').classes()).to.include('HelpPanel__section--active');
    expect(sectionOf(wrapper, 'demoForm-name').attributes('aria-current')).to.equal(undefined);

    await wrapper.setProps({ activeField: 'no-such-field' });

    expect(wrapper.findAll('.HelpPanel__section--active')).to.have.length(0);
    wrapper.unmount();
  });

  it('reports the section under the pointer', async () => {
    const wrapper = mountWith();

    await sectionOf(wrapper, 'demoForm-name').trigger('mouseenter');
    await sectionOf(wrapper, 'demoForm-name').trigger('mouseleave');

    expect(wrapper.emitted('hover-field')).to.deep.equal([['demoForm-name'], [null]]);
    wrapper.unmount();
  });

  it('reports the section in focus, also for a link inside it, and null elsewhere', async () => {
    const wrapper = mountWith();

    await sectionOf(wrapper, 'demoForm-type').trigger('focusin');
    await sectionOf(wrapper, 'demoForm-name').find('a').trigger('focusin');
    await wrapper.find('[data-testid="editDialog-docsClose"]').trigger('focusin');

    expect(wrapper.emitted('focus-field')).to.deep.equal([['demoForm-type'], ['demoForm-name'], [null]]);
    wrapper.unmount();
  });

  it('says which fields are not shown', () => {
    const wrapper = mountWith({ missingFields: ['demoForm-type'] });

    const note = wrapper.find('[data-testid="editDialog-docsSectionHidden-demoForm-type"]');
    expect(note.text()).to.equal('helpPanelFieldHidden');
    expect(wrapper.find('[data-testid="editDialog-docsSectionHidden-demoForm-name"]').exists())
      .to.equal(false);
    wrapper.unmount();
  });

  describe('section of a field that is not shown', () => {
    it('is disabled and out of the Tab order', () => {
      const wrapper = mountWith({ missingFields: ['demoForm-type'] });
      const disabled = sectionOf(wrapper, 'demoForm-type');
      const enabled = sectionOf(wrapper, 'demoForm-name');

      expect(disabled.classes()).to.include('HelpPanel__section--disabled');
      expect(disabled.attributes('aria-disabled')).to.equal('true');
      expect(disabled.attributes('tabindex')).to.equal(undefined);
      expect(enabled.classes()).to.not.include('HelpPanel__section--disabled');
      expect(enabled.attributes('aria-disabled')).to.equal(undefined);
      expect(enabled.attributes('tabindex')).to.equal('0');
      wrapper.unmount();
    });

    it('is not the active one, even for the active field', async () => {
      const wrapper = mountWith({ activeField: 'demoForm-type' });
      expect(sectionOf(wrapper, 'demoForm-type').classes()).to.include('HelpPanel__section--active');

      await wrapper.setProps({ missingFields: ['demoForm-type'] });

      expect(sectionOf(wrapper, 'demoForm-type').classes())
        .to.not.include('HelpPanel__section--active');
      expect(sectionOf(wrapper, 'demoForm-type').attributes('aria-current')).to.equal(undefined);

      await wrapper.setProps({ missingFields: [] });

      expect(sectionOf(wrapper, 'demoForm-type').classes()).to.include('HelpPanel__section--active');
      wrapper.unmount();
    });

    it('reports neither the pointer over it nor the focus inside it', async () => {
      const wrapper = mountWith({ missingFields: ['demoForm-name'] });

      await sectionOf(wrapper, 'demoForm-name').trigger('mouseenter');
      await sectionOf(wrapper, 'demoForm-name').trigger('focusin');
      await sectionOf(wrapper, 'demoForm-name').find('a').trigger('focusin');

      expect(wrapper.emitted('hover-field')).to.deep.equal([[null]]);
      expect(wrapper.emitted('focus-field')).to.deep.equal([[null], [null]]);
      wrapper.unmount();
    });
  });

  describe('followField', () => {
    // jsdom lays nothing out: a view of 100 px, the sections 300 px each, one under another.
    function layOut(wrapper, reducedMotion = false) {
      const body = wrapper.find('[data-testid="editDialog-docsBody"]').element;
      const calls = [];
      body.getBoundingClientRect = () => ({ top: 20, bottom: 120 });
      body.scrollTo = (to) => calls.push(to);
      wrapper.findAll('section').forEach((s, i) => {
        // eslint-disable-next-line no-param-reassign
        s.element.getBoundingClientRect = () => ({ top: 20 + 300 * i, bottom: 320 + 300 * i });
      });
      window.matchMedia = () => ({ matches: reducedMotion });
      return calls;
    }

    const matchMedia = window.matchMedia;

    afterEach(() => {
      window.matchMedia = matchMedia;
    });

    it('scrolls the body to the section of the field', async () => {
      const wrapper = mountWith();
      const calls = layOut(wrapper);

      await wrapper.setProps({ followField: 'demoForm-type' });

      expect(calls).to.deep.equal([{ top: 300, behavior: 'smooth' }]);
      wrapper.unmount();
    });

    it('does not animate for a user who asked for reduced motion', async () => {
      const wrapper = mountWith();
      const calls = layOut(wrapper, true);

      await wrapper.setProps({ followField: 'demoForm-type' });

      expect(calls).to.deep.equal([{ top: 300, behavior: 'auto' }]);
      wrapper.unmount();
    });

    it('does not scroll to a section that is in view, nor when the field is gone', async () => {
      const wrapper = mountWith();
      const calls = layOut(wrapper);
      wrapper.find('[data-testid="editDialog-docsBody"]').element
        .getBoundingClientRect = () => ({ top: 0, bottom: 5000 });

      await wrapper.setProps({ followField: 'demoForm-type' });
      await wrapper.setProps({ followField: null });

      expect(calls).to.deep.equal([]);
      wrapper.unmount();
    });

    it('ignores a field without a section', async () => {
      const wrapper = mountWith();
      const calls = layOut(wrapper);

      await wrapper.setProps({ followField: 'no-such-field' });

      expect(calls).to.deep.equal([]);
      wrapper.unmount();
    });
  });
});

describe('HelpContent', () => {
  function mountContent(blocks) {
    return mount(HelpContent, { props: { blocks } });
  }

  it('draws paragraphs, lists and code', () => {
    const wrapper = mountContent([...DOC.sections[0].blocks, ...DOC.sections[1].blocks]);

    expect(wrapper.find('p').text()).to.equal('A code, a bold and a link.');
    expect(wrapper.find('p code').text()).to.equal('code');
    expect(wrapper.find('p strong').text()).to.equal('bold');
    expect(wrapper.findAll('ul li').map((li) => li.text())).to.deep.equal(['one', 'two x']);
    expect(wrapper.find('pre code').element.textContent).to.equal('[web]\n10.0.0.1');
  });

  it('opens a link in a new tab without a reference to the opener', () => {
    const link = mountContent(DOC.sections[0].blocks).find('a');

    expect(link.attributes('href')).to.equal('https://example.com/a');
    expect(link.attributes('target')).to.equal('_blank');
    expect(link.attributes('rel')).to.equal('noopener noreferrer');
  });

  it('shows markup in the text of a document as text', () => {
    const markup = '<b onclick="alert(1)">bold</b><script>alert(2)</script>';
    const wrapper = mountContent([
      { type: 'p', inline: [text(markup), { type: 'code', text: markup }, { type: 'strong', text: markup }] },
      { type: 'ul', items: [[text(markup)]] },
      { type: 'code', text: markup },
    ]);

    expect(wrapper.find('b').exists()).to.equal(false);
    expect(wrapper.find('script').exists()).to.equal(false);
    expect(wrapper.find('p').element.textContent).to.equal(markup.repeat(3));
    expect(wrapper.find('li').element.textContent).to.equal(markup);
    expect(wrapper.find('pre').element.textContent).to.equal(markup);
  });

  const unsafeLinks = [
    // eslint-disable-next-line no-script-url
    'javascript:alert(1)',
    'http://example.com',
    '//example.com',
    '/project/1',
    'data:text/html,x',
    '',
    null,
  ];

  unsafeLinks.forEach((href) => {
    it(`shows a link to ${JSON.stringify(href)} as its text`, () => {
      const wrapper = mountContent([{ type: 'p', inline: [{ type: 'link', text: 'label', href }] }]);

      expect(wrapper.find('a').exists()).to.equal(false);
      expect(wrapper.find('p').text()).to.equal('label');
    });
  });
});
