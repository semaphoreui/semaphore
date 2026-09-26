import { expect } from 'chai';
import WorkflowHistory from '@/lib/workflowHistory';

const snapshot = (n) => ({ nodes: [{ id: 1, position_x: n }], edges: [] });

describe('lib/workflowHistory', () => {
  it('starts empty', () => {
    const h = new WorkflowHistory();
    expect(h.canUndo()).to.equal(false);
    expect(h.canRedo()).to.equal(false);
    expect(h.undo()).to.equal(null);
    expect(h.redo()).to.equal(null);
  });

  it('undo and redo round-trip over 20 steps', () => {
    const h = new WorkflowHistory();
    h.reset(snapshot(0));
    for (let i = 1; i <= 20; i += 1) h.push(snapshot(i));

    for (let i = 19; i >= 0; i -= 1) {
      expect(h.undo()).to.deep.equal(snapshot(i));
    }
    expect(h.canUndo()).to.equal(false);

    for (let i = 1; i <= 20; i += 1) {
      expect(h.redo()).to.deep.equal(snapshot(i));
    }
    expect(h.canRedo()).to.equal(false);
    expect(h.current()).to.deep.equal(snapshot(20));
  });

  it('ignores a push equal to the current state', () => {
    const h = new WorkflowHistory();
    h.reset(snapshot(0));
    expect(h.push(snapshot(0))).to.equal(false);
    expect(h.canUndo()).to.equal(false);
  });

  it('drops the redo stack on a new push', () => {
    const h = new WorkflowHistory();
    h.reset(snapshot(0));
    h.push(snapshot(1));
    h.undo();
    h.push(snapshot(2));
    expect(h.canRedo()).to.equal(false);
    expect(h.undo()).to.deep.equal(snapshot(0));
  });

  it('coalesces consecutive pushes with the same key into one step', () => {
    const h = new WorkflowHistory();
    h.reset(snapshot(0));
    h.push(snapshot(1), 'node-1');
    h.push(snapshot(2), 'node-1');
    h.push(snapshot(3), 'node-1');
    expect(h.undo()).to.deep.equal(snapshot(0));
    expect(h.redo()).to.deep.equal(snapshot(3));
  });

  it('starts a new step when the key changes', () => {
    const h = new WorkflowHistory();
    h.reset(snapshot(0));
    h.push(snapshot(1), 'node-1');
    h.push(snapshot(2), 'node-2');
    h.push(snapshot(3));
    h.push(snapshot(4), 'node-2');
    expect(h.undo()).to.deep.equal(snapshot(3));
    expect(h.undo()).to.deep.equal(snapshot(2));
    expect(h.undo()).to.deep.equal(snapshot(1));
  });

  it('keeps at most `limit` past states', () => {
    const h = new WorkflowHistory(3);
    h.reset(snapshot(0));
    for (let i = 1; i <= 10; i += 1) h.push(snapshot(i));
    expect(h.undo()).to.deep.equal(snapshot(9));
    expect(h.undo()).to.deep.equal(snapshot(8));
    expect(h.undo()).to.deep.equal(snapshot(7));
    expect(h.canUndo()).to.equal(false);
  });

  it('does not share state with the caller', () => {
    const h = new WorkflowHistory();
    const s = snapshot(0);
    h.reset(s);
    s.nodes[0].position_x = 99;
    expect(h.current()).to.deep.equal(snapshot(0));
  });
});
