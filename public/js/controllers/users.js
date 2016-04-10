define(function () {
	app.registerController('UsersCtrl', ['$scope', '$http', '$uibModal', function ($scope, $http, $modal) {
		$http.get('/users').success(function (users) {
			$scope.users = users;
		});

		$scope.addUser = function () {
			$modal.open({
				templateUrl: '/tpl/users/add.html'
			}).result.then(function (_user) {
				$http.post('/users', _user).success(function (user) {
					$scope.users.push(user);

					$http.post('/users/' + user.id + '/password', {
						password: _user.password
					}).error(function (_, status) {
						swal('Error', 'Setting password failed, API responded with HTTP ' + status, 'error');
					});
				}).error(function (_, status) {
					swal('Error', 'API responded with HTTP ' + status, 'error');
				});
			});
		}

		$scope.changePassword = function (user) {
			$modal.open({
				templateUrl: '/tpl/users/password.html'
			}).result.then(function (password) {
				$http.post('/users/' + user.id + '/password', {
					password: password
				}).error(function (_, status) {
					swal('Error', 'Setting password failed, API responded with HTTP ' + status, 'error');
				});
			});
		}
	}]);
});
