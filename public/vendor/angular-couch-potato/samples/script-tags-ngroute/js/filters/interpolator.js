define(['services/version'], function() {
  app.registerFilter(
    'interpolator',
    [
      'version',
      function (version) {
        return function(text) {
          return String(text).replace(/\%VERSION\%/mg, version);
        };
      }
    ]
  );
});
