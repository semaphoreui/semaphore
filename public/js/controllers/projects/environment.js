define(function () {
	app.registerController('ProjectEnvironmentCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', function ($scope, $http, $modal, Project, $rootScope) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/environment?sort=name&order=asc').success(function (environment) {
				$scope.environment = environment;
			});
		}

		$scope.remove = function (environment) {
			$http.delete(Project.getURL() + '/environment/' + environment.id).success(function () {
				$scope.reload();
			}).error(function (d) {
				if (!(d && d.inUse)) {
					swal('error', 'could not delete environment..', 'error');
					return;
				}

				swal({
					title: 'Environment in use',
					text: d.error,
					type: 'error',
					showCancelButton: true,
					confirmButtonColor: "#DD6B55",
					confirmButtonText: 'Mark as removed'
				}, function () {
					$http.delete(Project.getURL() + '/environment/' + environment.id + '?setRemoved=1').success(function () {
						$scope.reload();
					}).error(function () {
						swal('error', 'could not delete environment..', 'error');
					});
				});
			});
		}

		$scope.add = function () {
			var scope = $rootScope.$new();
			scope.env = {
				json: '{}'
			};

			$modal.open({
				templateUrl: '/tpl/projects/environment/add.html',
				scope: scope
			}).result.then(function (env) {
				$http.post(Project.getURL() + '/environment', env.environment)
				.success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('Error', 'Environment not added: ' + status, 'error');
				});
			});
		}

		$scope.editEnvironment = function (env) {
			var scope = $rootScope.$new();
			scope.env = env;

			$modal.open({
				templateUrl: '/tpl/projects/environment/add.html',
				scope: scope
			}).result.then(function (opts) {
				if (opts.remove) {
					return $scope.remove(env);
				}

				$http.put(Project.getURL() + '/environment/' + env.id, opts.environment)
				.success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('Error', 'Environment not updated: ' + status, 'error');
				});
			});
		}

		$scope.reload();
	}]);
});
