define(function () {
	app.registerController('ProjectKeysCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', function ($scope, $http, $modal, Project, $rootScope) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/keys?sort=name&order=asc').then(function (keys) {
				$scope.keys = keys.data;
			});
		}

		$scope.remove = function (key) {
			$http.delete(Project.getURL() + '/keys/' + key.id)
				.then(function () {
					$scope.reload();
				})
				.catch(function (response) {
					var d = response.data;

					if (!(d && d.inUse)) {
						SweetAlert.swal('error', 'could not delete key..', 'error');
						return;
					}

					SweetAlert.swal({
						title: 'Key in use',
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

						$http.delete(Project.getURL() + '/keys/' + key.id + '?setRemoved=1')
							.then(function () {
								swal.stopLoading();
								swal.close();

								$scope.reload();
							})
							.catch(function () {
								SweetAlert.swal('Error', 'Could not remove key..', 'error');
							});
					});
				});
		}

		$scope.add = function () {
			$modal.open({
				templateUrl: '/tpl/projects/keys/add.html'
			}).result.then(function (opts) {
				$http.post(Project.getURL() + '/keys', opts.key).then(function () {
					$scope.reload();
				}).catch(function (response) {
					SweetAlert.swal('error', 'could not add key:' + response.status, 'error');
				});
			}, function () {
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
					.then(function () {
						$scope.reload();
					}).catch(function (response) {
					SweetAlert.swal('Error', 'could not update key:' + response.status, 'error');
				});
			}, function () {
			});
		}

		$scope.reload();
	}]);
});
