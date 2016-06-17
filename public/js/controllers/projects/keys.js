define(function () {
	app.registerController('ProjectKeysCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', function ($scope, $http, $modal, Project, $rootScope) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/keys').success(function (keys) {
				$scope.keys = keys;
			});
		}

		$scope.remove = function (key) {
			$http.delete(Project.getURL() + '/keys/' + key.id).success(function () {
				$scope.reload();
			}).error(function (d) {
				if (!(d && d.inUse)) {
					swal('error', 'could not delete key..', 'error');
					return;
				}

				swal({
					title: 'Key in use',
					text: d.error,
					type: 'error',
					showCancelButton: true,
					confirmButtonColor: "#DD6B55",
					confirmButtonText: 'Mark as removed'
				}, function () {
					$http.delete(Project.getURL() + '/keys/' + key.id + '?setRemoved=1').success(function () {
						$scope.reload();
					}).error(function () {
						swal('error', 'could not remove key..', 'error');
					});
				});
			});
		}

		$scope.add = function () {
			$modal.open({
				templateUrl: '/tpl/projects/keys/add.html'
			}).result.then(function (key) {
				$http.post(Project.getURL() + '/keys', key).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('error', 'could not add key:' + status, 'error');
				});
			});
		}

		$scope.update = function (key) {
			var scope = $rootScope.$new();
			scope.key = key;

			$modal.open({
				templateUrl: '/tpl/projects/keys/add.html',
				scope: scope
			}).result.then(function (opts) {
				if (opts.delete) {
					$scope.remove(key);

					return;
				}

				$http.put(Project.getURL() + '/keys/' + key.id, opts.key)
				.success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('Error', 'could not update key:' + status, 'error');
				});
			});
		}

		$scope.reload();
	}]);
});