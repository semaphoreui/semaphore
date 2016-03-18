define(['services/myService2'], function () {
  app.registerController(
    'MyCtrl2',
    [
      '$scope', 'myService2',
      function($scope, myService2) {
        $scope.welcomeMessage = myService2.getMsg();
      }
    ]
  );
});
