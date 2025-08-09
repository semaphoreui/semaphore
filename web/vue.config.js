const webpack = require('webpack');
const { VuetifyPlugin } = require('webpack-plugin-vuetify');

module.exports = {
  configureWebpack: {
    plugins: [
      new webpack.DefinePlugin({
        'process.env.VUE_APP_BUILD_TYPE': JSON.stringify(process.env.VUE_APP_BUILD_TYPE),
      }),
      new VuetifyPlugin({
        autoImport: true,
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
    config.plugin('html')
      .tap((args) => {
        // eslint-disable-next-line no-param-reassign
        args[0].minify = false;
        return args;
      });

    // Alias Vue to the Vue 3 compatibility build
    config.resolve.alias.set('vue$', '@vue/compat');
  },
  publicPath: './',
  outputDir: '../api/public',
};
