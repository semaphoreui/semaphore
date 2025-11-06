const webpack = require('webpack');

module.exports = {
  configureWebpack: {
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
