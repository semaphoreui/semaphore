define(function () {
	app.registerController('ProjectTemplatesCtrl', ['$scope', '$http', '$uibModal', 'Project', function ($scope, $http, $modal, Project) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/templates').success(function (templates) {
				$scope.templates = templates;
			});
		}

		$scope.remove = function (environment) {
			$http.delete(Project.getURL() + '/templates/' + environment.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete template..', 'error');
			});
		}

		$scope.reload();
	}]);
});