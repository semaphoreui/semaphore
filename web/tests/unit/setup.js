// Shared setup for unit tests. Vitest runs it ahead of every spec file (`test.setupFiles` in
// vite.config.mjs). Specs can also import it explicitly.

// The jsdom polyfills come first: Vuetify reads some globals (CSS) as soon as it is imported.
import './polyfills';
import { config } from '@vue/test-utils';
import { createVuetify } from 'vuetify';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';
import { filters } from '@/plugins/filters';

// --- Vuetify -----------------------------------------------------------------
// Specs mount components that are not compiled with the automatic imports of the Vuetify plugin
// (template strings of stubs), so the test instance registers every component and directive
// globally. Every mount gets it
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
