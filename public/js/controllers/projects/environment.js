define(function () {
	app.registerController('ProjectEnvironmentCtrl', ['$scope', '$http', '$uibModal', 'Project', function ($scope, $http, $modal, Project) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/environment').success(function (environment) {
				$scope.environment = environment;
			});
		}

		$scope.remove = function (environment) {
			$http.delete(Project.getURL() + '/environment/' + environment.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete environment key..', 'error');
			});
		}

		$scope.reload();
	}]);
});