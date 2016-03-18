define(['app', 'services/findById'], function (app) {
  app.registerController(
    'contactsDetailController',
    [        '$scope', '$stateParams', 'something', 'findById',
    function ($scope,   $stateParams,   something, findById) {
      $scope.something = something;
      $scope.contact = findById.find($scope.contacts, $stateParams.contactId);
    }]
  );
});
