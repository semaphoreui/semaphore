import fs from 'node:fs';
import { createRequire } from 'node:module';
import { fileURLToPath, URL } from 'node:url';
import path from 'node:path';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import vuetify from 'vite-plugin-vuetify';

// The build type (`pro_selfhosted`, `pro_portal` or empty) comes from the environment of the build:
// Taskfile and CI set VUE_APP_BUILD_TYPE (the Vue CLI name, kept so that every caller works
// unchanged); VITE_BUILD_TYPE is accepted as well. The code reads it as
// `process.env.VUE_APP_BUILD_TYPE`, replaced at build time.
const buildType = process.env.VUE_APP_BUILD_TYPE || process.env.VITE_BUILD_TYPE || '';

// Backend the dev server proxies /api to (`stand web` points it at the stand of the work copy).
const devApiTarget = process.env.SEMAPHORE_DEV_API_URL || 'http://localhost:3000';

// The Go server embeds the build with `//go:embed public/*` (api/router.go), which silently skips
// files whose names start with `_` or `.`: no emitted file may be named that way.
function embeddable(name) {
  return name.replace(/^[_.]+/, '');
}

function assetName(info) {
  const file = (info.names && info.names[0]) || info.name || 'asset';
  return embeddable(path.basename(file, path.extname(file)));
}

// Order of the Vuetify styles. vite-plugin-vuetify appends the imports of the components a
// template uses to the end of the compiled single-file component, after the imports of its script
// (child components), and imports a component from the index of its folder, which loads the
// styles of every component of the folder. The webpack build imported the template ahead of the
// script, and webpack went past the index (Vuetify marks only styles as side effects) to the files
// of the components used. The component styles rely on their order (`.v-field .v-chip` and
// `.v-chip.v-chip--size-small` have the same specificity, for one): move the block to the top and
// import every component from its own file, so that the stylesheet comes out in the same order.
const vuetifyRoot = path.dirname(createRequire(import.meta.url).resolve('vuetify/package.json'));
const vuetifyComponents = path.join(vuetifyRoot, 'lib', 'components');

// Names in the order the index of the folder exports them, as webpack followed them.
function exportOrder(folder) {
  const index = fs.readFileSync(path.join(vuetifyComponents, folder, 'index.js'), 'utf8');
  return [...index.matchAll(/export \{ (\w+) \}/g)].map((m) => m[1]);
}

function vuetifyImport(line) {
  const m = /^import \{ ([^}]+) \} from "vuetify\/components\/(\w+)"$/.exec(line);
  if (!m) {
    return line;
  }
  const order = exportOrder(m[2]);
  const rank = (spec) => {
    const i = order.indexOf(spec.split(' as ')[0]);
    return i < 0 ? order.length : i;
  };
  return m[1].split(', ').sort((a, b) => rank(a) - rank(b)).map((spec) => {
    const name = spec.split(' as ')[0];
    return fs.existsSync(path.join(vuetifyComponents, m[2], `${name}.js`))
      ? `import { ${spec} } from "vuetify/lib/components/${m[2]}/${name}.js"`
      : `import { ${spec} } from "vuetify/components/${m[2]}"`;
  }).join('\n');
}

function vuetifyImportsFirst() {
  const marker = '\n\n/* Vuetify */\n';
  return {
    name: 'semaphore:vuetify-imports-first',
    transform(code) {
      const at = code.lastIndexOf(marker);
      if (at < 0) {
        return null;
      }
      const block = code.slice(at + marker.length).split('\n').map(vuetifyImport).join('\n');
      return { code: `/* Vuetify */\n${block}\n${code.slice(0, at)}\n`, map: null };
    },
  };
}

export default defineConfig(({ mode }) => ({
  // Relative asset URLs: index.html keeps `<base href="/">`, which the Go server replaces with the
  // address from `web_host` (api/router.go, serveFile), so the app also works under a sub-path.
  base: './',

  plugins: [
    vue(),
    // Vuetify components are imported automatically from templates, and the Vuetify styles are
    // compiled from Sass with the settings in src/scss/variables.scss. Unit tests (Vitest, mode
    // `test`) take no styles: the plain CSS imports of Vuetify are stubbed there, while the
    // virtual Sass modules of the plugin do not load in the test runner.
    vuetify(mode === 'test' ? { autoImport: true } : {
      autoImport: true,
      styles: { configFile: 'src/scss/variables.scss' },
    }),
    vuetifyImportsFirst(),
  ],

  define: {
    'process.env.VUE_APP_BUILD_TYPE': JSON.stringify(buildType),
    // Compile-time flags of the Vue 3 and vue-i18n bundler builds.
    __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'false',
    __VUE_I18N_FULL_INSTALL__: 'true',
    __VUE_I18N_LEGACY_API__: 'false',
    __INTLIFY_PROD_DEVTOOLS__: 'false',
  },

  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
    // Single-file components are imported without the extension in places.
    extensions: ['.mjs', '.js', '.json', '.vue'],
  },

  build: {
    outDir: '../api/public',
    // Vite empties only an outDir inside the project root by default. Emptying deletes the tracked
    // api/public/.gitkeep: the `build:fe` task of the Taskfile restores it.
    emptyOutDir: true,
    // CSS as Sass writes it: the default minifier (Lightning CSS) rounds numbers to six digits
    // (letter-spacing 0.0892857143em → 0.0892857em), which moves text by a sub-pixel.
    cssMinify: false,
    // Source maps next to the bundles, as Vue CLI wrote them.
    sourcemap: true,
    // One bundle of about 2 MB, as with Vue CLI: no size warning.
    chunkSizeWarningLimit: 10000,
    rolldownOptions: {
      // No report of the time the plugins take on every build.
      checks: { bundlerTimings: false },
      output: {
        // Dependencies in a chunk of their own (like `chunk-vendors` of Vue CLI): their styles then
        // come in a stylesheet ahead of the application styles. The Vuetify component styles are
        // virtual modules of vite-plugin-vuetify (compiled with the Sass settings), not files in
        // node_modules.
        codeSplitting: {
          groups: [{ name: 'vendor', test: /[\\/]node_modules[\\/]|^virtual:plugin-vuetify:/ }],
        },
        entryFileNames: (info) => `assets/${embeddable(info.name)}-[hash].js`,
        chunkFileNames: (info) => `assets/${embeddable(info.name)}-[hash].js`,
        assetFileNames: (info) => `assets/${assetName(info)}-[hash][extname]`,
      },
    },
  },

  server: {
    port: 8080,
    host: '0.0.0.0',
    proxy: {
      '^/api': {
        target: devApiTarget,
        // The task log and the live updates come over the WebSocket at /api/ws.
        ws: true,
      },
    },
  },

  test: {
    environment: 'jsdom',
    globals: true,
    include: ['tests/unit/**/*.spec.js'],
    setupFiles: ['tests/unit/setup.js'],
    server: {
      deps: {
        // Vuetify imports its CSS from its modules: let Vite transform it instead of Node.
        inline: ['vuetify'],
      },
    },
  },
}));
