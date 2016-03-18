define(['services/version'], function () {
  app.registerController(
    'MyCtrl1',
    [
      '$scope', 'version',
      function($scope, version) {
        $scope.scopedAppVersion = version;
      }
    ]
  );
});
