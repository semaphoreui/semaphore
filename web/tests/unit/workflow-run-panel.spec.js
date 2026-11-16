import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import WorkflowRunPanel from '@/components/WorkflowRunPanel.vue';

const WIDTH_KEY = 'WorkflowRunPanel__width';

function pressEscape() {
  window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', keyCode: 27, bubbles: true }));
}

describe('WorkflowRunPanel.vue', () => {
  const wrappers = [];

  afterEach(() => {
    while (wrappers.length) {
      wrappers.pop().unmount();
    }
    document.querySelectorAll('[data-app]').forEach((el) => el.remove());
    localStorage.clear();
  });

  function mountPanel(props = {}) {
    const app = document.createElement('div');
    app.setAttribute('data-app', 'true');
    document.body.appendChild(app);
    const mountPoint = document.createElement('div');
    app.appendChild(mountPoint);

    const wrapper = mount(WorkflowRunPanel, {
      attachTo: mountPoint,
      props: {
        modelValue: true, projectId: 7, workflowId: 3, runId: 9, ...props,
      },
      global: {
        stubs: {
          // The view of the run has its own spec; here it only shows which run it was given.
          WorkflowRunView: {
            props: {
              projectId: Number, workflowId: Number, runId: Number, panel: Boolean,
            },
            emits: ['close'],
            template: '<div data-testid="view" :data-run="workflowId + \'/\' + runId" :data-panel="panel" @click="$emit(\'close\')" />',
          },
          transition: false,
        },
      },
    });
    wrappers.push(wrapper);
    return wrapper;
  }

  function panelOf(wrapper) {
    return wrapper.find('[data-testid="workflowRunPanel"]');
  }

  // jsdom lays nothing out: the panel is given the width a browser would measure, and its handle
  // captures no pointer.
  function handleOf(wrapper, measuredWidth) {
    const handle = wrapper.find('[data-testid="workflowRunPanel-resize"]').element;
    handle.parentElement.getBoundingClientRect = () => ({ width: measuredWidth });
    handle.setPointerCapture = () => {};
    return handle;
  }

  async function drag(wrapper, handle, from, to) {
    handle.dispatchEvent(new MouseEvent('pointerdown', { clientX: from, bubbles: true }));
    handle.dispatchEvent(new MouseEvent('pointermove', { clientX: to, bubbles: true }));
    await wrapper.vm.$nextTick();
  }

  it('is not there until it is opened on a run', async () => {
    const closed = mountPanel({ modelValue: false });
    expect(panelOf(closed).exists()).to.equal(false);

    const noRun = mountPanel({ runId: null });
    expect(panelOf(noRun).exists()).to.equal(false);

    const noWorkflow = mountPanel({ workflowId: null });
    expect(panelOf(noWorkflow).exists()).to.equal(false);

    await closed.setProps({ modelValue: true });
    expect(panelOf(closed).exists()).to.equal(true);
  });

  it('shows the run it was opened on, as a panel', async () => {
    const wrapper = mountPanel();

    const view = wrapper.find('[data-testid="view"]');
    expect(view.attributes('data-run')).to.equal('3/9');
    expect(view.attributes('data-panel')).to.equal('true');

    await wrapper.setProps({ workflowId: 4, runId: 10 });
    expect(wrapper.find('[data-testid="view"]').attributes('data-run')).to.equal('4/10');
  });

  it('starts the view anew in another project, even for the same ids', async () => {
    const wrapper = mountPanel();
    const before = wrapper.find('[data-testid="view"]').element;

    await wrapper.setProps({ projectId: 8 });

    const after = wrapper.find('[data-testid="view"]').element;
    expect(after.getAttribute('data-run')).to.equal('3/9');
    expect(after).to.not.equal(before);
  });

  it('takes half of the window by default', () => {
    const wrapper = mountPanel();

    expect(panelOf(wrapper).element.style.width).to.equal('50%');
  });

  describe('resizing', () => {
    let innerWidth;

    beforeEach(() => {
      innerWidth = window.innerWidth;
      window.innerWidth = 1600;
    });

    afterEach(() => {
      window.innerWidth = innerWidth;
    });

    it('grows by what its left edge is dragged to the left, and shrinks back', async () => {
      const wrapper = mountPanel();
      const handle = handleOf(wrapper, 800);

      await drag(wrapper, handle, 800, 600);
      expect(panelOf(wrapper).element.style.width).to.equal('1000px');

      handle.dispatchEvent(new MouseEvent('pointermove', { clientX: 1000, bubbles: true }));
      await wrapper.vm.$nextTick();
      expect(panelOf(wrapper).element.style.width).to.equal('600px');
    });

    it('stays between the narrowest width and the width of the window', async () => {
      const wrapper = mountPanel();
      const handle = handleOf(wrapper, 800);

      await drag(wrapper, handle, 800, -5000);
      expect(panelOf(wrapper).element.style.width).to.equal('1600px');

      handle.dispatchEvent(new MouseEvent('pointermove', { clientX: 1590, bubbles: true }));
      await wrapper.vm.$nextTick();
      expect(panelOf(wrapper).element.style.width).to.equal('400px');
    });

    it('ignores the pointer when no drag is in progress', async () => {
      const wrapper = mountPanel();
      const handle = handleOf(wrapper, 800);

      await drag(wrapper, handle, 800, 600);
      handle.dispatchEvent(new MouseEvent('pointerup', { clientX: 600, bubbles: true }));
      handle.dispatchEvent(new MouseEvent('pointermove', { clientX: 100, bubbles: true }));
      await wrapper.vm.$nextTick();

      expect(panelOf(wrapper).element.style.width).to.equal('1000px');
    });

    it('remembers the width it was dragged to', async () => {
      const wrapper = mountPanel();
      const handle = handleOf(wrapper, 800);

      await drag(wrapper, handle, 800, 600);
      expect(localStorage.getItem(WIDTH_KEY)).to.equal(null);
      handle.dispatchEvent(new MouseEvent('pointerup', { clientX: 600, bubbles: true }));

      expect(localStorage.getItem(WIDTH_KEY)).to.equal('1000');
      expect(panelOf(mountPanel()).element.style.width).to.equal('1000px');
    });

    it('goes back to half of the window on a double click on its edge', async () => {
      localStorage.setItem(WIDTH_KEY, '1000');
      const wrapper = mountPanel();
      expect(panelOf(wrapper).element.style.width).to.equal('1000px');

      await wrapper.find('[data-testid="workflowRunPanel-resize"]').trigger('dblclick');

      expect(panelOf(wrapper).element.style.width).to.equal('50%');
      expect(localStorage.getItem(WIDTH_KEY)).to.equal(null);
    });

    it('does not trust a stored width below the narrowest one', () => {
      localStorage.setItem(WIDTH_KEY, '120');

      expect(panelOf(mountPanel()).element.style.width).to.equal('50%');
    });
  });

  describe('closing', () => {
    it('asks its owner to close it when the view asks', async () => {
      const wrapper = mountPanel();

      await wrapper.find('[data-testid="view"]').trigger('click');

      expect(wrapper.emitted('close')).to.have.length(1);
    });

    it('asks to be closed on Escape', () => {
      const wrapper = mountPanel();

      pressEscape();

      expect(wrapper.emitted('close')).to.have.length(1);
    });

    it('leaves Escape to a dialog that lies over it', () => {
      const wrapper = mountPanel();
      const dialog = document.createElement('div');
      dialog.className = 'v-overlay v-overlay--active v-dialog';
      document.querySelector('[data-app]').appendChild(dialog);

      pressEscape();

      expect(wrapper.emitted('close')).to.equal(undefined);
    });

    it('does not listen to Escape while it is closed', async () => {
      const wrapper = mountPanel();
      await wrapper.setProps({ modelValue: false });

      pressEscape();

      expect(wrapper.emitted('close')).to.equal(undefined);
    });
  });
});
