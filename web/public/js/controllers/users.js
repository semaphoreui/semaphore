define(function () {
	app.registerController('UsersCtrl', ['$scope', '$http', '$uibModal', '$rootScope', 'SweetAlert', function ($scope, $http, $modal, $rootScope, SweetAlert) {
		$http.get('/users').then(function (response) {
			$scope.users = response.data;
		});

		$scope.addUser = function () {
			var scope = $rootScope.$new();
			scope.user = {};

			$modal.open({
				templateUrl: '/tpl/users/add.html',
				scope: scope
			}).result.then(function (userData) {
				$http.post('/users', userData).then(function (response) {
					$scope.users.push(response.data);

					$http.post('/users/' + response.data.id + '/password', {
						password: userData.password
					}).catch(function (errorResponse) {
						SweetAlert.swal('Error', 'Setting password failed, API responded with HTTP ' + errorResponse.status, 'error');
					});
				}).catch(function (response) {
					SweetAlert.swal('Error', 'API responded with HTTP ' + response.status, 'error');
				});
			}, function () {});
		};
	}]);
});
