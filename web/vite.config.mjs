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

// Cascade layers. Vuetify 4 puts all its styles in layers (vuetify-core, vuetify-components,
// vuetify-overrides, vuetify-utilities, vuetify-final), and its utility classes (`mb-0`, `pa-2`,
// `d-flex`, `text-grey`) no longer carry `!important`: an unlayered rule would always beat them.
// The application styles therefore go to the layer `semaphore` between the components and the
// utilities: they override Vuetify components, as unlayered styles do, while a utility class on
// the element still wins over them, as the `!important` utilities of Vuetify 3 did. Other packages
// (drawflow, the mdi font) go to `semaphore-vendor` just below, so the application styles keep
// overriding them. A top-level `@layer` block of an application file stays where it is:
// `semaphore-base`, under the components, holds the element defaults that Vuetify components must
// override (assets/scss/components.scss).
const vuetifyLayersFile = path.join(vuetifyRoot, 'lib', 'styles', 'layers.css');
const vuetifyLayers = ['vuetify-core', 'vuetify-components', 'vuetify-overrides',
  'vuetify-utilities', 'vuetify-final'];
const layerOrder = ['vuetify-core', 'semaphore-base', 'vuetify-components', 'vuetify-overrides',
  'semaphore-vendor', 'semaphore', 'vuetify-utilities', 'vuetify-final'];

// The order of the layers is set by the first stylesheet that names them, and Vite bundles the CSS
// in an order of its own (vuetifyjs/vuetify#22752): the order of the layers, with the sub-layers
// of Vuetify (its layers.css), is declared in an inline `<style>` at the top of `<head>`, ahead of
// every stylesheet, in the dev server and in the build alike.
function vuetifyLayersFirst() {
  const layers = fs.readFileSync(vuetifyLayersFile, 'utf8');
  const topLevel = [...layers.matchAll(/^@layer ([\w-]+)/gm)].map((m) => m[1]);
  if (topLevel.join() !== vuetifyLayers.join()) {
    throw new Error(`Vuetify layers changed: ${topLevel.join(', ')} (${vuetifyLayersFile})`);
  }
  return {
    name: 'semaphore:vuetify-layers-first',
    transformIndexHtml() {
      return [{
        tag: 'style',
        children: `\n@layer ${layerOrder.join(', ')};\n${layers.trim()}\n`,
        injectTo: 'head-prepend',
      }];
    },
  };
}

// PostCSS plugin: wraps a stylesheet in a layer (see above); Vuetify styles are layered already.
function layerStyles() {
  return {
    postcssPlugin: 'semaphore-layers',
    Once(root, { AtRule }) {
      const file = (root.source && root.source.input && root.source.input.file) || '';
      if (file.startsWith(vuetifyRoot) || /plugin-vuetify/.test(file)) {
        return;
      }
      const layer = new AtRule({
        name: 'layer',
        params: /[\\/]node_modules[\\/]/.test(file) ? 'semaphore-vendor' : 'semaphore',
      });
      root.each((node) => {
        if (node.type === 'atrule' && ['charset', 'import', 'layer'].includes(node.name)) {
          return;
        }
        layer.append(node.remove());
      });
      if (layer.nodes && layer.nodes.length) {
        root.append(layer);
      }
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
    vuetifyLayersFirst(),
  ],

  css: {
    postcss: {
      plugins: [layerStyles()],
    },
  },

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
