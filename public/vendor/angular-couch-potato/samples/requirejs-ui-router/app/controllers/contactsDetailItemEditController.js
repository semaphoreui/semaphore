define(['app', 'services/findById'], function (app) {
  app.registerController(
    'contactsDetailItemEditController',
    [        '$scope', '$stateParams', '$state', 'findById',
    function ($scope,   $stateParams,   $state, findById) {
      $scope.item = findById.find($scope.contact.items, $stateParams.itemId);
      $scope.done = function () {
        $state.transitionTo('contacts.detail.item', $stateParams);
      };
    }]
  );
});
