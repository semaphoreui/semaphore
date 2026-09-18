// Shared setup for unit tests. Bundled ahead of every spec via
// `--include tests/unit/setup.js` in the `test:unit` npm script.
// Specs can also import it explicitly so they work when run directly.

import Vue from 'vue';
import Vuetify from 'vuetify';
import * as components from 'vuetify/lib/components';
import filtersPlugin from '@/plugins/filters';

// --- jsdom polyfills ---------------------------------------------------------
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

// --- Vuetify -----------------------------------------------------------------
// vuetify-loader is disabled in test mode (see vue.config.js), so register all
// components globally. Vuetify installs only once per bundle, so this has to
// run before any spec calls `localVue.use(Vuetify)`.
Vue.use(Vuetify, { components });

// Global template filters, normally installed by src/main.js.
Vue.use(filtersPlugin);
