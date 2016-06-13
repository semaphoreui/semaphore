define(function () {
	app.registerController('UsersCtrl', ['$scope', '$http', '$uibModal', '$rootScope', function ($scope, $http, $modal, $rootScope) {
		$http.get('/users').success(function (users) {
			$scope.users = users;
		});

		$scope.addUser = function () {
			var scope = $rootScope.$new();
			scope.user = {}

			$modal.open({
				templateUrl: '/tpl/users/add.html',
				scope: scope
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
	}]);
});
