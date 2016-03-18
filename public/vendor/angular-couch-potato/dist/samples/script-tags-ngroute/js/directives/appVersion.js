define(['services/version'], function () {
  app.registerDirective(
    'appVersion',
    [
      'version',
      function (versionValue) {
        return function (scope, elm, attrs) {
          elm.text(versionValue);
        };
      }
    ]
  );
});
