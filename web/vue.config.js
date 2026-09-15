const webpack = require('webpack');

module.exports = {
  configureWebpack: {
    performance: {
      hints: false,
    },
    plugins: [
      new webpack.DefinePlugin({
        'process.env.VUE_APP_BUILD_TYPE': JSON.stringify(process.env.VUE_APP_BUILD_TYPE),
      }),
    ],
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
      // In test mode babel transpiles modules to CommonJS, while vuetify-loader
      // prepends ES `import` statements to compiled templates. Mixing both makes
      // webpack drop the template's `render` export, so every SFC fails to mount
      // in unit tests. Components are registered by `Vue.use(Vuetify)` instead.
      config.plugins.delete('VuetifyLoaderPlugin');
    }

    config.plugin('html')
      .tap((args) => {
        // eslint-disable-next-line no-param-reassign
        args[0].minify = false;
        return args;
      });
  },
  transpileDependencies: [
    'vuetify',
  ],
  publicPath: './',
  outputDir: '../api/public',
};
