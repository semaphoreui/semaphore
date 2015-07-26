define(['app'], function (app) {
	app.registerController('AddUserCtrl', ['$scope', '$state', '$http', function($scope, $state, $http) {
		$scope.user = {};

		$scope.add = function () {
			$http.post('/users', $scope.user)
			.success(function (data) {
				$state.transitionTo('users.list');
			}).error(function () {
				alert('cannot add user.')
			});
		}
	}]);
});