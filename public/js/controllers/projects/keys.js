define(function () {
	app.registerController('ProjectKeysCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', function ($scope, $http, $modal, Project, $rootScope) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/keys?sort=name&order=asc').success(function (keys) {
				$scope.keys = keys;
			});

			$http.get(Project.getURL() + '/users?sort=name&order=asc').success(function (users) {
				$scope.project_user = null;
				$scope.users = users;
				$scope.users.push({"id":0, "username":"Public", "name":"Public"});
				for (var i = 0; i < users.length; i++) {
					if (users[i].id == $rootScope.user.id) {
						$scope.project_user = users[i];
						break;
					}
				}
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
			var scope = $rootScope.$new();
			scope.users =$scope.users
			scope.project_user=$scope.project_user

			$modal.open({
				templateUrl: '/tpl/projects/keys/add.html',
				scope: scope
			}).result.then(function (opts) {
				$http.post(Project.getURL() + '/keys', opts.key).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('error', 'could not add key:' + status, 'error');
				});
			});
		}

		$scope.update = function (key) {
			var scope = $rootScope.$new();
			scope.key = key;
			scope.users =$scope.users
			scope.project_user=$scope.project_user

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
