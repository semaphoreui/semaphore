import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import WorkflowNodeCard from '@/components/workflow/WorkflowNodeCard.vue';

const XSS_NAME = '<img src=x onerror=alert(1)>';

function makeStore(overrides = {}) {
  return {
    nodes: {},
    templates: [
      { id: 7, name: 'Deploy application', app: 'ansible' },
      { id: 8, name: XSS_NAME, app: 'terraform' },
    ],
    runs: {},
    problems: {},
    apps: null,
    canResolveApprovals: false,
    ...overrides,
  };
}

function mountCard(node, storeOverrides = {}, editable = true) {
  const localVue = createLocalVue();
  const store = makeStore(storeOverrides);
  store.nodes[node.id] = node;
  return mount(WorkflowNodeCard, {
    localVue,
    vuetify: new Vuetify(),
    propsData: { store, nodeId: node.id, editable },
    mocks: { $t: (k, params) => (params ? `${k}:${JSON.stringify(params)}` : k) },
    stubs: { 'v-tooltip': { template: '<div><slot name="activator" :on="{}" :attrs="{}" /><slot /></div>' } },
  });
}

describe('WorkflowNodeCard.vue', () => {
  describe('rendering by kind', () => {
    it('shows the template name and app for a task node', () => {
      const w = mountCard({ id: 1, kind: 'task', template_id: 7 });
      expect(w.find('.WorkflowNodeCard__title').text()).to.equal('Deploy application');
      expect(w.find('.WorkflowNodeCard__sub').text()).to.equal('Ansible');
      expect(w.find('.WorkflowNodeCard__tile--task').exists()).to.equal(true);
    });

    it('marks a task without a template as incomplete', () => {
      const w = mountCard({ id: 1, kind: 'task', template_id: null });
      const title = w.find('.WorkflowNodeCard__title');
      expect(title.text()).to.equal('workflowNodeIncomplete');
      expect(title.classes()).to.include('WorkflowNodeCard__title--placeholder');
    });

    it('renders an approval node with its timeout', () => {
      const w = mountCard({ id: 2, kind: 'approval', approval_timeout: 3600 });
      expect(w.find('.WorkflowNodeCard__title').text()).to.equal('workflowNodeKindApproval');
      expect(w.find('.WorkflowNodeCard__sub').text()).to.contain('workflowApprovalTimeoutShort');
      expect(w.find('.WorkflowNodeCard__sub').text()).to.contain('1h 0m');
      expect(w.find('.WorkflowNodeCard__tile--approval').exists()).to.equal(true);
    });

    it('renders a delay node with its duration', () => {
      const w = mountCard({ id: 3, kind: 'delay', delay_seconds: 90 });
      expect(w.find('.WorkflowNodeCard__title').text()).to.equal('workflowNodeKindDelay');
      expect(w.find('.WorkflowNodeCard__sub').text()).to.contain('1m 30s');
    });

    it('renders a note as a sticky without ports or plus handle', () => {
      const w = mountCard({ id: 4, kind: 'note', note: 'Ask ops first' });
      expect(w.classes()).to.include('WorkflowNodeCard--note');
      expect(w.find('.WorkflowNodeCard__noteText').text()).to.equal('Ask ops first');
      expect(w.find('.WorkflowNodeCard__plus').exists()).to.equal(false);
    });

    it('mentions "any parent" convergence', () => {
      const w = mountCard({
        id: 1, kind: 'task', template_id: 7, convergence_mode: 'any',
      });
      expect(w.find('.WorkflowNodeCard__sub').text()).to.equal('Ansible · workflowConvergenceAny');
    });
  });

  describe('escaping', () => {
    it('renders a hostile template name as text', () => {
      const w = mountCard({ id: 1, kind: 'task', template_id: 8 });
      expect(w.find('.WorkflowNodeCard__title').text()).to.equal(XSS_NAME);
      expect(w.find('img').exists()).to.equal(false);
    });

    it('renders a hostile note as text', () => {
      const w = mountCard({ id: 4, kind: 'note', note: '<script>alert(1)</script>' });
      expect(w.find('.WorkflowNodeCard__noteText').text()).to.equal('<script>alert(1)</script>');
      expect(w.find('script').exists()).to.equal(false);
    });
  });

  describe('editor state', () => {
    it('shows a problem badge with the message', () => {
      const w = mountCard({ id: 1, kind: 'task', template_id: null }, {
        problems: { 1: 'Every task node requires a template.' },
      });
      expect(w.find('.WorkflowNodeCard__badge').exists()).to.equal(true);
      expect(w.classes()).to.include('WorkflowNodeCard--problem');
      expect(w.text()).to.contain('Every task node requires a template.');
    });

    it('emits quick-add from the plus handle without starting a drag', async () => {
      const w = mountCard({ id: 1, kind: 'task', template_id: 7 });
      let mousedownReachedParent = false;
      w.element.addEventListener('mousedown', () => { mousedownReachedParent = true; });
      const plus = w.find('.WorkflowNodeCard__plus');
      await plus.trigger('mousedown');
      await plus.trigger('click');
      expect(mousedownReachedParent).to.equal(false);
      expect(w.emitted('quick-add')).to.deep.equal([[1]]);
    });

    it('hides the plus handle in read-only mode', () => {
      const w = mountCard({ id: 1, kind: 'task', template_id: 7 }, {}, false);
      expect(w.find('.WorkflowNodeCard__plus').exists()).to.equal(false);
    });
  });

  describe('run view', () => {
    it('shows a success icon and the duration', () => {
      const w = mountCard({ id: 1, kind: 'task', template_id: 7 }, {
        runs: {
          1: {
            status: 'success',
            start: '2026-09-26T10:00:00Z',
            end: '2026-09-26T10:02:14Z',
            taskId: 42,
          },
        },
      }, false);
      expect(w.find('.WorkflowNodeCard__sub').text()).to.equal('success · 2m 14s');
      expect(w.find('.WorkflowNodeCard__status .mdi-check-circle').exists()).to.equal(true);
      expect(w.classes()).to.include('WorkflowNodeCard--clickable');
    });

    it('dims a node that has not started while others have', () => {
      const w = mountCard({ id: 2, kind: 'task', template_id: 7 }, {
        runs: { 1: { status: 'success' } },
      }, false);
      expect(w.classes()).to.include('WorkflowNodeCard--dim');
      expect(w.find('.WorkflowNodeCard__sub').text()).to.equal('workflowNodeNotStarted');
    });

    it('pulses a running node with a spinner', () => {
      const w = mountCard({ id: 1, kind: 'task', template_id: 7 }, {
        runs: { 1: { status: 'running', start: new Date().toISOString() } },
      }, false);
      expect(w.classes()).to.include('WorkflowNodeCard--running');
      expect(w.find('.WorkflowNodeCard__status .v-progress-circular').exists()).to.equal(true);
      w.destroy();
    });

    it('counts down a waiting delay', () => {
      const resumeAt = new Date(Date.now() + 75 * 1000).toISOString();
      const w = mountCard({ id: 3, kind: 'delay', delay_seconds: 120 }, {
        runs: { 3: { status: 'waiting', resumeAt } },
      }, false);
      expect(w.find('.WorkflowNodeCard__sub').text()).to.contain('workflowDelayRemaining');
      expect(w.find('.WorkflowNodeCard__sub').text()).to.contain('1m 15s');
      expect(w.find('.WorkflowNodeCard__status .mdi-timer-sand').exists()).to.equal(true);
      w.destroy();
    });

    it('offers Approve / Reject on a pending approval when allowed', async () => {
      const w = mountCard({ id: 2, kind: 'approval', approval_message: 'Deploy to prod?' }, {
        runs: { 2: { status: 'pending', message: 'Deploy to prod?' } },
        canResolveApprovals: true,
      }, false);
      expect(w.classes()).to.include('WorkflowNodeCard--pending');
      expect(w.find('.WorkflowNodeCard__sub').text()).to.equal('Deploy to prod?');
      const buttons = w.findAll('.WorkflowNodeCard__actions button');
      expect(buttons.length).to.equal(2);
      await buttons.at(0).trigger('click');
      await buttons.at(1).trigger('click');
      expect(w.emitted('resolve-approval')).to.deep.equal([['approved'], ['rejected']]);
    });

    it('hides the approval actions without permission', () => {
      const w = mountCard({ id: 2, kind: 'approval' }, {
        runs: { 2: { status: 'pending' } },
        canResolveApprovals: false,
      }, false);
      expect(w.find('.WorkflowNodeCard__actions').exists()).to.equal(false);
    });

    it('emits select on a tap but not on a drag in read-only mode', async () => {
      const w = mountCard({ id: 1, kind: 'task', template_id: 7 }, {
        runs: { 1: { status: 'success', taskId: 42 } },
      }, false);
      await w.trigger('pointerdown', { clientX: 10, clientY: 10 });
      await w.trigger('pointerup', { clientX: 12, clientY: 11 });
      await w.trigger('pointerdown', { clientX: 10, clientY: 10 });
      await w.trigger('pointerup', { clientX: 40, clientY: 10 });
      expect(w.emitted('select')).to.deep.equal([[1]]);
    });
  });
});
