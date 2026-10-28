// Shared setup for unit tests. Bundled ahead of every spec via
// `--include tests/unit/setup.js` in the `test:unit` npm script.
// Specs can also import it explicitly so they work when run directly.

// The jsdom polyfills come first: Vuetify reads some globals (CSS) as soon as it is imported.
import './polyfills';
import { config } from '@vue/test-utils';
import { createVuetify } from 'vuetify';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';
import { filters } from '@/plugins/filters';

// --- Vuetify -----------------------------------------------------------------
// webpack-plugin-vuetify (automatic component imports) is off in test mode (see vue.config.js),
// so the test instance registers every component and directive globally. Every mount gets it
// through the global plugins of @vue/test-utils; a spec adds its own (i18n, router) per mount.
// A spec that needs a display of its own width (the display is measured when the instance is
// created) passes a fresh instance in its global plugins: the one installed last wins.
export default function createTestVuetify() {
  return createVuetify({ components, directives });
}

config.global.plugins.push(createTestVuetify());

// Global template formatters, normally set by src/main.js (`$filters`).
config.global.config.globalProperties = {
  ...config.global.config.globalProperties,
  $filters: filters,
};
