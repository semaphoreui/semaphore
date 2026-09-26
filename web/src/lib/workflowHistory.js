// Undo/redo stack of workflow model snapshots ({ nodes, edges }).
//
// Snapshots are stored serialized, so the caller may keep mutating the
// objects it passed in. A `key` groups consecutive pushes into one step:
// typing in a node's property field pushes with key `node-<id>` on every
// keystroke, and undo takes the whole edit back at once.

function serialize(snapshot) {
  return JSON.stringify(snapshot);
}

function deserialize(text) {
  return text == null ? null : JSON.parse(text);
}

export default class WorkflowHistory {
  constructor(limit = 50) {
    this.limit = limit;
    this.past = [];
    this.future = [];
    this.present = null;
    this.presentKey = null;
  }

  /** Starts over from `snapshot` (after load or save). */
  reset(snapshot) {
    this.past = [];
    this.future = [];
    this.present = snapshot == null ? null : serialize(snapshot);
    this.presentKey = null;
  }

  /**
   * Records a new state. Returns false when it equals the current one.
   * Consecutive pushes with the same non-null `key` replace the current
   * state instead of adding a step.
   */
  push(snapshot, key = null) {
    const next = serialize(snapshot);
    if (next === this.present) return false;

    if (key != null && key === this.presentKey && this.present != null) {
      this.present = next;
      return true;
    }

    if (this.present != null) {
      this.past.push(this.present);
      if (this.past.length > this.limit) this.past.shift();
    }
    this.present = next;
    this.presentKey = key;
    this.future = [];
    return true;
  }

  canUndo() {
    return this.past.length > 0;
  }

  canRedo() {
    return this.future.length > 0;
  }

  /** Steps back and returns the restored snapshot, or null. */
  undo() {
    if (!this.canUndo()) return null;
    this.future.push(this.present);
    this.present = this.past.pop();
    this.presentKey = null;
    return deserialize(this.present);
  }

  /** Steps forward and returns the restored snapshot, or null. */
  redo() {
    if (!this.canRedo()) return null;
    this.past.push(this.present);
    this.present = this.future.pop();
    this.presentKey = null;
    return deserialize(this.present);
  }

  current() {
    return deserialize(this.present);
  }
}
