define(function () {
	app.registerController('ProjectEnvironmentCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', function ($scope, $http, $modal, Project, $rootScope) {
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

		$scope.add = function () {
			var scope = $rootScope.$new();
			scope.env = {
				json: '{}'
			};

			$modal.open({
				templateUrl: '/tpl/projects/environment/add.html',
				scope: scope
			}).result.then(function (env) {
				$http.post(Project.getURL() + '/environment', env)
				.success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('Erorr', 'Environment not added: ' + status, 'error');
				});
			});
		}

		$scope.editEnvironment = function (env) {
			var scope = $rootScope.$new();
			scope.env = env;

			$modal.open({
				templateUrl: '/tpl/projects/environment/add.html',
				scope: scope
			}).result.then(function (v) {
				$http.put(Project.getURL() + '/environment/' + env.id, v)
				.success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('Erorr', 'Environment not updated: ' + status, 'error');
				});
			});
		}

		$scope.reload();
	}]);
});