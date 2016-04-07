define(function () {
	app.registerController('ProjectInventoryCtrl', ['$scope', '$http', '$uibModal', 'Project', function ($scope, $http, $modal, Project) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/inventory').success(function (inventory) {
				$scope.inventory = inventory;
			});
		}

		$scope.remove = function (environment) {
			$http.delete(Project.getURL() + '/inventory/' + environment.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete inventory..', 'error');
			});
		}

		$scope.reload();
	}]);
});