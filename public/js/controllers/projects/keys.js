define(function () {
	app.registerController('ProjectKeysCtrl', ['$scope', '$http', '$uibModal', 'Project', function ($scope, $http, $modal, Project) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/keys').success(function (keys) {
				$scope.keys = keys;
			});
		}

		$scope.remove = function (key) {
			$http.delete(Project.getURL() + '/keys/' + key.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete access key..', 'error');
			});
		}

		$scope.add = function () {
			$modal.open({
				templateUrl: '/tpl/projects/keysAdd.html'
			}).result.then(function (key) {
				$http.post(Project.getURL() + '/keys', key).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('error', 'could not add key:' + status, 'error');
				});
			});
		}

		$scope.reload();
	}]);
});