define(function () {
	app.registerController('ProjectRepositoriesCtrl', ['$scope', '$http', 'Project', '$uibModal', '$rootScope', function ($scope, $http, Project, $modal, $rootScope) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/repositories').success(function (repos) {
				$scope.repositories = repos;
			});
		}

		$scope.remove = function (repo) {
			$http.delete(Project.getURL() + '/repositories/' + repo.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete repository..', 'error');
			});
		}

		$scope.update = function (repo) {
			$scope.getKeys(function (keys) {
				var scope = $rootScope.$new();
				scope.keys = keys;
				scope.repo = JSON.parse(JSON.stringify(repo));

				$modal.open({
					templateUrl: '/tpl/projects/repositories/add.html',
					scope: scope
				}).result.then(function (repo) {
					$http.put(Project.getURL() + '/repositories/' + repo.id, repo).success(function () {
						$scope.reload();
					}).error(function (_, status) {
						swal('Error', 'Repository not updated: ' + status, 'error');
					})
				});
			});
		}

		$scope.add = function () {
			$scope.getKeys(function (keys) {
				var scope = $rootScope.$new();
				scope.keys = keys;

				$modal.open({
					templateUrl: '/tpl/projects/repositories/add.html',
					scope: scope
				}).result.then(function (repo) {
					$http.post(Project.getURL() + '/repositories', repo)
					.success(function () {
						$scope.reload();
					}).error(function (_, status) {
						swal('Erorr', 'Repository not added: ' + status, 'error');
					});
				});
			});
		}

		$scope.getKeys = function (cb) {
			$http.get(Project.getURL() + '/keys?type=ssh')
			.success(cb);
		}

		$scope.reload();
	}]);
});