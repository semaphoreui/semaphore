export default class Listenable {
  constructor() {
    this.listeners = {};
  }

  addListener(callback) {
    // eslint-disable-next-line symbol-description
    const id = Symbol();
    this.listeners[id] = callback;
    return id;
  }

  removeListener(id) {
    if (this.listeners[id] == null) {
      return false;
    }
    delete this.listeners[id];
    return true;
  }

  callListeners(data) {
    Object.getOwnPropertySymbols(this.listeners).forEach((id) => {
      const listener = this.listeners[id];
      listener(data);
    });
  }

  hasListeners() {
    return Object.keys(this.listeners).length > 0;
  }
}
