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
