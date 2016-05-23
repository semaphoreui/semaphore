define(function () {
	app.registerController('UserCtrl', ['$scope', '$http', '$uibModal', '$rootScope', 'user', '$state', function ($scope, $http, $modal, $rootScope, user, $state) {
		$scope.user = user.data;
		$scope.is_self = $scope.user.id == $rootScope.user.id;

		$scope.updatePassword = function (pwd) {
			$http.post('/users/' + $scope.user.id + '/password', {
				password: pwd
			}).success(function () {
				swal('OK', 'User profile & password were updated.');
			}).error(function (_, status) {
				swal('Error', 'Setting password failed, API responded with HTTP ' + status, 'error');
			});
		}

		$scope.updateUser = function () {
			var pwd = $scope.user.password;

			$http.put('/users/' + $scope.user.id, $scope.user).success(function () {
				if ($rootScope.user.id == $scope.user.id) {
					$rootScope.user = $scope.user;
				}

				if (pwd && pwd.length > 0) {
					$scope.updatePassword(pwd);
					return;
				}

				swal('OK', 'User has been updated!');
			}).error(function (_, status) {
				swal('Error', 'User profile could not be updated: ' + status, 'error');
			});
		}

		$scope.deleteUser = function () {
			$http.delete('/users/' + $scope.user.id).success(function () {
				$state.go('users.list');
			}).error(function (_, status) {
				swal('Error', 'User could not be deleted! ' + status, 'error');
			});
		}
	}]);
});
