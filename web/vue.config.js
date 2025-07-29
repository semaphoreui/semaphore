const { defineConfig } = require('@vue/cli-service');

module.exports = defineConfig({
  configureWebpack: {
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
  transpileDependencies: true,
  publicPath: './',
  outputDir: '../api/public',
});
