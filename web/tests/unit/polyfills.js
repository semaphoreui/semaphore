// jsdom polyfills of the unit tests, imported by setup.js before Vue, Vuetify and the app code.

// src/plugins/i18.js reads localStorage at import time.
function defineGlobal(name, value) {
  [global, typeof window !== 'undefined' ? window : null].forEach((target) => {
    if (!target) {
      return;
    }
    // jsdom exposes some properties (e.g. localStorage) as getters that throw
    // for opaque origins, so always (re)define them as plain values.
    Object.defineProperty(target, name, { value, configurable: true, writable: true });
  });
}

function hasGlobal(name) {
  try {
    return typeof global[name] !== 'undefined';
  } catch (e) {
    return false;
  }
}

if (!hasGlobal('localStorage')) {
  const store = new Map();
  defineGlobal('localStorage', {
    getItem: (k) => (store.has(k) ? store.get(k) : null),
    setItem: (k, v) => store.set(k, String(v)),
    removeItem: (k) => store.delete(k),
    clear: () => store.clear(),
  });
}

// Vuetify overlays use requestAnimationFrame/cancelAnimationFrame.
if (!hasGlobal('requestAnimationFrame')) {
  defineGlobal('requestAnimationFrame', (cb) => setTimeout(() => cb(Date.now()), 0));
  defineGlobal('cancelAnimationFrame', (id) => clearTimeout(id));
}

// Vuetify 3 probes CSS.supports() at import time; jsdom has no CSS object.
if (!hasGlobal('CSS')) {
  defineGlobal('CSS', { supports: () => false });
}

// Vuetify 3 measures overlays, slide groups and layouts with ResizeObserver; jsdom has none
// (and lays nothing out, so there is nothing to observe).
if (!hasGlobal('ResizeObserver')) {
  const noop = () => {};
  defineGlobal('ResizeObserver', class ResizeObserver {
    constructor() {
      Object.assign(this, { observe: noop, unobserve: noop, disconnect: noop });
    }
  });
}

// Vuetify 3 overlays follow window.visualViewport, which jsdom does not have.
if (!hasGlobal('visualViewport')) {
  defineGlobal('visualViewport', null);
}
