define(function () {
	app.registerController('ProjectRepositoriesCtrl', ['$scope', '$http', '$uibModal', 'Project', function ($scope, $http, $modal, Project) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/repositories').success(function (repositories) {
				$scope.repositories = repositories;
			});
		}

		$scope.remove = function (repo) {
			$http.delete(Project.getURL() + '/repositories/' + repo.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete repository..', 'error');
			});
		}

		$scope.reload();
	}]);
});