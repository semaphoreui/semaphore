define(function () {
	app.registerController('ProjectUsersCtrl', ['$scope', '$http', 'Project', '$uibModal', '$rootScope', function ($scope, $http, Project, $modal, $rootScope) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/users').success(function (users) {
				$scope.project_user = null;
				$scope.users = users;

				for (var i = 0; i < users.length; i++) {
					if (users[i].id == $scope.user.id) {
						$scope.project_user = users[i];
						break;
					}
				}
			});
		}

		$scope.remove = function (user) {
			$http.delete(Project.getURL() + '/users/' + user.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete user..', 'error');
			});
		}

		$scope.addUser = function () {
			$http.get('/users').success(function (users) {
				$scope.users.forEach(function (u) {
					for (var i = 0; i < users.length; i++) {
						if (u.id == users[i].id) {
							users.splice(i, 1);
							break;
						}
					}
				});

				var scope = $rootScope.$new();
				scope.users = users;

				$modal.open({
					templateUrl: '/tpl/projects/users/add.html',
					scope: scope
				}).result.then(function (user) {
					$http.post(Project.getURL() + '/users', user)
					.success(function () {
						$scope.reload();
					}).error(function (_, status) {
						swal('Erorr', 'User not added: ' + status, 'error');
					});
				});
			});
		}

		$scope.setAdmin = function (user) {
			var verb = $http.post;
			if (user.admin) verb = $http.delete;

			verb(Project.getURL() + '/users/' + user.id + '/admin').success(function () {
				$scope.reload();
			});
		}

		$scope.reload();
	}]);
});