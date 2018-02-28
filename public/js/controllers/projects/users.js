define(function () {
	app.registerController('ProjectUsersCtrl', ['$scope', '$http', 'Project', '$uibModal', '$rootScope', function ($scope, $http, Project, $modal, $rootScope) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/users?sort=name&order=asc').then(function (response) {
			  var users = response.data;
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
			$http.delete(Project.getURL() + '/users/' + user.id).then(function () {
				$scope.reload();
			}).catch(function () {
				swal('error', 'could not delete user..', 'error');
			});
		}

		$scope.addUser = function () {
			$http.get('/users').then(function (response) {
			  var users = response.data;
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
						.then(function () {
							$scope.reload();
						}).catch(function (response) {
							swal('Error', 'User not added: ' + response.status, 'error');
						});
				});
			});
		}

		$scope.setAdmin = function (user) {
			var verb = $http.post;
			if (user.admin) verb = $http.delete;

			var numAdmins = 0;
			this.users.forEach(function (user) {
				user.admin && numAdmins++;
			});

			if (user.admin && numAdmins == 1) {
				swal('Administrator Required', 'There must be at least one administrator on the project', 'error');

				return;
			}

			verb(Project.getURL() + '/users/' + user.id + '/admin').then(function () {
				$scope.reload();
			});
		}

		$scope.reload();
	}]);
});
