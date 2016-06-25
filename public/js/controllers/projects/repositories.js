define(function () {
	app.registerController('ProjectRepositoriesCtrl', ['$scope', '$http', 'Project', '$uibModal', '$rootScope', function ($scope, $http, Project, $modal, $rootScope) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/keys?type=ssh').success(function (keys) {
				$scope.sshKeys = keys;

				$http.get(Project.getURL() + '/repositories').success(function (repos) {
					repos.forEach(function (repo) {
						for (var i = 0; i < keys.length; i++) {
							if (repo.ssh_key_id == keys[i].id) {
								repo.ssh_key = keys[i];
								break;
							}
						}
					});

					$scope.repositories = repos;
				});
			});
		}

		$scope.remove = function (repo) {
			$http.delete(Project.getURL() + '/repositories/' + repo.id).success(function () {
				$scope.reload();
			}).error(function (d) {
				if (!(d && d.templatesUse)) {
					swal('error', 'could not delete repository..', 'error');
					return;
				}

				swal({
					title: 'Repository in use',
					text: d.error,
					type: 'error',
					showCancelButton: true,
					confirmButtonColor: "#DD6B55",
					confirmButtonText: 'Mark as removed'
				}, function () {
					$http.delete(Project.getURL() + '/repositories/' + repo.id + '?setRemoved=1').success(function () {
						$scope.reload();
					}).error(function () {
						swal('error', 'could not delete repository..', 'error');
					});
				});
			});
		}

		$scope.update = function (repo) {
			var scope = $rootScope.$new();
			scope.keys = $scope.sshKeys;
			scope.repo = JSON.parse(JSON.stringify(repo));

			$modal.open({
				templateUrl: '/tpl/projects/repositories/add.html',
				scope: scope
			}).result.then(function (opts) {
				if (opts.remove) {
					return $scope.remove(repo);
				}

				$http.put(Project.getURL() + '/repositories/' + repo.id, opts.repo).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('Error', 'Repository not updated: ' + status, 'error');
				});
			});
		}

		$scope.add = function () {
			var scope = $rootScope.$new();
			scope.keys = $scope.sshKeys;

			$modal.open({
				templateUrl: '/tpl/projects/repositories/add.html',
				scope: scope
			}).result.then(function (repo) {
				$http.post(Project.getURL() + '/repositories', repo.repo)
				.success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('Erorr', 'Repository not added: ' + status, 'error');
				});
			});
		}

		$scope.reload();
	}]);
});
