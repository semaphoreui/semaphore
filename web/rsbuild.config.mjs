import { defineConfig } from '@rsbuild/core';
import { pluginVue2 } from '@rsbuild/plugin-vue2';
import { pluginSass } from '@rsbuild/plugin-sass';

export default defineConfig({
  plugins: [
    pluginVue2(),
    pluginSass({
      sassLoaderOptions: {
        sassOptions: {
          // ponytail: vuetify 2 sass is frozen upstream; silencing beats fixing
          silenceDeprecations: ['import', 'global-builtin', 'slash-div', 'color-functions', 'legacy-js-api'],
        },
      },
    }),
  ],
  source: {
    entry: {
      index: './src/main.js',
    },
    define: {
      'process.env.VUE_APP_BUILD_TYPE': JSON.stringify(process.env.VUE_APP_BUILD_TYPE),
    },
  },
  resolve: {
    alias: {
      '@': './src',
    },
  },
  html: {
    template: './public/index.html',
  },
  output: {
    distPath: {
      root: '../api/public',
    },
    assetPrefix: './',
    cleanDistPath: true,
  },
  server: {
    port: 8080,
    proxy: {
      '/api': 'http://localhost:3000',
    },
  },
});
