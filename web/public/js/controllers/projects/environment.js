define(function () {
	app.registerController('ProjectEnvironmentCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', 'SweetAlert', function ($scope, $http, $modal, Project, $rootScope, SweetAlert) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/environment?sort=name&order=asc').then(function (environment) {
				$scope.environment = environment.data;
			});
		}

		$scope.remove = function (environment) {
			$http.delete(Project.getURL() + '/environment/' + environment.id)
				.then(function () {
					$scope.reload();
				})
				.catch(function (response) {
					var d = response.data;
					if (!(d && d.inUse)) {
						SweetAlert.swal('error', 'could not delete environment..', 'error');
						return;
					}

					SweetAlert.swal({
						title: 'Environment in use',
						text: d.error,
						icon: 'error',
						buttons: {
							cancel: true,
							confirm: {
								text: 'Mark as removed',
								closeModel: false,
								className: 'bg-danger',
							}
						}
					}).then(function (value) {
						if (!value) {
							return;
						}

						$http.delete(Project.getURL() + '/environment/' + environment.id + '?setRemoved=1')
							.then(function () {
								swal.stopLoading();
								swal.close();

								$scope.reload();
							})
							.catch(function () {
								SweetAlert.swal('Error', 'Could not delete environment..', 'error');
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
					.then(function () {
						$scope.reload();
					}).catch(function (response) {
					SweetAlert.swal('Error', 'Environment not added: ' + response.status, 'error');
				});
			}, function () {
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
					.then(function () {
						$scope.reload();
					}).catch(function (response) {
					SweetAlert.swal('Error', 'Environment not updated: ' + response.status, 'error');
				});
			}, function () {
			});
		}

		$scope.reload();
	}]);
});
