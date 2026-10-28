const path = require('path');
// eslint-disable-next-line import/no-extraneous-dependencies
const webpack = require('webpack');
// eslint-disable-next-line import/no-extraneous-dependencies
const { VuetifyPlugin } = require('webpack-plugin-vuetify');

const plugins = [
  new webpack.DefinePlugin({
    'process.env.VUE_APP_BUILD_TYPE': JSON.stringify(process.env.VUE_APP_BUILD_TYPE),
    // Compile-time flags of the Vue 3 and vue-i18n bundler builds.
    __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'false',
    __VUE_I18N_FULL_INSTALL__: 'true',
    __VUE_I18N_LEGACY_API__: 'false',
    __INTLIFY_PROD_DEVTOOLS__: 'false',
  }),
];

if (process.env.NODE_ENV !== 'test') {
  // Vuetify components are imported automatically from templates, and the Vuetify styles are
  // compiled from Sass with the settings in src/scss/variables.scss. In test mode babel
  // transpiles modules to CommonJS, while the plugin prepends ES `import` statements to
  // compiled templates; components are registered by the test setup instead.
  plugins.push(new VuetifyPlugin({
    autoImport: true,
    styles: { configFile: 'src/scss/variables.scss' },
  }));
}

module.exports = {
  configureWebpack: {
    performance: {
      hints: false,
    },
    plugins,
    devServer: {
      historyApiFallback: true,
      proxy: {
        '^/api': {
          target: 'http://localhost:3000',
        },
      },
    },
  },
  chainWebpack: (config) => {
    if (process.env.NODE_ENV === 'test') {
      // The test bundle (mode `none`) resolves package exports under the `production` condition:
      // Vue then comes as its production build, which ignores the component stubs of
      // @vue/test-utils. Resolve the development builds instead.
      config.resolve.set('conditionNames', ['webpack', 'development', 'module-sync', 'module', 'node', 'import', 'require']);
      // Unit tests stub components with `template` strings: the full build compiles them at run
      // time (the runtime-only build renders nothing for them).
      config.resolve.alias.set('vue$', path.resolve(__dirname, 'node_modules/vue/dist/vue.esm-bundler.js'));
    }
    config.plugin('html')
      .tap((args) => {
        // eslint-disable-next-line no-param-reassign
        args[0].minify = false;
        return args;
      });
  },
  css: {
    loaderOptions: {
      // sass-loader 13 calls the legacy JS API of Dart Sass; the warning is about the loader.
      sass: {
        sassOptions: { silenceDeprecations: ['legacy-js-api'] },
      },
    },
  },
  // Not in test mode: babel then turns modules into CommonJS, and webpack leaves the `require`
  // calls of a `"type": "module"` package (Vuetify 3) to Node at run time, where they fail.
  transpileDependencies: process.env.NODE_ENV === 'test' ? [] : [
    'vuetify',
  ],
  publicPath: './',
  outputDir: '../api/public',
};
