var config = require('./lib/config')

exports.config = {
  app_name : ['Semaphore'],
  license_key : config.credentials.newrelic_key,
  logging : {
    level : 'trace'
  },
  rules: {
    ignore: [
      '^/socket.io/.*/*-polling',
      '^/ping$'
    ]
  }
};
