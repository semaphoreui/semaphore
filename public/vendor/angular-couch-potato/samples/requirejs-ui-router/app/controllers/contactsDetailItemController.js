define(['app', 'services/findById'], function (app) {
  app.registerController(
    'contactsDetailItemController',
    [        '$scope', '$stateParams', '$state', 'findById',
    function ($scope,   $stateParams,   $state, findById) {
      $scope.item = findById.find($scope.contact.items, $stateParams.itemId);
      $scope.edit = function () {
        $state.transitionTo('contacts.detail.item.edit', $stateParams);
      };
    }]
  );
});
